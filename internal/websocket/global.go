package websocket

import (
	"sync"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"
)

var (
	globalHub *Hub
	once      sync.Once
)

// GetHub returns the global Hub instance (singleton)
func GetHub() *Hub {
	once.Do(func() {
		globalHub = NewHub()
		go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		globalHub.Run()
	}()
	})
	return globalHub
}

// Broadcast sends a message to all connected clients
func Broadcast(msgType string, data interface{}) {
	GetHub().Broadcast(msgType, data)
}
