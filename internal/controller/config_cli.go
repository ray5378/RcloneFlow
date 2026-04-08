package controller

import (
	"net/http"

	"rcloneflow/internal/rclone"
)

// ConfigDumpCLIHandler 替代 /api/config/dump（RC） → CLI dump
func ConfigDumpCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	configPath := r.URL.Query().Get("configPath")
	out, err := rclone.NewCLIConfig().Dump(configPath)
	if err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	WriteJSON(w, 200, map[string]any{"dump": out})
}

// ConfigDeleteCLIHandler 替代 /api/config/{name} DELETE → CLI delete
func ConfigDeleteCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	name := r.URL.Query().Get("name")
	configPath := r.URL.Query().Get("configPath")
	if name == "" { WriteJSON(w, 400, map[string]any{"error": "name required"}); return }
	cli := rclone.NewCLIConfig()
	// 延迟实现 Delete（如未实现，返回 501）
	WriteJSON(w, 501, map[string]any{"error": "not implemented", "name": name, "configPath": configPath})
}
