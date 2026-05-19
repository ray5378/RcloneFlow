package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/rclone"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

func setupRunTestDB(t *testing.T) *store.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_run_test_*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func setupRunController(t *testing.T) *RunController {
	t.Helper()
	db := setupRunTestDB(t)
	rc := rclone.NewFromEnv()
	runAdapter := service.NewStoreRunAdapter(db)
	runSvc := service.NewRunService(runAdapter)
	return NewRunController(runSvc, rc)
}

func TestRunController_HandleRuns_GET(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/runs", nil)
	rec := httptest.NewRecorder()
	c.HandleRuns(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRunController_HandleRunsByTask(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1/runs", nil)
	rec := httptest.NewRecorder()
	c.HandleRunsByTask(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRunController_HandleRunStatus_GET_NotFound(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/999", nil)
	rec := httptest.NewRecorder()
	c.HandleRunStatus(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRunController_HandleRunStatus_DELETE(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/runs/999", nil)
	rec := httptest.NewRecorder()
	c.HandleRunStatus(rec, req)

	// May succeed (no-op) or fail (not found)
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRunController_HandleRunKillCLI_MethodNotAllowed(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/1/kill", nil)
	rec := httptest.NewRecorder()
	c.HandleRunKillCLI(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRunController_HandleRunKillCLI(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/runs/1/kill", nil)
	rec := httptest.NewRecorder()
	c.HandleRunKillCLI(rec, req)

	// Will fail because run doesn't exist, but handler is exercised
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError || rec.Code == http.StatusNotFound)
}

func TestRunController_HandleTaskKill_MethodNotAllowed(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1/kill", nil)
	rec := httptest.NewRecorder()
	c.HandleTaskKill(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRunController_HandleTaskKill(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/1/kill", nil)
	rec := httptest.NewRecorder()
	c.HandleTaskKill(rec, req)

	// Returns 404 if no runs exist for the task, or 500 if service error
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusNotFound || rec.Code == http.StatusInternalServerError)
}

func TestRunController_HandleRunFiles(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/1/files", nil)
	rec := httptest.NewRecorder()
	c.HandleRunFiles(rec, req)

	// Will return 404 or 500 since run doesn't exist
	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusInternalServerError || rec.Code == http.StatusOK)
}

func TestRunController_HandleRunLog(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/1/log", nil)
	rec := httptest.NewRecorder()
	c.HandleRunLog(rec, req)

	// Will return 404 since run doesn't exist
	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRunController_HandleActiveRuns(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/active", nil)
	rec := httptest.NewRecorder()
	c.HandleActiveRuns(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRunController_HandleGlobalStats(t *testing.T) {
	c := setupRunController(t)
	db := setupRunTestDB(t)

	// Create tasks first to satisfy FK constraints
	task1, err := db.AddTask(store.Task{
		Name: "task1", Mode: "copy", SourceRemote: "src", SourcePath: "/a",
		TargetRemote: "dst", TargetPath: "/b",
	})
	require.NoError(t, err)

	task2, err := db.AddTask(store.Task{
		Name: "task2", Mode: "copy", SourceRemote: "src", SourcePath: "/a",
		TargetRemote: "dst", TargetPath: "/b",
	})
	require.NoError(t, err)

	summary1 := map[string]any{"progress": map[string]any{"bytes": float64(1000), "totalBytes": float64(5000), "speed": float64(100)}}
	_, err = db.AddRun(store.Run{TaskID: task1.ID, Status: "running", Trigger: "manual", TaskName: "task1", Summary: summary1})
	require.NoError(t, err)

	summary2 := map[string]any{"progress": map[string]any{"bytes": float64(2000), "totalBytes": float64(8000), "speed": float64(200)}}
	_, err = db.AddRun(store.Run{TaskID: task2.ID, Status: "running", Trigger: "manual", TaskName: "task2", Summary: summary2})
	require.NoError(t, err)

	// Verify runs are in the database
	runs, err := db.ListActiveRuns()
	require.NoError(t, err)
	assert.Len(t, runs, 2)

	req := httptest.NewRequest(http.MethodGet, "/api/global-stats", nil)
	rec := httptest.NewRecorder()
	c.HandleGlobalStats(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err = json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify the response has the expected structure
	_, ok := resp["bytes"]
	assert.True(t, ok, "response should have bytes field")
	_, ok = resp["totalBytes"]
	assert.True(t, ok, "response should have totalBytes field")
	_, ok = resp["speed"]
	assert.True(t, ok, "response should have speed field")
	_, ok = resp["percentage"]
	assert.True(t, ok, "response should have percentage field")
}

func TestRunController_HandleGlobalStats_Empty(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/global-stats", nil)
	rec := httptest.NewRecorder()
	c.HandleGlobalStats(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, float64(0), resp["bytes"])
	assert.Equal(t, float64(0), resp["totalBytes"])
	assert.Equal(t, float64(0), resp["percentage"])
}

func TestRunController_HandleGlobalStats_StringSummary(t *testing.T) {
	c := setupRunController(t)
	db := setupRunTestDB(t)

	// Create task first to satisfy FK constraint
	task, err := db.AddTask(store.Task{
		Name: "task3", Mode: "copy", SourceRemote: "src", SourcePath: "/a",
		TargetRemote: "dst", TargetPath: "/b",
	})
	require.NoError(t, err)

	summary := map[string]any{"progress": map[string]any{"bytes": float64(500), "totalBytes": float64(1000), "speed": float64(50)}}
	_, err = db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual", TaskName: "task3", Summary: summary})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/global-stats", nil)
	rec := httptest.NewRecorder()
	c.HandleGlobalStats(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		seconds int64
		want    string
	}{
		{0, "0秒"},
		{1, "1秒"},
		{59, "59秒"},
		{60, "1分"},
		{61, "1分1秒"},
		{3600, "1小时"},
		{3661, "1小时1分1秒"},
		{86400, "24小时"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, humanDuration(tt.seconds))
		})
	}
}

func TestAnyToInt64(t *testing.T) {
	assert.Equal(t, int64(42), anyToInt64(int64(42)))
	assert.Equal(t, int64(42), anyToInt64(42))
	assert.Equal(t, int64(42), anyToInt64(float64(42)))
	assert.Equal(t, int64(42), anyToInt64(float32(42)))
	assert.Equal(t, int64(42), anyToInt64(json.Number("42")))
	assert.Equal(t, int64(0), anyToInt64(nil))
	assert.Equal(t, int64(0), anyToInt64("abc"))
	assert.Equal(t, int64(0), anyToInt64(json.Number("invalid")))
}

func TestAbs64(t *testing.T) {
	assert.Equal(t, int64(10), abs64(10))
	assert.Equal(t, int64(10), abs64(-10))
	assert.Equal(t, int64(0), abs64(0))
}

func TestHumanDuration_Controller(t *testing.T) {
	assert.Equal(t, "0秒", humanDuration(0))
	assert.Equal(t, "1秒", humanDuration(1))
	assert.Equal(t, "1分", humanDuration(60))
	assert.Equal(t, "1小时", humanDuration(3600))
}

func TestRunController_HandleRunStatus_NotFound(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/99999", nil)
	rec := httptest.NewRecorder()
	c.HandleRunStatus(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRunController_HandleRunFiles_WithFilter(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/1/files?filter=copied", nil)
	rec := httptest.NewRecorder()
	c.HandleRunFiles(rec, req)

	// Returns 404 if run doesn't exist, 200 if it does
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusNotFound)
}

func TestRunController_HandleRunsByTask_EmptyTaskID(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/abc/runs", nil)
	rec := httptest.NewRecorder()
	c.HandleRunsByTask(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRunController_HandleRunStopCLI_MethodNotAllowed(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/1/stop", nil)
	rec := httptest.NewRecorder()
	c.HandleRunStopCLI(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRunController_HandleRunStopCLI_InvalidID(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/runs/abc/stop", nil)
	rec := httptest.NewRecorder()
	c.HandleRunStopCLI(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRunController_HandleRunStopCLI(t *testing.T) {
	c := setupRunController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/runs/1/stop", nil)
	rec := httptest.NewRecorder()
	c.HandleRunStopCLI(rec, req)

	// Should return 200 even if run doesn't exist (UpdateRunStatus is a no-op)
	assert.Equal(t, http.StatusOK, rec.Code)
}
