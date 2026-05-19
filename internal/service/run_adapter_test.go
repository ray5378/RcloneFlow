package service

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/store"
)

func setupRunAdapterDB(t *testing.T) *store.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_run_adapter_test_*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestFormatTime(t *testing.T) {
	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	got := formatTime(ts)
	assert.Equal(t, "2024-01-15T10:30:00Z", got)
}

func TestFormatOptTime_Nil(t *testing.T) {
	got := formatOptTime(nil)
	assert.Equal(t, "", got)
}

func TestFormatOptTime_Value(t *testing.T) {
	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	got := formatOptTime(&ts)
	assert.Equal(t, "2024-01-15T10:30:00Z", got)
}

func TestToRunRecord_Basic(t *testing.T) {
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	r := store.Run{
		ID:           1,
		TaskID:       10,
		Status:       "finished",
		Trigger:      "manual",
		CreatedAt:    start,
		TaskName:     "test-task",
		TaskMode:     "sync",
		SourceRemote: "src",
		SourcePath:   "/src",
		TargetRemote: "dst",
		TargetPath:   "/dst",
	}

	rec := toRunRecord(r)

	assert.Equal(t, int64(1), rec.ID)
	assert.Equal(t, int64(10), rec.TaskID)
	assert.Equal(t, "finished", rec.Status)
	assert.Equal(t, "manual", rec.Trigger)
	assert.Equal(t, "2024-01-15T10:00:00Z", rec.StartedAt)
	assert.Equal(t, "test-task", rec.TaskName)
	assert.Equal(t, "sync", rec.TaskMode)
}

func TestToRunRecord_WithFinishedAt(t *testing.T) {
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	r := store.Run{
		ID:         1,
		CreatedAt:  start,
		FinishedAt: &end,
	}

	rec := toRunRecord(r)

	assert.Equal(t, "2024-01-15T10:00:00Z", rec.StartedAt)
	assert.Equal(t, "2024-01-15T10:30:00Z", rec.FinishedAt)
}

func TestToRunRecord_WithSummary(t *testing.T) {
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	r := store.Run{
		ID:        1,
		CreatedAt: start,
		Summary: map[string]any{
			"bytes":    float64(1000),
			"total":    float64(2000),
			"finishedAt": "2024-01-15T10:30:00Z",
		},
	}

	rec := toRunRecord(r)

	assert.Contains(t, rec.Summary, `"bytes":1000`)
	assert.Equal(t, "2024-01-15T10:30:00Z", rec.FinishedAt)
}

func TestToRunRecords(t *testing.T) {
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	runs := []store.Run{
		{ID: 1, CreatedAt: start},
		{ID: 2, CreatedAt: start},
	}

	recs := toRunRecords(runs)

	assert.Len(t, recs, 2)
	assert.Equal(t, int64(1), recs[0].ID)
	assert.Equal(t, int64(2), recs[1].ID)
}

func TestStoreRunAdapter_NewStoreRunAdapter(t *testing.T) {
	db := setupRunAdapterDB(t)
	adapter := NewStoreRunAdapter(db)
	assert.NotNil(t, adapter)
}

func TestStoreRunAdapter_ListRuns(t *testing.T) {
	db := setupRunAdapterDB(t)
	adapter := NewStoreRunAdapter(db)

	records, total, err := adapter.ListRuns(1, 10)
	require.NoError(t, err)
	assert.Empty(t, records)
	assert.Equal(t, 0, total)
}

func TestStoreRunAdapter_ListRunsByTask(t *testing.T) {
	db := setupRunAdapterDB(t)
	adapter := NewStoreRunAdapter(db)

	records, err := adapter.ListRunsByTask(1)
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestStoreRunAdapter_ListActiveRuns(t *testing.T) {
	db := setupRunAdapterDB(t)
	adapter := NewStoreRunAdapter(db)

	records, err := adapter.ListActiveRuns()
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestStoreRunAdapter_GetRun_NotFound(t *testing.T) {
	db := setupRunAdapterDB(t)
	adapter := NewStoreRunAdapter(db)

	_, err := adapter.GetRun(999)
	assert.Error(t, err)
}

func TestStoreRunAdapter_DeleteRun(t *testing.T) {
	db := setupRunAdapterDB(t)
	adapter := NewStoreRunAdapter(db)

	err := adapter.DeleteRun(999)
	assert.NoError(t, err)
}

func TestStoreRunAdapter_DeleteAllRuns(t *testing.T) {
	db := setupRunAdapterDB(t)
	adapter := NewStoreRunAdapter(db)

	err := adapter.DeleteAllRuns()
	assert.NoError(t, err)
}

func TestStoreRunAdapter_DeleteRunsByTask(t *testing.T) {
	db := setupRunAdapterDB(t)
	adapter := NewStoreRunAdapter(db)

	err := adapter.DeleteRunsByTask(1)
	assert.NoError(t, err)
}

func TestStoreRunAdapter_CleanOldRuns(t *testing.T) {
	db := setupRunAdapterDB(t)
	adapter := NewStoreRunAdapter(db)

	deleted, err := adapter.CleanOldRuns(7)
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)
}
