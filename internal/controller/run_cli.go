package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	clirunner "rcloneflow/internal/runner/cli"
)

// 最小 CLI 运行器接入（临时）：
// - POST /api/cli/runs/start/{id}
// - POST /api/cli/runs/stop/{id}
// - GET  /api/cli/runs/progress/{id}

func StartRunCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	idStr := strings.TrimPrefix(r.URL.Path, "/api/cli/runs/start/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var req struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	if req.Src == "" || req.Dst == "" { WriteJSON(w, 400, map[string]any{"error": "src/dst 不能为空"}); return }
	opts := clirunner.StartOptions{
		RunID: id,
		WorkDir: "",
		CLI: clirunner.TaskCLIOptions{
			Src: req.Src,
			Dst: req.Dst,
			Transfers: 2,
			StatsInterval: "5s",
			JSONLog: false,
			LogLevel: "NOTICE",
		},
	}
	if _, err := clirunner.StartRun(opts); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

func StopRunCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	idStr := strings.TrimPrefix(r.URL.Path, "/api/cli/runs/stop/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := clirunner.StopRunByID(id); err != nil { WriteJSON(w, 400, map[string]any{"error": err.Error()}); return }
	WriteJSON(w, 200, map[string]any{"ok": true})
}

func ProgressRunCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	idStr := strings.TrimPrefix(r.URL.Path, "/api/cli/runs/progress/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	p, ok := clirunner.GetProgress(id)
	if !ok { WriteJSON(w, 200, map[string]any{"ok": true, "progress": nil}); return }
	WriteJSON(w, 200, map[string]any{"ok": true, "progress": p})
}
