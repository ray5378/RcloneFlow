package service

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"rcloneflow/internal/store"
)

func TestTaskService_UpdateTask_MergesOnlyProvidedFields(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-update-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	created, err := svc.CreateTask(store.Task{
		Name:         "task-a",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/from",
		TargetRemote: "dst",
		TargetPath:   "/to",
		Options:      json.RawMessage(`{"transfers":2}`),
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	err = svc.UpdateTask(created.ID, store.Task{
		Name:       " task-b ",
		SourcePath: "/updated-from",
	})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	updated, ok := svc.GetTask(created.ID)
	if !ok {
		t.Fatal("expected task to exist after update")
	}
	if updated.Name != " task-b " {
		t.Fatalf("expected updated name to be preserved, got %q", updated.Name)
	}
	if updated.Mode != "copy" {
		t.Fatalf("expected mode to remain copy, got %q", updated.Mode)
	}
	if updated.SourceRemote != "src" || updated.TargetRemote != "dst" || updated.TargetPath != "/to" {
		t.Fatalf("unexpected preserved task fields: %#v", updated)
	}
	if updated.SourcePath != "/updated-from" {
		t.Fatalf("expected source path update, got %q", updated.SourcePath)
	}
	if string(updated.Options) != `{"transfers":2}` {
		t.Fatalf("expected options unchanged, got %s", string(updated.Options))
	}
}

func TestTaskService_UpdateTask_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-update-notfound-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	if err := svc.UpdateTask(999, store.Task{Name: "missing"}); err != ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskService_UpdateTaskOptions_MergesAndOverrides(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-options-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	created, err := svc.CreateTask(store.Task{
		Name:         "opts-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/from",
		TargetRemote: "dst",
		TargetPath:   "/to",
		Options:      json.RawMessage(`{"transfers":2,"bufferSize":"16M"}`),
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	err = svc.UpdateTaskOptions(created.ID, map[string]any{
		"transfers":  4,
		"checkers":   8,
		"bufferSize": "32M",
	})
	if err != nil {
		t.Fatalf("UpdateTaskOptions() error = %v", err)
	}

	updated, ok := svc.GetTask(created.ID)
	if !ok {
		t.Fatal("expected task to exist after UpdateTaskOptions")
	}
	var opts map[string]any
	if err := json.Unmarshal(updated.Options, &opts); err != nil {
		t.Fatalf("json.Unmarshal(options) error = %v", err)
	}
	if opts["transfers"].(float64) != 4 {
		t.Fatalf("expected transfers overridden to 4, got %#v", opts["transfers"])
	}
	if opts["checkers"].(float64) != 8 {
		t.Fatalf("expected checkers=8, got %#v", opts["checkers"])
	}
	if opts["bufferSize"] != "32M" {
		t.Fatalf("expected bufferSize overridden to 32M, got %#v", opts["bufferSize"])
	}
}

func TestTaskService_UpdateTaskOptions_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-options-notfound-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	if err := svc.UpdateTaskOptions(999, map[string]any{"transfers": 1}); err != ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskService_RunTask_UsesRawOptionsWhenParsingFailsAndHonorsStreamingFlag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-runtask-*")
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
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "raw-options-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/from",
		TargetRemote: "dst",
		TargetPath:   "/to",
		Options:      json.RawMessage(`{"enableStreaming":false,"transfers":9,"header":"bad-shape"}`),
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	svc := NewTaskService(db, nil)
	result, err := svc.RunTask(context.Background(), task.ID, "manual")
	if err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}
	if !result.Started {
		t.Fatalf("expected started result, got %#v", result)
	}

	runs, err := db.ListRunsByTask(task.ID)
	if err != nil {
		t.Fatalf("ListRunsByTask() error = %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected run record")
	}
	if v, ok := runs[0].Summary["streamingEnabled"].(bool); !ok || v {
		t.Fatalf("expected streamingEnabled=false, got %#v", runs[0].Summary["streamingEnabled"])
	}
	eff, ok := runs[0].Summary["effectiveOptions"].(map[string]any)
	if !ok || eff == nil {
		t.Fatalf("expected effectiveOptions map, got %#v", runs[0].Summary["effectiveOptions"])
	}
	if eff["enableStreaming"] != false {
		t.Fatalf("expected effectiveOptions.enableStreaming=false, got %#v", eff["enableStreaming"])
	}
	if eff["transfers"].(float64) != 9 {
		t.Fatalf("expected raw transfers override to 9, got %#v", eff["transfers"])
	}
}

func TestTaskService_RunTask_PersistsTransferDefaultsWhenSettingsLoadSucceeds(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-settings-ok-*")
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
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "settings-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/from",
		TargetRemote: "dst",
		TargetPath:   "/to",
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	svc := NewTaskService(db, nil)
	result, err := svc.RunTask(context.Background(), task.ID, "manual")
	if err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}
	if !result.Started {
		t.Fatalf("expected started result, got %#v", result)
	}

	runs, err := db.ListRunsByTask(task.ID)
	if err != nil {
		t.Fatalf("ListRunsByTask() error = %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected run record")
	}
	defaults, ok := runs[0].Summary["transferDefaults"].(map[string]any)
	if !ok || defaults == nil {
		t.Fatalf("expected transferDefaults map, got %#v", runs[0].Summary["transferDefaults"])
	}
	if defaults["postVerifyMode"] != "mount" {
		t.Fatalf("expected default postVerifyMode=mount, got %#v", defaults["postVerifyMode"])
	}
	if defaults["postVerifyEnabled"] != true {
		t.Fatalf("expected default postVerifyEnabled=true, got %#v", defaults["postVerifyEnabled"])
	}
}

func TestTaskService_RunTask_ContinuesWhenSettingsLoadFails(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-settings-bad-*")
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
	if err := os.WriteFile(tmpDir+"/transfer_settings.json", []byte("{"), 0o644); err != nil {
		t.Fatalf("WriteFile(transfer_settings.json) error = %v", err)
	}

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "settings-bad-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/from",
		TargetRemote: "dst",
		TargetPath:   "/to",
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	svc := NewTaskService(db, nil)
	result, err := svc.RunTask(context.Background(), task.ID, "manual")
	if err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}
	if !result.Started {
		t.Fatalf("expected started result, got %#v", result)
	}

	runs, err := db.ListRunsByTask(task.ID)
	if err != nil {
		t.Fatalf("ListRunsByTask() error = %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected run record")
	}
	if _, ok := runs[0].Summary["transferDefaults"]; ok {
		t.Fatalf("expected no transferDefaults when settings load fails, got %#v", runs[0].Summary["transferDefaults"])
	}
}

