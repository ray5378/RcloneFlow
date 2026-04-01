package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "rcloneflow_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// 打开数据库
	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	
	if db == nil {
		t.Fatal("expected non-nil DB")
	}
}

func TestOpenCreatesDir(t *testing.T) {
	// 临时目录不存在
	tmpDir := filepath.Join(os.TempDir(), "rcloneflow_test_create", "subdir")
	defer os.RemoveAll(filepath.Dir(tmpDir))
	
	// 应该自动创建目录
	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	db.Close()
}

func TestTaskOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	
	// 添加任务
	task := Task{
		Name:        "test-task",
		Mode:        "copy",
		SourceRemote: "src",
		SourcePath:  "/source",
		TargetRemote: "dst",
		TargetPath:  "/target",
	}
	
	created, err := db.AddTask(task)
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	
	if created.ID == 0 {
		t.Error("expected non-zero task ID")
	}
	
	if created.Name != "test-task" {
		t.Errorf("expected Name test-task, got %s", created.Name)
	}
	
	// 列出任务
	tasks, err := db.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
	
	// 获取单个任务
	got, ok := db.GetTask(created.ID)
	if !ok {
		t.Error("expected GetTask to return true")
	}
	
	if got.Name != "test-task" {
		t.Errorf("expected Name test-task, got %s", got.Name)
	}
	
	// 更新任务
	got.Name = "updated-task"
	err = db.UpdateTask(created.ID, got)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	
	got, ok = db.GetTask(created.ID)
	if !ok {
		t.Error("expected GetTask to return true after update")
	}
	
	if got.Name != "updated-task" {
		t.Errorf("expected Name updated-task, got %s", got.Name)
	}
	
	// 删除任务
	err = db.DeleteTask(created.ID)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}
	
	tasks, err = db.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after delete, got %d", len(tasks))
	}
}

func TestScheduleOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	
	// 先创建一个任务
	task := Task{
		Name:        "scheduled-task",
		Mode:        "sync",
		SourceRemote: "src",
		TargetRemote: "dst",
	}
	createdTask, _ := db.AddTask(task)
	
	// 添加定时任务
	schedule := Schedule{
		TaskID: createdTask.ID,
		Spec:   "@every 5m",
	}
	
	created, err := db.AddSchedule(schedule)
	if err != nil {
		t.Fatalf("AddSchedule() error = %v", err)
	}
	
	if created.ID == 0 {
		t.Error("expected non-zero schedule ID")
	}
	
	// 列出定时任务
	schedules, err := db.ListSchedules()
	if err != nil {
		t.Fatalf("ListSchedules() error = %v", err)
	}
	
	if len(schedules) != 1 {
		t.Errorf("expected 1 schedule, got %d", len(schedules))
	}
	
	// 删除定时任务
	err = db.DeleteSchedule(created.ID)
	if err != nil {
		t.Fatalf("DeleteSchedule() error = %v", err)
	}
	
	schedules, err = db.ListSchedules()
	if err != nil {
		t.Fatalf("ListSchedules() error = %v", err)
	}
	
	if len(schedules) != 0 {
		t.Errorf("expected 0 schedules after delete, got %d", len(schedules))
	}
}

func TestRunOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	
	// 先创建一个任务
	task := Task{
		Name:        "run-test-task",
		Mode:        "copy",
		SourceRemote: "src",
		TargetRemote: "dst",
	}
	createdTask, _ := db.AddTask(task)
	
	// 添加运行记录
	run := Run{
		TaskID:   createdTask.ID,
		RcJobID:  123,
		Status:   "running",
		Trigger:  "manual",
	}
	
	created, err := db.AddRun(run)
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}
	
	if created.ID == 0 {
		t.Error("expected non-zero run ID")
	}
	
	// 列出运行记录
	runs, err := db.ListRuns()
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	
	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}
	
	// 更新运行状态
	db.UpdateRun(created.ID, func(r *Run) {
		r.Status = "finished"
		r.Summary = map[string]any{"files": 10}
	})
	
	runs, err = db.ListRuns()
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	
	if runs[0].Status != "finished" {
		t.Errorf("expected Status finished, got %s", runs[0].Status)
	}
}
