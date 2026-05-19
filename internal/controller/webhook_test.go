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
