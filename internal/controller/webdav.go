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
	maxSize, cleanupInterval := c.manager.GetCacheSettings()
	mode := c.manager.GetAccessMode()

	WriteJSON(w, http.StatusOK, map[string]any{
		"running":                running,
		"enabled":                enabled,
		"port":                   17870,
		"path":                   "/dav/",
		"mode":                   mode,
		"cache_max_size":         maxSize,
		"cache_cleanup_interval": cleanupInterval,
	})
}

type cacheSettingsRequest struct {
	CacheMaxSize         string `json:"cache_max_size"`
	CacheCleanupInterval string `json:"cache_cleanup_interval"`
}

func (c *WebdavController) HandleCacheSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		maxSize, cleanupInterval := c.manager.GetCacheSettings()
		WriteJSON(w, http.StatusOK, map[string]any{
			"cache_max_size":         maxSize,
			"cache_cleanup_interval": cleanupInterval,
		})
		return
	}

	if r.Method == http.MethodPost {
		body, _ := io.ReadAll(r.Body)
		var req cacheSettingsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求格式"})
			return
		}

		needsRestart := false

		if req.CacheMaxSize != "" {
			if err := c.manager.SetCacheMaxSize(req.CacheMaxSize); err != nil {
				logger.Error("保存缓存大小失败", zap.Error(err))
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			needsRestart = true
		}

		if req.CacheCleanupInterval != "" {
			if err := c.manager.SetCacheCleanupInterval(req.CacheCleanupInterval); err != nil {
				logger.Error("保存缓存清理间隔失败", zap.Error(err))
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}

		if needsRestart && c.manager.IsRunning() && c.manager.IsEnabled() {
			if err := c.manager.Restart(); err != nil {
				logger.Error("应用缓存设置后重启WebDAV失败", zap.Error(err))
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}

		WriteJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"message": "缓存设置已保存",
		})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (c *WebdavController) HandleCacheCleanup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "仅支持POST"})
		return
	}

	if err := c.manager.CacheCleanupNow(); err != nil {
		logger.Error("清理WebDAV缓存失败", zap.Error(err))
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "WebDAV缓存已清理",
	})
}

type modeRequest struct {
	Mode string `json:"mode"`
}

func (c *WebdavController) HandleMode(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		mode := c.manager.GetAccessMode()
		WriteJSON(w, http.StatusOK, map[string]any{"mode": mode})
		return
	}

	if r.Method == http.MethodPost {
		body, _ := io.ReadAll(r.Body)
		var req modeRequest
		if err := json.Unmarshal(body, &req); err != nil {
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求格式"})
			return
		}

		if req.Mode != "cache" && req.Mode != "stream" {
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "mode 必须为 cache 或 stream"})
			return
		}

		if err := c.manager.SetAccessMode(req.Mode); err != nil {
			logger.Error("保存 WebDAV 模式失败", zap.Error(err))
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		WriteJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"mode":    req.Mode,
			"message": "已切换访问模式",
		})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
