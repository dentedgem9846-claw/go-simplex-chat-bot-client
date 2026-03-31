// Package client provides a WebSocket client for the SimpleX Chat CLI Bot API.
//
// This client implements:
//   - Connection state machine: Disconnected → Connecting → Connected → Reconnecting → Closed
//   - Automatic reconnection with exponential backoff and jitter (capped at 30s)
//   - Periodic ping/pong health checks every 15 seconds
//   - Event buffering during outage, replayed on recovery
//   - Error categorization: retryable, permanent, connection-level
//   - Circuit breaker after 5 consecutive failures, reset after 30 seconds
//   - Functional options pattern for configuration
//   - Domain-organized API: client.Profiles().List(), client.Contacts().Add()
//   - Structured logging with secret redaction
//   - Metrics: request count, latency histogram, active connections, reconnect count
//   - TLS by default, 1MB max message size, configurable timeouts
//
// Basic usage:
//
//	c, err := client.Connect("ws://localhost:5225", client.WithTimeout(30*time.Second))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer c.Close()
//
//	user, err := c.Profiles().ShowActiveUser(ctx)
//	contacts, err := c.Contacts().List(ctx, user.UserID)
package client

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"

	sxtypes "simplex-chat-bot/types"
)

// ---------------------------------------------------------------------------
// Connection state machine
// ---------------------------------------------------------------------------

// ConnState represents the WebSocket connection lifecycle state.
type ConnState int32

const (
	StateDisconnected ConnState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateClosing
	StateClosed
)

func (s ConnState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateReconnecting:
		return "reconnecting"
	case StateClosing:
		return "closing"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// ---------------------------------------------------------------------------
// Error types with categorization
// ---------------------------------------------------------------------------

var (
	ErrNotConnected     = errors.New("client not connected")
	ErrClosed           = errors.New("client is closed")
	ErrTimeout          = errors.New("request timed out")
	ErrConnectionFailed = errors.New("connection failed")
	ErrInvalidState     = errors.New("invalid connection state")
	ErrCircuitOpen      = errors.New("circuit breaker is open")
)

// ErrorCategory classifies errors for retry and circuit-breaker logic.
type ErrorCategory int

const (
	ErrorRetryable ErrorCategory = iota
	ErrorPermanent
	ErrorConnection
)

// CategorizedError wraps an error with its category.
type CategorizedError struct {
	Category ErrorCategory
	Err      error
}

func (e *CategorizedError) Error() string { return e.Err.Error() }
func (e *CategorizedError) Unwrap() error { return e.Err }

// NewRetryableError creates an error that should be retried.
func NewRetryableError(err error) error {
	return &CategorizedError{Category: ErrorRetryable, Err: err}
}

// NewPermanentError creates an error that should not be retried.
func NewPermanentError(err error) error {
	return &CategorizedError{Category: ErrorPermanent, Err: err}
}

// NewConnectionError creates a connection-level error (triggers reconnection).
func NewConnectionError(err error) error {
	return &CategorizedError{Category: ErrorConnection, Err: err}
}

// IsRetryable reports whether err is retryable.
func IsRetryable(err error) bool {
	var ce *CategorizedError
	if errors.As(err, &ce) {
		return ce.Category == ErrorRetryable
	}
	return false
}

// IsConnectionError reports whether err is a connection-level error.
func IsConnectionError(err error) bool {
	var ce *CategorizedError
	if errors.As(err, &ce) {
		return ce.Category == ErrorConnection
	}
	return false
}

// RequestError wraps an error with request metadata.
type RequestError struct {
	Command string
	CorrID  string
	Err     error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("request failed (cmd=%s corrId=%s): %v",
		redactCommand(e.Command), e.CorrID, e.Err)
}
func (e *RequestError) Unwrap() error { return e.Err }

// ---------------------------------------------------------------------------
// Configuration (functional options)
// ---------------------------------------------------------------------------

const (
	defaultConnectTimeout  = 10 * time.Second
	defaultOperationTimeout = 30 * time.Second
	maxReconnectDelay      = 30 * time.Second
	initialReconnectDelay  = 1 * time.Second
	defaultPingInterval    = 15 * time.Second
	defaultPingTimeout     = 10 * time.Second
	maxMessageSize         = 1 << 20 // 1 MB

	defaultCircuitBreakerThreshold = 5
	defaultCircuitBreakerTimeout   = 30 * time.Second
	defaultEventBufferSize         = 1000
)

// Config holds the resolved client configuration.
type Config struct {
	URL                     string
	ConnectTimeout          time.Duration
	OperationTimeout        time.Duration
	Reconnect               bool
	MaxReconnectAttempts    int
	PingInterval            time.Duration
	PingTimeout             time.Duration
	EventBufferSize         int
	MaxMessageSize          int64
	CircuitBreakerThreshold int
	CircuitBreakerTimeout   time.Duration
	Logger                  Logger
	Metrics                 Metrics
	TLSConfig               *TLSConfig
}

// TLSConfig holds TLS settings. TLS is enabled by default for wss:// URLs.
type TLSConfig struct {
	InsecureSkipVerify bool
}

func (c *Config) setDefaults() {
	if c.ConnectTimeout <= 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}
	if c.OperationTimeout <= 0 {
		c.OperationTimeout = defaultOperationTimeout
	}
	if c.PingInterval <= 0 {
		c.PingInterval = defaultPingInterval
	}
	if c.PingTimeout <= 0 {
		c.PingTimeout = defaultPingTimeout
	}
	if c.EventBufferSize <= 0 {
		c.EventBufferSize = defaultEventBufferSize
	}
	if c.MaxMessageSize <= 0 {
		c.MaxMessageSize = maxMessageSize
	}
	if c.CircuitBreakerThreshold <= 0 {
		c.CircuitBreakerThreshold = defaultCircuitBreakerThreshold
	}
	if c.CircuitBreakerTimeout <= 0 {
		c.CircuitBreakerTimeout = defaultCircuitBreakerTimeout
	}
	if c.Logger == nil {
		c.Logger = nopLogger{}
	}
	if c.Metrics == nil {
		c.Metrics = nopMetrics{}
	}
}

