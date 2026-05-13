package service

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type runLifecycleDBMock struct {
	listRunsFn            func(page, pageSize int) ([]RunRecord, int, error)
	listRunsByTaskFn      func(taskID int64) ([]RunRecord, error)
	listActiveRunsFn      func() ([]RunRecord, error)
	getActiveRunByTaskIDFn func(taskID int64) (RunRecord, error)
	getRunFn              func(id int64) (RunRecord, error)
	updateRunFn           func(id int64, updateFn func(*RunRecord))
	deleteRunIDs          []int64
	deleteAllCalled       bool
	deleteRunsByTaskIDs   []int64
	cleanOldRunsFn        func(days int) (int64, error)
}

func (m *runLifecycleDBMock) ListRuns(page, pageSize int) ([]RunRecord, int, error) {
	if m.listRunsFn != nil {
		return m.listRunsFn(page, pageSize)
	}
	return nil, 0, nil
}
func (m *runLifecycleDBMock) ListRunsByTask(taskID int64) ([]RunRecord, error) {
	if m.listRunsByTaskFn != nil {
		return m.listRunsByTaskFn(taskID)
	}
	return nil, nil
}
func (m *runLifecycleDBMock) ListActiveRuns() ([]RunRecord, error) {
	if m.listActiveRunsFn != nil {
		return m.listActiveRunsFn()
	}
	return nil, nil
}
func (m *runLifecycleDBMock) GetActiveRunByTaskID(taskID int64) (RunRecord, error) {
	if m.getActiveRunByTaskIDFn != nil {
		return m.getActiveRunByTaskIDFn(taskID)
	}
	return RunRecord{}, nil
}
func (m *runLifecycleDBMock) GetRun(id int64) (RunRecord, error) {
	if m.getRunFn != nil {
		return m.getRunFn(id)
	}
	return RunRecord{}, nil
}
func (m *runLifecycleDBMock) UpdateRun(id int64, updateFn func(*RunRecord)) {
	if m.updateRunFn != nil {
		m.updateRunFn(id, updateFn)
	}
}
func (m *runLifecycleDBMock) DeleteRun(id int64) error {
	m.deleteRunIDs = append(m.deleteRunIDs, id)
	return nil
}
func (m *runLifecycleDBMock) DeleteAllRuns() error {
	m.deleteAllCalled = true
	return nil
}
func (m *runLifecycleDBMock) DeleteRunsByTask(taskID int64) error {
	m.deleteRunsByTaskIDs = append(m.deleteRunsByTaskIDs, taskID)
	return nil
}
func (m *runLifecycleDBMock) CleanOldRuns(days int) (int64, error) {
	if m.cleanOldRunsFn != nil {
		return m.cleanOldRunsFn(days)
	}
	return 0, nil
}

func runSummaryJSON(t *testing.T, summary map[string]any) string {
	t.Helper()
	bs, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return string(bs)
}

func TestRunService_BasicDelegationAndStatusMerge(t *testing.T) {
	record := RunRecord{ID: 9, Status: "running", Summary: `{"nested":{"old":1},"keep":"yes"}`}
	mock := &runLifecycleDBMock{
		listRunsFn: func(page, pageSize int) ([]RunRecord, int, error) {
			return []RunRecord{{ID: 1}}, 7, nil
		},
		listRunsByTaskFn: func(taskID int64) ([]RunRecord, error) {
			return []RunRecord{{ID: 2, TaskID: taskID}}, nil
		},
		listActiveRunsFn: func() ([]RunRecord, error) {
			return []RunRecord{{ID: 3, Status: "running"}}, nil
		},
		getActiveRunByTaskIDFn: func(taskID int64) (RunRecord, error) {
			return RunRecord{ID: 4, TaskID: taskID, Status: "running"}, nil
		},
		updateRunFn: func(id int64, updateFn func(*RunRecord)) {
			if id != 9 {
				t.Fatalf("unexpected update id %d", id)
			}
			updateFn(&record)
		},
	}
	svc := NewRunService(mock)

	runs, total, err := svc.ListRuns(2, 50)
	if err != nil || len(runs) != 1 || total != 7 {
		t.Fatalf("ListRuns() = (%v, %d, %v)", runs, total, err)
	}
	byTask, err := svc.ListRunsByTask(11)
	if err != nil || len(byTask) != 1 || byTask[0].TaskID != 11 {
		t.Fatalf("ListRunsByTask() = (%v, %v)", byTask, err)
	}
	active, err := svc.ListActiveRuns()
	if err != nil || len(active) != 1 || active[0].ID != 3 {
		t.Fatalf("ListActiveRuns() = (%v, %v)", active, err)
	}
	activeByTask, err := svc.GetActiveRunByTaskID(12)
	if err != nil || activeByTask.TaskID != 12 {
		t.Fatalf("GetActiveRunByTaskID() = (%v, %v)", activeByTask, err)
	}

	svc.UpdateRunStatus(9, map[string]any{
		"nested": map[string]any{"new": 2},
		"finished": true,
		"success": false,
		"error": "boom",
	})
	var merged map[string]any
	if err := json.Unmarshal([]byte(record.Summary), &merged); err != nil {
		t.Fatalf("Unmarshal(summary) error = %v", err)
	}
	if record.Status != "failed" || record.Error != "boom" {
		t.Fatalf("expected failed/boom, got %s/%s", record.Status, record.Error)
	}
	if merged["keep"] != "yes" {
		t.Fatalf("expected keep=yes, got %v", merged["keep"])
	}
	nested := merged["nested"].(map[string]any)
	if nested["old"].(float64) != 1 || nested["new"].(float64) != 2 {
		t.Fatalf("unexpected nested merge: %#v", nested)
	}

	svc.UpdateRunStatus(9, map[string]any{"finished": true, "success": true})
	if record.Status != "finished" || record.Error != "" {
		t.Fatalf("expected finished with cleared error, got %s/%s", record.Status, record.Error)
	}
}

