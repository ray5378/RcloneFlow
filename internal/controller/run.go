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
	if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	// 扁平化关键字段：bytes/totalBytes/speed/eta，从 run.Summary.progress 提取
	items := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		item := map[string]any{"id": run.ID, "taskId": run.TaskID, "status": run.Status}
		var progress map[string]any
		if run.Summary != "" {
			// run.Summary 在 service 层可能是字符串化的；这里容错解析
			if m, ok := any(run.Summary).(map[string]any); ok {
				if p, ok := m["progress"].(map[string]any); ok { progress = p }
			}
		}
		if progress != nil {
			item["progress"] = progress
			if v, ok := progress["bytes"]; ok { item["bytes"] = v }
			if v, ok := progress["totalBytes"]; ok { item["totalBytes"] = v }
			if v, ok := progress["speed"]; ok { item["speed"] = v }
			if v, ok := progress["eta"]; ok { item["eta"] = v }
		}
		items = append(items, item)
	}
	WriteJSON(w, 200, map[string]any{"runs": items})
}

// HandleGlobalStats 处理获取全局实时统计信息
func (c *RunController) HandleGlobalStats(w http.ResponseWriter, r *http.Request) {
	stats, err := c.rc.CoreStats(r.Context())
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, stats)
}

// HandleJobStatus 处理获取指定 Job 的状态
func (c *RunController) HandleJobStatus(w http.ResponseWriter, r *http.Request) {
	jobIdStr := r.PathValue("jobId")
	jobId, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid job id"})
		return
	}

	status, err := c.rc.JobStatus(r.Context(), jobId)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	WriteJSON(w, 200, status)
}

// HandleJobStop 处理停止指定的 Job
func (c *RunController) HandleJobStop(w http.ResponseWriter, r *http.Request) {
	jobIdStr := r.PathValue("jobId")
	jobId, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid job id"})
		return
	}

	if err := c.rc.JobStop(r.Context(), jobId); err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}

	// 更新数据库中该任务的状态为 stopped
	c.runSvc.UpdateRunStatusByJobId(jobId, "stopped", "用户手动停止")

	WriteJSON(w, 200, map[string]any{"success": true})
}
