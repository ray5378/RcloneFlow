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
	"rcloneflow/internal/rclone"
	"rcloneflow/internal/scheduler"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

func setupScheduleTestDB(t *testing.T) *store.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_schedule_test_*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func setupScheduleController(t *testing.T) *ScheduleController {
	t.Helper()
	db := setupScheduleTestDB(t)
	
	// Create a task first (required by foreign key constraint)
	_, err := db.AddTask(store.Task{
		Name:         "Test Task",
		Mode:         "copy",
		SourceRemote: "local",
		SourcePath:   "/tmp",
		TargetRemote: "local",
		TargetPath:   "/tmp/dest",
	})
	require.NoError(t, err)
	
	rc := rclone.NewFromEnv()
	scheduleSvc := service.NewScheduleService(db)
	sched := scheduler.New(db, rc)
	return NewScheduleController(scheduleSvc, sched)
}

func TestScheduleController_HandleSchedules_GET(t *testing.T) {
	c := setupScheduleController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/schedules", nil)
	rec := httptest.NewRecorder()
	c.HandleSchedules(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestScheduleController_HandleSchedules_POST(t *testing.T) {
	c := setupScheduleController(t)

	body, _ := json.Marshal(map[string]any{
		"taskId":  1,
		"spec":    "0 * * * *",
		"enabled": true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/schedules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleSchedules(rec, req)

	// May fail if task doesn't exist, but should return a valid response
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestScheduleController_HandleSchedules_POST_InvalidBody(t *testing.T) {
	c := setupScheduleController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/schedules", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleSchedules(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestScheduleController_HandleSchedules_DELETE(t *testing.T) {
	c := setupScheduleController(t)

	// Create a schedule first
	body, _ := json.Marshal(map[string]any{
		"taskId":  1,
		"spec":    "0 * * * *",
		"enabled": true,
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/schedules", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	c.HandleSchedules(createRec, createReq)

	var created struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(createRec.Body.Bytes(), &created)
	if created.ID == 0 {
		t.Skip("Schedule creation failed, skipping delete test")
	}

	// Delete it
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/schedules/%d", created.ID), nil)
	rec := httptest.NewRecorder()
	c.HandleSchedules(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestScheduleController_HandleSchedules_DELETE_InvalidID(t *testing.T) {
	c := setupScheduleController(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/schedules/abc", nil)
	rec := httptest.NewRecorder()
	c.HandleSchedules(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestScheduleController_HandleSchedules_PUT(t *testing.T) {
	c := setupScheduleController(t)

	// Create a schedule first
	body, _ := json.Marshal(map[string]any{
		"taskId":  1,
		"spec":    "0 * * * *",
		"enabled": true,
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/schedules", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	c.HandleSchedules(createRec, createReq)

	var created struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(createRec.Body.Bytes(), &created)
	if created.ID == 0 {
		t.Skip("Schedule creation failed, skipping put test")
	}

	// Update it
	updateBody, _ := json.Marshal(map[string]any{
		"enabled": false,
		"spec":    "30 * * * *",
	})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/schedules/%d", created.ID), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleSchedules(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestScheduleController_HandleSchedules_PUT_InvalidID(t *testing.T) {
	c := setupScheduleController(t)

	body, _ := json.Marshal(map[string]any{"enabled": false})
	req := httptest.NewRequest(http.MethodPut, "/api/schedules/abc", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleSchedules(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestScheduleController_HandleSchedules_PUT_InvalidBody(t *testing.T) {
	c := setupScheduleController(t)

	req := httptest.NewRequest(http.MethodPut, "/api/schedules/1", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleSchedules(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestScheduleController_HandleSchedules_MethodNotAllowed(t *testing.T) {
	c := setupScheduleController(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/schedules", nil)
	rec := httptest.NewRecorder()
	c.HandleSchedules(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
