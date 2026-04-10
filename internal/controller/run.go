package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

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

// HandleRunKillCLI 强制终止指定 run（优先内部 runner；否则按 PID 逐级信号）
func (c *RunController) HandleRunKillCLI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/runs/"), "/kill")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 读出 run，尝试从 summary 取 pid
	runs, err := c.runSvc.ListRuns()
	if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	for _, run := range runs {
		if run.ID != id { continue }
		if killRunBySummary(run) { WriteJSON(w, 200, map[string]any{"killed": true}); return }
		break
	}
	WriteJSON(w, 404, map[string]any{"error": "run not found or no pid"})
}

// HandleTaskKill 强制终止某任务的当前 rclone 进程（按最近 run 定位，兼容空窗期）
func (c *RunController) HandleTaskKill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(405); return }
	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	idStr = strings.TrimSuffix(idStr, "/kill")
	tid, _ := strconv.ParseInt(strings.Trim(idStr, "/"), 10, 64)
	if tid == 0 { WriteJSON(w, 400, map[string]any{"error":"invalid task id"}); return }
	// 找到该任务最近的 run（running/finalizing 优先，找不到就按开始时间最近）
	runs, err := c.runSvc.ListRunsByTask(tid)
	if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	var candidate *service.RunRecord
	for i := range runs {
		r := runs[i]
		if r.Status == "running" || r.Status == "finalizing" { candidate = &r; break }
	}
	if candidate == nil {
		// 回退：取最近一条
		if len(runs) > 0 { candidate = &runs[0] }
	}
	if candidate == nil { WriteJSON(w, 404, map[string]any{"error":"no runs for task"}); return }
	if killRunBySummary(*candidate) { WriteJSON(w, 200, map[string]any{"killed": true, "runId": candidate.ID}); return }
	WriteJSON(w, 404, map[string]any{"error":"pid not found"})
}

func killRunBySummary(run service.RunRecord) bool {
	var pid int
	var sum map[string]any
	switch v := any(run.Summary).(type) {
	case map[string]any:
		sum = v
	case string:
		if v != "" { _ = json.Unmarshal([]byte(v), &sum) }
	}
	if sum != nil {
		if p, ok := sum["pid"].(float64); ok { pid = int(p) }
		if p2, ok := sum["pid"].(int); ok { pid = p2 }
	}
	if pid > 0 {
		_ = syscall.Kill(pid, syscall.SIGINT); time.Sleep(2*time.Second)
		_ = syscall.Kill(pid, syscall.SIGTERM); time.Sleep(2*time.Second)
		_ = syscall.Kill(pid, syscall.SIGKILL)
		return true
	}
	return false
}

// HandleRunLog 统一提供 stderr 单文件下载
func (c *RunController) HandleRunLog(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/runs/"), "/log")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	runs, err := c.runSvc.ListRuns()
	if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	for _, run := range runs {
		if run.ID == id {
			// Summary 里优先取 stderrFile
			if s, ok := any(run.Summary).(string); ok && s != "" {
				var m map[string]any
				if json.Unmarshal([]byte(s), &m) == nil {
					if p, ok := m["stderrFile"].(string); ok && p != "" { http.ServeFile(w, r, p); return }
				}
			}
			if m, ok := any(run.Summary).(map[string]any); ok {
				if p, ok := m["stderrFile"].(string); ok && p != "" { http.ServeFile(w, r, p); return }
			}
			// 回退路径（标准位置）
			base := "/app/data/logs"
			http.ServeFile(w, r, base+"/run-"+idStr+"-stderr.log")
			return
		}
	}
	WriteJSON(w, 404, map[string]any{"error": "log not found"})
}

// HandleActiveRuns 处理获取所有运行中的任务及其实时状态
func (c *RunController) HandleActiveRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := c.runSvc.ListActiveRuns()
	if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	// 扁平化关键字段：bytes/totalBytes/speed/eta，从 run.Summary.progress 提取
	items := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		// 兼容前端旧形状：每项包含 runRecord + realtimeStatus + derivedProgress
		var progress map[string]any
		// 兼容两种形态：store 层可能已将 summary 反序列化成 map，也可能是原始 JSON 字符串
		switch v := any(run.Summary).(type) {
		case map[string]any:
			if p, ok := v["progress"].(map[string]any); ok { progress = p }
		case string:
			if v != "" {
				var m map[string]any
				if json.Unmarshal([]byte(v), &m) == nil {
					if p, ok := m["progress"].(map[string]any); ok { progress = p }
				}
			}
		}
		bytes := int64(0)
		if v, ok := progress["bytes"].(float64); ok { bytes = int64(v) }
		total := int64(0)
		if v, ok := progress["totalBytes"].(float64); ok { total = int64(v) }
		speed := int64(0)
		if v, ok := progress["speed"].(float64); ok { speed = int64(v) }
		var eta any
		if v, ok := progress["eta"]; ok { eta = v }
		percentage := 0.0
		if total > 0 { percentage = float64(bytes) / float64(total) * 100 }

		item := map[string]any{
			"runRecord": map[string]any{
				"id": run.ID,
				"taskId": run.TaskID,
				"status": run.Status,
				"rcJobId": 0,
				"bytesTransferred": run.BytesTransferred,
				"error": run.Error,
			},
			// 兼容旧前端：在 realtimeStatus 中也提供扁平字段
			"realtimeStatus": map[string]any{
				"progress": progress,
				"error": run.Error,
				"bytes": bytes,
				"totalBytes": total,
				"speed": speed,
				"eta": eta,
				"percentage": percentage,
			},
			"derivedProgress": map[string]any{
				"bytes": bytes,
				"totalBytes": total,
				"speed": speed,
				"eta": eta,
				"percentage": percentage,
			},
		}
		items = append(items, item)
	}
	// 前端期望返回数组
	WriteJSON(w, 200, items)
}

// HandleGlobalStats 处理获取全局实时统计信息
func (c *RunController) HandleGlobalStats(w http.ResponseWriter, r *http.Request) {
	// 先尝试从 RC 获取
	stats, err := c.rc.CoreStats(r.Context())
	if err == nil {
		// 如果 RC 可用直接返回
		WriteJSON(w, 200, stats)
		return
	}
	// 回退：聚合本地 CLI Runner 的活动任务进度
	runs, e2 := c.runSvc.ListActiveRuns()
	if e2 != nil {
		WriteJSON(w, 500, map[string]any{"error": e2.Error()})
		return
	}
	var bytesSum, totalSum, speedSum float64
	for _, run := range runs {
		// 兼容 summary 为 map 或 string
		var p map[string]any
		switch v := any(run.Summary).(type) {
		case map[string]any:
			if pp, ok := v["progress"].(map[string]any); ok { p = pp }
		case string:
			if v != "" {
				var m map[string]any
				if json.Unmarshal([]byte(v), &m) == nil {
					if pp, ok := m["progress"].(map[string]any); ok { p = pp }
				}
			}
		}
		if p != nil {
			if v, ok := p["bytes"].(float64); ok { bytesSum += v }
			if v, ok := p["totalBytes"].(float64); ok { totalSum += v }
			if v, ok := p["speed"].(float64); ok { speedSum += v }
		}
	}
	percentage := 0.0
	if totalSum > 0 { percentage = (bytesSum / totalSum) * 100 }
	WriteJSON(w, 200, map[string]any{
		"bytes": bytesSum,
		"totalBytes": totalSum,
		"speed": speedSum,
		"speedAvg": speedSum, // 简化：无历史窗口，先返回当前合计
		"eta": nil, // CLI 模式无法可靠聚合 ETA，这里暂置空
		"percentage": percentage,
	})
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
