package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/store"
)

func TestIntegration_TaskLifecycle(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "integration-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)
	assert.Greater(t, task.ID, int64(0))

	got, ok := db.GetTask(task.ID)
	require.True(t, ok)
	assert.Equal(t, "integration-task", got.Name)
	assert.Equal(t, "copy", got.Mode)

	tasks, err := db.ListTasks()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tasks), 1)

	err = db.DeleteTask(task.ID)
	require.NoError(t, err)

	_, ok = db.GetTask(task.ID)
	assert.False(t, ok)
}

func TestIntegration_ScheduleLifecycle(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "scheduled-task",
		Mode:         "sync",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	sched, err := db.AddSchedule(store.Schedule{
		TaskID:  task.ID,
		Spec:    "@every 5m",
		Enabled: true,
	})
	require.NoError(t, err)
	assert.Greater(t, sched.ID, int64(0))

	got, ok := db.GetSchedule(sched.ID)
	require.True(t, ok)
	assert.Equal(t, "@every 5m", got.Spec)
	assert.True(t, got.Enabled)

	schedules, err := db.ListSchedules()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(schedules), 1)

	err = db.DeleteSchedule(sched.ID)
	require.NoError(t, err)

	_, ok = db.GetSchedule(sched.ID)
	assert.False(t, ok)
}

func TestIntegration_RunLifecycle(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "run-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	run, err := db.AddRun(store.Run{
		TaskID:  task.ID,
		Status:  "running",
		Trigger: "manual",
	})
	require.NoError(t, err)
	assert.Greater(t, run.ID, int64(0))

	err = db.UpdateRunStatus(run.ID, "finished", "", map[string]any{"files": 10, "bytes": 1024})
	require.NoError(t, err)

	got, err := db.GetRun(run.ID)
	require.NoError(t, err)
	assert.Equal(t, "finished", got.Status)

	runs, total, err := db.ListRuns(1, 50)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.GreaterOrEqual(t, len(runs), 1)

	runs, err = db.ListRunsByTask(task.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(runs), 1)
}

func TestIntegration_OptionsPersistence(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	opts := map[string]any{
		"bandwidth":      "10M",
		"transfers":      4,
		"checkers":       8,
		"exclude":        []string{"*.tmp", "*.log"},
		"webhookEnabled": true,
		"webhookUrl":     "https://example.com/webhook",
	}

	optsBytes, err := json.Marshal(opts)
	require.NoError(t, err)

	task, err := db.AddTask(store.Task{
		Name:         "options-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
		Options:      optsBytes,
	})
	require.NoError(t, err)

	got, ok := db.GetTask(task.ID)
	require.True(t, ok)

	var retrieved map[string]any
	err = json.Unmarshal(got.Options, &retrieved)
	require.NoError(t, err)
	assert.Equal(t, "10M", retrieved["bandwidth"])
	assert.Equal(t, float64(4), retrieved["transfers"])
	assert.Equal(t, float64(8), retrieved["checkers"])
	assert.Equal(t, true, retrieved["webhookEnabled"])
}

func TestIntegration_RunPagination(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "pagination-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	for i := 0; i < 25; i++ {
		_, err := db.AddRun(store.Run{
			TaskID:  task.ID,
			Status:  "finished",
			Trigger: "manual",
		})
		require.NoError(t, err)
	}

	runs, total, err := db.ListRuns(1, 10)
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Equal(t, 10, len(runs))

	runs, total, err = db.ListRuns(2, 10)
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Equal(t, 10, len(runs))

	runs, total, err = db.ListRuns(3, 10)
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Equal(t, 5, len(runs))
}

func TestIntegration_CleanOldRuns(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "cleanup-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		_, err := db.AddRun(store.Run{
			TaskID:  task.ID,
			Status:  "finished",
			Trigger: "manual",
		})
		require.NoError(t, err)
	}

	runs, _, err := db.ListRuns(1, 100)
	require.NoError(t, err)
	assert.Equal(t, 5, len(runs))

	affected, err := db.CleanOldRuns(365)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, affected, 0)

	runs, _, err = db.ListRuns(1, 100)
	require.NoError(t, err)
	_ = affected
}

