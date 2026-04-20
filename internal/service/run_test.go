package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type runServiceDBMock struct {
	runsByID     map[int64]RunRecord
	runsByTask   map[int64][]RunRecord
	deletedRun   []int64
	deletedTasks []int64
}

func (m *runServiceDBMock) ListRuns(page, pageSize int) ([]RunRecord, int, error) { return nil, 0, nil }
func (m *runServiceDBMock) ListRunsByTask(taskId int64) ([]RunRecord, error) {
	return m.runsByTask[taskId], nil
}
func (m *runServiceDBMock) ListActiveRuns() ([]RunRecord, error) { return nil, nil }
func (m *runServiceDBMock) GetActiveRunByTaskID(taskID int64) (RunRecord, error) { return RunRecord{}, nil }
func (m *runServiceDBMock) GetRun(id int64) (RunRecord, error) { return m.runsByID[id], nil }
func (m *runServiceDBMock) UpdateRun(id int64, updateFn func(*RunRecord))          {}
func (m *runServiceDBMock) DeleteRun(id int64) error {
	m.deletedRun = append(m.deletedRun, id)
	return nil
}
func (m *runServiceDBMock) DeleteAllRuns() error { return nil }
func (m *runServiceDBMock) DeleteRunsByTask(taskId int64) error {
	m.deletedTasks = append(m.deletedTasks, taskId)
	return nil
}
func (m *runServiceDBMock) CleanOldRuns(days int) (int64, error) { return 0, nil }

func TestRunService_DeleteRun_RemovesLogFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logDir := filepath.Join(tmpDir, "logs", "task-a-0421")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	logPath := filepath.Join(logDir, "0001.log")
	if err := os.WriteFile(logPath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	summaryBytes, _ := json.Marshal(map[string]any{"stderrFile": logPath})
	mock := &runServiceDBMock{runsByID: map[int64]RunRecord{1: {ID: 1, TaskID: 9, Summary: string(summaryBytes)}}}

	svc := NewRunService(mock)
	if err := svc.DeleteRun(1); err != nil {
		t.Fatalf("DeleteRun() error = %v", err)
	}
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("expected log removed, got err=%v", err)
	}
	if _, err := os.Stat(logDir); !os.IsNotExist(err) {
		t.Fatalf("expected empty log dir removed, got err=%v", err)
	}
	if len(mock.deletedRun) != 1 || mock.deletedRun[0] != 1 {
		t.Fatalf("expected deleted run id 1, got %#v", mock.deletedRun)
	}
}

func TestRunService_DeleteRunsByTask_RemovesAllKnownLogs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logDir1 := filepath.Join(tmpDir, "logs", "task-a-0421")
	logDir2 := filepath.Join(tmpDir, "logs", "task-a-0422")
	if err := os.MkdirAll(logDir1, 0o755); err != nil {
		t.Fatalf("MkdirAll(logDir1) error = %v", err)
	}
	if err := os.MkdirAll(logDir2, 0o755); err != nil {
		t.Fatalf("MkdirAll(logDir2) error = %v", err)
	}
	log1 := filepath.Join(logDir1, "0001.log")
	log2 := filepath.Join(logDir2, "0002.log")
	if err := os.WriteFile(log1, []byte("one"), 0o644); err != nil {
		t.Fatalf("WriteFile(log1) error = %v", err)
	}
	if err := os.WriteFile(log2, []byte("two"), 0o644); err != nil {
		t.Fatalf("WriteFile(log2) error = %v", err)
	}
	sum1, _ := json.Marshal(map[string]any{"stderrFile": log1})
	sum2, _ := json.Marshal(map[string]any{"stderrFile": log2})
	mock := &runServiceDBMock{runsByTask: map[int64][]RunRecord{9: {{ID: 1, TaskID: 9, Summary: string(sum1)}, {ID: 2, TaskID: 9, Summary: string(sum2)}}}}

	svc := NewRunService(mock)
	if err := svc.DeleteRunsByTask(9); err != nil {
		t.Fatalf("DeleteRunsByTask() error = %v", err)
	}
	if _, err := os.Stat(log1); !os.IsNotExist(err) {
		t.Fatalf("expected log1 removed, got err=%v", err)
	}
	if _, err := os.Stat(log2); !os.IsNotExist(err) {
		t.Fatalf("expected log2 removed, got err=%v", err)
	}
	if _, err := os.Stat(logDir1); !os.IsNotExist(err) {
		t.Fatalf("expected logDir1 removed, got err=%v", err)
	}
	if _, err := os.Stat(logDir2); !os.IsNotExist(err) {
		t.Fatalf("expected logDir2 removed, got err=%v", err)
	}
	if len(mock.deletedTasks) != 1 || mock.deletedTasks[0] != 9 {
		t.Fatalf("expected deleted task id 9, got %#v", mock.deletedTasks)
	}
}
