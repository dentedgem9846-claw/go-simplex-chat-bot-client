package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	sxtypes "simplex-chat-bot/types"
)

// mockServer creates a test WebSocket server that responds to commands
func mockServer(t *testing.T, handler func(ws *websocket.Conn, msg sxtypes.Command)) (string, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			t.Fatalf("accept websocket: %v", err)
		}
		defer ws.Close(websocket.StatusNormalClosure, "")

		for {
			var msg sxtypes.Command
			err := wsjson.Read(r.Context(), ws, &msg)
			if err != nil {
				var closeErr websocket.CloseError
				if errors.As(err, &closeErr) {
					return
				}
				t.Logf("read error: %v", err)
				return
			}
			handler(ws, msg)
		}
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	return wsURL, func() {
		server.Close()
	}
}

func TestNew(t *testing.T) {
	c := New("ws://localhost:5225")

	if c == nil {
		t.Fatal("expected client to be created")
	}
	if c.url != "ws://localhost:5225" {
		t.Errorf("expected url to be ws://localhost:5225, got %s", c.url)
	}
	if c.config.OperationTimeout != defaultOperationTimeout {
		t.Errorf("expected default timeout %v, got %v", defaultOperationTimeout, c.config.OperationTimeout)
	}

	// Test with options
	c2 := New("ws://localhost:5226",
		WithOperationTimeout(10*time.Second),
		WithReconnect(true),
	)
	if c2.config.OperationTimeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", c2.config.OperationTimeout)
	}
	if !c2.config.Reconnect {
		t.Error("expected reconnect to be true")
	}
}

func TestConnect(t *testing.T) {
	ctx := context.Background()

	// Test successful connection
	wsURL, cleanup := mockServer(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		// Echo back a response
		resp := sxtypes.Response{
			CorrID: msg.CorrID,
			Resp:   json.RawMessage(`{"type":"ok"}`),
		}
		wsjson.Write(context.Background(), ws, resp)
	})
	defer cleanup()

	c := New(wsURL)
	err := c.Connect(ctx)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer c.Close()

	if c.conn == nil {
		t.Error("expected connection to be established")
	}
}

func TestConnect_InvalidURL(t *testing.T) {
	ctx := context.Background()
	c := New("ws://[invalid")
	err := c.Connect(ctx)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestClose(t *testing.T) {
	ctx := context.Background()

	wsURL, cleanup := mockServer(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		resp := sxtypes.Response{
			CorrID: msg.CorrID,
			Resp:   json.RawMessage(`{"type":"ok"}`),
		}
		wsjson.Write(context.Background(), ws, resp)
	})
	defer cleanup()

	c := New(wsURL)
	err := c.Connect(ctx)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	err = c.Close()
	if err != nil {
		t.Errorf("close failed: %v", err)
	}

	if !c.closed.Load() {
		t.Error("expected client to be marked as closed")
	}

	// Double close should be safe
	err = c.Close()
	if err != nil {
		t.Errorf("second close failed: %v", err)
	}
}

func TestSend(t *testing.T) {
	ctx := context.Background()

	wsURL, cleanup := mockServer(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		// Verify the command structure
		if msg.CorrID == "" {
			t.Error("expected corrId to be set")
		}
		if msg.Cmd == "" {
			t.Error("expected cmd to be set")
		}

		// Echo back a response
		resp := sxtypes.Response{
			CorrID: msg.CorrID,
			Resp:   json.RawMessage(`{"type":"testResponse","data":"hello"}`),
		}
		wsjson.Write(context.Background(), ws, resp)
	})
	defer cleanup()

	c := New(wsURL)
	err := c.Connect(ctx)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer c.Close()

	resp, err := c.Send(ctx, "/test command")
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if result["type"] != "testResponse" {
		t.Errorf("expected type testResponse, got %v", result["type"])
	}
}

func TestSend_Timeout(t *testing.T) {
	ctx := context.Background()

	// Server that doesn't respond
	wsURL, cleanup := mockServer(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		// Don't respond - let it timeout
	})
	defer cleanup()

	c := New(wsURL, WithOperationTimeout(100*time.Millisecond))
	err := c.Connect(ctx)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer c.Close()

	_, err = c.Send(ctx, "/test")
	if err == nil {
		t.Error("expected timeout error")
	}
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got: %v", err)
	}
}

func TestSend_NotConnected(t *testing.T) {
	ctx := context.Background()
	c := New("ws://localhost:5225")

	_, err := c.Send(ctx, "/test")
	if err == nil {
		t.Error("expected error when not connected")
	}
	if !errors.Is(err, ErrNotConnected) {
		t.Errorf("expected ErrNotConnected, got: %v", err)
	}
}

func TestSend_Closed(t *testing.T) {
	ctx := context.Background()

	wsURL, cleanup := mockServer(t, func(ws *websocket.Conn, msg sxtypes.Command) {})
	defer cleanup()

	c := New(wsURL)
	err := c.Connect(ctx)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	c.Close()

	err = c.Connect(ctx)
	if err == nil {
		t.Error("expected error when connecting to closed client")
	}
}

func TestSubscribe(t *testing.T) {
	ctx := context.Background()

	wsURL, cleanup := mockServer(t, func(ws *websocket.Conn, msg sxtypes.Command) {
		// Send an event (no corrId)
		event := sxtypes.Response{
			Resp: json.RawMessage(`{"type":"newChatItems","user":{"userId":1}}`),
		}
		wsjson.Write(context.Background(), ws, event)
	})
	defer cleanup()

	c := New(wsURL)

	// Subscribe to events
	sub := c.Subscribe(EventFilterAll)

	err := c.Connect(ctx)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer c.Close()

	// Send a command to trigger the server to send an event
	c.Send(ctx, "/test")

	select {
	case event := <-sub.Channel():
		if event.Type != "newChatItems" {
			t.Errorf("expected event type newChatItems, got %v", event.Type)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for event")
	}

	sub.Unsubscribe()
}

func TestGenerateCorrID(t *testing.T) {
	// Generate multiple IDs and verify they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateCorrID()
		if len(id) != 16 {
			t.Errorf("expected corrId length 16, got %d", len(id))
		}
		if ids[id] {
			t.Error("generated duplicate corrId")
		}
		ids[id] = true
	}
}
