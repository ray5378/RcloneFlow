package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/store"
)

func setupBisyncHandlerTest(t *testing.T) (*TaskController, int64, string) {
	t.Helper()
	dataDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dataDir)

	c := setupTaskController(t)

	task, err := c.taskSvc.CreateTask(store.Task{
		Name: "bisync-test", Mode: "bisync",
		SourceRemote: "src", SourcePath: "/src",
		TargetRemote: "dst", TargetPath: "/dst",
	})
	require.NoError(t, err)

	bisyncDir := filepath.Join(dataDir, "bisync", task.Name)
	require.NoError(t, os.MkdirAll(bisyncDir, 0o755))

	return c, task.ID, bisyncDir
}

// ---------------------------------------------------------------------------
// HandleBisyncLstFiles (GET)
// ---------------------------------------------------------------------------

func TestTaskController_HandleBisyncLstFiles_MethodNotAllowed(t *testing.T) {
	c, _, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/1/bisync/lst-files", nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncLstFiles(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestTaskController_HandleBisyncLstFiles_InvalidID(t *testing.T) {
	c := setupTaskController(t)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/abc/bisync/lst-files", nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncLstFiles(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncLstFiles_Success(t *testing.T) {
	c, taskID, bisyncDir := setupBisyncHandlerTest(t)

	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, "path1.lst"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, "path2.lst"), []byte("content2"), 0644))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d/bisync/lst-files", taskID), nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncLstFiles(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	versions, ok := resp["versions"]
	assert.True(t, ok)
	assert.NotNil(t, versions)
}

// ---------------------------------------------------------------------------
// HandleBisyncDeleteLst (POST)
// ---------------------------------------------------------------------------

func TestTaskController_HandleBisyncDeleteLst_MethodNotAllowed(t *testing.T) {
	c, _, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1/bisync/delete-lst", nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncDeleteLst(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestTaskController_HandleBisyncDeleteLst_InvalidID(t *testing.T) {
	c := setupTaskController(t)
	body := `{"versionId":"v1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/abc/bisync/delete-lst", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleBisyncDeleteLst(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncDeleteLst_BadBody(t *testing.T) {
	c, taskID, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/bisync/delete-lst", taskID), bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleBisyncDeleteLst(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncDeleteLst_Success(t *testing.T) {
	c, taskID, bisyncDir := setupBisyncHandlerTest(t)

	// Create a file that looks like a backup version to delete
	backupFile := "20260401_120000.path1.lst-old"
	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, backupFile), []byte("old"), 0644))

	body := fmt.Sprintf(`{"versionId":"%s"}`, "20260401_120000")
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/bisync/delete-lst", taskID), bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleBisyncDeleteLst(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["ok"])
}

// ---------------------------------------------------------------------------
// HandleBisyncRollbackLst (POST)
// ---------------------------------------------------------------------------

func TestTaskController_HandleBisyncRollbackLst_MethodNotAllowed(t *testing.T) {
	c, _, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1/bisync/rollback-lst", nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncRollbackLst(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestTaskController_HandleBisyncRollbackLst_InvalidID(t *testing.T) {
	c := setupTaskController(t)
	body := `{"versionId":"v1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/abc/bisync/rollback-lst", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleBisyncRollbackLst(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncRollbackLst_BadBody(t *testing.T) {
	c, taskID, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/bisync/rollback-lst", taskID), bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleBisyncRollbackLst(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncRollbackLst_Success(t *testing.T) {
	c, taskID, bisyncDir := setupBisyncHandlerTest(t)

	// Place current lst files + a backup; rollback copies backup over current
	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, "path1.lst"), []byte("current"), 0644))
	backupFile := "20260401_120000.path1.lst-old"
	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, backupFile), []byte("backup"), 0644))

	body := `{"versionId":"20260401_120000"}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/bisync/rollback-lst", taskID), bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleBisyncRollbackLst(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["ok"])
}

// ---------------------------------------------------------------------------
// HandleBisyncLstContent (GET)
// ---------------------------------------------------------------------------

func TestTaskController_HandleBisyncLstContent_MethodNotAllowed(t *testing.T) {
	c, _, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/1/bisync/lst-content", nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncLstContent(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestTaskController_HandleBisyncLstContent_InvalidID(t *testing.T) {
	c := setupTaskController(t)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/abc/bisync/lst-content?file=path1.lst", nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncLstContent(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncLstContent_MissingFile(t *testing.T) {
	c, taskID, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d/bisync/lst-content", taskID), nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncLstContent(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncLstContent_Success(t *testing.T) {
	c, taskID, bisyncDir := setupBisyncHandlerTest(t)

	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, "path1.lst"), []byte("file content"), 0644))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d/bisync/lst-content?file=path1.lst", taskID), nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncLstContent(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "file content", resp["content"])
}

// ---------------------------------------------------------------------------
// HandleBisyncResync (POST)
// ---------------------------------------------------------------------------

func TestTaskController_HandleBisyncResync_MethodNotAllowed(t *testing.T) {
	c, _, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1/bisync/resync", nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncResync(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestTaskController_HandleBisyncResync_InvalidID(t *testing.T) {
	c := setupTaskController(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/abc/bisync/resync", nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncResync(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncResync_Success(t *testing.T) {
	c, taskID, bisyncDir := setupBisyncHandlerTest(t)

	// Create lst files so BackupCurrentLstFiles doesn't warn
	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, "path1.lst"), []byte("content"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, "path2.lst"), []byte("content"), 0644))

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/bisync/resync", taskID), nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncResync(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["ok"])
}

// ---------------------------------------------------------------------------
// HandleBisyncResolveConflict (POST)
// ---------------------------------------------------------------------------

func TestTaskController_HandleBisyncResolveConflict_MethodNotAllowed(t *testing.T) {
	c, _, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/1/bisync/resolve-conflict", nil)
	rec := httptest.NewRecorder()
	c.HandleBisyncResolveConflict(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestTaskController_HandleBisyncResolveConflict_InvalidID(t *testing.T) {
	c := setupTaskController(t)
	body := `{"versionId":"v1","keepFile":"conflict1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/abc/bisync/resolve-conflict", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleBisyncResolveConflict(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncResolveConflict_BadBody(t *testing.T) {
	c, taskID, _ := setupBisyncHandlerTest(t)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/bisync/resolve-conflict", taskID), bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleBisyncResolveConflict(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestTaskController_HandleBisyncResolveConflict_Success(t *testing.T) {
	c, taskID, bisyncDir := setupBisyncHandlerTest(t)

	// Place two conflict files
	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, "somefile.conflict1"), []byte("keep this"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(bisyncDir, "somefile.conflict2"), []byte("delete this"), 0644))

	body := `{"versionId":"somefile","keepFile":"conflict1"}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/bisync/resolve-conflict", taskID), bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleBisyncResolveConflict(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["ok"])

	// Verify on disk: conflict2 should be gone, conflict1 renamed to somefile
	_, err1 := os.Stat(filepath.Join(bisyncDir, "somefile.conflict1"))
	assert.True(t, os.IsNotExist(err1), "conflict1 should be renamed")
	_, err2 := os.Stat(filepath.Join(bisyncDir, "somefile.conflict2"))
	assert.True(t, os.IsNotExist(err2), "conflict2 should be deleted")
	data, err := os.ReadFile(filepath.Join(bisyncDir, "somefile"))
	require.NoError(t, err)
	assert.Equal(t, "keep this", string(data))
}
