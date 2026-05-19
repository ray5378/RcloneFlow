package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"rcloneflow/internal/store"
)

func TestRunService_UpdateRunStatus_MergeSummary(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	run, _ := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual", Summary: map[string]any{"existing": "value"}})

	svc := NewRunService(NewStoreRunAdapter(db))
	svc.UpdateRunStatus(run.ID, map[string]any{
		"finished": true,
		"success":  true,
		"bytes":    1024,
	})

	adapter := NewStoreRunAdapter(db)
	updated, _ := adapter.GetRun(run.ID)
	if updated.Status != "finished" {
		t.Errorf("expected status finished, got %s", updated.Status)
	}
	if updated.Summary == "" {
		t.Fatal("expected non-empty summary")
	}
	var summary map[string]any
	json.Unmarshal([]byte(updated.Summary), &summary)
	if summary["existing"] != "value" {
		t.Errorf("expected existing key preserved, got %v", summary["existing"])
	}
}

func TestRunService_UpdateRunStatus_Failed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	run, _ := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	svc.UpdateRunStatus(run.ID, map[string]any{
		"finished": true,
		"success":  false,
		"error":    "connection timeout",
	})

	adapter := NewStoreRunAdapter(db)
	updated, _ := adapter.GetRun(run.ID)
	if updated.Status != "failed" {
		t.Errorf("expected status failed, got %s", updated.Status)
	}
}

func TestRunService_UpdateRunStatus_NotFinished(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	run, _ := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	svc.UpdateRunStatus(run.ID, map[string]any{
		"bytes": 500,
	})

	adapter := NewStoreRunAdapter(db)
	updated, _ := adapter.GetRun(run.ID)
	if updated.Status != "running" {
		t.Errorf("expected status running, got %s", updated.Status)
	}
}

func TestRunService_CleanOldRuns_ZeroDays(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewRunService(NewStoreRunAdapter(db))
	deleted, err := svc.CleanOldRuns(0)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

func TestRunService_CleanOldRuns_NegativeDays(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewRunService(NewStoreRunAdapter(db))
	deleted, err := svc.CleanOldRuns(-5)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted for negative days, got %d", deleted)
	}
}

func TestDeepMerge(t *testing.T) {
	a := map[string]any{"a": 1, "b": map[string]any{"c": 2}}
	b := map[string]any{"b": map[string]any{"d": 3}, "e": 4}
	result := deepMerge(a, b)

	if result["a"] != 1 {
		t.Errorf("expected a=1, got %v", result["a"])
	}
	if result["e"] != 4 {
		t.Errorf("expected e=4, got %v", result["e"])
	}
	nested := result["b"].(map[string]any)
	if nested["c"] != 2 {
		t.Errorf("expected b.c=2, got %v", nested["c"])
	}
	if nested["d"] != 3 {
		t.Errorf("expected b.d=3, got %v", nested["d"])
	}
}

func TestDeepMerge_NilA(t *testing.T) {
	b := map[string]any{"x": 1}
	result := deepMerge(nil, b)
	if result["x"] != 1 {
		t.Errorf("expected x=1, got %v", result["x"])
	}
}

