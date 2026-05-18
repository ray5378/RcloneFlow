package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"rcloneflow/internal/active_transfer"
	"rcloneflow/internal/service"
)

type ActiveTransferController struct {
	mgr    *active_transfer.Manager
	runSvc *service.RunService
}

func NewActiveTransferController(mgr *active_transfer.Manager, runSvc *service.RunService) *ActiveTransferController {
	return &ActiveTransferController{mgr: mgr, runSvc: runSvc}
}

func (c *ActiveTransferController) ensureStateFromRun(taskID int64, run service.RunRecord) bool {
	if c.mgr == nil {
		return false
	}
	if _, ok := c.mgr.GetByTaskID(taskID); ok {
		return true
	}
	st, ok := active_transfer.SnapshotFromSummary(run.Summary)
	if !ok || st == nil {
		return false
	}
	c.mgr.RestoreState(st)
	return true
}

func (c *ActiveTransferController) HandleOverview(w http.ResponseWriter, r *http.Request) {
	taskID, ok := parseActiveTransferTaskID(r.URL.Path, "/active-transfer")
	if !ok {
		WriteJSON(w, 400, map[string]any{"error": "invalid task id"})
		return
	}
	run, err := c.runSvc.GetActiveRunByTaskID(taskID)
	if err != nil || run.ID <= 0 {
		WriteJSON(w, 404, map[string]any{"error": "active run not found"})
		return
	}
	bytes, total, speed, eta, percentage := extractProgress(run.Summary)
	if _, ok := c.mgr.GetByTaskID(taskID); !ok {
		_ = c.ensureStateFromRun(taskID, run)
	}
	resp, ok := c.mgr.BuildSummary(taskID, bytes, total, speed, eta, percentage)
	if !ok {
		WriteJSON(w, 404, map[string]any{"error": "active transfer not found"})
		return
	}
	WriteJSON(w, 200, resp)
}

func (c *ActiveTransferController) HandleCompleted(w http.ResponseWriter, r *http.Request) {
	taskID, ok := parseActiveTransferTaskID(r.URL.Path, "/active-transfer/completed")
	if !ok {
		WriteJSON(w, 400, map[string]any{"error": "invalid task id"})
		return
	}
	if _, ok := c.mgr.GetByTaskID(taskID); !ok {
		if run, err := c.runSvc.GetActiveRunByTaskID(taskID); err == nil && run.ID > 0 {
			_ = c.ensureStateFromRun(taskID, run)
		}
	}
	WriteJSON(w, 200, c.mgr.ListCompleted(taskID, queryInt(r, "offset", 0), queryInt(r, "limit", 100)))
}

func (c *ActiveTransferController) HandlePending(w http.ResponseWriter, r *http.Request) {
	taskID, ok := parseActiveTransferTaskID(r.URL.Path, "/active-transfer/pending")
	if !ok {
		WriteJSON(w, 400, map[string]any{"error": "invalid task id"})
		return
	}
	if _, ok := c.mgr.GetByTaskID(taskID); !ok {
		if run, err := c.runSvc.GetActiveRunByTaskID(taskID); err == nil && run.ID > 0 {
			_ = c.ensureStateFromRun(taskID, run)
		}
	}
	WriteJSON(w, 200, c.mgr.ListPending(taskID, queryInt(r, "offset", 0), queryInt(r, "limit", 100)))
}

func parseActiveTransferTaskID(path, suffix string) (int64, bool) {
	prefix := "/api/tasks/"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return 0, false
	}
	mid := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	mid = strings.Trim(mid, "/")
	id, err := strconv.ParseInt(mid, 10, 64)
	return id, err == nil && id > 0
}

func queryInt(r *http.Request, key string, def int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func extractProgress(summary string) (bytes, total, speed, eta int64, percentage float64) {
	if strings.TrimSpace(summary) == "" {
		return
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(summary), &raw); err != nil {
		return
	}
	prog, _ := raw["progress"].(map[string]any)
	if prog == nil {
		return
	}
	bytes = anyInt64(prog["bytes"])
	total = anyInt64(prog["totalBytes"])
	speed = anyInt64(prog["speed"])
	eta = anyInt64(prog["eta"])
	percentage = anyFloat64(prog["percentage"])
	return
}

func anyInt64(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int64:
		return x
	case int:
		return int64(x)
	default:
		return 0
	}
}

func anyFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case int:
		return float64(x)
	default:
		return 0
	}
}