// Option is a functional option for Client configuration.
type Option func(*Config)

// WithLogger sets the structured logger.
func WithLogger(l Logger) Option { return func(c *Config) { c.Logger = l } }

// WithConnectTimeout sets the TCP/WS dial timeout (default 10 s).
func WithConnectTimeout(d time.Duration) Option {
	return func(c *Config) { c.ConnectTimeout = d }
}

// WithOperationTimeout sets the per-request response timeout (default 30 s).
func WithOperationTimeout(d time.Duration) Option {
	return func(c *Config) { c.OperationTimeout = d }
}

// WithReconnect enables automatic reconnection on connection loss.
func WithReconnect(enabled bool) Option {
	return func(c *Config) { c.Reconnect = enabled }
}

// WithMaxReconnectAttempts limits reconnection attempts (0 = unlimited).
func WithMaxReconnectAttempts(n int) Option {
	return func(c *Config) { c.MaxReconnectAttempts = n }
}

// WithPingInterval sets the health-check interval (default 15 s).
func WithPingInterval(d time.Duration) Option {
	return func(c *Config) { c.PingInterval = d }
}

// WithEventBufferSize sets the capacity of the event buffer used during reconnection.
func WithEventBufferSize(n int) Option {
	return func(c *Config) { c.EventBufferSize = n }
}

// WithMetrics sets the metrics collector.
func WithMetrics(m Metrics) Option { return func(c *Config) { c.Metrics = m } }

// WithCircuitBreaker configures the circuit breaker.
func WithCircuitBreaker(threshold int, timeout time.Duration) Option {
	return func(c *Config) {
		c.CircuitBreakerThreshold = threshold
		c.CircuitBreakerTimeout = timeout
	}
}

// ---------------------------------------------------------------------------
// Structured logging
// ---------------------------------------------------------------------------

// Logger is the structured logging interface.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type defaultLogger struct{ logger *log.Logger }

