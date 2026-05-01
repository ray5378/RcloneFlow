package controller

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rcloneflow/internal/rclone"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

// TaskController 任务控制器
type TaskController struct {
	taskSvc     *service.TaskService
	scheduleSvc *service.ScheduleService
	runSvc      *service.RunService
	rc          *rclone.Client
}

func (c *TaskController) Service() *service.TaskService { return c.taskSvc }

// NewTaskController 创建任务控制器
func NewTaskController(taskSvc *service.TaskService, scheduleSvc *service.ScheduleService, runSvc *service.RunService, rc *rclone.Client) *TaskController {
	return &TaskController{
		taskSvc:     taskSvc,
		scheduleSvc: scheduleSvc,
		runSvc:      runSvc,
		rc:          rc,
	}
}

// HandleTasks 处理任务列表和创建
func (c *TaskController) HandleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasks, err := c.taskSvc.ListTasks()
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, tasks)

	case http.MethodPost:
		var req store.Task
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		t, err := c.taskSvc.CreateTask(req)
		if err != nil {
			if errors.Is(err, service.ErrTaskNameExists) {
				WriteJSON(w, 409, map[string]any{"error": err.Error()})
				return
			}
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, t)

	case http.MethodPut:
		var req struct {
			ID   int64      `json:"id"`
			Task store.Task `json:"task"`
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		if err := c.taskSvc.UpdateTask(req.ID, req.Task); err != nil {
			if errors.Is(err, service.ErrTaskNameExists) {
				WriteJSON(w, 409, map[string]any{"error": err.Error()})
				return
			}
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, nil)

	case http.MethodPatch:
		// PATCH /api/tasks { id, options }
		var req struct {
			ID      int64          `json:"id"`
			Options map[string]any `json:"options"`
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": "invalid body"})
			return
		}
		if req.ID == 0 {
			WriteJSON(w, 400, map[string]any{"error": "missing id"})
			return
		}
		if err := c.taskSvc.UpdateTaskOptions(req.ID, req.Options); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"ok": true})

	default:
		w.WriteHeader(405)
	}
}

// HandleTaskActions 处理任务操作（删除、运行）
func (c *TaskController) resolveLogPath(run service.RunRecord) (string, bool) {
	if s, ok := any(run.Summary).(string); ok && s != "" {
		var m map[string]any
		if json.Unmarshal([]byte(s), &m) == nil {
			if p, ok := m["stderrFile"].(string); ok && p != "" {
				return p, true
			}
		}
	}
	if m, ok := any(run.Summary).(map[string]any); ok {
		if p, ok := m["stderrFile"].(string); ok && p != "" {
			return p, true
		}
	}
	return "", false
}