func TestRunService_DeleteAndCleanupFlows(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-run-cleanup-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mkRun := func(id int64, dir, name string) RunRecord {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		logPath := filepath.Join(dir, name)
		if err := os.WriteFile(logPath, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		return RunRecord{ID: id, Summary: runSummaryJSON(t, map[string]any{"stderrFile": logPath})}
	}

	run1 := mkRun(1, filepath.Join(tmpDir, "task-a"), "1.log")
	run2 := mkRun(2, filepath.Join(tmpDir, "task-b"), "2.log")
	run3 := mkRun(3, filepath.Join(tmpDir, "task-c"), "3.log")
	pages := map[int][]RunRecord{1: {run1, run2, run3}}

	mock := &runLifecycleDBMock{
		getRunFn: func(id int64) (RunRecord, error) {
			if id == 1 {
				return run1, nil
			}
			return RunRecord{}, errors.New("not found")
		},
		listRunsFn: func(page, pageSize int) ([]RunRecord, int, error) {
			return pages[page], 3, nil
		},
		listRunsByTaskFn: func(taskID int64) ([]RunRecord, error) {
			return []RunRecord{run2, run3}, nil
		},
	}
	svc := NewRunService(mock)

	if err := svc.DeleteRun(1); err != nil {
		t.Fatalf("DeleteRun() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "task-a", "1.log")); !os.IsNotExist(err) {
		t.Fatalf("expected log file removed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "task-a")); !os.IsNotExist(err) {
		t.Fatalf("expected empty dir removed, stat err = %v", err)
	}

	if err := svc.DeleteAllRuns(); err != nil {
		t.Fatalf("DeleteAllRuns() error = %v", err)
	}
	if !mock.deleteAllCalled {
		t.Fatal("expected DeleteAllRuns delegation")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "task-b", "2.log")); !os.IsNotExist(err) {
		t.Fatalf("expected task-b log removed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "task-c", "3.log")); !os.IsNotExist(err) {
		t.Fatalf("expected task-c log removed, stat err = %v", err)
	}

	run4 := mkRun(4, filepath.Join(tmpDir, "task-d"), "4.log")
	run5 := mkRun(5, filepath.Join(tmpDir, "task-e"), "5.log")
	mock.listRunsByTaskFn = func(taskID int64) ([]RunRecord, error) {
		return []RunRecord{run4, run5}, nil
	}
	if err := svc.DeleteRunsByTask(88); err != nil {
		t.Fatalf("DeleteRunsByTask() error = %v", err)
	}
	if len(mock.deleteRunsByTaskIDs) != 1 || mock.deleteRunsByTaskIDs[0] != 88 {
		t.Fatalf("unexpected DeleteRunsByTask calls: %#v", mock.deleteRunsByTaskIDs)
	}
}

func TestRunService_CleanOldRuns(t *testing.T) {
	now := time.Now()
	oldA := RunRecord{ID: 10, StartedAt: now.AddDate(0, 0, -10).Format(time.RFC3339)}
	oldB := RunRecord{ID: 11, StartedAt: now.AddDate(0, 0, -8).Format(time.RFC3339)}
	newer := RunRecord{ID: 12, StartedAt: now.AddDate(0, 0, -2).Format(time.RFC3339)}
	badTime := RunRecord{ID: 13, StartedAt: "not-a-time"}

	mock := &runLifecycleDBMock{
		listRunsFn: func(page, pageSize int) ([]RunRecord, int, error) {
			if page == 1 {
				return []RunRecord{oldA, newer, badTime, oldB}, 4, nil
			}
			return nil, 4, nil
		},
	}
	svc := NewRunService(mock)

	deleted, err := svc.CleanOldRuns(7)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if deleted != 2 {
		t.Fatalf("expected 2 deletions, got %d", deleted)
	}
	if len(mock.deleteRunIDs) != 2 || mock.deleteRunIDs[0] != 10 || mock.deleteRunIDs[1] != 11 {
		t.Fatalf("unexpected deleted run ids: %#v", mock.deleteRunIDs)
	}

	deleted, err = svc.CleanOldRuns(0)
	if err != nil || deleted != 0 {
		t.Fatalf("CleanOldRuns(0) = (%d, %v)", deleted, err)
	}
}
