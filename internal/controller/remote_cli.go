package controller

import (
	"encoding/json"
	"net/http"

	"rcloneflow/internal/rclone"
)

// RemoteCreateCLIHandler 使用 rclone CLI 写入受控 config（最小占位）。
func RemoteCreateCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	var req struct {
		ConfigPath string         `json:"configPath"`
		Name       string         `json:"name"`
		Type       string         `json:"type"`
		Params     map[string]any `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	cli := rclone.NewCLIConfig()
	if err := cli.Create(req.ConfigPath, req.Name, req.Type, req.Params); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// RemoteTestCLIHandler 简单列举根路径验证连通性。
func RemoteTestCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	var req struct { Name string `json:"name"`; ConfigPath string `json:"configPath"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	cli := rclone.NewCLIConfig()
	if err := cli.TestRemote(req.Name, req.ConfigPath); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// RemoteDumpCLIHandler 输出受控 config 内容（调试用）。
func RemoteDumpCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	configPath := r.URL.Query().Get("configPath")
	cli := rclone.NewCLIConfig()
	out, err := cli.Dump(configPath)
	if err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	WriteJSON(w, 200, map[string]any{"dump": out})
}
