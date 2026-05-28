package websocket

import (
	"testing"
	"time"
)

func TestHubConstants(t *testing.T) {
	assert := func(cond bool, msg string) {
		if !cond {
			t.Fatal(msg)
		}
	}
	assert(pingInterval > 0, "pingInterval must be positive")
	assert(readWriteTimeout > pingInterval, "readWriteTimeout must be greater than pingInterval")
	assert(pingInterval < 60*time.Second, "pingInterval must be less than 60s")
}

func TestHubBroadcast_TwoClients(t *testing.T) {
	h := NewHub()
	go h.Run()

	c1 := NewClient(h, nil)
	c2 := NewClient(h, nil)
	h.register <- c1
	h.register <- c2
	time.Sleep(20 * time.Millisecond)

	if n := h.ClientCount(); n != 2 {
		t.Fatalf("ClientCount=%d want 2", n)
	}

	h.Broadcast("ping_test", map[string]any{"msg": "hello"})
	time.Sleep(30 * time.Millisecond)

	h.Stop()
}

func TestHubBroadcast_EmptyHub(t *testing.T) {
	h := NewHub()
	go h.Run()

	h.Broadcast("test", map[string]any{"data": 123})

	h.Stop()
}

func TestHubRegisterUnregister_ClientCount(t *testing.T) {
	h := NewHub()
	go h.Run()

	c1 := NewClient(h, nil)
	c2 := NewClient(h, nil)

	h.register <- c1
	time.Sleep(10 * time.Millisecond)
	if n := h.ClientCount(); n != 1 {
		t.Fatalf("after c1 register: ClientCount=%d want 1", n)
	}

	h.register <- c2
	time.Sleep(10 * time.Millisecond)
	if n := h.ClientCount(); n != 2 {
		t.Fatalf("after c2 register: ClientCount=%d want 2", n)
	}

	h.unregister <- c1
	time.Sleep(10 * time.Millisecond)
	if n := h.ClientCount(); n != 1 {
		t.Fatalf("after c1 unregister: ClientCount=%d want 1", n)
	}

	h.Stop()
}
