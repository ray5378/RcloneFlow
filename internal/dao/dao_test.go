package dao

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"rcloneflow/internal/store"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	tmpDir, err := os.MkdirTemp("", "dao_test_*")
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

func TestTaskDAO_CreateAndGetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	dao := NewTaskDAO(db)
	task, err := dao.Create(store.Task{
		Name:         "test-task",
		Mode:         "copy",
		SourceRemote: "local",
		SourcePath:   "/src",
		TargetRemote: "gdrive",
		TargetPath:   "/dst",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if task.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := dao.GetByID(task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "test-task" {
		t.Errorf("expected task name test-task, got %s", got.Name)
	}
}

func TestScheduleDAO_CreateAndDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	dao := NewScheduleDAO(db)
	s, err := dao.Create(store.Schedule{TaskID: 1, Spec: "@every 5m", Enabled: true})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if s.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	all, err := dao.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(all))
	}

	if err := dao.Delete(s.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestRunDAO_CreateGetAllAndUpdateStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	dao := NewRunDAO(db)
	run, err := dao.Create(store.Run{TaskID: 1, Status: "running", Trigger: "manual"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if run.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	all, err := dao.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 run, got %d", len(all))
	}

	if err := dao.UpdateStatus(run.ID, "finished", "", map[string]any{"files": 10}); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, err := dao.GetByID(run.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Status != "finished" {
		t.Errorf("expected status finished, got %s", got.Status)
	}
}
