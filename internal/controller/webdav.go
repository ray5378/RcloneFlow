package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"rcloneflow/internal/logger"
	"rcloneflow/internal/service"
	"rcloneflow/internal/webdavserver"

	"go.uber.org/zap"
)

type WebdavController struct {
	manager *webdavserver.Manager
	authSvc *service.AuthService
}

func NewWebdavController(manager *webdavserver.Manager, authSvc *service.AuthService) *WebdavController {
	return &WebdavController{manager: manager, authSvc: authSvc}
}

type webdavStartRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *WebdavController) HandleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "仅支持POST"})
		return
	}

	body, _ := io.ReadAll(r.Body)
	var req webdavStartRequest
	if err := json.Unmarshal(body, &req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求格式"})
		return
	}

	if req.Username == "" || req.Password == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "用户名和密码不能为空"})
		return
	}

	if _, _, err := c.authSvc.Login(req.Username, req.Password); err != nil {
		WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "用户名或密码错误"})
		return
	}

	if err := c.manager.Start(req.Username, req.Password); err != nil {
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