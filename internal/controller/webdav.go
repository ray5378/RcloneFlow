package controller

import (
	"encoding/json"
	"io"
	"net/http"

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

type credentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *WebdavController) HandleSaveCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "仅支持POST"})
		return
	}

	body, _ := io.ReadAll(r.Body)
	var req credentialsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求格式"})
		return
	}

	if req.Username == "" || req.Password == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "用户名和密码不能为空"})
		return
	}

	if err := c.manager.SetCredentials(req.Username, req.Password); err != nil {
		logger.Error("保存WebDAV凭证失败", zap.Error(err))
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "WebDAV凭证已保存",
	})
}

func (c *WebdavController) HandleGetCredentials(w http.ResponseWriter, r *http.Request) {
	username, password, err := c.manager.GetCredentials()
	if err != nil {
		WriteJSON(w, http.StatusOK, map[string]any{
			"ok":       false,
			"username": "",
			"password": "",
		})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"username": username,
		"password": password,
	})
}

func (c *WebdavController) HandleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "仅支持POST"})
		return
	}

	if err := c.manager.Start(); err != nil {
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