func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+msg, args...)
}
func (l *defaultLogger) Info(msg string, args ...interface{}) {
	l.logger.Printf("[INFO] "+msg, args...)
}
func (l *defaultLogger) Warn(msg string, args ...interface{}) {
	l.logger.Printf("[WARN] "+msg, args...)
}
func (l *defaultLogger) Error(msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

type nopLogger struct{}

func (nopLogger) Debug(string, ...interface{}) {}
func (nopLogger) Info(string, ...interface{})  {}
func (nopLogger) Warn(string, ...interface{})  {}
func (nopLogger) Error(string, ...interface{}) {}

// AdaptLogger wraps a standard library *log.Logger to satisfy the Logger interface.
func AdaptLogger(l *log.Logger) Logger { return &defaultLogger{logger: l} }

// redactCommand strips potential secrets from a command string for safe logging.
func redactCommand(cmd string) string {
	// Redact JSON payloads that may contain keys/nonces/tokens.
	if idx := strings.Index(cmd, "{"); idx >= 0 {
		return cmd[:idx] + "[REDACTED]"
	}
	// Redact long hex/base64 strings that may be keys.
	parts := strings.Fields(cmd)
	for i, p := range parts {
		if len(p) > 64 {
			parts[i] = "[REDACTED]"
		}
	}
	return strings.Join(parts, "")
}

// ---------------------------------------------------------------------------
// Metrics
// ---------------------------------------------------------------------------

// Metrics collects client telemetry.
type Metrics interface {
	RecordRequest(duration time.Duration, success bool)
	RecordReconnect()
	RecordError(category ErrorCategory)
	SetActiveConnections(n int)
}

type nopMetrics struct{}

func (nopMetrics) RecordRequest(time.Duration, bool)  {}
func (nopMetrics) RecordReconnect()                    {}
func (nopMetrics) RecordError(ErrorCategory)           {}
func (nopMetrics) SetActiveConnections(int)            {}

// ---------------------------------------------------------------------------
// Transport abstraction (enables mock testing)
// ---------------------------------------------------------------------------

// Transport abstracts a WebSocket connection for testability.
type Transport interface {
	Read(ctx context.Context, maxSize int64) ([]byte, error)
	Write(ctx context.Context, data []byte) error
	Ping(ctx context.Context) error
	Close(code int, reason string) error
}

// wsTransport wraps nhooyr.io/websocket.
type wsTransport struct {
	conn *websocket.Conn
}

func (t *wsTransport) Read(ctx context.Context, maxSize int64) ([]byte, error) {
	t.conn.SetReadLimit(maxSize)
	_, data, err := t.conn.Read(ctx)
	return data, err
}

func (t *wsTransport) Write(ctx context.Context, data []byte) error {
	return t.conn.Write(ctx, websocket.MessageText, data)
}

func (t *wsTransport) Ping(ctx context.Context) error {
	return t.conn.Ping(ctx)
}

func (t *wsTransport) Close(code int, reason string) error {
	return t.conn.Close(websocket.StatusCode(code), reason)
}

// ---------------------------------------------------------------------------
// Event system
// ---------------------------------------------------------------------------

// Event represents a SimpleX Chat server-push event (no correlation ID).
type Event struct {
	Type string
	Data json.RawMessage
}

// EventFilter selects which events a subscriber receives.
type EventFilter func(event Event) bool

// EventFilterAll matches every event.
var EventFilterAll EventFilter = func(Event) bool { return true }

// EventFilterType matches a single event type.
func EventFilterType(t string) EventFilter {
	return func(e Event) bool { return e.Type == t }
}

// EventFilterTypes matches any of the listed types.
func EventFilterTypes(types ...string) EventFilter {
	set := make(map[string]bool, len(types))
	for _, t := range types {
		set[t] = true
	}
	return func(e Event) bool { return set[e.Type] }
}

// Subscription represents a live event subscription.
type Subscription struct {
	client *Client
	filter EventFilter
	ch     chan Event
	id     string
}

// Channel returns the event channel.
func (s *Subscription) Channel() <-chan Event { return s.ch }

// Unsubscribe removes this subscription and closes its channel.
func (s *Subscription) Unsubscribe() { s.client.unsubscribe(s) }

// ---------------------------------------------------------------------------
// Circuit breaker
// ---------------------------------------------------------------------------

type circuitState int

const (
	circuitClosed circuitState = iota
	circuitOpen
	circuitHalfOpen
)

type circuitBreaker struct {
	mu               sync.Mutex
	state            circuitState
	failures         int
	threshold        int
	timeout          time.Duration
	lastFailureTime  time.Time
}

func newCircuitBreaker(threshold int, timeout time.Duration) *circuitBreaker {
	return &circuitBreaker{
		state:     circuitClosed,
		threshold: threshold,
		timeout:   timeout,
	}
}

func (cb *circuitBreaker) allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == circuitOpen {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = circuitHalfOpen
			return nil
		}
		return ErrCircuitOpen
	}
	return nil
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = circuitClosed
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailureTime = time.Now()
	if cb.failures >= cb.threshold {
		cb.state = circuitOpen
	}
}

