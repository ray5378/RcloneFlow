package websocket

import "sync"

var (
	globalHub *Hub
	once      sync.Once
)

// GetHub returns the global Hub instance (singleton)
func GetHub() *Hub {
	once.Do(func() {
		globalHub = NewHub()
		go globalHub.Run()
	})
	return globalHub
}

// Broadcast sends a message to all connected clients
func Broadcast(msgType string, data interface{}) {
	GetHub().Broadcast(msgType, data)
}
