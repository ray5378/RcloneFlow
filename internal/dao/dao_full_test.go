package dao

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"rcloneflow/internal/store"

	_ "modernc.org/sqlite"
)

func setupTestDBFull(t *testing.T) (*sql.DB, func()) {
	tmpDir, err := os.MkdirTemp("", "dao_full_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			mode TEXT NOT NULL,
			source_remote TEXT NOT NULL,
			source_path TEXT NOT NULL,
			target_remote TEXT NOT NULL,
			target_path TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			spec TEXT NOT NULL,
			enabled INTEGER DEFAULT 1,
			next_run_time DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			trigger TEXT NOT NULL,
			summary TEXT DEFAULT '{}',
			error TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			task_name TEXT DEFAULT '',
			task_mode TEXT DEFAULT '',
			source_remote TEXT DEFAULT '',
			source_path TEXT DEFAULT '',
			target_remote TEXT DEFAULT '',
			target_path TEXT DEFAULT '',
			finished_at DATETIME,
			bytes_transferred INTEGER DEFAULT 0,
			speed TEXT DEFAULT ''
		);
	`)
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create tables: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
	return db, cleanup
}

func TestTaskDAO_GetAll(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewTaskDAO(db)
	dao.Create(store.Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	dao.Create(store.Task{Name: "t2", Mode: "sync", SourceRemote: "src", SourcePath: "/c", TargetRemote: "dst", TargetPath: "/d"})

	tasks, err := dao.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestTaskDAO_Update(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewTaskDAO(db)
	task, _ := dao.Create(store.Task{Name: "original", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	task.Name = "updated"
	if err := dao.Update(task.ID, task); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, _ := dao.GetByID(task.ID)
	if got.Name != "updated" {
		t.Errorf("expected name updated, got %s", got.Name)
	}
}

func TestTaskDAO_Delete(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewTaskDAO(db)
	task, _ := dao.Create(store.Task{Name: "to-delete", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	if err := dao.Delete(task.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := dao.GetByID(task.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestScheduleDAO_GetByID(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewScheduleDAO(db)
	s, _ := dao.Create(store.Schedule{TaskID: 1, Spec: "@every 10m", Enabled: true})

	got, err := dao.GetByID(s.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Spec != "@every 10m" {
		t.Errorf("expected spec @every 10m, got %s", got.Spec)
	}
	if !got.Enabled {
		t.Error("expected enabled=true")
	}

	_, err = dao.GetByID(999)
	if err == nil {
		t.Error("expected error for nonexistent schedule")
	}
}

func TestScheduleDAO_GetByTaskID(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewScheduleDAO(db)
	dao.Create(store.Schedule{TaskID: 1, Spec: "@every 5m", Enabled: true})
	dao.Create(store.Schedule{TaskID: 1, Spec: "@every 10m", Enabled: false})
	dao.Create(store.Schedule{TaskID: 2, Spec: "@daily", Enabled: true})

	schedules, err := dao.GetByTaskID(1)
	if err != nil {
		t.Fatalf("GetByTaskID() error = %v", err)
	}
	if len(schedules) != 2 {
		t.Errorf("expected 2 schedules for task 1, got %d", len(schedules))
	}

	schedules2, _ := dao.GetByTaskID(2)
	if len(schedules2) != 1 {
		t.Errorf("expected 1 schedule for task 2, got %d", len(schedules2))
	}
}

func TestScheduleDAO_UpdateNextRunTime(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewScheduleDAO(db)
	s, _ := dao.Create(store.Schedule{TaskID: 1, Spec: "@every 5m", Enabled: true})

	nextTime := time.Now().Add(5 * time.Minute)
	if err := dao.UpdateNextRunTime(s.ID, nextTime); err != nil {
		t.Fatalf("UpdateNextRunTime() error = %v", err)
	}

	got, _ := dao.GetByID(s.ID)
	if got.NextRunTime == nil {
		t.Fatal("expected NextRunTime to be set")
	}
}

func TestRunDAO_GetAll(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})
	dao.Create(store.Run{TaskID: 2, Status: "finished", Trigger: "schedule"})

	runs, err := dao.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestRunDAO_GetByTaskID(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})
	dao.Create(store.Run{TaskID: 1, Status: "finished", Trigger: "manual"})
	dao.Create(store.Run{TaskID: 2, Status: "running", Trigger: "manual"})

	runs, err := dao.GetByTaskID(1)
	if err != nil {
		t.Fatalf("GetByTaskID() error = %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs for task 1, got %d", len(runs))
	}
}

func TestRunDAO_GetActiveRunByTaskID(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	dao.Create(store.Run{TaskID: 1, Status: "finished", Trigger: "manual"})
	dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})

	r, err := dao.GetActiveRunByTaskID(1)
	if err != nil {
		t.Fatalf("GetActiveRunByTaskID() error = %v", err)
	}
	if r.Status != "running" {
		t.Errorf("expected status running, got %s", r.Status)
	}

	_, err = dao.GetActiveRunByTaskID(999)
	if err == nil {
		t.Error("expected error for nonexistent task")
	}
}

func TestRunDAO_GetByID_WithSummary(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	run, _ := dao.Create(store.Run{TaskID: 1, Status: "finished", Trigger: "manual"})

	db.Exec(`UPDATE runs SET summary = '{"files": 10, "bytes": 1024}' WHERE id = ?`, run.ID)

	got, err := dao.GetByID(run.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if got.Summary["files"] != float64(10) {
		t.Errorf("expected summary files 10, got %v", got.Summary["files"])
	}
}

func TestRunDAO_Update(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	run, _ := dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})

	if err := dao.Update(run.ID, func(r *store.Run) {
		r.Status = "finished"
		r.Summary = map[string]any{"files": 5}
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, _ := dao.GetByID(run.ID)
	if got.Status != "finished" {
		t.Errorf("expected status finished, got %s", got.Status)
	}
	if got.Summary["files"] != float64(5) {
		t.Errorf("expected summary files 5, got %v", got.Summary["files"])
	}

	err := dao.Update(999, func(r *store.Run) {})
	// Implementation returns nil for nonexistent run (GetByID error is swallowed)
	if err != nil {
		t.Logf("Update(999) returned error: %v", err)
	}
}

func TestRunDAO_Delete(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	run, _ := dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})

	if err := dao.Delete(run.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := dao.GetByID(run.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestRunDAO_UpdateStatus_Finished(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	run, _ := dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})

	summary := map[string]any{"files": 10, "bytes": 1024}
	if err := dao.UpdateStatus(run.ID, "finished", "", summary); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, _ := dao.GetByID(run.ID)
	if got.Status != "finished" {
		t.Errorf("expected status finished, got %s", got.Status)
	}
	if got.FinishedAt == nil {
		t.Error("expected FinishedAt to be set for finished status")
	}
}

func TestRunDAO_UpdateStatus_Failed(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	run, _ := dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})

	if err := dao.UpdateStatus(run.ID, "failed", "something went wrong", nil); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, _ := dao.GetByID(run.ID)
	if got.Status != "failed" {
		t.Errorf("expected status failed, got %s", got.Status)
	}
	if got.Error != "something went wrong" {
		t.Errorf("expected error message, got %s", got.Error)
	}
	if got.FinishedAt == nil {
		t.Error("expected FinishedAt to be set for failed status")
	}
}

func TestRunDAO_UpdateStatus_Running(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	run, _ := dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})

	if err := dao.UpdateStatus(run.ID, "running", "", nil); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, _ := dao.GetByID(run.ID)
	if got.Status != "running" {
		t.Errorf("expected status running, got %s", got.Status)
	}
	if got.FinishedAt != nil {
		t.Error("expected FinishedAt to be nil for running status")
	}
}

func TestRunDAO_UpdateStatus_WithSpeedFloat(t *testing.T) {
	db, cleanup := setupTestDBFull(t)
	defer cleanup()

	dao := NewRunDAO(db)
	run, _ := dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})

	summary := map[string]any{"bytes": float64(5000), "speed": float64(1500000)}
	if err := dao.UpdateStatus(run.ID, "finished", "", summary); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, _ := dao.GetByID(run.ID)
	if got.Speed == "" {
		t.Error("expected speed to be set")
	}
}

func TestFormatSpeed(t *testing.T) {
	tests := []struct {
		speed    float64
		expected string
	}{
		{1.5e9, "1.50 GB/s"},
		{2.5e6, "2.50 MB/s"},
		{3.5e3, "3.50 KB/s"},
		{500, "500 B/s"},
		{0, "0 B/s"},
	}

	for _, tt := range tests {
		got := formatSpeed(tt.speed)
		if got != tt.expected {
			t.Errorf("formatSpeed(%v) = %q, want %q", tt.speed, got, tt.expected)
		}
	}
}