func TestTaskService_UpdateTask_Full(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	task, _ := svc.CreateTask(store.Task{Name: "original", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	update := store.Task{Name: "updated-name"}
	if err := svc.UpdateTask(task.ID, update); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	got, ok := db.GetTask(task.ID)
	if !ok {
		t.Fatal("expected task to exist")
	}
	if got.Name != "updated-name" {
		t.Errorf("expected name updated-name, got %s", got.Name)
	}
	if got.Mode != "copy" {
		t.Errorf("expected mode preserved as copy, got %s", got.Mode)
	}
}

func TestTaskService_UpdateTask_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	err = svc.UpdateTask(999, store.Task{Name: "test"})
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskService_UpdateTaskOptions_Merge(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	task, _ := svc.CreateTask(store.Task{
		Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b",
		Options: json.RawMessage(`{"bwLimit": "10M"}`),
	})

	opts := map[string]any{"transfers": 4}
	if err := svc.UpdateTaskOptions(task.ID, opts); err != nil {
		t.Fatalf("UpdateTaskOptions() error = %v", err)
	}

	got, _ := db.GetTask(task.ID)
	var merged map[string]any
	json.Unmarshal(got.Options, &merged)
	if merged["bwLimit"] != "10M" {
		t.Errorf("expected bwLimit preserved, got %v", merged["bwLimit"])
	}
	if merged["transfers"] != float64(4) {
		t.Errorf("expected transfers=4, got %v", merged["transfers"])
	}
}

func TestTaskService_UpdateTaskOptions_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	err = svc.UpdateTaskOptions(999, map[string]any{"key": "value"})
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskService_ExportTasks_All(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	task, _ := svc.CreateTask(store.Task{Name: "export-task", Mode: "sync", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddSchedule(store.Schedule{TaskID: task.ID, Spec: "@daily", Enabled: true})

	result, err := svc.ExportTasks()
	if err != nil {
		t.Fatalf("ExportTasks() error = %v", err)
	}

	tasks := result["tasks"].([]map[string]any)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0]["name"] != "export-task" {
		t.Errorf("expected name export-task, got %v", tasks[0]["name"])
	}

	schedules := result["schedules"].([]map[string]any)
	if len(schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(schedules))
	}
	if schedules[0]["taskName"] != "export-task" {
		t.Errorf("expected taskName export-task, got %v", schedules[0]["taskName"])
	}
}

func TestTaskService_ImportTasks_Create(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	data := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "imported-task",
				"mode":         "copy",
				"sourceRemote": "src",
				"sourcePath":   "/a",
				"targetRemote": "dst",
				"targetPath":   "/b",
			},
		},
	}

	imported, skipped, overwritten, err := svc.ImportTasks(data, "")
	if err != nil {
		t.Fatalf("ImportTasks() error = %v", err)
	}
	if imported != 1 {
		t.Errorf("expected 1 imported, got %d", imported)
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}
	if overwritten != 0 {
		t.Errorf("expected 0 overwritten, got %d", overwritten)
	}

	tasks, _ := db.ListTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task after import, got %d", len(tasks))
	}
}

func TestTaskService_ImportTasks_SkipDuplicate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	svc.CreateTask(store.Task{Name: "existing", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	data := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "Existing",
				"mode":         "copy",
				"sourceRemote": "src",
				"sourcePath":   "/a",
				"targetRemote": "dst",
				"targetPath":   "/b",
			},
		},
	}

	imported, skipped, _, err := svc.ImportTasks(data, "")
	if err != nil {
		t.Fatalf("ImportTasks() error = %v", err)
	}
	if imported != 0 {
		t.Errorf("expected 0 imported, got %d", imported)
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}
}

func TestTaskService_ImportTasks_Overwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	svc.CreateTask(store.Task{Name: "existing", Mode: "copy", SourceRemote: "src", SourcePath: "/old", TargetRemote: "dst", TargetPath: "/old"})

	data := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "existing",
				"mode":         "sync",
				"sourceRemote": "new-src",
				"targetRemote": "new-dst",
			},
		},
	}

	_, _, overwritten, err := svc.ImportTasks(data, "overwrite")
	if err != nil {
		t.Fatalf("ImportTasks() error = %v", err)
	}
	if overwritten != 1 {
		t.Errorf("expected 1 overwritten, got %d", overwritten)
	}

	tasks, _ := db.ListTasks()
	if tasks[0].Mode != "sync" {
		t.Errorf("expected mode sync after overwrite, got %s", tasks[0].Mode)
	}
}

func TestTaskService_ImportTasks_InvalidData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	data := map[string]any{
		"tasks": "not-an-array",
	}

	_, _, _, err = svc.ImportTasks(data, "")
	if err == nil {
		t.Error("expected error for invalid tasks data")
	}
}

func TestTaskService_ImportTasks_SkipInvalidTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	data := map[string]any{
		"tasks": []any{
			map[string]any{"name": ""},
			map[string]any{"name": "no-mode", "sourceRemote": "src", "targetRemote": "dst"},
			map[string]any{"name": "valid", "mode": "copy", "sourceRemote": "src", "targetRemote": "dst"},
		},
	}

	imported, skipped, _, err := svc.ImportTasks(data, "")
	if err != nil {
		t.Fatalf("ImportTasks() error = %v", err)
	}
	if imported != 1 {
		t.Errorf("expected 1 imported, got %d", imported)
	}
	if skipped != 2 {
		t.Errorf("expected 2 skipped, got %d", skipped)
	}
}