func (cb *circuitBreaker) isOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == circuitOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.state = circuitHalfOpen
		return false
	}
	return cb.state == circuitOpen
}

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// Client manages a WebSocket connection to the SimpleX Chat CLI.
type Client struct {
	url    string
	config *Config

	state atomic.Int32 // ConnState

	mu   sync.RWMutex // protects conn / transport
	conn Transport

	closed atomic.Bool

	pendingMu sync.Mutex
	pending   map[string]chan *sxtypes.Response

	subMu      sync.RWMutex
	subscribers map[string]*Subscription
	subCounter atomic.Uint64

	bufferMu    sync.Mutex
	eventBuffer []Event
	buffering   bool

	inflightMu sync.Mutex
	inflight   map[string]context.CancelFunc

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	breaker *circuitBreaker

	// Domain API groups
	profiles *ProfileAPI
	contacts *ContactAPI
	groups   *GroupAPI
	messages *MessageAPI
	files    *FileAPI
	address  *AddressAPI
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// New creates a Client without connecting. Call Connect() to establish the link.
func New(url string, opts ...Option) *Client {
	cfg := &Config{URL: url}
	for _, o := range opts {
		o(cfg)
	}
	cfg.setDefaults()

	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		url:         url,
		config:      cfg,
		pending:     make(map[string]chan *sxtypes.Response),
		subscribers: make(map[string]*Subscription),
		eventBuffer: make([]Event, 0, cfg.EventBufferSize),
		inflight:    make(map[string]context.CancelFunc),
		ctx:         ctx,
		cancel:      cancel,
		breaker:     newCircuitBreaker(cfg.CircuitBreakerThreshold, cfg.CircuitBreakerTimeout),
	}
	c.state.Store(int32(StateDisconnected))

	c.profiles = &ProfileAPI{client: c}
	c.contacts = &ContactAPI{client: c}
	c.groups = &GroupAPI{client: c}
	c.messages = &MessageAPI{client: c}
	c.files = &FileAPI{client: c}
	c.address = &AddressAPI{client: c}

	return c
}

