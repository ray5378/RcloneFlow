package controller

import (
	"net/http"
	"strconv"
	"strings"

	"rcloneflow/internal/service"
)

// ScheduleController 定时任务控制器
type ScheduleController struct {
	scheduleSvc *service.ScheduleService
}

// NewScheduleController 创建定时任务控制器
func NewScheduleController(scheduleSvc *service.ScheduleService) *ScheduleController {
	return &ScheduleController{scheduleSvc: scheduleSvc}
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
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		item, err := c.scheduleSvc.CreateSchedule(req.TaskID, req.Spec)
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
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
			Enabled bool `json:"enabled"`
		}
		if err := DecodeRequest(r, &req); err != nil {
			WriteJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		if err := c.scheduleSvc.SetScheduleEnabled(id, req.Enabled); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"enabled": req.Enabled})

	default:
		w.WriteHeader(405)
	}
}