func TestTaskService_ClearAllTasks_All(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	svc.CreateTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	svc.CreateTask(store.Task{Name: "t2", Mode: "sync", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	if err := svc.ClearAllTasks(); err != nil {
		t.Fatalf("ClearAllTasks() error = %v", err)
	}

	tasks, _ := db.ListTasks()
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after clear, got %d", len(tasks))
	}
}

func TestTaskService_CreateTask_DuplicateName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	svc.CreateTask(store.Task{Name: "unique-name", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	_, err = svc.CreateTask(store.Task{Name: "Unique-Name", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != ErrTaskNameExists {
		t.Errorf("expected ErrTaskNameExists, got %v", err)
	}
}

func TestTaskService_EnsureTaskNameUnique_EmptyName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	svc.CreateTask(store.Task{Name: "existing", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	err = svc.ensureTaskNameUnique("", 0)
	if err != nil {
		t.Errorf("expected no error for empty name, got %v", err)
	}
}

func TestCleanupRunLog_NoSummary(t *testing.T) {
	run := RunRecord{ID: 1, Summary: ""}
	cleanupRunLog(run)
}

func TestCleanupRunLog_InvalidSummary(t *testing.T) {
	run := RunRecord{ID: 1, Summary: "not-json"}
	cleanupRunLog(run)
}

func TestCleanupRunLog_NoStderrFile(t *testing.T) {
	summary, _ := json.Marshal(map[string]any{"files": 10})
	run := RunRecord{ID: 1, Summary: string(summary)}
	cleanupRunLog(run)
}

func TestCleanupRunLog_RemovesFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_cleanup_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logDir := filepath.Join(tmpDir, "logs")
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "test.log")
	os.WriteFile(logPath, []byte("log"), 0644)

	summary, _ := json.Marshal(map[string]any{"stderrFile": logPath})
	run := RunRecord{ID: 1, Summary: string(summary)}
	cleanupRunLog(run)

	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("expected log file to be removed")
	}
	if _, err := os.Stat(logDir); !os.IsNotExist(err) {
		t.Error("expected empty log dir to be removed")
	}
}

func TestRunService_DeleteRun_WithLogCleanup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logDir := filepath.Join(tmpDir, "logs", "task-1")
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "run.log")
	os.WriteFile(logPath, []byte("log"), 0644)

	summaryBytes, _ := json.Marshal(map[string]any{"stderrFile": logPath})
	mock := &runServiceDBMock{
		runsByID: map[int64]RunRecord{1: {ID: 1, TaskID: 1, Summary: string(summaryBytes)}},
	}

	svc := NewRunService(mock)
	if err := svc.DeleteRun(1); err != nil {
		t.Fatalf("DeleteRun() error = %v", err)
	}
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("expected log file to be removed")
	}
}

func TestRunService_DeleteAllRuns_Error(t *testing.T) {
	mock := &runServiceDBMock{
		listRunsPages: map[int][]RunRecord{1: {{ID: 1}}},
		listRunsTotal: 1,
	}
	mock.listRunsPages[2] = nil

	svc := NewRunService(mock)
	err := svc.DeleteAllRuns()
	if err != nil {
		t.Fatalf("DeleteAllRuns() error = %v", err)
	}
}

func TestTaskService_ImportTasks_WithSchedules(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	data := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "sched-task",
				"mode":         "copy",
				"sourceRemote": "src",
				"sourcePath":   "/a",
				"targetRemote": "dst",
				"targetPath":   "/b",
			},
		},
		"schedules": []any{
			map[string]any{
				"taskName": "sched-task",
				"spec":     "@daily",
				"enabled":  true,
			},
		},
	}

	_, _, _, err = svc.ImportTasks(data, "")
	if err != nil {
		t.Fatalf("ImportTasks() error = %v", err)
	}

	schedules, _ := db.ListSchedules()
	if len(schedules) != 1 {
		t.Errorf("expected 1 schedule after import, got %d", len(schedules))
	}
	if schedules[0].Spec != "@daily" {
		t.Errorf("expected spec @daily, got %s", schedules[0].Spec)
	}
}

func TestTaskService_ImportTasks_ScheduleMissingTask(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	data := map[string]any{
		"tasks": []any{},
		"schedules": []any{
			map[string]any{
				"taskName": "nonexistent",
				"spec":     "@daily",
				"enabled":  true,
			},
		},
	}

	_, _, _, err = svc.ImportTasks(data, "")
	if err != nil {
		t.Fatalf("ImportTasks() error = %v", err)
	}

	schedules, _ := db.ListSchedules()
	if len(schedules) != 0 {
		t.Errorf("expected 0 schedules for missing task, got %d", len(schedules))
	}
}

func TestTaskService_ImportTasks_ScheduleDuplicate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	task, _ := svc.CreateTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddSchedule(store.Schedule{TaskID: task.ID, Spec: "@daily", Enabled: true})

	data := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "t1",
				"mode":         "copy",
				"sourceRemote": "src",
				"sourcePath":   "/a",
				"targetRemote": "dst",
				"targetPath":   "/b",
			},
		},
		"schedules": []any{
			map[string]any{
				"taskName": "t1",
				"spec":     "@daily",
				"enabled":  true,
			},
		},
	}

	_, _, _, err = svc.ImportTasks(data, "")
	if err != nil {
		t.Fatalf("ImportTasks() error = %v", err)
	}

	schedules, _ := db.ListSchedules()
	if len(schedules) != 1 {
		t.Errorf("expected 1 schedule (no duplicate), got %d", len(schedules))
	}
}

