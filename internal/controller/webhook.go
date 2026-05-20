package controller

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/service"
)

// WebhookController 处理外部触发
type WebhookController struct {
	taskSvc *service.TaskService
	secret  string
}

func NewWebhookController(taskSvc *service.TaskService) *WebhookController {
	return &WebhookController{taskSvc: taskSvc, secret: os.Getenv("WEBHOOK_SECRET")}
}

func readSettingsWebhookSecret() string {
	if v := os.Getenv("WEBHOOK_SECRET"); v != "" {
		return v
	}
	dataDir := os.Getenv("APP_DATA_DIR")
	if dataDir == "" {
		dataDir = "."
	}
	fp := filepath.Join(dataDir, "settings.json")
	b, err := os.ReadFile(fp)
	if err != nil || len(b) == 0 {
		return ""
	}
	var m map[string]string
	if json.Unmarshal(b, &m) != nil {
		return ""
	}
	return strings.TrimSpace(m["WEBHOOK_SECRET"])
}

// HandleTrigger 外部 webhook 触发任务
// 支持两种方式：
// 1) 直接按任务ID触发：/webhook/{taskId}
// 2) 按自定义ID匹配任务 options.webhookId：/webhook/{customId}
// 如果设置了 WEBHOOK_SECRET 环境变量，则必须通过 query 参数 ?secret=xxx 或 Header X-Webhook-Secret 提供
func (c *WebhookController) HandleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	secret := readSettingsWebhookSecret()
	if secret != "" {
		provided := r.URL.Query().Get("secret")
		if provided == "" {
			provided = r.Header.Get("X-Webhook-Secret")
		}
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			WriteJSON(w, 401, map[string]any{"error": "unauthorized"})
			return
		}
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/webhook/"), "/")
	if id == "" {
		WriteJSON(w, 400, map[string]any{"error": "missing id"})
		return
	}

	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	bodyText := string(bodyBytes)

	shouldTrigger := func(opts map[string]any) bool {
		matchText := strings.TrimSpace(toString(opts["webhookMatchText"]))
		if matchText == "" {
			return true
		}
		return strings.Contains(bodyText, matchText)
	}

	// 优先：数字则按任务ID直接触发
	if tid, err := strconv.ParseInt(id, 10, 64); err == nil {
		if t, ok := c.taskSvc.GetTask(tid); ok {
			var opts map[string]any
			if len(t.Options) > 0 {
				if err := json.Unmarshal(t.Options, &opts); err != nil {
					logger.Error("unmarshal task options", zap.Error(err))
				}
			}
			if !shouldTrigger(opts) {
				WriteJSON(w, 200, map[string]any{"ok": true, "triggered": false, "reason": "webhook_match_not_hit"})
				return
			}
			result, err := c.taskSvc.RunTask(r.Context(), t.ID, "webhook")
			if err != nil {
				WriteJSON(w, 500, map[string]any{"error": "internal error"})
				return
			}
			WriteJSON(w, 200, result)
			return
		}
		WriteJSON(w, 404, map[string]any{"error": "task not found"})
		return
	}

	// 否则：按自定义 webhookId 匹配
	tasks, err := c.taskSvc.ListTasks()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": "internal error"})
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
		if !shouldTrigger(opts) {
			WriteJSON(w, 200, map[string]any{"ok": true, "triggered": false, "reason": "webhook_match_not_hit"})
			return
		}
		result, err := c.taskSvc.RunTask(r.Context(), t.ID, "webhook")
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": "internal error"})
			return
		}
		WriteJSON(w, 200, result)
		return
	}
	WriteJSON(w, 404, map[string]any{"error": "webhook id not found"})
}

func toString(v any) string {
	s, _ := v.(string)
	return s
}
