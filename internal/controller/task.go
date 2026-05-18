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
		var req struct {
			ID             int64           `json:"id"`
			Options        map[string]any  `json:"options"`
			Orders         map[int64]int64 `json:"orders"`
			PriorityTaskID int64           `json:"priorityTaskId"`
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": "invalid body"})
			return
		}
		if len(req.Orders) > 0 {
			if err := c.taskSvc.UpdateTaskSortOrders(req.Orders, req.PriorityTaskID); err != nil {
				if errors.Is(err, service.ErrTaskNotFound) {
					WriteJSON(w, 404, map[string]any{"error": err.Error()})
					return
				}
				WriteJSON(w, 500, map[string]any{"error": err.Error()})
				return
			}
			WriteJSON(w, 200, map[string]any{"ok": true})
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
func (c *TaskController) HandleTaskActions(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/api/tasks/")

	if p == "export" && r.Method == http.MethodGet {
		data, err := c.taskSvc.ExportTasks()
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, data)
		return
	}

	if p == "import" && r.Method == http.MethodPost {
		var req struct {
			Tasks            []any          `json:"tasks"`
			Schedules        []any          `json:"schedules"`
			ConflictStrategy string         `json:"conflictStrategy"`
			RawData          map[string]any `json:"-"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req.RawData); err != nil {
			WriteJSON(w, 400, map[string]any{"error": "无效的请求体"})
			return
		}
		strategy, _ := req.RawData["conflictStrategy"].(string)
		if strategy != "skip" && strategy != "overwrite" {
			strategy = "skip"
		}
		imported, skipped, overwritten, err := c.taskSvc.ImportTasks(req.RawData, strategy)
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{
			"imported":    imported,
			"skipped":     skipped,
			"overwritten": overwritten,
		})
		return
	}

	if p == "clear" && r.Method == http.MethodDelete {
		if err := c.taskSvc.ClearAllTasks(); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"ok": true})
		return
	}

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

	result, err := c.RunTask(r.Context(), id, "manual")
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, result)
}

// RunTask 运行指定任务
func (c *TaskController) RunTask(ctx context.Context, taskID int64, trigger string) (service.TaskRunResult, error) {
	return c.taskSvc.RunTask(ctx, taskID, trigger)
}
