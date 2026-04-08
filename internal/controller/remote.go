package controller

import (
	"context"
	"encoding/json"
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

// HandleRemotes 处理远程存储列表和创建（CLI 等价实现）
func (c *RemoteController) HandleRemotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 使用 CLI dump 枚举名称，保留 version 字段（从 rc.Version 获取或留空）
		names, err := rclone.NewCLIConfig().ListNames("")
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		version, _ := c.rc.Version(r.Context())
		WriteJSON(w, 200, map[string]any{"remotes": names, "version": version})

	case http.MethodPost, http.MethodPut:
		var req struct {
			Name       string         `json:"name"`
			Type       string         `json:"type"`
			Parameters map[string]any `json:"parameters"`
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		if req.Name == "" || req.Type == "" {
			WriteJSON(w, 400, map[string]any{"error": "name/type required"})
			return
		}
		if err := rclone.NewCLIConfig().Create("", req.Name, req.Type, req.Parameters); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"ok": true})

	default:
		w.WriteHeader(405)
	}
}

// HandleRemoteConfig 获取单个存储配置（CLI dump 解析）
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
	out, err := rclone.NewCLIConfig().Dump("")
	if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	var all map[string]map[string]any
	if err := json.Unmarshal([]byte(out), &all); err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	cfg, ok := all[name]
	if !ok { WriteJSON(w, 404, map[string]any{"error": "remote not found"}); return }
	WriteJSON(w, 200, cfg)
}

// HandleRemoteTest 测试远程存储（CLI lsd）
func (c *RemoteController) HandleRemoteTest(w http.ResponseWriter, r *http.Request) {
	var req struct{ Name string `json:"name"` }
	if err := DecodeRequest(r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if err := rclone.NewCLIConfig().TestRemote(req.Name, ""); err != nil {
		WriteJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandleProviders 获取所有存储提供商（静态最小子集）
func (c *RemoteController) HandleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	WriteJSON(w, 200, map[string]any{"providers": rclone.StaticProviders()})
}

// HandleConfigDump 获取所有存储配置（CLI dump 解析为对象）
func (c *RemoteController) HandleConfigDump(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	out, err := rclone.NewCLIConfig().Dump("")
	if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	var m map[string]map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	WriteJSON(w, 200, m)
}

// HandleConfigActions 获取/删除单个存储配置（CLI）
func (c *RemoteController) HandleConfigActions(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/config/")

	switch r.Method {
	case http.MethodGet:
		out, err := rclone.NewCLIConfig().Dump("")
		if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
		var all map[string]map[string]any
		if err := json.Unmarshal([]byte(out), &all); err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
		if cfg, ok := all[name]; ok { WriteJSON(w, 200, cfg); return }
		WriteJSON(w, 404, map[string]any{"error": "remote not found"})
		return

	case http.MethodDelete:
		if err := rclone.NewCLIConfig().Delete("", name); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return

	default:
		w.WriteHeader(405)
	}
}

// HandleUsage 获取存储使用量（暂仍走 rc）
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

// HandleFsInfo 获取文件系统信息（暂仍走 rc）
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

// RunTask 运行任务（仍保留 rc 入口供兼容；实际任务走 TaskService 的 CLI runner）
func (c *RemoteController) RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string, opts *adapter.TaskOptions) (int64, error) {
	return c.rc.RunTask(ctx, taskID, mode, srcRemote, srcPath, dstRemote, dstPath, trigger, opts)
}
