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

func TestGetTask(t *testing.T) {
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
	
	// 获取不存在的任务
	_, ok := db.GetTask(999)
	if ok {
		t.Error("expected GetTask(999) to return false")
	}
	
	// 添加并获取任务
	task := Task{
		Name:        "get-test-task",
		Mode:        "sync",
		SourceRemote: "src",
		TargetRemote: "dst",
	}
	created, _ := db.AddTask(task)
	
	got, ok := db.GetTask(created.ID)
	if !ok {
		t.Error("expected GetTask to return true for existing task")
	}
	
	if got.Name != "get-test-task" {
		t.Errorf("expected Name get-test-task, got %s", got.Name)
	}
}

func TestGetSchedule(t *testing.T) {
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
	
	// 获取不存在的定时任务
	_, ok := db.GetSchedule(999)
	if ok {
		t.Error("expected GetSchedule(999) to return false")
	}
	
	// 添加并获取定时任务
	task := Task{
		Name:        "schedule-get-test",
		Mode:        "copy",
		SourceRemote: "src",
		TargetRemote: "dst",
	}
	createdTask, _ := db.AddTask(task)
	
	schedule, _ := db.AddSchedule(Schedule{
		TaskID: createdTask.ID,
		Spec:   "@every 10m",
	})
	
	got, ok := db.GetSchedule(schedule.ID)
	if !ok {
		t.Error("expected GetSchedule to return true for existing schedule")
	}
	
	if got.Spec != "@every 10m" {
		t.Errorf("expected Spec @every 10m, got %s", got.Spec)
	}
}

func TestListRunningRuns(t *testing.T) {
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
	
	// 添加任务和运行记录
	task := Task{
		Name:        "running-test",
		Mode:        "copy",
		SourceRemote: "src",
		TargetRemote: "dst",
	}
	createdTask, _ := db.AddTask(task)
	
	// 添加一个running状态的运行
	db.AddRun(Run{
		TaskID:  createdTask.ID,
		RcJobID: 100,
		Status:  "running",
		Trigger: "manual",
	})
	
	// 添加一个finished状态的运行
	db.AddRun(Run{
		TaskID:  createdTask.ID,
		RcJobID: 200,
		Status:  "finished",
		Trigger: "manual",
	})
	
	// 获取running运行
	runs, err := db.ListRunningRuns()
	if err != nil {
		t.Fatalf("ListRunningRuns() error = %v", err)
	}
	
	if len(runs) != 1 {
		t.Errorf("expected 1 running run, got %d", len(runs))
	}
	
	if runs[0].Status != "running" {
		t.Errorf("expected status running, got %s", runs[0].Status)
	}
}

func TestUpdateRunStatus(t *testing.T) {
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
	
	task := Task{
		Name:        "update-status-test",
		Mode:        "move",
		SourceRemote: "src",
		TargetRemote: "dst",
	}
	createdTask, _ := db.AddTask(task)
	
	run, _ := db.AddRun(Run{
		TaskID:  createdTask.ID,
		RcJobID: 300,
		Status:  "running",
		Trigger: "schedule",
	})
	
	// 更新状态
	summary := map[string]any{"files_copied": 50, "bytes_transferred": 1024000}
	err = db.UpdateRunStatus(run.ID, "finished", "", summary)
	if err != nil {
		t.Fatalf("UpdateRunStatus() error = %v", err)
	}
	
	// 验证更新
	runs, _ := db.ListRuns()
	if runs[0].Status != "finished" {
		t.Errorf("expected status finished, got %s", runs[0].Status)
	}
}

func TestMigrationCreatesVersionTable(t *testing.T) {
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
	
	// 验证迁移版本表存在且有记录
	var version int
	err = db.db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err != nil {
		t.Fatalf("failed to get migration version: %v", err)
	}
	
	if version < 1 {
		t.Errorf("expected migration version >= 1, got %d", version)
	}
}
