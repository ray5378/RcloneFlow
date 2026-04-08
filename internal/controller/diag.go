package controller

import (
	"encoding/json"
	"net/http"

	"rcloneflow/internal/rclone"
)

// DiagRcloneHandler 运行 rclone -vv <cmd> 并返回 stdout/stderr（仅管理员使用）。
func (c *RemoteController) DiagRcloneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	var req struct {
		Kind string `json:"kind"` // "webdav" | "smb"
		Args map[string]string `json:"args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	switch req.Kind {
	case "webdav":
		// 仅做最小连通性检查：URL 必填；用户名可选，密码不走命令行
		url := req.Args["url"]
		user := req.Args["user"]
		out, errb, err := rclone.TestWebDAV(url, user)
		WriteJSON(w, 200, map[string]any{"stdout": out, "stderr": errb, "ok": err == nil})
	case "smb":
		host := req.Args["host"]
		share := req.Args["share"]
		user := req.Args["user"]
		out, errb, err := rclone.TestSMBRoot(host, share, user)
		WriteJSON(w, 200, map[string]any{"stdout": out, "stderr": errb, "ok": err == nil})
	default:
		WriteJSON(w, 400, map[string]any{"error": "unknown kind"})
	}
}
