package controller

import (
	"net/http"
	"strings"

	"rcloneflow/internal/auth"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/webdavserver"

	"go.uber.org/zap"
)

type WebdavController struct {
	manager *webdavserver.Manager
}

func NewWebdavController(manager *webdavserver.Manager) *WebdavController {
	return &WebdavController{manager: manager}
}

func extractUsernameFromRequest(r *http.Request) string {
	tokenStr := r.Header.Get("Authorization")
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	if tokenStr == "" {
		return "admin"
	}
	claims, err := auth.ValidateToken(tokenStr)
	if err != nil {
		return "admin"
	}
	if claims.Username != "" {
		return claims.Username
	}
	return "admin"
}

func (c *WebdavController) HandleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "仅支持POST"})
		return
	}

	username := extractUsernameFromRequest(r)

	if err := c.manager.Start(username); err != nil {
		logger.Error("启动WebDAV失败", zap.Error(err))
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"message":  "WebDAV服务已启动",
		"dav_path": "/dav/",
	})
}

func (c *WebdavController) HandleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "仅支持POST"})
		return
	}

	if err := c.manager.Stop(); err != nil {
		logger.Error("停止WebDAV失败", zap.Error(err))
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "WebDAV服务已停止",
	})
}

func (c *WebdavController) HandleStatus(w http.ResponseWriter, r *http.Request) {
	running := c.manager.IsRunning()
	enabled := c.manager.IsEnabled()

	WriteJSON(w, http.StatusOK, map[string]any{
		"running": running,
		"enabled": enabled,
		"port":    17870,
		"path":    "/dav/",
	})
}