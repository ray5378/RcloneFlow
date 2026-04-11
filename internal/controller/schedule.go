package controller

import (
	"net/http"
	"strconv"
	"strings"

	"rcloneflow/internal/scheduler"
	"rcloneflow/internal/service"
)

// ScheduleController 定时任务控制器
type ScheduleController struct {
	scheduleSvc *service.ScheduleService
	sched      *scheduler.Scheduler
}

// NewScheduleController 创建定时任务控制器
func NewScheduleController(scheduleSvc *service.ScheduleService, sched *scheduler.Scheduler) *ScheduleController {
	return &ScheduleController{scheduleSvc: scheduleSvc, sched: sched}
}

// HandleSchedules 处理定时任务列表、创建和删除
func (c *ScheduleController) HandleSchedules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		schedules, err := c.scheduleSvc.ListSchedules()
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, schedules)

	case http.MethodPost:
		var req struct {
			TaskID int64  `json:"taskId"`
			Spec   string `json:"spec"`
			Enabled bool  `json:"enabled"`
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		item, err := c.scheduleSvc.CreateSchedule(req.TaskID, req.Spec, req.Enabled)
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		// 添加到调度器
		if req.Enabled {
			c.sched.AddSchedule(item)
		}
		WriteJSON(w, 200, item)

	case http.MethodDelete:
		p := strings.TrimPrefix(r.URL.Path, "/api/schedules/")
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			WriteJSON(w, 400, map[string]any{"error": "invalid schedule id"})
			return
		}
		if err := c.scheduleSvc.DeleteSchedule(id); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"deleted": true})

	case http.MethodPut:
		p := strings.TrimPrefix(r.URL.Path, "/api/schedules/")
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			WriteJSON(w, 400, map[string]any{"error": "invalid schedule id"})
			return
		}
		var req struct {
			Enabled bool   `json:"enabled"`
			Spec    string `json:"spec"`
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		// 可选：更新 spec
		if strings.TrimSpace(req.Spec) != "" {
			if err := c.scheduleSvc.UpdateSpec(id, req.Spec); err != nil {
				WriteJSON(w, 500, map[string]any{"error": err.Error()})
				return
			}
		}
		// 更新启用状态
		if err := c.scheduleSvc.SetScheduleEnabled(id, req.Enabled); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		// 运行时热重载（强制单一生效）：先移除旧 entry，再按最新状态重建
		if s, ok := c.sched.DB().GetSchedule(id); ok {
			// 无条件先移除旧 entry，避免旧规则残留
			c.sched.RemoveSchedule(id)
			if s.Enabled {
				_ = c.sched.AddSchedule(s)
			}
		}
		WriteJSON(w, 200, map[string]any{"enabled": req.Enabled, "spec": req.Spec})

	default:
		w.WriteHeader(405)
	}
}
