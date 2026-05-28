package controller

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/runnercli"
	"rcloneflow/internal/service"
)

type RunController struct {
	runSvc *service.RunService
	rc     *adapter.RcloneClient
}

func NewRunController(runSvc *service.RunService, rc *adapter.RcloneClient) *RunController {
	return &RunController{
		runSvc: runSvc,
		rc:     rc,
	}
}

func (c *RunController) HandleRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		if err := c.runSvc.DeleteAllRuns(); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return
	}

	page := 1
	pageSize := 50
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	runs, total, err := c.runSvc.ListRuns(page, pageSize)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(runs))
	for _, r := range runs {
		var sum map[string]any
		switch v := any(r.Summary).(type) {
		case string:
			if v != "" {
				if err := json.Unmarshal([]byte(v), &sum); err != nil {
					logger.Error("unmarshal run summary for listing", zap.Error(err))
				}
			}
		case map[string]any:
			sum = v
		}
		sum = ensureHistoricalFinalSummary(r, sum)
		out = append(out, buildLightRunObject(r, sum))
	}
	WriteJSON(w, 200, map[string]any{
		"runs":     out,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (c *RunController) HandleRunsByTask(w http.ResponseWriter, r *http.Request) {
	taskId, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/runs/task/"), 10, 64)

	if r.Method == http.MethodDelete {
		if err := c.runSvc.DeleteRunsByTask(taskId); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return
	}

	runs, err := c.runSvc.ListRunsByTask(taskId)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		var sum map[string]any
		if run.Summary != "" {
			if err := json.Unmarshal([]byte(run.Summary), &sum); err != nil {
				logger.Error("unmarshal run summary for task listing", zap.Error(err))
			}
		}
		sum = ensureHistoricalFinalSummary(run, sum)
		out = append(out, buildLightRunObject(run, sum))
	}
	WriteJSON(w, 200, out)
}

func (c *RunController) HandleRunStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/runs/"), 10, 64)

	if r.Method == http.MethodDelete {
		if err := c.runSvc.DeleteRun(id); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return
	}

	run, err := c.runSvc.GetRun(id)
	if err != nil {
		WriteJSON(w, 404, map[string]any{"error": "run not found"})
		return
	}

	var sum map[string]any
	switch v := any(run.Summary).(type) {
	case string:
		if v != "" {
			if err := json.Unmarshal([]byte(v), &sum); err != nil {
				logger.Error("unmarshal run summary for status", zap.Error(err))
			}
		}
	case map[string]any:
		sum = v
	}
	sum = ensureHistoricalFinalSummary(run, sum)
	WriteJSON(w, 200, buildLightRunObject(run, sum))
}

func (c *RunController) HandleRunStopCLI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	idStr = strings.TrimSuffix(idStr, "/stop")
	id, _ := strconv.ParseInt(strings.Trim(idStr, "/"), 10, 64)
	if id <= 0 {
		WriteJSON(w, 400, map[string]any{"error": "invalid id"})
		return
	}
	c.runSvc.UpdateRunStatus(id, map[string]any{"finished": true, "success": false, "error": "stopped by user"})
	WriteJSON(w, 200, map[string]any{"stopped": true})
}

func (c *RunController) HandleRunKillCLI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/runs/"), "/kill")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	run, err := c.runSvc.GetRun(id)
	if err != nil {
		WriteJSON(w, 200, map[string]any{"killed": false})
		return
	}
	if killRunBySummary(run) {
		WriteJSON(w, 200, map[string]any{"killed": true})
		return
	}
	WriteJSON(w, 200, map[string]any{"killed": false})
}

func (c *RunController) HandleTaskKill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/kill")
	tid, _ := strconv.ParseInt(strings.Trim(idStr, "/"), 10, 64)
	if tid == 0 {
		WriteJSON(w, 400, map[string]any{"error": "invalid task id"})
		return
	}
	runs, err := c.runSvc.ListRunsByTask(tid)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	var candidate *service.RunRecord
	for i := range runs {
		r := runs[i]
		if r.Status == "running" || r.Status == "finalizing" {
			candidate = &r
			break
		}
	}
	if candidate == nil {
		if len(runs) > 0 {
			candidate = &runs[0]
		}
	}
	if candidate == nil {
		WriteJSON(w, 404, map[string]any{"error": "no runs for task"})
		return
	}
	runnercli.CancelRun(candidate.ID)
	if killRunBySummary(*candidate) {
		WriteJSON(w, 200, map[string]any{"killed": true, "runId": candidate.ID})
		return
	}
	WriteJSON(w, 200, map[string]any{"killed": false})
}

// HandleRunLog 统一提供 stderr 单文件下载

func (c *RunController) HandleRunLog(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/runs/"), "/log")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	run, err := c.runSvc.GetRun(id)
	if err != nil {
		WriteJSON(w, 404, map[string]any{"error": "run not found"})
		return
	}
	if s, ok := any(run.Summary).(string); ok && s != "" {
		var m map[string]any
		if json.Unmarshal([]byte(s), &m) == nil {
			if p, ok := m["stderrFile"].(string); ok && p != "" && isSafeLogPath(p) {
				http.ServeFile(w, r, p)
				return
			}
		}
	}
	if m, ok := any(run.Summary).(map[string]any); ok {
		if p, ok := m["stderrFile"].(string); ok && p != "" && isSafeLogPath(p) {
			http.ServeFile(w, r, p)
			return
		}
	}
	r.URL.RawQuery = ""
	base := "/app/data/logs"
	sanitize := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			return s
		}
		inv := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}_-]+`)
		s = inv.ReplaceAllString(s, "_")
		r := []rune(s)
		if len(r) > 60 {
			s = string(r[:60])
		}
		return s
	}
	if run.TaskName != "" {
		san := sanitize(run.TaskName)
		entries, _ := os.ReadDir(base)
		var best string
		var bestMod int64
		for _, ent := range entries {
			if !ent.IsDir() {
				continue
			}
			name := ent.Name()
			if !strings.HasPrefix(name, san+"-") {
				continue
			}
			sub := filepath.Join(base, name)
			files, _ := os.ReadDir(sub)
			for _, f := range files {
				if f.IsDir() || !strings.HasSuffix(f.Name(), ".log") {
					continue
				}
				fi, _ := f.Info()
				if fi != nil {
					mod := fi.ModTime().Unix()
					if mod > bestMod {
						bestMod = mod
						best = filepath.Join(sub, f.Name())
					}
				}
			}
		}
		if best != "" {
			http.ServeFile(w, r, best)
			return
		}
	}
	WriteJSON(w, 404, map[string]any{"error": "log not found"})
}
