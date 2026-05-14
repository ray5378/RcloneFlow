package websocket

import (
	"testing"
	"time"
)

func TestHubBroadcastRemovesStaleClientsSafely(t *testing.T) {
	h := NewHub()
	go h.Run()

	good := &Client{hub: h, send: make(chan []byte, 1)}
	stale := &Client{hub: h, send: make(chan []byte, 1)}
	stale.send <- []byte("full")

	h.register <- good
	h.register <- stale
	time.Sleep(20 * time.Millisecond)

	h.Broadcast("test", map[string]any{"ok": true})
	time.Sleep(50 * time.Millisecond)

	if got := h.ClientCount(); got != 1 {
		t.Fatalf("clientCount=%d want 1", got)
	}

	select {
	case <-good.send:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("expected good client to receive broadcast")
	}

	h.mu.RLock()
	_, staleStillPresent := h.clients[stale]
	h.mu.RUnlock()
	if staleStillPresent {
		t.Fatalf("expected stale client to be removed from hub")
	}
}
