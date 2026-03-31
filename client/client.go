// Package client provides a WebSocket client for the SimpleX Chat CLI Bot API.
//
// The Client manages a WebSocket connection to a running SimpleX Chat CLI instance,
// handling command dispatch, response correlation, event subscriptions, and automatic
// reconnection with exponential backoff.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	sxtypes "simplex-chat-bot/types"
)

const (
	maxReconnectDelay  = 30 * time.Second
	initialReconnDelay = 1 * time.Second
	defaultTimeout     = 30 * time.Second
)

// EventHandler is a callback function for asynchronous events from the CLI.
type EventHandler func(event json.RawMessage)

// Client manages a WebSocket connection to the SimpleX Chat CLI.
type Client struct {
	url     string
	options *Options

	mu     sync.RWMutex
	conn   *websocket.Conn
	closed atomic.Bool

	pendingMu sync.Mutex
	pending   map[string]chan *sxtypes.Response

	eventHandlersMu sync.RWMutex
	eventHandlers   []EventHandler

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
}

// Options configures the Client.
type Options struct {
	// Logger for diagnostic output. Defaults to a no-op logger.
	Logger *log.Logger
	// Timeout for individual command responses. Defaults to 30 seconds.
	Timeout time.Duration
	// Reconnect enables automatic reconnection on connection loss.
	Reconnect bool
}

// New creates a new Client connected to the given WebSocket URL (e.g., "ws://localhost:5225").
//
// The context controls the lifetime of the client; cancelling it will close the connection.
func New(ctx context.Context, url string, opts *Options) *Client {
	if opts == nil {
		opts = &Options{}
	}
	if opts.Timeout <= 0 {
		opts.Timeout = defaultTimeout
	}

	ctx, cancel := context.WithCancel(ctx)

	c := &Client{
		url:     url,
		options: opts,
		pending: make(map[string]chan *sxtypes.Response),
		ctx:     ctx,
		cancel:  cancel,
	}

	return c
}

// Connect establishes the WebSocket connection and starts the read loop.
func (c *Client) Connect(ctx context.Context) error {
	if c.closed.Load() {
		return errors.New("client is closed")
	}

	conn, _, err := websocket.Dial(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	c.wg.Add(1)
	go c.readLoop()

	return nil
}

// Close cleanly shuts down the connection and cancels in-flight operations.
func (c *Client) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}

	c.cancel()

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn != nil {
		_ = conn.Close(websocket.StatusNormalClosure, "client closing")
	}

	c.wg.Wait()

	c.pendingMu.Lock()
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
	c.pendingMu.Unlock()

	return nil
}

// OnEvent registers a handler that will be called for every asynchronous event.
func (c *Client) OnEvent(handler EventHandler) {
	c.eventHandlersMu.Lock()
	c.eventHandlers = append(c.eventHandlers, handler)
	c.eventHandlersMu.Unlock()
}

// Send dispatches a raw command string and returns the correlated response.
func (c *Client) Send(ctx context.Context, cmd string) (json.RawMessage, error) {
	corrID := generateCorrID()

	ch := make(chan *sxtypes.Response, 1)
	c.pendingMu.Lock()
	c.pending[corrID] = ch
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, corrID)
		c.pendingMu.Unlock()
	}()

	command := sxtypes.Command{
		CorrID: corrID,
		Cmd:    cmd,
	}

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, errors.New("not connected")
	}

	if err := wsjson.Write(ctx, conn, command); err != nil {
		return nil, fmt.Errorf("write command: %w", err)
	}

	timeout := c.options.Timeout
	if deadline, ok := ctx.Deadline(); ok {
		if t := time.Until(deadline); t < timeout {
			timeout = t
		}
	}

	select {
	case resp, ok := <-ch:
		if !ok {
			return nil, errors.New("connection closed while waiting for response")
		}
		return resp.Resp, nil
	case <-time.After(timeout):
		return nil, errors.New("command timed out")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) readLoop() {
	defer c.wg.Done()

	for {
		if c.closed.Load() {
			return
		}

		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		var raw json.RawMessage
		err := wsjson.Read(c.ctx, conn, &raw)
		if err != nil {
			if c.closed.Load() {
				return
			}

			if c.options.Logger != nil {
				c.options.Logger.Printf("read error: %v", err)
			}

			if c.options.Reconnect {
				c.reconnect()
				continue
			}
			return
		}

		c.dispatchMessage(raw)
	}
}

func (c *Client) dispatchMessage(raw json.RawMessage) {
	var envelope sxtypes.Response
	if err := json.Unmarshal(raw, &envelope); err != nil {
		if c.options.Logger != nil {
			c.options.Logger.Printf("unmarshal envelope: %v", err)
		}
		return
	}

	if envelope.CorrID != "" {
		c.pendingMu.Lock()
		ch, ok := c.pending[envelope.CorrID]
		if ok {
			ch <- &envelope
		}
		c.pendingMu.Unlock()
		return
	}

	c.eventHandlersMu.RLock()
	handlers := make([]EventHandler, len(c.eventHandlers))
	copy(handlers, c.eventHandlers)
	c.eventHandlersMu.RUnlock()

	for _, h := range handlers {
		h(envelope.Resp)
	}
}

func (c *Client) reconnect() {
	delay := initialReconnDelay

	for attempt := 0; ; attempt++ {
		if c.closed.Load() {
			return
		}

		select {
		case <-c.ctx.Done():
			return
		case <-time.After(delay + time.Duration(rand.Int63n(int64(delay)/2))):
		}

		if delay < maxReconnectDelay {
			delay *= 2
			if delay > maxReconnectDelay {
				delay = maxReconnectDelay
			}
		}

		if c.options.Logger != nil {
			c.options.Logger.Printf("reconnecting (attempt %d)...", attempt+1)
		}

		ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
		conn, _, err := websocket.Dial(ctx, c.url, nil)
		cancel()

		if err != nil {
			if c.options.Logger != nil {
				c.options.Logger.Printf("reconnect failed: %v", err)
			}
			continue
		}

		c.mu.Lock()
		oldConn := c.conn
		c.conn = conn
		c.mu.Unlock()

		if oldConn != nil {
			_ = oldConn.Close(websocket.StatusGoingAway, "reconnecting")
		}

		if c.options.Logger != nil {
			c.options.Logger.Printf("reconnected")
		}

		return
	}
}

func generateCorrID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 12)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
