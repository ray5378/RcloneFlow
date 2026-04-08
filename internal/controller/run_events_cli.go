package controller

import (
	"net/http"
	"strconv"
	"strings"

	clirunner "rcloneflow/internal/runner/cli"
)

// RunEventsCLIHandler 返回内存采样的进度事件。
func RunEventsCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	idStr := strings.TrimPrefix(r.URL.Path, "/api/cli/runs/events/")
	runID, _ := strconv.ParseInt(idStr, 10, 64)
	evs := clirunner.ListEvents(runID)
	WriteJSON(w, 200, map[string]any{"events": evs})
}
