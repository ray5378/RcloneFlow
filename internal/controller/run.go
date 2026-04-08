package controller

import (
	"net/http"
	"strconv"
	"strings"

	"rcloneflow/internal/rclone"
	"rcloneflow/internal/service"
	clirunner "rcloneflow/internal/runner/cli"
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

	// CLI 改造后：不再依赖 RC 的全局统计；此处置空或后续以 CLI 聚合替代
	var globalStats map[string]any

	// 同步每个运行中任务的实时状态
	type ActiveRun struct {
		RunRecord       service.RunRecord `json:"runRecord"`
		RealTimeStatus  map[string]any    `json:"realtimeStatus,omitempty"`
		GroupStats      map[string]any    `json:"groupStats,omitempty"`
		GlobalStats     map[string]any    `json:"globalStats,omitempty"`
		DerivedProgress map[string]any    `json:"derivedProgress,omitempty"`
	}

	activeRuns := make([]ActiveRun, 0, len(runs))
	for _, run := range runs {
		active := ActiveRun{RunRecord: run, GlobalStats: globalStats}

		// CLI 改造：从内存态获取 DerivedProgress，使用 RcJobID 作为运行标识
		if p := (service.NewCLIRunAdapter()).GetDerivedProgress(run.RcJobID); p != nil {
			active.DerivedProgress = p
		}

		// 派生一个更稳定的前端进度对象：优先 job/group stats，其次 job/status，最后兜底全局 stats
		derived := map[string]any{}
		if active.GroupStats != nil {
			for _, key := range []string{"bytes", "totalBytes", "total_bytes", "speed", "speedAvg", "speed_avg", "eta", "percentage", "checks", "transfers", "totalTransfers", "group"} {
				if v, ok := active.GroupStats[key]; ok {
					derived[key] = v
				}
			}
		}
		if active.RealTimeStatus != nil {
			for _, key := range []string{"bytes", "totalBytes", "total_bytes", "speed", "speedAvg", "speed_avg", "eta", "percentage", "group"} {
				if _, exists := derived[key]; !exists {
					if v, ok := active.RealTimeStatus[key]; ok {
						derived[key] = v
					}
				}
			}
			if progress, ok := active.RealTimeStatus["progress"]; ok {
				derived["progress"] = progress
			}
		}
		if globalStats != nil {
			for _, key := range []string{"bytes", "totalBytes", "total_bytes", "speed", "speedAvg", "speed_avg", "eta", "percentage"} {
				if _, exists := derived[key]; !exists {
					if v, ok := globalStats[key]; ok {
						derived[key] = v
					}
				}
			}
		}
		if len(derived) > 0 {
			active.DerivedProgress = derived
		}

		activeRuns = append(activeRuns, active)
	}

	WriteJSON(w, 200, activeRuns)
}

// HandleGlobalStats 处理获取全局实时统计信息
func (c *RunController) HandleGlobalStats(w http.ResponseWriter, r *http.Request) {
	// CLI 改造后暂无全局统计，返回空结构或后续聚合实现
	WriteJSON(w, 200, map[string]any{"ok": true})
}

// HandleJobStatus 处理获取指定 Job 的状态
func (c *RunController) HandleJobStatus(w http.ResponseWriter, r *http.Request) {
	// CLI 改造：job 概念由 runID 代替，这里保持兼容，返回内存态 DerivedProgress
	jobIdStr := r.PathValue("jobId")
	runID, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil { WriteJSON(w, 400, map[string]any{"error": "invalid id"}); return }
	p, ok := service.NewCLIRunAdapter().GetDerivedProgress(runID), true
	if !ok || p == nil { WriteJSON(w, 200, map[string]any{"progress": nil}); return }
	WriteJSON(w, 200, p)
}

// HandleJobStop 处理停止指定的 Job
func (c *RunController) HandleJobStop(w http.ResponseWriter, r *http.Request) {
	jobIdStr := r.PathValue("jobId")
	runID, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil { WriteJSON(w, 400, map[string]any{"error": "invalid id"}); return }
	if err := clirunner.StopRunByID(runID); err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	c.runSvc.UpdateRunStatusByJobId(runID, "stopped", "用户手动停止")
	WriteJSON(w, 200, map[string]any{"success": true})
}
