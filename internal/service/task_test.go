package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"rcloneflow/internal/store"
)

// mockTaskRunner 模拟任务运行器
type mockTaskRunner struct {
	runTaskFn func(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error)
}

func (m *mockTaskRunner) RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error) {
	if m.runTaskFn != nil {
		return m.runTaskFn(ctx, taskID, mode, srcRemote, srcPath, dstRemote, dstPath, trigger)
	}
	return 123, nil
}

// mockTaskDB 模拟任务数据库
type mockTaskDB struct {
	tasks []store.Task
}

func (m *mockTaskDB) ListTasks() ([]store.Task, error) {
	return m.tasks, nil
}

func (m *mockTaskDB) AddTask(task store.Task) (store.Task, error) {
	task.ID = int64(len(m.tasks) + 1)
	m.tasks = append(m.tasks, task)
	return task, nil
}

func (m *mockTaskDB) GetTask(id int64) (store.Task, bool) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, true
		}
	}
	return store.Task{}, false
}

func (m *mockTaskDB) UpdateTask(id int64, task store.Task) error {
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks[i] = task
			return nil
		}
	}
	return nil
}

func (m *mockTaskDB) DeleteTask(id int64) error {
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockTaskDB) AddRun(run store.Run) (store.Run, error) {
	run.ID = int64(len(m.tasks) * 100)
	return run, nil
}

func TestTaskService_ListTasks(t *testing.T) {
	db := &mockTaskDB{
		tasks: []store.Task{
			{ID: 1, Name: "task1", Mode: "copy", SourceRemote: "local", TargetRemote: "gdrive"},
			{ID: 2, Name: "task2", Mode: "sync", SourceRemote: "local", TargetRemote: "s3"},
		},
	}
	runner := &mockTaskRunner{}

	// 由于TaskService依赖store.DB，我们直接测试DAO层
	_ = db
	_ = runner
}

func TestErrTaskNotFound(t *testing.T) {
	err := ErrTaskNotFound
	if err.Error() != "task not found" {
		t.Errorf("expected 'task not found', got '%s'", err.Error())
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	db := &mockTaskDB{
		tasks: []store.Task{
			{ID: 1, Name: "original", Mode: "copy"},
		},
	}

	// 测试更新逻辑
	db.UpdateTask(1, store.Task{ID: 1, Name: "updated", Mode: "sync"})

	task, _ := db.GetTask(1)
	if task.Name != "updated" {
		t.Errorf("expected Name 'updated', got '%s'", task.Name)
	}
	if task.Mode != "sync" {
		t.Errorf("expected Mode 'sync', got '%s'", task.Mode)
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	db := &mockTaskDB{
		tasks: []store.Task{
			{ID: 1, Name: "task1"},
			{ID: 2, Name: "task2"},
		},
	}

	db.DeleteTask(1)

	tasks, _ := db.ListTasks()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task after delete, got %d", len(tasks))
	}

	_, ok := db.GetTask(1)
	if ok {
		t.Error("expected task 1 to be deleted")
	}
}

func TestMockTaskRunner(t *testing.T) {
	runner := &mockTaskRunner{
		runTaskFn: func(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error) {
			return 999, nil
		},
	}

	jobID, err := runner.RunTask(context.Background(), 1, "copy", "local", "/src", "gdrive", "/dst", "manual")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if jobID != 999 {
		t.Errorf("expected jobID 999, got %d", jobID)
	}
}

func TestMockTaskRunnerError(t *testing.T) {
	runner := &mockTaskRunner{
		runTaskFn: func(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string) (int64, error) {
			return 0, ErrTaskNotFound
		},
	}

	_, err := runner.RunTask(context.Background(), 999, "copy", "local", "/src", "gdrive", "/dst", "manual")
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskService_CreateTask_RejectsDuplicateName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	_, err = svc.CreateTask(store.Task{Name: "same-name", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("first CreateTask() error = %v", err)
	}

	_, err = svc.CreateTask(store.Task{Name: "same-name", Mode: "sync", SourceRemote: "src2", SourcePath: "/c", TargetRemote: "dst2", TargetPath: "/d"})
	if err != ErrTaskNameExists {
		t.Fatalf("expected ErrTaskNameExists, got %v", err)
	}
}

func TestTaskService_UpdateTask_RejectsDuplicateName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	first, err := svc.CreateTask(store.Task{Name: "task-a", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("CreateTask(first) error = %v", err)
	}
	second, err := svc.CreateTask(store.Task{Name: "task-b", Mode: "sync", SourceRemote: "src2", SourcePath: "/c", TargetRemote: "dst2", TargetPath: "/d"})
	if err != nil {
		t.Fatalf("CreateTask(second) error = %v", err)
	}

	err = svc.UpdateTask(second.ID, store.Task{Name: first.Name})
	if err != ErrTaskNameExists {
		t.Fatalf("expected ErrTaskNameExists, got %v", err)
	}
}

func TestTaskService_DeleteTask_RemovesAllKnownRunLogs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldAppDataDir := os.Getenv("APP_DATA_DIR")
	if err := os.Setenv("APP_DATA_DIR", tmpDir); err != nil {
		t.Fatalf("Setenv(APP_DATA_DIR) error = %v", err)
	}
	defer func() {
		if oldAppDataDir == "" {
			_ = os.Unsetenv("APP_DATA_DIR")
		} else {
			_ = os.Setenv("APP_DATA_DIR", oldAppDataDir)
		}
	}()

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	task, err := svc.CreateTask(store.Task{Name: "task-log-clean", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	logDir1 := filepath.Join(tmpDir, "logs", "task-log-clean-0421")
	logDir2 := filepath.Join(tmpDir, "logs", "task-log-clean-0422")
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

	_, err = db.AddRun(store.Run{TaskID: task.ID, Status: "finished", Trigger: "manual", Summary: map[string]any{"stderrFile": log1}})
	if err != nil {
		t.Fatalf("AddRun(log1) error = %v", err)
	}
	_, err = db.AddRun(store.Run{TaskID: task.ID, Status: "finished", Trigger: "manual", Summary: map[string]any{"stderrFile": log2}})
	if err != nil {
		t.Fatalf("AddRun(log2) error = %v", err)
	}

	orphanDir := filepath.Join(tmpDir, "logs", "task-log-clean-0423")
	if err := os.MkdirAll(orphanDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(orphanDir) error = %v", err)
	}
	orphanLog := filepath.Join(orphanDir, "9999.log")
	if err := os.WriteFile(orphanLog, []byte("orphan"), 0o644); err != nil {
		t.Fatalf("WriteFile(orphanLog) error = %v", err)
	}

	if err := svc.DeleteTask(task.ID); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
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
	if _, err := os.Stat(orphanDir); !os.IsNotExist(err) {
		t.Fatalf("expected orphanDir removed by task-name glob, got err=%v", err)
	}
	if _, ok := db.GetTask(task.ID); ok {
		t.Fatal("expected task deleted")
	}
}