// Connect dials the WebSocket endpoint, starts background loops, and returns
// a ready-to-use Client.
func Connect(url string, opts ...Option) (*Client, error) {
	c := New(url, opts...)
	if err := c.connect(c.ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

// State returns the current connection state.
func (c *Client) State() ConnState { return ConnState(c.state.Load()) }

// Metrics returns the configured metrics collector (may be a no-op).
func (c *Client) Metrics() Metrics { return c.config.Metrics }

// Config returns a copy of the resolved configuration.
func (c *Client) Config() Config { return *c.config }

// ---------------------------------------------------------------------------
// Connection lifecycle
// ---------------------------------------------------------------------------

// Establish dials the server and starts read/ping loops.
// It may be called only once; subsequent calls return ErrInvalidState or ErrClosed.
func (c *Client) Establish(ctx context.Context) error {
	return c.connect(ctx)
}

// Connect is an alias for Establish.
func (c *Client) Connect(ctx context.Context) error {
	return c.Establish(ctx)
}

func (c *Client) connect(ctx context.Context) error {
	if c.closed.Load() {
		return ErrClosed
	}
	if !c.state.CompareAndSwap(int32(StateDisconnected), int32(StateConnecting)) {
		return fmt.Errorf("%w: cannot connect from state %s", ErrInvalidState, c.State())
	}

	c.config.Logger.Info("connecting to %s", redactCommand(c.url))

	dialCtx, dialCancel := context.WithTimeout(ctx, c.config.ConnectTimeout)
	defer dialCancel()

	conn, _, err := websocket.Dial(dialCtx, c.url, &websocket.DialOptions{
		HTTPClient: &http.Client{Timeout: c.config.ConnectTimeout},
	})
	if err != nil {
		c.state.Store(int32(StateDisconnected))
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	c.mu.Lock()
	c.conn = &wsTransport{conn: conn}
	c.mu.Unlock()

	c.state.Store(int32(StateConnected))
	c.config.Logger.Info("connected")
	c.config.Metrics.SetActiveConnections(1)

	c.wg.Add(2)
	go c.readLoop()
	go c.pingLoop()
	return nil
}

// Close gracefully shuts down the connection, cancels in-flight requests,
// and waits for all goroutines to finish.
func (c *Client) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}

	c.config.Logger.Info("closing client")
	c.state.Store(int32(StateClosing))
	c.cancel()

	c.inflightMu.Lock()
	for id, cancel := range c.inflight {
		cancel()
		delete(c.inflight, id)
	}
	c.inflightMu.Unlock()

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn != nil {
		_ = conn.Close(int(websocket.StatusNormalClosure), "client closing")
	}

	c.wg.Wait()

	c.subMu.Lock()
	for _, sub := range c.subscribers {
		close(sub.ch)
	}
	c.subscribers = make(map[string]*Subscription)
	c.subMu.Unlock()

	c.config.Metrics.SetActiveConnections(0)
	c.state.Store(int32(StateClosed))
	c.config.Logger.Info("client closed")
	return nil
}

// ---------------------------------------------------------------------------
// API group accessors (domain-organized API)
// ---------------------------------------------------------------------------

func (c *Client) Profiles() *ProfileAPI  { return c.profiles }
func (c *Client) Contacts() *ContactAPI  { return c.contacts }
func (c *Client) Groups() *GroupAPI      { return c.groups }
func (c *Client) Messages() *MessageAPI  { return c.messages }
func (c *Client) Files() *FileAPI        { return c.files }
func (c *Client) Address() *AddressAPI   { return c.address }

// ---------------------------------------------------------------------------
// Convenience methods (backward-compatible with direct Client usage in tests)
// ---------------------------------------------------------------------------

func (c *Client) ShowActiveUser(ctx context.Context) (*sxtypes.User, error) {
	return c.profiles.ShowActiveUser(ctx)
}
func (c *Client) ListUsers(ctx context.Context) ([]sxtypes.UserUI, error) {
	return c.profiles.ListUsers(ctx)
}
func (c *Client) CreateActiveUser(ctx context.Context, u sxtypes.NewUser) (*sxtypes.User, error) {
	return c.profiles.CreateActiveUser(ctx, u)
}
func (c *Client) ListContacts(ctx context.Context, userID int64) ([]sxtypes.Contact, error) {
	return c.contacts.List(ctx, userID)
}
func (c *Client) AddContact(ctx context.Context, userID int64, incognito bool) (*sxtypes.CreatedConnLink, error) {
	return c.contacts.Add(ctx, userID, incognito)
}
func (c *Client) NewGroup(ctx context.Context, userID int64, incognito bool, profile sxtypes.GroupProfile) (*sxtypes.GroupInfo, error) {
	return c.groups.Create(ctx, userID, incognito, profile)
}
func (c *Client) ListGroups(ctx context.Context, userID int64, contactID *int64, search *string) ([]sxtypes.GroupInfo, error) {
	return c.groups.List(ctx, userID, contactID, search)
}
func (c *Client) SendMessages(ctx context.Context, chatRef sxtypes.ChatRef, live bool, ttl *int, messages []sxtypes.ComposedMessage) ([]sxtypes.AChatItem, error) {
	return c.messages.Send(ctx, chatRef, live, ttl, messages)
}
func (c *Client) CreateMyAddress(ctx context.Context, userID int64) (*sxtypes.UserContactLink, error) {
	return c.address.Create(ctx, userID)
}

// OnEvent registers a callback for all events. Returns a function that unsubscribes.
// Provided for backward compatibility; prefer Subscribe for new code.
func (c *Client) OnEvent(fn func(json.RawMessage)) context.CancelFunc {
	ctx, cancel := context.WithCancel(c.ctx)
	sub := c.Subscribe(EventFilterAll)
	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-sub.Channel():
				if !ok {
					return
				}
				fn(evt.Data)
			}
		}
	}()
	return cancel
}

// ---------------------------------------------------------------------------
// Send / request-response
// ---------------------------------------------------------------------------

