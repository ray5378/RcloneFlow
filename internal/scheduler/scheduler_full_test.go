package scheduler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/store"
)

func setupSchedulerDB(t *testing.T) *store.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_sched_test_*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNew(t *testing.T) {
	db := setupSchedulerDB(t)
	s := New(db, nil)
	assert.NotNil(t, s)
	assert.NotNil(t, s.cron)
	assert.NotNil(t, s.entries)
	assert.Equal(t, db, s.DB())
}

func TestNewWithRunner(t *testing.T) {
	db := setupSchedulerDB(t)
	runner := &mockRunner{}
	s := NewWithRunner(db, runner)
	assert.NotNil(t, s)
	assert.NotNil(t, s.cron)
	assert.Equal(t, db, s.DB())
}

type mockRunner struct {
	calledWithTaskID int64
	calledWithTrigger string
}

func (m *mockRunner) RunTask(ctx context.Context, taskID int64, trigger string) error {
	m.calledWithTaskID = taskID
	m.calledWithTrigger = trigger
	return nil
}

func TestScheduler_AddSchedule_Disabled(t *testing.T) {
	db := setupSchedulerDB(t)
	s := New(db, nil)

	schedule := store.Schedule{
		ID:      1,
		TaskID:  10,
		Spec:    "0|*|*|*|*",
		Enabled: false,
	}
	err := s.AddSchedule(schedule)
	require.NoError(t, err)
}

func TestScheduler_AddSchedule_InvalidSpec(t *testing.T) {
	db := setupSchedulerDB(t)
	s := New(db, nil)

	schedule := store.Schedule{
		ID:      1,
		TaskID:  10,
		Spec:    "invalid",
		Enabled: true,
	}
	err := s.AddSchedule(schedule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cron spec")
}

func TestScheduler_AddSchedule_Valid(t *testing.T) {
	db := setupSchedulerDB(t)
	s := New(db, nil)

	schedule := store.Schedule{
		ID:      1,
		TaskID:  10,
		Spec:    "0|12|*|*|*",
		Enabled: true,
	}
	err := s.AddSchedule(schedule)
	require.NoError(t, err)
	assert.Len(t, s.entries, 1)
}

func TestScheduler_RemoveSchedule(t *testing.T) {
	db := setupSchedulerDB(t)
	s := New(db, nil)

	schedule := store.Schedule{
		ID:      1,
		TaskID:  10,
		Spec:    "0|12|*|*|*",
		Enabled: true,
	}
	err := s.AddSchedule(schedule)
	require.NoError(t, err)

	s.RemoveSchedule(1)
	assert.Empty(t, s.entries)
}

func TestScheduler_RemoveSchedule_NonExistent(t *testing.T) {
	db := setupSchedulerDB(t)
	s := New(db, nil)

	s.RemoveSchedule(999)
	assert.Empty(t, s.entries)
}

func TestScheduler_Start_Stop(t *testing.T) {
	db := setupSchedulerDB(t)
	s := New(db, nil)

	err := s.Start()
	require.NoError(t, err)

	s.Stop()
}

func TestScheduler_Start_WithDisabledSchedule(t *testing.T) {
	db := setupSchedulerDB(t)
	// Add a task first to satisfy foreign key constraint
	_, err := db.AddTask(store.Task{
		Name:         "test-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/src",
		TargetRemote: "dst",
		TargetPath:   "/dst",
	})
	require.NoError(t, err)

	_, err = db.AddSchedule(store.Schedule{
		TaskID:  1,
		Spec:    "0|12|*|*|*",
		Enabled: false,
	})
	require.NoError(t, err)

	s := New(db, nil)
	err = s.Start()
	require.NoError(t, err)
	s.Stop()
}

func TestParseSpecToCron_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0|12|*|*|*", "0 0 12 * * *"},
		{"30|8|1|1|*", "0 30 8 1 1 *"},
		{"0,30|9,17|*|*|1,5", "0 0,30 9,17 * * 1,5"},
		{"*|*|*|*|*", "0 * * * * *"},
		{"|*|*|*|*", "0 * * * * *"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseSpecToCron(tt.input)
			assert.True(t, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseSpecToCron_Invalid(t *testing.T) {
	invalidSpecs := []string{
		"0|12|*",
		"0|12|*|*",
		"0|12|*|*|*|*",
		"60|12|*|*|*",
		"0|24|*|*|*",
		"0|12|32|*|*",
		"0|12|*|13|*",
		"0|12|*|*|7",
		"abc|12|*|*|*",
	}
	for _, spec := range invalidSpecs {
		t.Run(spec, func(t *testing.T) {
			_, ok := ParseSpecToCron(spec)
			assert.False(t, ok, "expected invalid spec: %s", spec)
		})
	}
}

func TestCalcNextRun_Valid(t *testing.T) {
	next, err := CalcNextRun("0 0 12 * * *")
	require.NoError(t, err)
	assert.True(t, next.After(time.Now()))
}

func TestCalcNextRun_Invalid(t *testing.T) {
	_, err := CalcNextRun("invalid")
	assert.Error(t, err)
}

func TestScheduler_DB(t *testing.T) {
	db := setupSchedulerDB(t)
	s := New(db, nil)
	assert.Equal(t, db, s.DB())
}