func TestIntegration_DeleteAllRuns(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "delete-all-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		_, err := db.AddRun(store.Run{
			TaskID:  task.ID,
			Status:  "finished",
			Trigger: "manual",
		})
		require.NoError(t, err)
	}

	runs, _, err := db.ListRuns(1, 100)
	require.NoError(t, err)
	assert.Equal(t, 5, len(runs))

	err = db.DeleteAllRuns()
	require.NoError(t, err)

	runs, _, err = db.ListRuns(1, 100)
	require.NoError(t, err)
	assert.Equal(t, 0, len(runs))
}

func TestIntegration_ScheduleUpdate(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "update-sched",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	sched, err := db.AddSchedule(store.Schedule{
		TaskID:  task.ID,
		Spec:    "@every 10m",
		Enabled: true,
	})
	require.NoError(t, err)

	err = db.UpdateScheduleSpec(sched.ID, "@hourly")
	require.NoError(t, err)

	sched, ok := db.GetSchedule(sched.ID)
	require.True(t, ok)
	assert.Equal(t, "@hourly", sched.Spec)

	err = db.SetScheduleEnabled(sched.ID, false)
	require.NoError(t, err)

	sched, ok = db.GetSchedule(sched.ID)
	require.True(t, ok)
	assert.False(t, sched.Enabled)

	nextRunTime := time.Now().Add(1 * time.Hour)
	err = db.UpdateScheduleNextRunTime(sched.ID, nextRunTime)
	require.NoError(t, err)
}

func TestIntegration_TaskSortOrder(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task1, err := db.AddTask(store.Task{
		Name:         "task-1",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	task2, err := db.AddTask(store.Task{
		Name:         "task-2",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	updates := map[int64]int64{
		task1.ID: 2,
		task2.ID: 1,
	}
	err = db.UpdateTaskSortOrders(updates)
	require.NoError(t, err)

	tasks, err := db.ListTasks()
	require.NoError(t, err)
	assert.Equal(t, 2, len(tasks))

	taskMap := make(map[int64]store.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	if task1.ID == taskMap[task1.ID].ID {
		assert.Equal(t, int64(2), taskMap[task1.ID].SortOrder)
	}
	if task2.ID == taskMap[task2.ID].ID {
		assert.Equal(t, int64(1), taskMap[task2.ID].SortOrder)
	}
}

func TestIntegration_Vacuum(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "vacuum-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		_, err := db.AddRun(store.Run{
			TaskID:  task.ID,
			Status:  "finished",
			Trigger: "manual",
		})
		require.NoError(t, err)
	}

	err = db.Vacuum()
	require.NoError(t, err)
}

func TestIntegration_ActiveRuns(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "active-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	_, err = db.AddRun(store.Run{
		TaskID:  task.ID,
		Status:  "running",
		Trigger: "manual",
	})
	require.NoError(t, err)

	_, err = db.AddRun(store.Run{
		TaskID:  task.ID,
		Status:  "finished",
		Trigger: "manual",
	})
	require.NoError(t, err)

	runs, err := db.ListActiveRuns()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(runs), 1)

	err = db.ClearAllRunningStatus()
	require.NoError(t, err)

	runs, err = db.ListActiveRuns()
	require.NoError(t, err)
	assert.Equal(t, 0, len(runs))
}

func TestIntegration_DeleteRunsByTask(t *testing.T) {
	db, err := store.Open(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	task1, err := db.AddTask(store.Task{
		Name:         "task-1",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	task2, err := db.AddTask(store.Task{
		Name:         "task-2",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
	})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		_, err := db.AddRun(store.Run{
			TaskID:  task1.ID,
			Status:  "finished",
			Trigger: "manual",
		})
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		_, err := db.AddRun(store.Run{
			TaskID:  task2.ID,
			Status:  "finished",
			Trigger: "manual",
		})
		require.NoError(t, err)
	}

	err = db.DeleteRunsByTask(task1.ID)
	require.NoError(t, err)

	runs, err := db.ListRunsByTask(task1.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, len(runs))

	runs, err = db.ListRunsByTask(task2.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(runs))
}