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
	baseDir, err := os.MkdirTemp("", "rcloneflow_test_create_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(baseDir)

	tmpDir := filepath.Join(baseDir, "subdir")

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
		Name:         "test-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/source",
		TargetRemote: "dst",
		TargetPath:   "/target",
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

func TestReorderTasksPersistsOrder(t *testing.T) {
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

	first, err := db.AddTask(Task{Name: "task-1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/1"})
	if err != nil {
		t.Fatalf("AddTask(first) error = %v", err)
	}
	second, err := db.AddTask(Task{Name: "task-2", Mode: "copy", SourceRemote: "src", SourcePath: "/b", TargetRemote: "dst", TargetPath: "/2"})
	if err != nil {
		t.Fatalf("AddTask(second) error = %v", err)
	}
	third, err := db.AddTask(Task{Name: "task-3", Mode: "copy", SourceRemote: "src", SourcePath: "/c", TargetRemote: "dst", TargetPath: "/3"})
	if err != nil {
		t.Fatalf("AddTask(third) error = %v", err)
	}

	if err := db.ReorderTasks([]int64{second.ID, third.ID, first.ID}); err != nil {
		t.Fatalf("ReorderTasks() error = %v", err)
	}

	tasks, err := db.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	gotIDs := []int64{tasks[0].ID, tasks[1].ID, tasks[2].ID}
	wantIDs := []int64{second.ID, third.ID, first.ID}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Fatalf("unexpected order after reorder: got %v want %v", gotIDs, wantIDs)
		}
		if tasks[i].SortIndex != int64(i+1) {
			t.Fatalf("unexpected sort index at pos %d: got %d want %d", i, tasks[i].SortIndex, i+1)
		}
	}
}

func TestListTasks_NormalizesMissingSortIndexes(t *testing.T) {
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

	first, err := db.AddTask(Task{Name: "task-a", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/1"})
	if err != nil {
		t.Fatalf("AddTask(first) error = %v", err)
	}
	second, err := db.AddTask(Task{Name: "task-b", Mode: "copy", SourceRemote: "src", SourcePath: "/b", TargetRemote: "dst", TargetPath: "/2"})
	if err != nil {
		t.Fatalf("AddTask(second) error = %v", err)
	}
	third, err := db.AddTask(Task{Name: "task-c", Mode: "copy", SourceRemote: "src", SourcePath: "/c", TargetRemote: "dst", TargetPath: "/3"})
	if err != nil {
		t.Fatalf("AddTask(third) error = %v", err)
	}

	if _, err := db.db.Exec(`UPDATE tasks SET sort_index = 0 WHERE id IN (?, ?)`, first.ID, third.ID); err != nil {
		t.Fatalf("force invalid sort_index error = %v", err)
	}

	tasks, err := db.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	gotIDs := []int64{tasks[0].ID, tasks[1].ID, tasks[2].ID}
	wantIDs := []int64{second.ID, first.ID, third.ID}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Fatalf("unexpected normalized order: got %v want %v", gotIDs, wantIDs)
		}
		if tasks[i].SortIndex != int64(i+1) {
			t.Fatalf("unexpected normalized sort index at pos %d: got %d want %d", i, tasks[i].SortIndex, i+1)
		}
	}

	firstReloaded, ok := db.GetTask(first.ID)
	if !ok {
		t.Fatalf("expected first task to still exist")
	}
	thirdReloaded, ok := db.GetTask(third.ID)
	if !ok {
		t.Fatalf("expected third task to still exist")
	}
	if firstReloaded.SortIndex <= 0 || thirdReloaded.SortIndex <= 0 {
		t.Fatalf("expected normalized sort indexes persisted, got first=%d third=%d", firstReloaded.SortIndex, thirdReloaded.SortIndex)
	}
}

