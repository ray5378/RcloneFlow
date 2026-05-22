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
// 安全密钥：每个任务必须设置 webhookSecret，外部请求需通过 ?secret=xxx 或 X-Webhook-Secret 头提供
// 全局 WEBHOOK_SECRET 作为可选的服务级网关，独立于任务级密钥
func (c *WebhookController) HandleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// 全局服务级密钥（可选，独立于任务密钥）
	globalSecret := readSettingsWebhookSecret()
	providedSecret := r.URL.Query().Get("secret")
	if providedSecret == "" {
		providedSecret = r.Header.Get("X-Webhook-Secret")
	}
	if globalSecret != "" {
		if providedSecret == "" || subtle.ConstantTimeCompare([]byte(providedSecret), []byte(globalSecret)) != 1 {
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

	// shouldTrigger 先验证任务密钥，再验证匹配字段
	shouldTrigger := func(opts map[string]any) (bool, string) {
		taskSecret := strings.TrimSpace(toString(opts["webhookSecret"]))
		if taskSecret == "" {
			return false, "secret_not_set"
		}
		if globalSecret == "" {
			if providedSecret == "" || subtle.ConstantTimeCompare([]byte(providedSecret), []byte(taskSecret)) != 1 {
				return false, "secret_mismatch"
			}
		}
		matchText := strings.TrimSpace(toString(opts["webhookMatchText"]))
		if matchText == "" {
			return true, ""
		}
		if !strings.Contains(bodyText, matchText) {
			return false, "webhook_match_not_hit"
		}
		return true, ""
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
			ok, reason := shouldTrigger(opts)
			if !ok {
				if reason == "secret_not_set" || reason == "secret_mismatch" {
					WriteJSON(w, 401, map[string]any{"error": "unauthorized"})
				} else {
					WriteJSON(w, 200, map[string]any{"ok": true, "triggered": false, "reason": reason})
				}
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
		ok, reason := shouldTrigger(opts)
		if !ok {
			if reason == "secret_not_set" || reason == "secret_mismatch" {
				WriteJSON(w, 401, map[string]any{"error": "unauthorized"})
			} else {
				WriteJSON(w, 200, map[string]any{"ok": true, "triggered": false, "reason": reason})
			}
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