func (c *TaskController) buildActiveRunItems() ([]map[string]any, error) {
	runs, err := c.runSvc.ListActiveRuns()
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		var summary map[string]any
		var progress map[string]any
		switch v := any(run.Summary).(type) {
		case map[string]any:
			summary = v
			if p, ok := v["progress"].(map[string]any); ok {
				progress = p
			}
		case string:
			if v != "" {
				var m map[string]any
				if json.Unmarshal([]byte(v), &m) == nil {
					summary = m
					if p, ok := m["progress"].(map[string]any); ok {
						progress = p
					}
				}
			}
		}
		bytes := int64(0)
		if v, ok := progress["bytes"].(float64); ok {
			bytes = int64(v)
		}
		progressTotal := int64(0)
		if v, ok := progress["totalBytes"].(float64); ok {
			progressTotal = int64(v)
		}
		total := progressTotal
		if total > 0 && bytes > total {
			bytes = total
		}
		speed := int64(0)
		if v, ok := progress["speed"].(float64); ok {
			speed = int64(v)
		}
		eta := float64(0)
		if v, ok := progress["eta"].(float64); ok {
			eta = v
		}
		pct := 0.0
		if v, ok := progress["percentage"].(float64); ok {
			pct = v
		} else if total > 0 {
			pct = float64(bytes) / float64(total) * 100
		}
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		completedFiles := float64(0)
		if v, ok := progress["completedFiles"].(float64); ok {
			completedFiles = v
		}
		if completedFiles <= 0 {
			if logPath, ok := c.resolveLogPath(run); ok && logPath != "" {
				if n := countCompletedFilesFromLog(logPath); n > 0 {
					completedFiles = float64(n)
				}
			}
		}
		totalCount := float64(0)
		if v, ok := progress["plannedFiles"].(float64); ok {
			totalCount = v
		}
		if totalCount <= 0 && summary != nil {
			if pf, ok := summary["preflight"].(map[string]any); ok {
				if v, ok2 := pf["totalCount"].(float64); ok2 {
					totalCount = v
				}
			}
		}
		if totalCount > 0 && completedFiles > totalCount {
			completedFiles = totalCount
		}
		phase := "transferring"
		if total == 0 && bytes == 0 {
			phase = "preparing"
		}
		if summary != nil {
			if fw, ok := summary["finishWait"].(map[string]any); ok {
				if en, ok2 := fw["enabled"].(bool); ok2 && en {
					if done, ok3 := fw["done"].(bool); !ok3 || !done {
						phase = "finalizing"
					}
				}
			}
		}
		progressLine := ""
		if summary != nil {
			if v, ok := summary["progressLine"].(string); ok {
				progressLine = v
			}
		}
		stable := map[string]any{
			"bytes":          bytes,
			"totalBytes":     total,
			"speed":          speed,
			"eta":            eta,
			"percentage":     pct,
			"phase":          phase,
			"lastUpdatedAt":  time.Now().Format(time.RFC3339),
			"completedFiles": completedFiles,
			"totalCount":     totalCount,
		}
		calcPct := 0.0
		if total > 0 {
			calcPct = float64(bytes) / float64(total) * 100
		}
		pctMismatch := total > 0 && math.Abs(calcPct-pct) > 1.5
		countMismatch := totalCount > 0 && completedFiles > totalCount
		etaMismatch := eta > 0 && speed > 0 && total > bytes && math.Abs((float64(total-bytes)/float64(speed))-eta) > 300
		progressMismatch := pctMismatch || countMismatch || etaMismatch
		progressCheck := map[string]any{
			"ok":            !progressMismatch,
			"pctMismatch":   pctMismatch,
			"countMismatch": countMismatch,
			"etaMismatch":   etaMismatch,
			"calcPct":       calcPct,
		}
		item := map[string]any{
			"runRecord": map[string]any{
				"id":               run.ID,
				"taskId":           run.TaskID,
				"status":           run.Status,
				"bytesTransferred": run.BytesTransferred,
				"error":            run.Error,
				"startedAt":        run.StartedAt,
				"finishedAt":       run.FinishedAt,
			},
			"progress":         stable,
			"progressLine":     progressLine,
			"progressSource":   "summary.progress",
			"progressMismatch": progressMismatch,
			"progressCheck":    progressCheck,
		}
		if start, err := time.Parse(time.RFC3339, run.StartedAt); err == nil {
			dur := int64(time.Since(start).Seconds())
			if dur < 0 {
				dur = 0
			}
			item["runRecord"].(map[string]any)["durationSeconds"] = dur
			item["runRecord"].(map[string]any)["durationText"] = humanDuration(dur)
		}
		items = append(items, item)
	}
	return items, nil
}

func (c *TaskController) HandleBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	tasks, err := c.taskSvc.ListTasks()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	activeRuns, err := c.buildActiveRunItems()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{
		"tasks":      tasks,
		"activeRuns": activeRuns,
	})
}

func (c *TaskController) HandleTaskActions(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/api/tasks/")

	// DELETE /api/tasks/{id}
	if r.Method == http.MethodDelete {
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			WriteJSON(w, 400, map[string]any{"error": "invalid task id"})
			return
		}
		if err := c.taskSvc.DeleteTask(id); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return
	}

	// POST /api/tasks/{id}/run
	if !strings.HasSuffix(p, "/run") {
		w.WriteHeader(404)
		return
	}
	idStr := strings.TrimSuffix(p, "/run")
	id, _ := strconv.ParseInt(strings.Trim(idStr, "/"), 10, 64)

	if err := c.RunTask(r.Context(), id, "manual"); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, map[string]any{"started": true})
}

// RunTask 运行指定任务
func (c *TaskController) RunTask(ctx context.Context, taskID int64, trigger string) error {
	return c.taskSvc.RunTask(ctx, taskID, trigger)
}
