package controller

import (
	"encoding/json"
	"net/http"

	"rcloneflow/internal/rclone"
)

// BrowserListCLIHandler 使用 rclone lsjson 列目录。
func BrowserListCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	var req struct {
		Fs   string `json:"fs"`
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	if req.Fs == "" { WriteJSON(w, 400, map[string]any{"error": "fs 不能为空"}); return }
	items, err := rclone.LsJSON(req.Fs, req.Path)
	if err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	// 返回结构与现有前端期望的最小子集保持一致
	resp := make([]map[string]any, 0, len(items))
	for _, it := range items {
		resp = append(resp, map[string]any{
			"name": it.Name,
			"path": it.Path,
			"size": it.Size,
			"isDir": it.IsDir,
		})
	}
	WriteJSON(w, 200, map[string]any{"items": resp})
}
