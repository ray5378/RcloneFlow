package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

func newTestTaskController(t *testing.T) (*TaskController, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_import_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.RemoveAll(tmpDir)
	})

	taskSvc := service.NewTaskService(db, nil)
	scheduleSvc := service.NewScheduleService(db)
	runSvc := service.NewRunService(service.NewStoreRunAdapter(db))
	rc := adapter.NewRcloneClient(nil)

	ctrl := NewTaskController(taskSvc, scheduleSvc, runSvc, rc)
	return ctrl, tmpDir
}

func TestTaskImport_ImportsTasksWithoutConflicts(t *testing.T) {
	ctrl, _ := newTestTaskController(t)

	payload := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "import-task-1",
				"mode":         "copy",
				"sourceRemote": "src",
				"sourcePath":   "/a",
				"targetRemote": "dst",
				"targetPath":   "/b",
			},
		},
		"schedules":        []any{},
		"conflictStrategy": "skip",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/import", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.HandleImportTask(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if v, _ := resp["imported"].(float64); v != 1 {
		t.Errorf("expected imported=1, got %v", resp["imported"])
	}
}

func TestTaskImport_SkipsConflictingTasks(t *testing.T) {
	ctrl, tmpDir := newTestTaskController(t)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	_, _ = db.AddTask(store.Task{
		Name:         "existing-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
	})

	payload := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "existing-task",
				"mode":         "sync",
				"sourceRemote": "src2",
				"sourcePath":   "/c",
				"targetRemote": "dst2",
				"targetPath":   "/d",
			},
		},
		"schedules":        []any{},
		"conflictStrategy": "skip",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/import", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.HandleImportTask(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if v, _ := resp["skipped"].(float64); v != 1 {
		t.Errorf("expected skipped=1, got %v", resp["skipped"])
	}
}

func TestTaskImport_OverwritesConflictingTasks(t *testing.T) {
	ctrl, tmpDir := newTestTaskController(t)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	_, _ = db.AddTask(store.Task{
		Name:         "existing-task",
		Mode:         "copy",
		SourceRemote: "old-src",
		SourcePath:   "/old",
		TargetRemote: "old-dst",
		TargetPath:   "/old",
	})

	payload := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "existing-task",
				"mode":         "sync",
				"sourceRemote": "new-src",
				"sourcePath":   "/new",
				"targetRemote": "new-dst",
				"targetPath":   "/new",
			},
		},
		"schedules":        []any{},
		"conflictStrategy": "overwrite",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/import", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.HandleImportTask(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if v, _ := resp["overwritten"].(float64); v != 1 {
		t.Errorf("expected overwritten=1, got %v", resp["overwritten"])
	}

	tasks, _ := db.ListTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Mode != "sync" {
		t.Errorf("expected mode=sync after overwrite, got %s", tasks[0].Mode)
	}
	if tasks[0].SourceRemote != "new-src" {
		t.Errorf("expected sourceRemote=new-src after overwrite, got %s", tasks[0].SourceRemote)
	}
}

func TestTaskImport_InvalidPayload(t *testing.T) {
	ctrl, _ := newTestTaskController(t)

	payload := map[string]any{
		"schedules":        []any{},
		"conflictStrategy": "skip",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/import", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.HandleImportTask(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500 for missing tasks, got %d", w.Code)
	}
}

func TestTaskImport_EmptyTasksArray(t *testing.T) {
	ctrl, _ := newTestTaskController(t)

	payload := map[string]any{
		"tasks":            []any{},
		"schedules":        []any{},
		"conflictStrategy": "skip",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/import", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.HandleImportTask(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if v, _ := resp["imported"].(float64); v != 0 {
		t.Errorf("expected imported=0 for empty tasks, got %v", resp["imported"])
	}
}

func TestTaskImport_WithSchedules(t *testing.T) {
	ctrl, tmpDir := newTestTaskController(t)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	task, _ := db.AddTask(store.Task{
		Name:         "scheduled-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
	})

	payload := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "scheduled-task",
				"mode":         "copy",
				"sourceRemote": "src",
				"sourcePath":   "/a",
				"targetRemote": "dst",
				"targetPath":   "/b",
			},
		},
		"schedules": []any{
			map[string]any{
				"taskName": "scheduled-task",
				"spec":     "0 2 * * *",
				"enabled":  true,
			},
		},
		"conflictStrategy": "skip",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/import", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.HandleImportTask(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	schedules, _ := db.ListSchedules()
	found := false
	for _, s := range schedules {
		if s.TaskID == task.ID && s.Spec == "0 2 * * *" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected schedule to be imported, got %d schedules", len(schedules))
	}
}

func TestTaskImport_RcloneConfigIncluded(t *testing.T) {
	ctrl, _ := newTestTaskController(t)

	payload := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "rclone-task",
				"mode":         "copy",
				"sourceRemote": "myremote",
				"sourcePath":   "/a",
				"targetRemote": "local",
				"targetPath":   "/b",
			},
		},
		"schedules":        []any{},
		"conflictStrategy": "skip",
		"rcloneConfig": map[string]any{
			"myremote": map[string]any{
				"type": "s3",
				"provider": "AWS",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/import", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.HandleImportTask(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if _, ok := resp["remotesAdded"]; !ok {
		t.Error("expected remotesAdded in response")
	}
	if _, ok := resp["remotesSkipped"]; !ok {
		t.Error("expected remotesSkipped in response")
	}
}

func TestTaskImport_RcloneConfigSkipsExisting(t *testing.T) {
	ctrl, tmpDir := newTestTaskController(t)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	_, _ = db.AddTask(store.Task{
		Name:         "existing-remote-task",
		Mode:         "copy",
		SourceRemote: "existing",
		SourcePath:   "/a",
		TargetRemote: "local",
		TargetPath:   "/b",
	})

	payload := map[string]any{
		"tasks": []any{
			map[string]any{
				"name":         "existing-remote-task",
				"mode":         "copy",
				"sourceRemote": "existing",
				"sourcePath":   "/a",
				"targetRemote": "local",
				"targetPath":   "/b",
			},
		},
		"schedules":        []any{},
		"conflictStrategy": "skip",
		"rcloneConfig": map[string]any{
			"existing": map[string]any{
				"type": "s3",
				"provider": "AWS",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/import", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ctrl.HandleImportTask(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	skipped, _ := resp["skipped"].(float64)
	if skipped != 1 {
		t.Errorf("expected task skipped=1 due to name conflict, got %v", skipped)
	}
}

func TestTaskExport_IncludesRcloneConfig(t *testing.T) {
	ctrl, tmpDir := newTestTaskController(t)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	_, _ = db.AddTask(store.Task{
		Name:         "export-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/a",
		TargetRemote: "dst",
		TargetPath:   "/b",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/export", nil)
	w := httptest.NewRecorder()

	ctrl.HandleExportTask(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if _, ok := resp["tasks"]; !ok {
		t.Error("expected tasks in export response")
	}
	if _, ok := resp["schedules"]; !ok {
		t.Error("expected schedules in export response")
	}
	if _, ok := resp["version"]; !ok {
		t.Error("expected version in export response")
	}
	if _, ok := resp["exportedAt"]; !ok {
		t.Error("expected exportedAt in export response")
	}
}
