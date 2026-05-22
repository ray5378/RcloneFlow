package websocket

import (
	"sync"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"
)

var (
	globalHub   *Hub
	initHubOnce sync.Once
	hubMu       sync.RWMutex
)

func InitHub(h *Hub) {
	hubMu.Lock()
	defer hubMu.Unlock()
	globalHub = h
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		h.Run()
	}()
}

func GetHub() *Hub {
	initHubOnce.Do(func() {
		hubMu.RLock()
		existing := globalHub
		hubMu.RUnlock()
		if existing == nil {
			hub := NewHub()
			InitHub(hub)
		}
	})
	hubMu.RLock()
	defer hubMu.RUnlock()
	return globalHub
}

func ResetHubForTest(h *Hub) {
	hubMu.Lock()
	defer hubMu.Unlock()
	if globalHub != nil {
		globalHub.Stop()
	}
	globalHub = h
	initHubOnce = sync.Once{}
	if h != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("goroutine panic", zap.Any("panic", r))
				}
			}()
			h.Run()
		}()
	}
}

func Broadcast(msgType string, data interface{}) {
	hub := GetHub()
	if hub != nil {
		hub.Broadcast(msgType, data)
	}
}
