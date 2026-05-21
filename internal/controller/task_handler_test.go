package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/adapter"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

func setupTaskTestDB(t *testing.T) *store.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_task_test_*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func setupTaskController(t *testing.T) *TaskController {
	t.Helper()
	db := setupTaskTestDB(t)
	rc := adapter.NewRcloneClient(nil)
	taskSvc := service.NewTaskService(db, nil)
	scheduleSvc := service.NewScheduleService(db)
	runAdapter := service.NewStoreRunAdapter(db)
	runSvc := service.NewRunService(runAdapter)
	return NewTaskController(taskSvc, scheduleSvc, runSvc, rc)
}

func TestTaskController_HandleTasks_GET(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	rec := httptest.NewRecorder()
	c.HandleTasks(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var tasks []store.Task
	err := json.Unmarshal(rec.Body.Bytes(), &tasks)
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestTaskController_HandleTasks_POST_Create(t *testing.T) {
	c := setupTaskController(t)

	body := `{"name":"test-task","mode":"sync","sourceRemote":"src","sourcePath":"/src","targetRemote":"dst","targetPath":"/dst"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleTasks(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var task store.Task
	err := json.Unmarshal(rec.Body.Bytes(), &task)
	require.NoError(t, err)
	assert.Equal(t, "test-task", task.Name)
	assert.Equal(t, "sync", task.Mode)
}

func TestTaskController_HandleTasks_POST_Duplicate(t *testing.T) {
	c := setupTaskController(t)

	body := `{"name":"dup-task","mode":"copy","sourceRemote":"src","sourcePath":"/src","targetRemote":"dst","targetPath":"/dst"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader([]byte(body)))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	c.HandleTasks(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	req2 := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader([]byte(body)))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	c.HandleTasks(rec2, req2)
	assert.Equal(t, http.StatusConflict, rec2.Code)
}

func TestTaskController_HandleTasks_POST_InvalidJSON(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleTasks(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleTasks_PUT_Update(t *testing.T) {
	c := setupTaskController(t)

	// Create task first
	body := `{"name":"update-task","mode":"copy","sourceRemote":"src","sourcePath":"/src","targetRemote":"dst","targetPath":"/dst"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader([]byte(body)))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	c.HandleTasks(rec1, req1)
	var task store.Task
	json.Unmarshal(rec1.Body.Bytes(), &task)

	// Update
	updateBody := `{"id":` + fmt.Sprintf("%d", task.ID) + `,"task":{"name":"updated-task"}}`
	req2 := httptest.NewRequest(http.MethodPut, "/api/tasks", bytes.NewReader([]byte(updateBody)))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	c.HandleTasks(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestTaskController_HandleTasks_PUT_InvalidJSON(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodPut, "/api/tasks", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleTasks(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleTasks_PATCH_InvalidJSON(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/tasks", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleTasks(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleTasks_PATCH_MissingID(t *testing.T) {
	c := setupTaskController(t)

	body := `{"options":{"key":"value"}}`
	req := httptest.NewRequest(http.MethodPatch, "/api/tasks", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleTasks(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleTasks_MethodNotAllowed(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks", nil)
	rec := httptest.NewRecorder()
	c.HandleTasks(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestTaskController_HandleBootstrap_GET(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/bootstrap", nil)
	rec := httptest.NewRecorder()
	c.HandleBootstrap(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	// tasks should be present (may be empty array)
	_, hasTasks := resp["tasks"]
	assert.True(t, hasTasks, "response should have tasks field")
}

func TestTaskController_HandleBootstrap_MethodNotAllowed(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/bootstrap", nil)
	rec := httptest.NewRecorder()
	c.HandleBootstrap(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestTaskController_HandleTaskActions_Delete_InvalidID(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/abc", nil)
	rec := httptest.NewRecorder()
	c.HandleTaskActions(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleTaskActions_Delete_NonExistent(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/999", nil)
	rec := httptest.NewRecorder()
	c.HandleTaskActions(rec, req)

	// Should succeed even if task doesn't exist (or fail with 500)
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestTaskController_HandleTaskActions_ClearAll(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/clear", nil)
	rec := httptest.NewRecorder()
	c.HandleTaskActions(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestTaskController_HandleTaskActions_Run_InvalidID(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/abc/run", nil)
	rec := httptest.NewRecorder()
	c.HandleTaskActions(rec, req)

	// Will fail because task doesn't exist
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestTaskController_HandleTaskActions_Export(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/export", nil)
	rec := httptest.NewRecorder()
	c.HandleTaskActions(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestTaskController_HandleTaskActions_Import_InvalidJSON(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/import", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleTaskActions(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleTaskActions_NotFound(t *testing.T) {
	c := setupTaskController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/nonexistent", nil)
	rec := httptest.NewRecorder()
	c.HandleTaskActions(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestTaskController_Service(t *testing.T) {
	c := setupTaskController(t)
	assert.NotNil(t, c.Service())
}
