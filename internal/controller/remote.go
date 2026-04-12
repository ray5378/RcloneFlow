package controller

import (
	"context"
	"net/http"
	"strings"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/rclone"
)

// RemoteController 远程存储控制器
type RemoteController struct {
	rc *rclone.Client
}

// NewRemoteController 创建远程存储控制器
func NewRemoteController(rc *rclone.Client) *RemoteController {
	return &RemoteController{rc: rc}
}

// Healthz 健康检查
func (c *RemoteController) Healthz(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandleRemotes 处理远程存储列表和创建
func (c *RemoteController) HandleRemotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		remotes, err := c.rc.ListRemotes(r.Context())
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		version, _ := c.rc.Version(r.Context())
		WriteJSON(w, 200, map[string]any{"remotes": remotes, "version": version})

	case http.MethodPost:
		var req struct {
			Name       string         `json:"name"`
			Type       string         `json:"type"`
			Parameters map[string]any `json:"parameters"`
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		if err := c.rc.CreateRemote(r.Context(), req.Name, req.Type, req.Parameters); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"created": true})

	case http.MethodPut:
		var req struct {
			Name       string         `json:"name"`
			Type       string         `json:"type"`
			Parameters map[string]any `json:"parameters"`
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		if err := c.rc.CreateRemote(r.Context(), req.Name, req.Type, req.Parameters); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"updated": true})

	default:
		w.WriteHeader(405)
	}
}

// HandleRemoteConfig 获取单个存储配置
func (c *RemoteController) HandleRemoteConfig(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/remotes/config/")
	if name == "" {
		WriteJSON(w, 400, map[string]any{"error": "name required"})
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	cfg, err := c.rc.GetConfig(r.Context(), name)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, cfg)
}

// HandleRemoteTest 测试远程存储
func (c *RemoteController) HandleRemoteTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	items, err := c.rc.ListPath(r.Context(), req.Name+":", "")
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"ok": true, "count": len(items)})
}

// HandleProviders 获取所有存储提供商
func (c *RemoteController) HandleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	providers, err := c.rc.GetProviders(r.Context())
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"providers": providers})
}

// HandleConfigDump 获取所有存储配置
func (c *RemoteController) HandleConfigDump(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	config, err := c.rc.DumpConfig(r.Context())
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, config)
}

// HandleConfigActions 获取/删除单个存储配置
func (c *RemoteController) HandleConfigActions(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/config/")

	switch r.Method {
	case http.MethodGet:
		cfg, err := c.rc.GetConfig(r.Context(), name)
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, cfg)

	case http.MethodDelete:
		if err := c.rc.DeleteRemote(r.Context(), name); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})

	default:
		w.WriteHeader(405)
	}
}

// HandleUsage 获取存储使用量
func (c *RemoteController) HandleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	fs := strings.TrimPrefix(r.URL.Path, "/api/usage/")
	if fs == "" {
		WriteJSON(w, 400, map[string]any{"error": "fs parameter required"})
		return
	}
	usage, err := c.rc.GetUsage(r.Context(), fs)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, usage)
}

// HandleFsInfo 获取文件系统信息
func (c *RemoteController) HandleFsInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	fs := strings.TrimPrefix(r.URL.Path, "/api/fsinfo/")
	if fs == "" {
		WriteJSON(w, 400, map[string]any{"error": "fs parameter required"})
		return
	}
	info, err := c.rc.GetFsInfo(r.Context(), fs)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, info)
}

// RcloneClient 获取rclone客户端（供其他控制器使用）
func (c *RemoteController) RcloneClient() *rclone.Client {
	return c.rc
}

// RunTask 运行任务
func (c *RemoteController) RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string, opts *adapter.TaskOptions) (int64, error) {
	return c.rc.RunTask(ctx, taskID, mode, srcRemote, srcPath, dstRemote, dstPath, trigger, opts)
}
