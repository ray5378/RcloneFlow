package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

// WebhookController 处理无需鉴权的外部触发
type WebhookController struct {
	taskSvc *service.TaskService
}

func NewWebhookController(taskSvc *service.TaskService) *WebhookController { return &WebhookController{taskSvc: taskSvc} }

// HandleTrigger 外部 webhook 触发任务
// 路径：/webhook/{id}?token=xxx 支持 GET/POST
// 行为：遍历任务，匹配 task.Options.webhookId == {id}，若该任务配置了 webhookSecret 则校验 token；通过则触发任务运行（trigger=webhook）
func (c *WebhookController) HandleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/webhook/"), "/")
	if id == "" { WriteJSON(w, 400, map[string]any{"error":"missing webhook id"}); return }
	token := r.URL.Query().Get("token")

	tasks, err := c.taskSvc.ListTasks()
	if err != nil { WriteJSON(w, 500, map[string]any{"error": err.Error()}); return }
	for _, t := range tasks {
		if len(t.Options) == 0 { continue }
		var opts map[string]any
		if json.Unmarshal(t.Options, &opts) != nil { continue }
		wid, _ := opts["webhookId"].(string)
		if wid == "" || wid != id { continue }
		secret, _ := opts["webhookSecret"].(string)
		if secret != "" && token != secret { WriteJSON(w, 403, map[string]any{"error":"invalid token"}); return }
		// 触发任务
		if err := c.taskSvc.RunTask(r.Context(), t.ID, "webhook"); err != nil {
			WriteJSON(w, 500, map[string]any{"error": err.Error()});
			return
		}
		WriteJSON(w, 200, map[string]any{"started": true, "taskId": t.ID})
		return
	}
	WriteJSON(w, 404, map[string]any{"error":"webhook id not found"})
}
