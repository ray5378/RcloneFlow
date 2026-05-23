package scheduler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/adapter"
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
		{"04,03,06|17,19|*|*|*", "0 04,03,06 17,19 * * *"},
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
		"",
		"*|*|*",
		"0|12|*",
		"0|12|*|*",
		"0|12|*|*|*|*",
		"60|12|*|*|*",
		"61|*|*|*|*",
		"0|24|*|*|*",
		"*|24|*|*|*",
		"0|12|32|*|*",
		"*|*|0|*|*",
		"0|12|*|13|*",
		"*|*|*|13|*",
		"0|12|*|*|7",
		"*|*|*|*|7",
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

func TestTaskRunner_RunTask_TaskNotFound(t *testing.T) {
	db := setupSchedulerDB(t)
	runner := &taskRunner{db: db, rc: nil}
	err := runner.RunTask(context.Background(), 999, "manual")
	assert.NoError(t, err)
}

func TestTaskRunner_RunTask_RCError(t *testing.T) {
	db := setupSchedulerDB(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rclone error", 500)
	}))
	defer ts.Close()

	rc := adapter.NewRcloneClient(&adapter.RcloneConfig{BaseURL: ts.URL})

	task := store.Task{
		Name:         "test-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/path",
		TargetRemote: "dst",
		TargetPath:   "/path",
		Options:      json.RawMessage(`{"transfers":4}`),
	}
	created, err := db.AddTask(task)
	require.NoError(t, err)
	id := created.ID
	require.NoError(t, err)

	runner := &taskRunner{db: db, rc: rc}
	err = runner.RunTask(context.Background(), id, "scheduled")
	assert.Error(t, err)

	runs, err := db.ListRunsByTask(id)
	require.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, "failed", runs[0].Status)
}

func TestTaskRunner_RunTask_Success(t *testing.T) {
	db := setupSchedulerDB(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jobid": 42})
	}))
	defer ts.Close()

	rc := adapter.NewRcloneClient(&adapter.RcloneConfig{BaseURL: ts.URL})

	task := store.Task{
		Name:         "success-task",
		Mode:         "sync",
		SourceRemote: "src",
		SourcePath:   "/data",
		TargetRemote: "dst",
		TargetPath:   "/backup",
	}
	created2, err := db.AddTask(task)
	require.NoError(t, err)
	id2 := created2.ID

	runner := &taskRunner{db: db, rc: rc}
	err = runner.RunTask(context.Background(), id2, "scheduled")
	assert.NoError(t, err)

	runs, err := db.ListRunsByTask(id2)
	require.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, "running", runs[0].Status)
	assert.Equal(t, "scheduled", runs[0].Trigger)
}