func TestTaskService_DeleteTask_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	err = svc.DeleteTask(999)
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestRunService_ListRuns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	runs, total, err := svc.ListRuns(1, 50)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 total, got %d", total)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}
}

func TestRunService_GetRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	run, _ := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual"})

	adapter := NewStoreRunAdapter(db)
	got, err := adapter.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Status != "running" {
		t.Errorf("expected status running, got %s", got.Status)
	}
}

func TestRunService_ListActiveRuns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual"})
	db.AddRun(store.Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	runs, err := svc.ListActiveRuns()
	if err != nil {
		t.Fatalf("ListActiveRuns() error = %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 active run, got %d", len(runs))
	}
}

func TestRunService_GetActiveRunByTaskID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	got, err := svc.GetActiveRunByTaskID(task.ID)
	if err != nil {
		t.Fatalf("GetActiveRunByTaskID() error = %v", err)
	}
	if got.Status != "running" {
		t.Errorf("expected status running, got %s", got.Status)
	}
}

func TestRunService_ListRunsByTask(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	t1, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	t2, _ := db.AddTask(store.Task{Name: "t2", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(store.Run{TaskID: t1.ID, Status: "running", Trigger: "manual"})
	db.AddRun(store.Run{TaskID: t1.ID, Status: "finished", Trigger: "manual"})
	db.AddRun(store.Run{TaskID: t2.ID, Status: "running", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	runs, err := svc.ListRunsByTask(t1.ID)
	if err != nil {
		t.Fatalf("ListRunsByTask() error = %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs for task1, got %d", len(runs))
	}
}

func TestTaskService_UpdateTask_FullUpdate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	task, _ := svc.CreateTask(store.Task{Name: "original", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	full := store.Task{
		Name:         "full-update",
		Mode:         "sync",
		SourceRemote: "new-src",
		SourcePath:   "/new-src",
		TargetRemote: "new-dst",
		TargetPath:   "/new-dst",
	}
	if err := svc.UpdateTask(task.ID, full); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	got, _ := db.GetTask(task.ID)
	if got.Name != "full-update" {
		t.Errorf("expected name full-update, got %s", got.Name)
	}
	if got.Mode != "sync" {
		t.Errorf("expected mode sync, got %s", got.Mode)
	}
	if got.SourceRemote != "new-src" {
		t.Errorf("expected sourceRemote new-src, got %s", got.SourceRemote)
	}
}

func TestTaskService_UpdateTask_DuplicateName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	svc.CreateTask(store.Task{Name: "task-a", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	task2, _ := svc.CreateTask(store.Task{Name: "task-b", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	err = svc.UpdateTask(task2.ID, store.Task{Name: "Task-A"})
	if err != ErrTaskNameExists {
		t.Errorf("expected ErrTaskNameExists, got %v", err)
	}
}

func TestTaskService_ExportTasks_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	result, err := svc.ExportTasks()
	if err != nil {
		t.Fatalf("ExportTasks() error = %v", err)
	}
	if result["version"] != 1 {
		t.Errorf("expected version 1, got %v", result["version"])
	}
	if _, ok := result["exportedAt"].(string); !ok {
		t.Error("expected exportedAt to be a string")
	}
}

func TestTaskService_ImportTasks_WithTaskOptions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	opts := map[string]any{"bwLimit": "10M"}
	data := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "opt-task",
				"mode":         "copy",
				"sourceRemote": "src",
				"sourcePath":   "/a",
				"targetRemote": "dst",
				"targetPath":   "/b",
				"options":      opts,
			},
		},
	}

	_, _, _, err = svc.ImportTasks(data, "")
	if err != nil {
		t.Fatalf("ImportTasks() error = %v", err)
	}

	tasks, _ := db.ListTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	var gotOpts map[string]any
	json.Unmarshal(tasks[0].Options, &gotOpts)
	if gotOpts["bwLimit"] != "10M" {
		t.Errorf("expected bwLimit 10M, got %v", gotOpts["bwLimit"])
	}
}

func TestTaskService_ImportTasks_ScheduleInvalidData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	data := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "t1",
				"mode":         "copy",
				"sourceRemote": "src",
				"sourcePath":   "/a",
				"targetRemote": "dst",
				"targetPath":   "/b",
			},
		},
		"schedules": []any{
			"not-a-map",
			map[string]any{"taskName": "", "spec": "@daily"},
			map[string]any{"taskName": "t1", "spec": ""},
		},
	}

	_, _, _, err = svc.ImportTasks(data, "")
	if err != nil {
		t.Fatalf("ImportTasks() error = %v", err)
	}

	schedules, _ := db.ListSchedules()
	if len(schedules) != 0 {
		t.Errorf("expected 0 schedules, got %d", len(schedules))
	}
}

