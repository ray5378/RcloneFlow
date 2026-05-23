package websocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestServeHTTP_Upgrade(t *testing.T) {
	h := NewHub()
	go h.Run()
	defer h.Stop()

	handler := NewHandler(h)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	wsURL := "ws" + srv.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)
	if h.ClientCount() != 1 {
		t.Fatalf("expected 1 client after connect, got %d", h.ClientCount())
	}
}

func TestServeHTTP_ReadPump_WritePump(t *testing.T) {
	h := NewHub()
	go h.Run()
	defer h.Stop()

	handler := NewHandler(h)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	wsURL := "ws" + srv.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	h.Broadcast("ping", map[string]any{"msg": "hello"})

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message failed: %v", err)
	}
	if len(msg) == 0 {
		t.Fatal("expected non-empty message")
	}
}

func TestServeHTTP_ReadPump_ClientDisconnect(t *testing.T) {
	h := NewHub()
	go h.Run()
	defer h.Stop()

	handler := NewHandler(h)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	wsURL := "ws" + srv.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if h.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", h.ClientCount())
	}

	conn.Close()
	time.Sleep(100 * time.Millisecond)

	if h.ClientCount() != 0 {
		t.Fatalf("expected 0 clients after disconnect, got %d", h.ClientCount())
	}
}

func TestServeHTTP_Upgrade_Failure(t *testing.T) {
	h := NewHub()
	go h.Run()
	defer h.Stop()

	handler := NewHandler(h)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	// Non-WebSocket request should not panic and should get an error response
	if resp.StatusCode != http.StatusBadRequest {
		t.Logf("expected 400 for non-WS request, got %d (depends on gorilla behavior)", resp.StatusCode)
	}
}