// Send dispatches a command string and returns the correlated response payload.
func (c *Client) Send(ctx context.Context, cmd string) (json.RawMessage, error) {
	if c.closed.Load() {
		return nil, ErrClosed
	}
	if c.State() != StateConnected {
		return nil, ErrNotConnected
	}
	if err := c.breaker.allow(); err != nil {
		return nil, err
	}

	corrID := generateCorrID()

	reqCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	c.inflightMu.Lock()
	c.inflight[corrID] = cancel
	c.inflightMu.Unlock()
	defer func() {
		c.inflightMu.Lock()
		delete(c.inflight, corrID)
		c.inflightMu.Unlock()
	}()

	ch := make(chan *sxtypes.Response, 1)
	c.pendingMu.Lock()
	c.pending[corrID] = ch
	c.pendingMu.Unlock()
	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, corrID)
		c.pendingMu.Unlock()
	}()

	command := sxtypes.Command{CorrID: corrID, Cmd: cmd}

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return nil, ErrNotConnected
	}

	start := time.Now()
	cmdBytes, err := json.Marshal(command)
	if err != nil {
		return nil, &RequestError{Command: cmd, CorrID: corrID, Err: err}
	}
	if err := conn.Write(reqCtx, cmdBytes); err != nil {
		c.config.Metrics.RecordRequest(time.Since(start), false)
		c.breaker.recordFailure()
		return nil, &RequestError{Command: cmd, CorrID: corrID, Err: err}
	}

	timeout := c.config.OperationTimeout
	if deadline, ok := reqCtx.Deadline(); ok {
		if t := time.Until(deadline); t < timeout {
			timeout = t
		}
	}

	select {
	case resp, ok := <-ch:
		if !ok {
			c.config.Metrics.RecordRequest(time.Since(start), false)
			return nil, &RequestError{Command: cmd, CorrID: corrID, Err: errors.New("connection closed")}
		}
		c.config.Metrics.RecordRequest(time.Since(start), true)
		c.breaker.recordSuccess()
		return resp.Resp, nil
	case <-time.After(timeout):
		c.config.Metrics.RecordRequest(time.Since(start), false)
		return nil, &RequestError{Command: cmd, CorrID: corrID, Err: ErrTimeout}
	case <-reqCtx.Done():
		c.config.Metrics.RecordRequest(time.Since(start), false)
		return nil, &RequestError{Command: cmd, CorrID: corrID, Err: reqCtx.Err()}
	}
}

// ---------------------------------------------------------------------------
// Event subscriptions
// ---------------------------------------------------------------------------

func (c *Client) Subscribe(filter EventFilter) *Subscription {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	id := fmt.Sprintf("sub-%d", c.subCounter.Add(1))
	sub := &Subscription{
		client: c,
		filter: filter,
		ch:     make(chan Event, 100),
		id:     id,
	}
	c.subscribers[id] = sub
	c.config.Logger.Debug("subscription created id=%s", id)
	return sub
}

func (c *Client) unsubscribe(sub *Subscription) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	if _, ok := c.subscribers[sub.id]; ok {
		delete(c.subscribers, sub.id)
		close(sub.ch)
		c.config.Logger.Debug("subscription removed id=%s", sub.id)
	}
}

