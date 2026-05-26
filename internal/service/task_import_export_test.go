package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/store"
)

func TestExportTasks_WithBisyncOptions(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	svc := NewTaskService(db, nil)

	bisyncOpts := map[string]any{
		"maxDelete":       "100",
		"checkAccess":     true,
		"conflictResolve": "none",
	}
	bisyncBytes, _ := json.Marshal(bisyncOpts)

	task := store.Task{
		Name:          "bisync-task",
		Mode:          "bisync",
		SourceRemote:  "remoteA",
		SourcePath:    "/src",
		TargetRemote:  "remoteB",
		TargetPath:    "/dst",
		BisyncOptions: bisyncBytes,
	}
	created, err := svc.CreateTask(task)
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	result, err := svc.ExportTasks()
	require.NoError(t, err)

	tasksData, ok := result["tasks"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, tasksData, 1)

	taskData := tasksData[0]
	exportedOpts, ok := taskData["bisyncOptions"].(map[string]any)
	require.True(t, ok, "bisyncOptions should be present in export")
	assert.Equal(t, "100", exportedOpts["maxDelete"])
	assert.Equal(t, true, exportedOpts["checkAccess"])
	assert.Equal(t, "none", exportedOpts["conflictResolve"])
}

func TestExportTasks_WithoutBisyncOptions(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	svc := NewTaskService(db, nil)

	task := store.Task{
		Name:          "normal-task",
		Mode:          "copy",
		SourceRemote:  "remoteA",
		SourcePath:    "/src",
		TargetRemote:  "remoteB",
		TargetPath:    "/dst",
	}
	created, err := svc.CreateTask(task)
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	result, err := svc.ExportTasks()
	require.NoError(t, err)

	tasksData, ok := result["tasks"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, tasksData, 1)

	taskData := tasksData[0]
	_, hasBisyncOpts := taskData["bisyncOptions"]
	assert.False(t, hasBisyncOpts, "bisyncOptions should not be present")
}

func TestImportTasks_WithBisyncOptions(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	svc := NewTaskService(db, nil)

	importData := map[string]any{
		"version": 1,
		"tasks": []any{
			map[string]any{
				"name":         "imported-bisync",
				"mode":         "bisync",
				"sourceRemote": "remoteA",
				"sourcePath":   "/src",
				"targetRemote": "remoteB",
				"targetPath":   "/dst",
				"bisyncOptions": map[string]any{
					"maxDelete":       "200",
					"checkAccess":     false,
					"conflictResolve": "path1",
				},
			},
		},
	}

	imported, skipped, overwritten, err := svc.ImportTasks(importData, "skip")
	require.NoError(t, err)
	assert.Equal(t, 1, imported)
	assert.Equal(t, 0, skipped)
	assert.Equal(t, 0, overwritten)

	tasks, err := svc.ListTasks()
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "imported-bisync", tasks[0].Name)

	var opts map[string]any
	err = json.Unmarshal(tasks[0].BisyncOptions, &opts)
	require.NoError(t, err)
	assert.Equal(t, "200", opts["maxDelete"])
	assert.Equal(t, false, opts["checkAccess"])
	assert.Equal(t, "path1", opts["conflictResolve"])
}

func TestImportTasks_OverwritePreservesBisyncOptions(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	svc := NewTaskService(db, nil)

	// 创建已有任务（有 bisyncOptions）
	existingOpts, _ := json.Marshal(map[string]any{"maxDelete": "50"})
	_, err = svc.CreateTask(store.Task{
		Name:          "existing-task",
		Mode:          "bisync",
		SourceRemote:  "remoteA",
		SourcePath:    "/src",
		TargetRemote:  "remoteB",
		TargetPath:    "/dst",
		BisyncOptions: existingOpts,
	})
	require.NoError(t, err)

	// 导入时不带 bisyncOptions → 应保留原有
	importData := map[string]any{
		"version": 1,
		"tasks": []any{
			map[string]any{
				"name":         "existing-task",
				"mode":         "bisync",
				"sourceRemote": "remoteA",
				"sourcePath":   "/src",
				"targetRemote": "remoteB",
				"targetPath":   "/dst",
			},
		},
	}

	imported, skipped, overwritten, err := svc.ImportTasks(importData, "overwrite")
	require.NoError(t, err)
	assert.Equal(t, 0, imported)
	assert.Equal(t, 0, skipped)
	assert.Equal(t, 1, overwritten)

	tasks, err := svc.ListTasks()
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	var opts map[string]any
	err = json.Unmarshal(tasks[0].BisyncOptions, &opts)
	require.NoError(t, err)
	assert.Equal(t, "50", opts["maxDelete"])
}
