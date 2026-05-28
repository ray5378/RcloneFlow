package controller

import (
	"context"
	"net/http"
	"strings"

	"rcloneflow/internal/adapter"
)

// RemoteController 远程存储控制器
type RemoteController struct {
	rc *adapter.RcloneClient
}

// NewRemoteController 创建远程存储控制器
func NewRemoteController(rc *adapter.RcloneClient) *RemoteController {
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
		dump, _ := c.rc.DumpConfig(r.Context())
		descriptions := make(map[string]string, len(remotes))
		for _, name := range remotes {
			if cfg, ok := dump[name]; ok {
				if d, ok := cfg["description"].(string); ok && d != "" {
					descriptions[name] = d
				}
			}
		}
		versionResp, _ := c.rc.Version(r.Context())
		version := ""
		if versionResp != nil {
			version = versionResp.Version
		}
		WriteJSON(w, 200, map[string]any{"remotes": remotes, "descriptions": descriptions, "version": version})

	case http.MethodPost:
		var req struct {
			Name       string         `json:"name"`
			Type       string         `json:"type"`
			Parameters map[string]any `json:"parameters"`
		}
		if err := DecodeRequest(w, r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		if err := c.rc.CreateRemote(r.Context(), &adapter.CreateRemoteRequest{
			Name:       req.Name,
			Type:       req.Type,
			Parameters: req.Parameters,
		}); err != nil {
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
		if err := DecodeRequest(w, r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		if err := c.rc.CreateRemote(r.Context(), &adapter.CreateRemoteRequest{
			Name:       req.Name,
			Type:       req.Type,
			Parameters: req.Parameters,
		}); err != nil {
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

// HandleRemoteDescription 更新远程存储备注
func (c *RemoteController) HandleRemoteDescription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := DecodeRequest(w, r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	if req.Name == "" {
		WriteJSON(w, 400, map[string]any{"error": "name required"})
		return
	}
	if err := c.rc.UpdateRemoteDescription(r.Context(), req.Name, req.Description); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"updated": true})
}

// HandleRemoteTest 测试远程存储
func (c *RemoteController) HandleRemoteTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := DecodeRequest(w, r, &req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	items, err := c.rc.ListPath(r.Context(), req.Name+":", "")
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	result := make([]map[string]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"Name":     item.Name,
			"Path":     item.Path,
			"IsDir":    item.IsDir,
			"MimeType": item.MimeType,
			"ModTime":  item.ModTime,
			"Size":     item.Size,
		}
	}
	WriteJSON(w, 200, map[string]any{"ok": true, "count": len(result)})
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
	result := make([]map[string]any, len(providers))
	for i, p := range providers {
		result[i] = map[string]any{
			"Name":      p.Name,
			"Hangul":    p.Hangul,
			"Prefix":    p.Prefix,
			"OpenURL":   p.OpenURL,
			"HashTypes": p.HashTypes,
			"Options":   p.Options,
		}
	}
	WriteJSON(w, 200, map[string]any{"providers": result})
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
	WriteJSON(w, 200, map[string]any{"config": config})
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
	about, err := c.rc.GetUsage(r.Context(), fs)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	usage := map[string]any{
		"used":  about.Used,
		"free":  about.Free,
		"total": about.Used + about.Free,
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
	result := map[string]any{
		"name":      info.Name,
		"precision": info.Precision,
		"root":      info.Root,
	}
	if info.Features != nil {
		result["features"] = info.Features
	}
	WriteJSON(w, 200, result)
}

// RcloneClient 获取rclone客户端（供其他控制器使用）
func (c *RemoteController) RcloneClient() *adapter.RcloneClient {
	return c.rc
}

// RunTask 运行任务
func (c *RemoteController) RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string, opts *adapter.TaskOptions) (int64, error) {
	src := srcRemote + ":" + strings.TrimPrefix(srcPath, "/")
	dst := dstRemote + ":" + strings.TrimPrefix(dstPath, "/")
	return c.rc.StartJob(ctx, mode, src, dst, opts)
}
