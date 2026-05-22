package websocket

import (
	"net/http"

	"go.uber.org/zap"
	"github.com/gorilla/websocket"
	"rcloneflow/internal/logger"
)

// CheckOrigin 允许所有来源：本服务前后端部署在同一容器/主机内，
// WebSocket 升级请求不携带 cookie（JWT 在 Authorization header），不存在 CSRF 风险。
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler handles WebSocket upgrade requests
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeHTTP handles WebSocket upgrade requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := NewClient(h.hub, conn)
	h.hub.register <- client
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		client.WritePump()
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic", zap.Any("panic", r))
			}
		}()
		client.ReadPump()
	}()
}
