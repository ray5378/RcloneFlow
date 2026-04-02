package controller

import (
	"net/http"
	"strconv"
	"strings"

	"rcloneflow/internal/rclone"
	"rcloneflow/internal/service"
)

// RunController 运行记录控制器
type RunController struct {
	runSvc *service.RunService
	rc     *rclone.Client
}

// NewRunController 创建运行记录控制器
func NewRunController(runSvc *service.RunService, rc *rclone.Client) *RunController {
	return &RunController{
		runSvc: runSvc,
		rc:     rc,
	}
}

// HandleRuns 处理运行记录列表
func (c *RunController) HandleRuns(w http.ResponseWriter, r *http.Request) {
	// DELETE /api/runs - 删除所有历史记录
	if r.Method == http.MethodDelete {
		if err := c.runSvc.DeleteAllRuns(); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return
	}

	runs, err := c.runSvc.ListRuns()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, runs)
}

// HandleRunsByTask 处理按任务ID删除历史记录
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

	// GET 请求 - 获取该任务的历史记录
	runs, err := c.runSvc.ListRunsByTask(taskId)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, runs)
}

// HandleRunStatus 处理运行状态查询
func (c *RunController) HandleRunStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/runs/"), 10, 64)
	
	// DELETE 请求
	if r.Method == http.MethodDelete {
		if err := c.runSvc.DeleteRun(id); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})
		return
	}
	
	// GET 请求
	runs, err := c.runSvc.ListRuns()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	for _, run := range runs {
		if run.ID != id {
			continue
		}
		if run.RcJobID > 0 {
			st, err := c.rc.JobStatus(r.Context(), run.RcJobID)
			if err == nil {
				c.runSvc.UpdateRunStatus(run.ID, st)
				WriteJSON(w, 200, st)
				return
			}
		}
		WriteJSON(w, 200, run)
		return
	}
	WriteJSON(w, 404, map[string]any{"error": "run not found"})
}

// HandleActiveRuns 处理获取所有运行中的任务及其实时状态
func (c *RunController) HandleActiveRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := c.runSvc.ListActiveRuns()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	// 同步每个运行中任务的实时状态
	type ActiveRun struct {
		RunRecord service.RunRecord `json:"runRecord"`
		RealTimeStatus map[string]any `json:"realtimeStatus,omitempty"`
	}

	activeRuns := make([]ActiveRun, 0, len(runs))
	for _, run := range runs {
		active := ActiveRun{RunRecord: run}

		// 如果有rcJobID，获取实时状态
		if run.RcJobID > 0 {
			st, err := c.rc.JobStatus(r.Context(), run.RcJobID)
			if err == nil {
				active.RealTimeStatus = st
				// 更新数据库状态
				c.runSvc.UpdateRunStatus(run.ID, st)
			}
		}
		activeRuns = append(activeRuns, active)
	}

	WriteJSON(w, 200, activeRuns)
}