func TestTaskService_UpdateTaskOptions_EmptyTask(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	task, _ := svc.CreateTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	if err := svc.UpdateTaskOptions(task.ID, map[string]any{"newKey": "newValue"}); err != nil {
		t.Fatalf("UpdateTaskOptions() error = %v", err)
	}

	got, _ := db.GetTask(task.ID)
	var opts map[string]any
	json.Unmarshal(got.Options, &opts)
	if opts["newKey"] != "newValue" {
		t.Errorf("expected newKey newValue, got %v", opts["newKey"])
	}
}

func TestTaskService_UpdateTaskOptions_MergeNested(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	task, _ := svc.CreateTask(store.Task{
		Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b",
		Options: json.RawMessage(`{"existing": {"nested": true}}`),
	})

	if err := svc.UpdateTaskOptions(task.ID, map[string]any{"new": "value"}); err != nil {
		t.Fatalf("UpdateTaskOptions() error = %v", err)
	}

	got, _ := db.GetTask(task.ID)
	var opts map[string]any
	json.Unmarshal(got.Options, &opts)
	if opts["new"] != "value" {
		t.Errorf("expected new=value, got %v", opts["new"])
	}
	existing := opts["existing"].(map[string]any)
	if existing["nested"] != true {
		t.Errorf("expected existing.nested=true, got %v", existing["nested"])
	}
}

func TestTaskService_GetTask_ByID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasksvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	task, _ := svc.CreateTask(store.Task{Name: "get-task", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	got, ok := svc.GetTask(task.ID)
	if !ok {
		t.Fatal("expected GetTask to return true")
	}
	if got.Name != "get-task" {
		t.Errorf("expected name get-task, got %s", got.Name)
	}

	_, ok = svc.GetTask(999)
	if ok {
		t.Error("expected GetTask to return false for nonexistent")
	}
}

func TestRunService_DeleteRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	run, _ := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	if err := svc.DeleteRun(run.ID); err != nil {
		t.Fatalf("DeleteRun() error = %v", err)
	}

	_, err = db.GetRun(run.ID)
	if err == nil {
		t.Error("expected run to be deleted")
	}
}

func TestRunService_DeleteAllRuns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual"})
	db.AddRun(store.Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	if err := svc.DeleteAllRuns(); err != nil {
		t.Fatalf("DeleteAllRuns() error = %v", err)
	}

	_, total, _ := db.ListRuns(1, 50)
	if total != 0 {
		t.Errorf("expected 0 runs after delete all, got %d", total)
	}
}

func TestRunService_DeleteRunsByTask(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	t1, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	t2, _ := db.AddTask(store.Task{Name: "t2", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(store.Run{TaskID: t1.ID, Status: "running", Trigger: "manual"})
	db.AddRun(store.Run{TaskID: t1.ID, Status: "finished", Trigger: "manual"})
	db.AddRun(store.Run{TaskID: t2.ID, Status: "running", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	if err := svc.DeleteRunsByTask(t1.ID); err != nil {
		t.Fatalf("DeleteRunsByTask() error = %v", err)
	}

	_, total, _ := db.ListRuns(1, 50)
	if total != 1 {
		t.Errorf("expected 1 run remaining, got %d", total)
	}
}

func TestRunService_CleanOldRuns_InvalidDate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(store.Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	deleted, err := svc.CleanOldRuns(5)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted for recent run, got %d", deleted)
	}
}

func TestRunService_CleanOldRuns_NoOldRuns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(store.Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	svc := NewRunService(NewStoreRunAdapter(db))
	deleted, err := svc.CleanOldRuns(30)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

func TestRunService_UpdateRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runsvc_*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	run, _ := db.AddRun(store.Run{TaskID: task.ID, Status: "running", Trigger: "manual"})

	adapter := NewStoreRunAdapter(db)
	adapter.UpdateRun(run.ID, func(r *RunRecord) {
		r.Status = "finished"
	})

	got, _ := adapter.GetRun(run.ID)
	if got.Status != "finished" {
		t.Errorf("expected status finished, got %s", got.Status)
	}
}
