package runnercli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"rcloneflow/internal/store"
)

func TestStoreDBAdapter_UpdateTask(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := store.Open(tmpDir)
	assert.NoError(t, err)
	defer db.Close()

	adapter := &StoreDBAdapter{DB: db}

	task := store.Task{
		Name:         "test-update-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
	}

	created, err := db.AddTask(task)
	assert.NoError(t, err)

	updatedTask := created
	updatedTask.Name = "updated-name"

	err = adapter.UpdateTask(created.ID, updatedTask)
	assert.NoError(t, err)

	retrieved, ok := db.GetTask(created.ID)
	assert.True(t, ok)
	assert.Equal(t, "updated-name", retrieved.Name)
}

func TestStoreDBAdapter_GetTask(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := store.Open(tmpDir)
	assert.NoError(t, err)
	defer db.Close()

	adapter := &StoreDBAdapter{DB: db}

	task, err := db.AddTask(store.Task{
		Name:         "test-get-task",
		Mode:         "sync",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
	})
	assert.NoError(t, err)

	retrieved, ok := adapter.GetTask(task.ID)
	assert.True(t, ok)
	assert.Equal(t, "test-get-task", retrieved.Name)
}

func TestStoreDBAdapter_GetTask_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := store.Open(tmpDir)
	assert.NoError(t, err)
	defer db.Close()

	adapter := &StoreDBAdapter{DB: db}

	_, ok := adapter.GetTask(99999)
	assert.False(t, ok)
}

func TestBuildBisyncFinalSummary(t *testing.T) {
	summary, files1To2, files2To1 := buildBisyncFinalSummary("")

	assert.NotNil(t, summary)
	assert.NotNil(t, files1To2)
	assert.NotNil(t, files2To1)

	assert.Equal(t, 0, summary.Path1ToPath2["copied"])
	assert.Equal(t, 0, summary.Path1ToPath2["deleted"])
	assert.Equal(t, 0, summary.Path1ToPath2["skipped"])
	assert.Equal(t, 0, summary.Path1ToPath2["failed"])
	assert.Equal(t, 0, summary.Path1ToPath2["total"])

	assert.Equal(t, 0, summary.Path2ToPath1["copied"])
	assert.Equal(t, 0, summary.Path2ToPath1["deleted"])
	assert.Equal(t, 0, summary.Path2ToPath1["skipped"])
	assert.Equal(t, 0, summary.Path2ToPath1["failed"])
	assert.Equal(t, 0, summary.Path2ToPath1["total"])

	assert.Empty(t, files1To2)
	assert.Empty(t, files2To1)
}

func TestBuildBisyncFinalSummary_WithPath(t *testing.T) {
	summary, files1To2, files2To1 := buildBisyncFinalSummary("/nonexistent/path/bisync.log")

	assert.NotNil(t, summary)
	assert.NotNil(t, files1To2)
	assert.NotNil(t, files2To1)
	assert.Empty(t, files1To2)
	assert.Empty(t, files2To1)
}

func TestSanitizeFilenameForBisync(t *testing.T) {
	testCases := []struct {
		input    string
		taskID   int64
		expected string
	}{
		{"valid-name", 1, "valid-name"},
		{"name with spaces", 1, "name_with_spaces"},
		{"special@#$chars", 1, "special_chars"},
		{"中文文件名", 1, "中文文件名"},
		{"", 123, "task-123"},
		{"a", 999, "a"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := sanitizeFilenameForBisync(tc.input, tc.taskID)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSanitizeFilenameForBisync_Truncate(t *testing.T) {
	longName := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := sanitizeFilenameForBisync(longName, 1)
	assert.LessOrEqual(t, len(result), 60)
}

func TestWSBroadcaster_Broadcast(t *testing.T) {
	broadcaster := &WSBroadcaster{}
	
	assert.NotNil(t, broadcaster)
	
	broadcaster.Broadcast("test-event", map[string]any{"key": "value"})
}