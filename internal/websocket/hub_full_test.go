package websocket

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	h := NewHub()
	if h == nil {
		t.Fatal("expected non-nil hub")
	}
	if h.clients == nil {
		t.Error("expected non-nil clients map")
	}
	if h.broadcast == nil {
		t.Error("expected non-nil broadcast channel")
	}
	if h.register == nil {
		t.Error("expected non-nil register channel")
	}
	if h.unregister == nil {
		t.Error("expected non-nil unregister channel")
	}
}

func TestNewClient(t *testing.T) {
	h := NewHub()
	c := NewClient(h, nil)
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.hub != h {
		t.Error("expected client hub to match")
	}
	if c.send == nil {
		t.Error("expected non-nil send channel")
	}
}

func TestClientCount(t *testing.T) {
	h := NewHub()
	go h.Run()

	c1 := &Client{hub: h, send: make(chan []byte, 1)}
	c2 := &Client{hub: h, send: make(chan []byte, 1)}

	h.register <- c1
	h.register <- c2
	time.Sleep(20 * time.Millisecond)

	if h.ClientCount() != 2 {
		t.Errorf("expected 2 clients, got %d", h.ClientCount())
	}

	h.unregister <- c1
	time.Sleep(20 * time.Millisecond)

	if h.ClientCount() != 1 {
		t.Errorf("expected 1 client after unregister, got %d", h.ClientCount())
	}

	h.unregister <- c2
	time.Sleep(20 * time.Millisecond)
}

func TestRemoveClientNil(t *testing.T) {
	h := NewHub()
	h.removeClient(nil)
}

func TestRemoveClientNotInHub(t *testing.T) {
	h := NewHub()
	c := &Client{hub: h, send: make(chan []byte, 1)}
	h.removeClient(c)
}

func TestBroadcastMarshalError(t *testing.T) {
	h := NewHub()
	go h.Run()

	c := &Client{hub: h, send: make(chan []byte, 1)}
	h.register <- c
	time.Sleep(20 * time.Millisecond)

	h.Broadcast("test", make(chan int))
	time.Sleep(20 * time.Millisecond)

	if h.ClientCount() != 1 {
		t.Errorf("expected 1 client after marshal error, got %d", h.ClientCount())
	}

	h.unregister <- c
	time.Sleep(20 * time.Millisecond)
}

func TestBroadcastToMultipleClients(t *testing.T) {
	h := NewHub()
	go h.Run()

	c1 := &Client{hub: h, send: make(chan []byte, 4)}
	c2 := &Client{hub: h, send: make(chan []byte, 4)}

	h.register <- c1
	h.register <- c2
	time.Sleep(20 * time.Millisecond)

	h.Broadcast("ping", map[string]any{"msg": "hello"})
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-c1.send:
		if len(msg) == 0 {
			t.Error("expected non-empty message for c1")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("c1 did not receive broadcast")
	}

	select {
	case msg := <-c2.send:
		if len(msg) == 0 {
			t.Error("expected non-empty message for c2")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("c2 did not receive broadcast")
	}

	h.unregister <- c1
	h.unregister <- c2
	time.Sleep(20 * time.Millisecond)
}

func TestRemoveClientClosesSendChannel(t *testing.T) {
	h := NewHub()
	go h.Run()

	c := &Client{hub: h, send: make(chan []byte, 1)}
	h.register <- c
	time.Sleep(20 * time.Millisecond)

	h.unregister <- c
	time.Sleep(20 * time.Millisecond)

	select {
	case _, ok := <-c.send:
		if ok {
			t.Error("expected send channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("send channel was not closed")
	}
}

func TestGetHub(t *testing.T) {
	h := GetHub()
	if h == nil {
		t.Fatal("GetHub() returned nil")
	}
	// Second call should return same instance
	h2 := GetHub()
	if h != h2 {
		t.Error("GetHub() should return same singleton instance")
	}
}

func TestBroadcast(t *testing.T) {
	h := NewHub()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	c := &Client{hub: h, send: make(chan []byte, 1)}
	h.register <- c
	time.Sleep(20 * time.Millisecond)

	h.Broadcast("test", map[string]any{"msg": "hello"})
	time.Sleep(20 * time.Millisecond)

	select {
	case data := <-c.send:
		if len(data) == 0 {
			t.Error("expected non-empty broadcast data")
		}
	default:
		t.Error("expected broadcast message in send channel")
	}
}

func TestNewHandler(t *testing.T) {
	h := NewHub()
	handler := NewHandler(h)
	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}
	if handler.hub != h {
		t.Error("NewHandler() hub mismatch")
	}
}

func TestBroadcast_Global(t *testing.T) {
	// GetHub returns a singleton, so we test Broadcast through it
	h := GetHub()
	if h == nil {
		t.Fatal("GetHub() returned nil")
	}
	// Broadcast should not panic
	Broadcast("test", map[string]any{"msg": "hello"})
}

func TestHubStop(t *testing.T) {
	h := NewHub()
	done := make(chan struct{})
	go func() {
		h.Run()
		close(done)
	}()

	c := &Client{hub: h, send: make(chan []byte, 1)}
	h.register <- c
	time.Sleep(20 * time.Millisecond)

	if h.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", h.ClientCount())
	}

	h.Stop()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run() did not return after Stop()")
	}

	// Second Stop should be a no-op
	h.Stop()

	// Client send channel should be closed after hub stops
	select {
	case _, ok := <-c.send:
		if ok {
			t.Error("expected send channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("send channel was not closed")
	}
}

func TestInitHub(t *testing.T) {
	h := NewHub()
	InitHub(h)
	time.Sleep(20 * time.Millisecond)
	if GetHub() != h {
		t.Error("InitHub should set global hub")
	}
	h.Stop()
}

func TestResetHubForTest(t *testing.T) {
	h1 := NewHub()
	ResetHubForTest(h1)
	if GetHub() != h1 {
		t.Fatal("ResetHubForTest should set the hub")
	}
	time.Sleep(20 * time.Millisecond)

	h2 := NewHub()
	ResetHubForTest(h2)
	if GetHub() != h2 {
		t.Fatal("ResetHubForTest should replace the hub")
	}
	time.Sleep(20 * time.Millisecond)

	ResetHubForTest(nil)
	if GetHub() == nil {
		t.Fatal("GetHub should still return non-nil after ResetHubForTest(nil)")
	}

	GetHub().Stop()
}