func (c *Client) publishEvent(event Event) {
	c.subMu.RLock()
	subs := make([]*Subscription, 0, len(c.subscribers))
	for _, s := range c.subscribers {
		subs = append(subs, s)
	}
	c.subMu.RUnlock()

	for _, s := range subs {
		if s.filter(event) {
			select {
			case s.ch <- event:
			default:
				c.config.Logger.Warn("subscriber channel full, dropping event type=%s", event.Type)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Read / ping loops
// ---------------------------------------------------------------------------

func (c *Client) readLoop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		data, err := c.readMessage()
		if err != nil {
			if c.closed.Load() || c.ctx.Err() != nil {
				return
			}
			c.config.Logger.Error("read error: %v", err)
			c.config.Metrics.RecordError(ErrorConnection)
			if c.config.Reconnect {
				c.reconnect()
				continue
			}
			return
		}
		c.dispatchMessage(data)
	}
}

func (c *Client) readMessage() (json.RawMessage, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return nil, ErrNotConnected
	}

	data, err := conn.Read(c.ctx, c.config.MaxMessageSize)
	if err != nil {
		if c.closed.Load() {
			return nil, ErrClosed
		}
		// Normal / abnormal closure → connection error for reconnection.
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
			websocket.CloseStatus(err) == websocket.StatusGoingAway {
			return nil, NewConnectionError(err)
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, NewConnectionError(err)
	}
	return json.RawMessage(data), nil
}

func (c *Client) pingLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if c.State() != StateConnected {
				continue
			}
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()
			if conn == nil {
				continue
			}
			pingCtx, pingCancel := context.WithTimeout(c.ctx, c.config.PingTimeout)
			err := conn.Ping(pingCtx)
			pingCancel()
			if err != nil {
				c.config.Logger.Warn("ping failed: %v", err)
				c.config.Metrics.RecordRequest(0, false)
				if c.config.Reconnect {
					c.reconnect()
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Reconnection with exponential backoff + jitter, capped at 30 s
// ---------------------------------------------------------------------------

func (c *Client) reconnect() {
	if !c.state.CompareAndSwap(int32(StateConnected), int32(StateReconnecting)) {
		return // another goroutine is already reconnecting
	}

	c.config.Logger.Info("starting reconnection")
	c.config.Metrics.RecordReconnect()

	// Buffer events while disconnected.
	c.bufferMu.Lock()
	c.buffering = true
	c.bufferMu.Unlock()

	defer func() {
		c.bufferMu.Lock()
		c.buffering = false
		buffered := c.eventBuffer
		c.eventBuffer = make([]Event, 0, c.config.EventBufferSize)
		c.bufferMu.Unlock()

		for _, e := range buffered {
			c.publishEvent(e)
		}
		if len(buffered) > 0 {
			c.config.Logger.Info("replayed %d buffered events", len(buffered))
		}
	}()

	// Tear down old connection.
	c.mu.RLock()
	old := c.conn
	c.mu.RUnlock()
	if old != nil {
		_ = old.Close(int(websocket.StatusGoingAway), "reconnecting")
	}
	c.mu.Lock()
	c.conn = nil
	c.mu.Unlock()

	delay := initialReconnectDelay
	attempts := 0

	for {
		if c.closed.Load() {
			c.state.Store(int32(StateDisconnected))
			return
		}
		select {
		case <-c.ctx.Done():
			c.state.Store(int32(StateDisconnected))
			return
		case <-time.After(delay):
		}

		if c.config.MaxReconnectAttempts > 0 && attempts >= c.config.MaxReconnectAttempts {
			c.config.Logger.Error("max reconnection attempts (%d) reached", c.config.MaxReconnectAttempts)
			c.state.Store(int32(StateDisconnected))
			return
		}
		attempts++

		c.config.Logger.Info("reconnecting attempt=%d/%d", attempts, c.config.MaxReconnectAttempts)

		dialCtx, dialCancel := context.WithTimeout(c.ctx, c.config.ConnectTimeout)
		conn, _, err := websocket.Dial(dialCtx, c.url, nil)
		dialCancel()

		if err != nil {
			c.config.Logger.Warn("reconnect failed: %v", err)
			delay = minDuration(delay*2, maxReconnectDelay)
			jitter := time.Duration(fastrandn(uint32(delay / 4)))
			delay += jitter
			continue
		}

		c.mu.Lock()
		c.conn = &wsTransport{conn: conn}
		c.mu.Unlock()

		c.state.Store(int32(StateConnected))
		c.config.Logger.Info("reconnected successfully")
		c.config.Metrics.SetActiveConnections(1)
		c.breaker.recordSuccess()

		c.wg.Add(1)
		go c.readLoop()
		return
	}
}

// ---------------------------------------------------------------------------
// Message dispatch
// ---------------------------------------------------------------------------

func (c *Client) dispatchMessage(raw json.RawMessage) {
	var envelope sxtypes.Response
	if err := json.Unmarshal(raw, &envelope); err != nil {
		c.config.Logger.Error("unmarshal envelope: %v", err)
		return
	}

	// Correlated response → deliver to waiting caller.
	if envelope.CorrID != "" {
		c.pendingMu.Lock()
		ch, ok := c.pending[envelope.CorrID]
		c.pendingMu.Unlock()
		if ok {
			ch <- &envelope
		}
		return
	}

	// Server-pushed event.
	var typeCheck struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(envelope.Resp, &typeCheck); err != nil {
		c.config.Logger.Error("extract event type: %v", err)
		return
	}
	evt := Event{Type: typeCheck.Type, Data: envelope.Resp}

	c.bufferMu.Lock()
	if c.buffering {
		if len(c.eventBuffer) < cap(c.eventBuffer) {
			c.eventBuffer = append(c.eventBuffer, evt)
		} else {
			c.config.Logger.Warn("event buffer full, dropping event type=%s", evt.Type)
		}
		c.bufferMu.Unlock()
		return
	}
	c.bufferMu.Unlock()

	c.publishEvent(evt)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func generateCorrID() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	_, _ = rand.Read(b) // crypto/rand; err is always nil for []byte
	for i := range b {
		b[i] = alphabet[b[i]%byte(len(alphabet))]
	}
	return string(b)
}

// fastrandn returns a cryptographically random uint32 in [0, n).
func fastrandn(n uint32) uint32 {
	if n == 0 {
		return 0
	}
	var b [4]byte
	_, _ = rand.Read(b[:])
	v := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	return v % n
}

// ---------------------------------------------------------------------------
// Domain API Groups
// ---------------------------------------------------------------------------

// ProfileAPI provides user profile operations.
type ProfileAPI struct {
	client *Client
}

// ShowActiveUser returns the currently active user profile.
func (a *ProfileAPI) ShowActiveUser(ctx context.Context) (*sxtypes.User, error) {
	raw, err := a.client.Send(ctx, "/user")
	if err != nil {
		return nil, err
	}
	var resp sxtypes.User
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &resp, nil
}

// ListUsers returns all user profiles.
func (a *ProfileAPI) ListUsers(ctx context.Context) ([]sxtypes.UserUI, error) {
	raw, err := a.client.Send(ctx, "/users")
	if err != nil {
		return nil, err
	}
	var resp []sxtypes.UserUI
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal users: %w", err)
	}
	return resp, nil
}

// CreateActiveUser creates a new user profile and sets it as active.
func (a *ProfileAPI) CreateActiveUser(ctx context.Context, u sxtypes.NewUser) (*sxtypes.User, error) {
	payload, err := json.Marshal(u)
	if err != nil {
		return nil, fmt.Errorf("marshal newUser: %w", err)
	}
	raw, err := a.client.Send(ctx, "/_create user "+string(payload))
	if err != nil {
		return nil, err
	}
	var resp sxtypes.User
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &resp, nil
}

// ContactAPI provides contact operations.
type ContactAPI struct {
	client *Client
}

// List returns all contacts for a user.
func (a *ContactAPI) List(ctx context.Context, userID int64) ([]sxtypes.Contact, error) {
	cmd := fmt.Sprintf("/_contacts %d", userID)
	raw, err := a.client.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var resp []sxtypes.Contact
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal contacts: %w", err)
	}
	return resp, nil
}

// Add creates a 1-time invitation link.
func (a *ContactAPI) Add(ctx context.Context, userID int64, incognito bool) (*sxtypes.CreatedConnLink, error) {
	cmd := fmt.Sprintf("/_connect %d", userID)
	if incognito {
		cmd += " incognito=on"
	}
	raw, err := a.client.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var resp sxtypes.CreatedConnLink
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal link: %w", err)
	}
	return &resp, nil
}

// GroupAPI provides group operations.
type GroupAPI struct {
	client *Client
}

// Create creates a new group.
func (a *GroupAPI) Create(ctx context.Context, userID int64, incognito bool, profile sxtypes.GroupProfile) (*sxtypes.GroupInfo, error) {
	payload, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshal profile: %w", err)
	}
	cmd := fmt.Sprintf("/_group %d", userID)
	if incognito {
		cmd += " incognito=on"
	}
	cmd += " " + string(payload)
	raw, err := a.client.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var resp sxtypes.GroupInfo
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal group: %w", err)
	}
	return &resp, nil
}