func TestTaskService_RunTask_SingletonSuccessCreatesRunningRecord(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-singleton-ok-*")
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
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	task, err := db.AddTask(store.Task{
		Name:         "singleton-ok-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/from",
		TargetRemote: "dst",
		TargetPath:   "/to",
		Options:      json.RawMessage(`{"singletonMode":true}`),
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	svc := NewTaskService(db, nil)
	result, err := svc.RunTask(context.Background(), task.ID, "manual")
	if err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}
	if !result.Started {
		t.Fatalf("expected started result, got %#v", result)
	}

	runs, err := db.ListRunsByTask(task.ID)
	if err != nil {
		t.Fatalf("ListRunsByTask() error = %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run record, got %d", len(runs))
	}
	if runs[0].Status != "running" {
		t.Fatalf("expected running status, got %q", runs[0].Status)
	}
	eff, ok := runs[0].Summary["effectiveOptions"].(map[string]any)
	if !ok || eff == nil {
		t.Fatalf("expected effectiveOptions map, got %#v", runs[0].Summary["effectiveOptions"])
	}
	if eff["singletonMode"] != true {
		t.Fatalf("expected singletonMode=true, got %#v", eff["singletonMode"])
	}
}

func TestTaskService_RunTask_ReturnsErrorWhenSingletonAcquireOrAddRunFails(t *testing.T) {
	t.Run("singleton TryAcquireRun error", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-singleton-err-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		db, err := store.Open(tmpDir)
		if err != nil {
			t.Fatalf("store.Open() error = %v", err)
		}
		task, err := db.AddTask(store.Task{
			Name:         "singleton-error-task",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from",
			TargetRemote: "dst",
			TargetPath:   "/to",
			Options:      json.RawMessage(`{"singletonMode":true}`),
		})
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close() error = %v", err)
		}

		svc := NewTaskService(db, nil)
		_, err = svc.RunTask(context.Background(), task.ID, "manual")
		if err == nil {
			t.Fatal("expected RunTask() error after closed DB in singleton mode")
		}
	})

	t.Run("non-singleton AddRun error", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-addrun-err-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		db, err := store.Open(tmpDir)
		if err != nil {
			t.Fatalf("store.Open() error = %v", err)
		}
		task, err := db.AddTask(store.Task{
			Name:         "addrun-error-task",
			Mode:         "copy",
			SourceRemote: "src",
			SourcePath:   "/from",
			TargetRemote: "dst",
			TargetPath:   "/to",
		})
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close() error = %v", err)
		}

		svc := NewTaskService(db, nil)
		_, err = svc.RunTask(context.Background(), task.ID, "manual")
		if err == nil {
			t.Fatal("expected RunTask() error after closed DB in non-singleton mode")
		}
	})
}

func TestTaskService_RunTask_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-tasksvc-runtask-notfound-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	svc := NewTaskService(db, nil)
	_, err = svc.RunTask(context.Background(), 999, "manual")
	if err != ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}