func TestListTasks_NormalizesDuplicateSortIndexes(t *testing.T) {
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

	first, err := db.AddTask(Task{Name: "dup-task-a", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/1"})
	if err != nil {
		t.Fatalf("AddTask(first) error = %v", err)
	}
	second, err := db.AddTask(Task{Name: "dup-task-b", Mode: "copy", SourceRemote: "src", SourcePath: "/b", TargetRemote: "dst", TargetPath: "/2"})
	if err != nil {
		t.Fatalf("AddTask(second) error = %v", err)
	}
	third, err := db.AddTask(Task{Name: "dup-task-c", Mode: "copy", SourceRemote: "src", SourcePath: "/c", TargetRemote: "dst", TargetPath: "/3"})
	if err != nil {
		t.Fatalf("AddTask(third) error = %v", err)
	}

	if _, err := db.db.Exec(`UPDATE tasks SET sort_index = 1 WHERE id IN (?, ?)`, first.ID, second.ID); err != nil {
		t.Fatalf("force duplicate sort_index error = %v", err)
	}

	tasks, err := db.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	gotIDs := []int64{tasks[0].ID, tasks[1].ID, tasks[2].ID}
	wantIDs := []int64{second.ID, first.ID, third.ID}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Fatalf("unexpected normalized duplicate order: got %v want %v", gotIDs, wantIDs)
		}
		if tasks[i].SortIndex != int64(i+1) {
			t.Fatalf("unexpected normalized sort index at pos %d: got %d want %d", i, tasks[i].SortIndex, i+1)
		}
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
		Name:         "scheduled-task",
		Mode:         "sync",
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

	task := Task{
		Name:         "run-test-task",
		Mode:         "copy",
		SourceRemote: "src",
		TargetRemote: "dst",
	}
	createdTask, _ := db.AddTask(task)

	run := Run{
		TaskID:  createdTask.ID,
		Status:  "running",
		Trigger: "manual",
	}

	created, err := db.AddRun(run)
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}
	if created.ID == 0 {
		t.Error("expected non-zero run ID")
	}

	runs, _, err := db.ListRuns(1, 50)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}

	db.UpdateRun(created.ID, func(r *Run) {
		r.Status = "finished"
		r.Summary = map[string]any{"files": 10}
	})

	runs, _, err = db.ListRuns(1, 50)
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

	_, ok := db.GetTask(999)
	if ok {
		t.Error("expected GetTask(999) to return false")
	}

	task := Task{
		Name:         "get-test-task",
		Mode:         "sync",
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

	_, ok := db.GetSchedule(999)
	if ok {
		t.Error("expected GetSchedule(999) to return false")
	}

	task := Task{
		Name:         "schedule-get-test",
		Mode:         "copy",
		SourceRemote: "src",
		TargetRemote: "dst",
	}
	createdTask, _ := db.AddTask(task)

	schedule, _ := db.AddSchedule(Schedule{TaskID: createdTask.ID, Spec: "@every 10m"})
	got, ok := db.GetSchedule(schedule.ID)
	if !ok {
		t.Error("expected GetSchedule to return true for existing schedule")
	}
	if got.Spec != "@every 10m" {
		t.Errorf("expected Spec @every 10m, got %s", got.Spec)
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
		Name:         "update-status-test",
		Mode:         "move",
		SourceRemote: "src",
		TargetRemote: "dst",
	}
	createdTask, _ := db.AddTask(task)

	run, _ := db.AddRun(Run{TaskID: createdTask.ID, Status: "running", Trigger: "schedule"})
	summary := map[string]any{"files_copied": 50, "bytes_transferred": 1024000}
	if err = db.UpdateRunStatus(run.ID, "finished", "", summary); err != nil {
		t.Fatalf("UpdateRunStatus() error = %v", err)
	}

	runs, _, _ := db.ListRuns(1, 50)
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

	var version int
	if err = db.db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		t.Fatalf("failed to get migration version: %v", err)
	}
	if version < 3 {
		t.Errorf("expected migration version >= 3, got %d", version)
	}
}

func TestTaskNameUniqueIndex(t *testing.T) {
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

	_, err = db.AddTask(Task{Name: "dup-name", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	if err != nil {
		t.Fatalf("first AddTask() error = %v", err)
	}

	_, err = db.AddTask(Task{Name: "Dup-Name", Mode: "sync", SourceRemote: "src2", SourcePath: "/c", TargetRemote: "dst2", TargetPath: "/d"})
	if err == nil {
		t.Fatal("expected duplicate task name insert to fail")
	}
}
