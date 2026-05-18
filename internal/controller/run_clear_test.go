package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

func TestHandleRuns_DeleteAllRuns_ClearsDatabase(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-clear-runs-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}

	// Create tasks first (runs have FK to tasks)
	task, err := db.AddTask(store.Task{
		Name:         "test-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/src",
		TargetRemote: "dst",
		TargetPath:   "/dst",
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	// Insert test runs
	for i := 1; i <= 5; i++ {
		_, err := db.AddRun(store.Run{
			TaskID:   task.ID,
			Status:   "finished",
			Trigger:  "manual",
			Summary:  map[string]any{"test": true},
			TaskName: "test-task",
		})
		if err != nil {
			t.Fatalf("AddRun() error = %v", err)
		}
	}

	// Verify runs were inserted
	runs, total, err := db.ListRuns(1, 100)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if total != 5 {
		t.Fatalf("expected 5 runs before delete, got %d", total)
	}
	if len(runs) != 5 {
		t.Fatalf("expected 5 runs in list, got %d", len(runs))
	}

	// Create controller with real service
	runSvc := service.NewRunService(service.NewStoreRunAdapter(db))
	ctrl := NewRunController(runSvc, nil)

	// Call DELETE /api/runs
	req := httptest.NewRequest(http.MethodDelete, "/api/runs", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRuns(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200, body=%s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["deleted"] != true {
		t.Fatalf("expected deleted=true, got %#v", resp["deleted"])
	}

	// Verify database is actually empty
	runsAfter, totalAfter, err := db.ListRuns(1, 100)
	if err != nil {
		t.Fatalf("ListRuns() after delete error = %v", err)
	}
	if totalAfter != 0 {
		t.Fatalf("expected 0 runs after delete, got %d", totalAfter)
	}
	if len(runsAfter) != 0 {
		t.Fatalf("expected empty list after delete, got %d items", len(runsAfter))
	}
}

func TestHandleRuns_DeleteAllRuns_EmptyDatabase(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-clear-runs-empty-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}

	runSvc := service.NewRunService(service.NewStoreRunAdapter(db))
	ctrl := NewRunController(runSvc, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/runs", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRuns(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200, body=%s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["deleted"] != true {
		t.Fatalf("expected deleted=true, got %#v", resp["deleted"])
	}
}

func TestHandleRuns_DeleteAllRuns_DoesNotAffectTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-clear-runs-tasks-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}

	// Insert a task
	task, err := db.AddTask(store.Task{
		Name:         "test-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/src",
		TargetRemote: "dst",
		TargetPath:   "/dst",
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	// Insert a run for that task
	_, err = db.AddRun(store.Run{
		TaskID:   task.ID,
		Status:   "finished",
		Trigger:  "manual",
		Summary:  map[string]any{"test": true},
		TaskName: "test-task",
	})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}

	runSvc := service.NewRunService(service.NewStoreRunAdapter(db))
	ctrl := NewRunController(runSvc, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/runs", nil)
	w := httptest.NewRecorder()
	ctrl.HandleRuns(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", w.Code)
	}

	// Verify task still exists
	tasks, err := db.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task after clearing runs, got %d", len(tasks))
	}
	if tasks[0].Name != "test-task" {
		t.Fatalf("expected task name 'test-task', got %s", tasks[0].Name)
	}
}