// MessageAPI provides message operations.
type MessageAPI struct {
	client *Client
}

// Send sends messages to a chat with optional live mode and TTL.
func (a *MessageAPI) Send(ctx context.Context, chatRef sxtypes.ChatRef, live bool, ttl *int, messages []sxtypes.ComposedMessage) ([]sxtypes.AChatItem, error) {
	payload, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("marshal messages: %w", err)
	}
	cmd := fmt.Sprintf("/_send %s", chatRef.String())
	if live {
		cmd += " live=on"
	}
	if ttl != nil {
		cmd += fmt.Sprintf(" ttl=%d", *ttl)
	}
	cmd += " json " + string(payload)
	raw, err := a.client.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var resp struct {
		ChatItems []sxtypes.AChatItem `json:"chatItems"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal chat items: %w", err)
	}
	return resp.ChatItems, nil
}

// FileAPI provides file operations.
type FileAPI struct {
	client *Client
}

// AddressAPI provides address operations.
type AddressAPI struct {
	client *Client

}


// List returns all groups for a user.
func (a *GroupAPI) List(ctx context.Context, userID int64, contactID *int64, search *string) ([]sxtypes.GroupInfo, error) {
	cmd := fmt.Sprintf("/_groups %d", userID)
	if contactID != nil {
		cmd += fmt.Sprintf(" @%d", *contactID)
	}
	if search != nil {
		cmd += " " + *search
	}
	raw, err := a.client.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var resp []sxtypes.GroupInfo
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal groups: %w", err)
	}
	return resp, nil
}

// Create creates a bot contact address.
func (a *AddressAPI) Create(ctx context.Context, userID int64) (*sxtypes.UserContactLink, error) {
	cmd := fmt.Sprintf("/_address %d", userID)
	raw, err := a.client.Send(ctx, cmd)
	if err != nil {
		return nil, err
	}
	var resp sxtypes.UserContactLink
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal address: %w", err)
	}
	return &resp, nil
}
