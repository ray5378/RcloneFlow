package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

func setupWebhookTest(t *testing.T) *service.TaskService {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_webhook_test_*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	ts := service.NewTaskService(db, nil)
	return ts
}

func TestWebhookController_HandleTrigger_MissingID(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	req := httptest.NewRequest(http.MethodGet, "/webhook/", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 400, rec.Code)
}

func TestWebhookController_HandleTrigger_MethodNotAllowed(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	req := httptest.NewRequest(http.MethodDelete, "/webhook/1", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestWebhookController_HandleTrigger_TaskNotFound_Numeric(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	req := httptest.NewRequest(http.MethodGet, "/webhook/999", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 404, rec.Code)
	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "task not found", resp["error"])
}

func TestWebhookController_HandleTrigger_WebhookIdNotFound(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	req := httptest.NewRequest(http.MethodGet, "/webhook/nonexistent-id", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 404, rec.Code)
	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "webhook id not found", resp["error"])
}

func TestWebhookController_HandleTrigger_POST_Method(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	req := httptest.NewRequest(http.MethodPost, "/webhook/1", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 404, rec.Code)
}

func TestToString(t *testing.T) {
	assert.Equal(t, "hello", toString("hello"))
	assert.Equal(t, "", toString(123))
	assert.Equal(t, "", toString(nil))
	assert.Equal(t, "", toString(map[string]any{}))
}

func TestWebhookController_HandleTrigger_WithGlobalSecret(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	t.Setenv("WEBHOOK_SECRET", "secret123")
	defer t.Setenv("WEBHOOK_SECRET", "")

	req := httptest.NewRequest(http.MethodGet, "/webhook/1", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 401, rec.Code)
}

func TestWebhookController_HandleTrigger_WithGlobalSecret_Valid(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	t.Setenv("WEBHOOK_SECRET", "secret123")
	defer t.Setenv("WEBHOOK_SECRET", "")

	req := httptest.NewRequest(http.MethodGet, "/webhook/1?secret=secret123", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 404, rec.Code)
}

func TestWebhookController_HandleTrigger_WithGlobalSecret_Header(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	t.Setenv("WEBHOOK_SECRET", "secret123")
	defer t.Setenv("WEBHOOK_SECRET", "")

	req := httptest.NewRequest(http.MethodGet, "/webhook/1", nil)
	req.Header.Set("X-Webhook-Secret", "secret123")
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 404, rec.Code)
}

func TestWebhookController_HandleTrigger_WithTaskSecret(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	opts := map[string]any{"webhookSecret": "tasksecret"}
	optsBytes, _ := json.Marshal(opts)
	task, err := svc.CreateTask(store.Task{
		Name:         "webhook-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
		Options:      optsBytes,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/webhook/"+string(rune(task.ID+'0')), nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 401, rec.Code)
}

func TestWebhookController_HandleTrigger_WithTaskSecret_Valid(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	opts := map[string]any{"webhookSecret": "tasksecret"}
	optsBytes, _ := json.Marshal(opts)
	task, err := svc.CreateTask(store.Task{
		Name:         "webhook-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
		Options:      optsBytes,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/webhook/"+string(rune(task.ID+'0'))+"?secret=tasksecret", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestWebhookController_HandleTrigger_WithWebhookId(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	opts := map[string]any{"webhookId": "my-custom-id"}
	optsBytes, _ := json.Marshal(opts)
	_, err := svc.CreateTask(store.Task{
		Name:         "webhook-custom",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
		Options:      optsBytes,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/webhook/my-custom-id", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestWebhookController_HandleTrigger_WithWebhookIdAndSecret(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	opts := map[string]any{"webhookId": "custom-with-secret", "webhookSecret": "secret456"}
	optsBytes, _ := json.Marshal(opts)
	_, err := svc.CreateTask(store.Task{
		Name:         "webhook-secret",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
		Options:      optsBytes,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/webhook/custom-with-secret?secret=secret456", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestWebhookController_HandleTrigger_WithWebhookMatchText(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	opts := map[string]any{"webhookId": "match-test", "webhookMatchText": "trigger"}
	optsBytes, _ := json.Marshal(opts)
	_, err := svc.CreateTask(store.Task{
		Name:         "match-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
		Options:      optsBytes,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/webhook/match-test", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 200, rec.Code)
	var resp map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["triggered"].(bool))
}

func TestWebhookController_HandleTrigger_WithWebhookMatchText_Matched(t *testing.T) {
	svc := setupWebhookTest(t)
	ctrl := NewWebhookController(svc)

	opts := map[string]any{"webhookId": "match-test-2", "webhookMatchText": "trigger"}
	optsBytes, _ := json.Marshal(opts)
	_, err := svc.CreateTask(store.Task{
		Name:         "match-task-2",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
		Options:      optsBytes,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/webhook/match-test-2", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleTrigger(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestReadSettingsWebhookSecret_FromEnv(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", "env-secret")
	defer t.Setenv("WEBHOOK_SECRET", "")

	result := readSettingsWebhookSecret()
	assert.Equal(t, "env-secret", result)
}

func TestReadSettingsWebhookSecret_FromSettingsFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)
	t.Setenv("WEBHOOK_SECRET", "")
	defer t.Setenv("APP_DATA_DIR", "")
	defer t.Setenv("WEBHOOK_SECRET", "")

	settings := map[string]string{"WEBHOOK_SECRET": "file-secret"}
	settingsBytes, _ := json.Marshal(settings)
	err := os.WriteFile(tmpDir+"/settings.json", settingsBytes, 0644)
	require.NoError(t, err)

	result := readSettingsWebhookSecret()
	assert.Equal(t, "file-secret", result)
}

func TestReadSettingsWebhookSecret_SettingsFileMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)
	t.Setenv("WEBHOOK_SECRET", "")
	defer t.Setenv("APP_DATA_DIR", "")
	defer t.Setenv("WEBHOOK_SECRET", "")

	result := readSettingsWebhookSecret()
	assert.Equal(t, "", result)
}

func TestReadSettingsWebhookSecret_SettingsFileInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)
	t.Setenv("WEBHOOK_SECRET", "")
	defer t.Setenv("APP_DATA_DIR", "")
	defer t.Setenv("WEBHOOK_SECRET", "")

	err := os.WriteFile(tmpDir+"/settings.json", []byte("invalid json"), 0644)
	require.NoError(t, err)

	result := readSettingsWebhookSecret()
	assert.Equal(t, "", result)
}
