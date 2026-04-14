package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"rcloneflow/internal/service"
)

// WebhookController 处理无需鉴权的外部触发
type WebhookController struct {
	taskSvc *service.TaskService
}

func NewWebhookController(taskSvc *service.TaskService) *WebhookController {
	return &WebhookController{taskSvc: taskSvc}
}

// HandleTrigger 外部 webhook 触发任务（无需秘钥）
// 支持两种方式：
// 1) 直接按任务ID触发：/webhook/{taskId}
// 2) 按自定义ID匹配任务 options.webhookId：/webhook/{customId}
func (c *WebhookController) HandleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/webhook/"), "/")
	if id == "" {
		WriteJSON(w, 400, map[string]any{"error": "missing id"})
		return
	}

	// 优先：数字则按任务ID直接触发
	if tid, err := strconv.ParseInt(id, 10, 64); err == nil {
		if t, ok := c.taskSvc.GetTask(tid); ok {
			if err := c.taskSvc.RunTask(r.Context(), t.ID, "webhook"); err != nil {
				WriteJSON(w, 500, map[string]any{"error": err.Error()})
				return
			}
			WriteJSON(w, 200, map[string]any{"started": true, "taskId": t.ID})
			return
		}
		WriteJSON(w, 404, map[string]any{"error": "task not found"})
		return
	}

	// 否则：按自定义 webhookId 匹配（无密钥）
	tasks, err := c.taskSvc.ListTasks()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": err.Error()})
		return
	}
	for _, t := range tasks {
		if len(t.Options) == 0 {
			continue
		}
		var opts map[string]any
		if json.Unmarshal(t.Options, &opts) != nil {
			continue
		}
		wid, _ := opts["webhookId"].(string)
		if wid == "" || wid != id {
			continue
		}
		if err := c.taskSvc.RunTask(r.Context(), t.ID, "webhook"); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		WriteJSON(w, 200, map[string]any{"started": true, "taskId": t.ID})
		return
	}
	WriteJSON(w, 404, map[string]any{"error": "webhook id not found"})
}
