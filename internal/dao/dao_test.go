package dao

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	tmpDir, err := os.MkdirTemp("", "dao_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open db: %v", err)
	}
	
	// 创建表
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
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			rc_job_id INTEGER DEFAULT 0,
			status TEXT NOT NULL,
			trigger TEXT NOT NULL,
			summary TEXT DEFAULT '{}',
			error TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			finished_at DATETIME
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

// TestTaskDAO_Create 测试创建任务
func TestTaskDAO_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewTaskDAO(db)
	
	task := Task{
		Name:         "test-task",
		Mode:         "copy",
		SourceRemote: "local",
		SourcePath:   "/src",
		TargetRemote: "gdrive",
		TargetPath:   "/dst",
	}
	
	created, err := dao.Create(task)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	
	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}
	
	if created.Name != "test-task" {
		t.Errorf("expected Name 'test-task', got '%s'", created.Name)
	}
}

// TestTaskDAO_GetByID 测试获取任务
func TestTaskDAO_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewTaskDAO(db)
	
	// 先创建
	dao.Create(Task{
		Name:         "test",
		Mode:         "copy",
		SourceRemote: "local",
		SourcePath:   "/src",
		TargetRemote: "gdrive",
		TargetPath:   "/dst",
	})
	
	// 获取
	task, ok := dao.GetByID(1)
	if !ok {
		t.Fatal("expected to get task")
	}
	
	if task.Name != "test" {
		t.Errorf("expected Name 'test', got '%s'", task.Name)
	}
}

// TestTaskDAO_GetByID_NotFound 测试获取不存在的任务
func TestTaskDAO_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewTaskDAO(db)
	
	_, ok := dao.GetByID(999)
	if ok {
		t.Error("expected not found")
	}
}

// TestTaskDAO_GetAll 测试获取所有任务
func TestTaskDAO_GetAll(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewTaskDAO(db)
	
	dao.Create(Task{Name: "task1", Mode: "copy", SourceRemote: "a", SourcePath: "/a", TargetRemote: "b", TargetPath: "/b"})
	dao.Create(Task{Name: "task2", Mode: "sync", SourceRemote: "c", SourcePath: "/c", TargetRemote: "d", TargetPath: "/d"})
	
	tasks, err := dao.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

// TestTaskDAO_Update 测试更新任务
func TestTaskDAO_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewTaskDAO(db)
	
	dao.Create(Task{Name: "original", Mode: "copy", SourceRemote: "a", SourcePath: "/a", TargetRemote: "b", TargetPath: "/b"})
	
	err := dao.Update(1, Task{
		Name:         "updated",
		Mode:         "sync",
		SourceRemote: "c",
		SourcePath:   "/c",
		TargetRemote: "d",
		TargetPath:   "/d",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	
	task, _ := dao.GetByID(1)
	if task.Name != "updated" {
		t.Errorf("expected Name 'updated', got '%s'", task.Name)
	}
	if task.Mode != "sync" {
		t.Errorf("expected Mode 'sync', got '%s'", task.Mode)
	}
}

// TestTaskDAO_Delete 测试删除任务
func TestTaskDAO_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewTaskDAO(db)
	
	dao.Create(Task{Name: "task1", Mode: "copy", SourceRemote: "a", SourcePath: "/a", TargetRemote: "b", TargetPath: "/b"})
	dao.Create(Task{Name: "task2", Mode: "copy", SourceRemote: "c", SourcePath: "/c", TargetRemote: "d", TargetPath: "/d"})
	
	err := dao.Delete(1)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	
	tasks, _ := dao.GetAll()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

// TestScheduleDAO_Create 测试创建定时任务
func TestScheduleDAO_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewScheduleDAO(db)
	
	schedule, err := dao.Create(Schedule{TaskID: 1, Spec: "@every 5m"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	
	if schedule.ID == 0 {
		t.Error("expected non-zero ID")
	}
	
	if schedule.Spec != "@every 5m" {
		t.Errorf("expected Spec '@every 5m', got '%s'", schedule.Spec)
	}
}

// TestScheduleDAO_GetByID 测试获取定时任务
func TestScheduleDAO_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewScheduleDAO(db)
	dao.Create(Schedule{TaskID: 1, Spec: "@every 10m"})
	
	schedule, ok := dao.GetByID(1)
	if !ok {
		t.Fatal("expected to get schedule")
	}
	
	if schedule.Spec != "@every 10m" {
		t.Errorf("expected Spec '@every 10m', got '%s'", schedule.Spec)
	}
}

// TestScheduleDAO_GetAll 测试获取所有定时任务
func TestScheduleDAO_GetAll(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewScheduleDAO(db)
	
	dao.Create(Schedule{TaskID: 1, Spec: "@every 5m"})
	dao.Create(Schedule{TaskID: 2, Spec: "@every 10m"})
	
	schedules, err := dao.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	
	if len(schedules) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(schedules))
	}
}

// TestScheduleDAO_Delete 测试删除定时任务
func TestScheduleDAO_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewScheduleDAO(db)
	
	dao.Create(Schedule{TaskID: 1, Spec: "@every 5m"})
	dao.Create(Schedule{TaskID: 2, Spec: "@every 10m"})
	
	err := dao.Delete(1)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	
	schedules, _ := dao.GetAll()
	if len(schedules) != 1 {
		t.Errorf("expected 1 schedule, got %d", len(schedules))
	}
}

// TestRunDAO_Create 测试创建运行记录
func TestRunDAO_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewRunDAO(db)
	
	run, err := dao.Create(Run{
		TaskID:  1,
		RcJobID: 123,
		Status: "running",
		Trigger: "manual",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	
	if run.ID == 0 {
		t.Error("expected non-zero ID")
	}
	
	if run.RcJobID != 123 {
		t.Errorf("expected RcJobID 123, got %d", run.RcJobID)
	}
}

// TestRunDAO_GetAll 测试获取所有运行记录
func TestRunDAO_GetAll(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewRunDAO(db)
	
	dao.Create(Run{TaskID: 1, RcJobID: 1, Status: "running", Trigger: "manual"})
	dao.Create(Run{TaskID: 2, RcJobID: 2, Status: "finished", Trigger: "schedule"})
	
	runs, err := dao.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

// TestRunDAO_GetRunning 测试获取运行中的任务
func TestRunDAO_GetRunning(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewRunDAO(db)
	
	dao.Create(Run{TaskID: 1, RcJobID: 1, Status: "running", Trigger: "manual"})
	dao.Create(Run{TaskID: 2, RcJobID: 2, Status: "finished", Trigger: "schedule"})
	dao.Create(Run{TaskID: 3, RcJobID: 0, Status: "running", Trigger: "manual"}) // 无rc_job_id
	
	running, err := dao.GetRunning()
	if err != nil {
		t.Fatalf("GetRunning() error = %v", err)
	}
	
	if len(running) != 1 {
		t.Errorf("expected 1 running job, got %d", len(running))
	}
}

// TestRunDAO_UpdateStatus 测试更新运行状态
func TestRunDAO_UpdateStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	dao := NewRunDAO(db)
	
	dao.Create(Run{TaskID: 1, RcJobID: 1, Status: "running", Trigger: "manual"})
	
	err := dao.UpdateStatus(1, "finished", "", map[string]any{"files": 10})
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	
	runs, _ := dao.GetAll()
	if runs[0].Status != "finished" {
		t.Errorf("expected Status 'finished', got '%s'", runs[0].Status)
	}
}
