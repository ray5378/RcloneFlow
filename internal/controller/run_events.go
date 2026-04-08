package controller

import (
	"net/http"
	"strconv"
	"strings"

	clirunner "rcloneflow/internal/runner/cli"
)

// HandleRunEvents 正式事件查询接口：GET /api/runs/events/{id}
func (c *RunController) HandleRunEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { w.WriteHeader(http.StatusMethodNotAllowed); return }
	idStr := strings.TrimPrefix(r.URL.Path, "/api/runs/events/")
	runID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil { WriteJSON(w, 400, map[string]any{"error": "invalid id"}); return }
	evs := clirunner.ListEvents(runID)
	WriteJSON(w, 200, map[string]any{"events": evs})
}
