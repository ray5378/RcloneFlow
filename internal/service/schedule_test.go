package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/store"
)

func setupScheduleDB(t *testing.T) *store.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_schedule_svc_test_*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestScheduleService_New(t *testing.T) {
	db := setupScheduleDB(t)
	svc := NewScheduleService(db)
	assert.NotNil(t, svc)
}

func TestScheduleService_ListSchedules(t *testing.T) {
	db := setupScheduleDB(t)
	svc := NewScheduleService(db)

	// 先添加Task
	task1, err := db.AddTask(store.Task{Name: "Test Task 1"})
	require.NoError(t, err)
	task2, err := db.AddTask(store.Task{Name: "Test Task 2"})
	require.NoError(t, err)

	// 先添加一些测试数据
	sched1, err := db.AddSchedule(store.Schedule{
		TaskID:  task1.ID,
		Spec:    "0|9|*|*|*",
		Enabled: true,
	})
	require.NoError(t, err)
	sched2, err := db.AddSchedule(store.Schedule{
		TaskID:  task2.ID,
		Spec:    "30|18|*|*|*",
		Enabled: false,
	})
	require.NoError(t, err)

	schedules, err := svc.ListSchedules()
	require.NoError(t, err)
	assert.Len(t, schedules, 2)

	// 检查是否包含这两个schedule
	schedMap := make(map[int64]store.Schedule)
	for _, s := range schedules {
		schedMap[s.ID] = s
	}

	assert.Equal(t, sched1.ID, schedMap[sched1.ID].ID)
	assert.Equal(t, sched1.TaskID, schedMap[sched1.ID].TaskID)
	assert.Equal(t, "0|9|*|*|*", schedMap[sched1.ID].Spec)
	assert.True(t, schedMap[sched1.ID].Enabled)

	assert.Equal(t, sched2.ID, schedMap[sched2.ID].ID)
	assert.Equal(t, sched2.TaskID, schedMap[sched2.ID].TaskID)
	assert.Equal(t, "30|18|*|*|*", schedMap[sched2.ID].Spec)
	assert.False(t, schedMap[sched2.ID].Enabled)
}

func TestScheduleService_CreateSchedule(t *testing.T) {
	db := setupScheduleDB(t)
	svc := NewScheduleService(db)

	// 先添加Task
	task, err := db.AddTask(store.Task{Name: "Test Task"})
	require.NoError(t, err)

	sched, err := svc.CreateSchedule(task.ID, "0|10|*|*|*", true)
	require.NoError(t, err)
	assert.NotZero(t, sched.ID)
	assert.Equal(t, task.ID, sched.TaskID)
	assert.Equal(t, "0|10|*|*|*", sched.Spec)
	assert.True(t, sched.Enabled)

	// 验证数据库中有这个schedule
	dbScheds, err := db.ListSchedules()
	require.NoError(t, err)
	assert.Len(t, dbScheds, 1)
	assert.Equal(t, sched.ID, dbScheds[0].ID)
}

func TestScheduleService_UpdateSpec(t *testing.T) {
	db := setupScheduleDB(t)
	svc := NewScheduleService(db)

	// 先添加Task
	task, err := db.AddTask(store.Task{Name: "Test Task"})
	require.NoError(t, err)

	sched, err := db.AddSchedule(store.Schedule{
		TaskID:  task.ID,
		Spec:    "0|9|*|*|*",
		Enabled: true,
	})
	require.NoError(t, err)

	err = svc.UpdateSpec(sched.ID, "30|18|*|*|*")
	require.NoError(t, err)

	// 验证修改生效
	updatedSched, ok := db.GetSchedule(sched.ID)
	require.True(t, ok)
	assert.Equal(t, "30|18|*|*|*", updatedSched.Spec)
}

func TestScheduleService_DeleteSchedule(t *testing.T) {
	db := setupScheduleDB(t)
	svc := NewScheduleService(db)

	// 先添加Task
	task, err := db.AddTask(store.Task{Name: "Test Task"})
	require.NoError(t, err)

	sched, err := db.AddSchedule(store.Schedule{
		TaskID:  task.ID,
		Spec:    "0|9|*|*|*",
		Enabled: true,
	})
	require.NoError(t, err)

	err = svc.DeleteSchedule(sched.ID)
	require.NoError(t, err)

	// 验证删除生效
	_, ok := db.GetSchedule(sched.ID)
	assert.False(t, ok)
}

func TestScheduleService_SetScheduleEnabled(t *testing.T) {
	db := setupScheduleDB(t)
	svc := NewScheduleService(db)

	// 先添加Task
	task, err := db.AddTask(store.Task{Name: "Test Task"})
	require.NoError(t, err)

	sched, err := db.AddSchedule(store.Schedule{
		TaskID:  task.ID,
		Spec:    "0|9|*|*|*",
		Enabled: true,
	})
	require.NoError(t, err)

	err = svc.SetScheduleEnabled(sched.ID, false)
	require.NoError(t, err)

	// 验证禁用生效
	updatedSched, ok := db.GetSchedule(sched.ID)
	require.True(t, ok)
	assert.False(t, updatedSched.Enabled)

	err = svc.SetScheduleEnabled(sched.ID, true)
	require.NoError(t, err)

	// 验证启用生效
	updatedSched, ok = db.GetSchedule(sched.ID)
	require.True(t, ok)
	assert.True(t, updatedSched.Enabled)
}
