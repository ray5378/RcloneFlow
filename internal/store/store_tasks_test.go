package store

import (
	"encoding/json"
	"os"
	"testing"
)

func openTasksDB(t *testing.T) *DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_tasks_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })
	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return db
}

func TestAddTask_WithOptions(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	task := Task{
		Name:         "task-with-options",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/src",
		TargetRemote: "dst",
		TargetPath:   "/dst",
		Options:      json.RawMessage(`{"transfers": 8, "checkers": 16}`),
	}

	created, err := db.AddTask(task)
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	if created.ID == 0 {
		t.Error("expected non-zero task ID")
	}

	got, ok := db.GetTask(created.ID)
	if !ok {
		t.Fatal("expected GetTask to return true")
	}
	if got.Options == nil {
		t.Fatal("expected non-nil Options")
	}
	var opts map[string]any
	if err := json.Unmarshal(got.Options, &opts); err != nil {
		t.Fatalf("unmarshal Options: %v", err)
	}
	if opts["transfers"] != float64(8) {
		t.Errorf("expected transfers=8, got %v", opts["transfers"])
	}
	if opts["checkers"] != float64(16) {
		t.Errorf("expected checkers=16, got %v", opts["checkers"])
	}
}

func TestAddTask_WithBisyncOptions(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	task := Task{
		Name:         "bisync-task",
		Mode:         "bisync",
		SourceRemote: "src",
		SourcePath:   "/src",
		TargetRemote: "dst",
		TargetPath:   "/dst",
		BisyncOptions: json.RawMessage(`{"resync": true, "maxDelete": "5", "conflictResolve": "newer"}`),
	}

	created, err := db.AddTask(task)
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	got, ok := db.GetTask(created.ID)
	if !ok {
		t.Fatal("expected GetTask to return true")
	}
	if got.BisyncOptions == nil {
		t.Fatal("expected non-nil BisyncOptions")
	}
	var bisync map[string]any
	if err := json.Unmarshal(got.BisyncOptions, &bisync); err != nil {
		t.Fatalf("unmarshal BisyncOptions: %v", err)
	}
	if bisync["resync"] != true {
		t.Errorf("expected resync=true, got %v", bisync["resync"])
	}
	if bisync["maxDelete"] != "5" {
		t.Errorf("expected maxDelete=5, got %v", bisync["maxDelete"])
	}
}

func TestAddTask_SortOrderIncrements(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	makeTask := func(name string) Task {
		return Task{
			Name: name, Mode: "copy",
			SourceRemote: "src", SourcePath: "/a",
			TargetRemote: "dst", TargetPath: "/b",
		}
	}

	t1, _ := db.AddTask(makeTask("t1"))
	t2, _ := db.AddTask(makeTask("t2"))
	t3, _ := db.AddTask(makeTask("t3"))

	if t1.SortOrder >= t2.SortOrder || t2.SortOrder >= t3.SortOrder {
		t.Errorf("sort orders should increment: got %d, %d, %d", t1.SortOrder, t2.SortOrder, t3.SortOrder)
	}

	tasks, _ := db.ListTasks()
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
	for i := 1; i < len(tasks); i++ {
		if tasks[i-1].SortOrder > tasks[i].SortOrder {
			t.Error("tasks should be sorted by sort_order ascending")
		}
	}
}

func TestAddTask_NilOptionsStored(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	task := Task{
		Name:         "no-options-task",
		Mode:         "copy",
		SourceRemote: "src",
		TargetRemote: "dst",
	}

	created, _ := db.AddTask(task)
	got, ok := db.GetTask(created.ID)
	if !ok {
		t.Fatal("expected GetTask to return true")
	}
	if got.Options != nil {
		t.Error("expected nil Options for task without options")
	}
	if got.BisyncOptions != nil {
		t.Error("expected nil BisyncOptions for task without bisync options")
	}
}

func TestUpdateTask_SourcePathTargetPath(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	task := Task{
		Name:         "path-task",
		Mode:         "copy",
		SourceRemote: "src:",
		SourcePath:   "/data/incoming",
		TargetRemote: "dst:",
		TargetPath:   "/data/archive",
	}
	created, _ := db.AddTask(task)

	created.SourcePath = "/data/new"
	created.TargetPath = "/backup/new"
	if err := db.UpdateTask(created.ID, created); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	got, _ := db.GetTask(created.ID)
	if got.SourcePath != "/data/new" {
		t.Errorf("expected SourcePath /data/new, got %s", got.SourcePath)
	}
	if got.TargetPath != "/backup/new" {
		t.Errorf("expected TargetPath /backup/new, got %s", got.TargetPath)
	}
}

func TestUpdateTask_NotExist(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	err := db.UpdateTask(99999, Task{Name: "ghost", Mode: "copy"})
	if err != nil {
		t.Fatalf("UpdateTask on non-existent should not panic, got error: %v", err)
	}

	_, ok := db.GetTask(99999)
	if ok {
		t.Error("non-existent task should not appear after update")
	}
}

func TestDeleteTask_NotExist(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	err := db.DeleteTask(99999)
	if err != nil {
		t.Fatalf("DeleteTask on non-existent should not fail: %v", err)
	}
}

func TestUpdateTaskSortOrders_EmptyMap(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	makeTask := func(name string) Task {
		return Task{
			Name: name, Mode: "copy",
			SourceRemote: "src", SourcePath: "/a",
			TargetRemote: "dst", TargetPath: "/b",
		}
	}
	db.AddTask(makeTask("t1"))
	db.AddTask(makeTask("t2"))

	if err := db.UpdateTaskSortOrders(map[int64]int64{}); err != nil {
		t.Fatalf("UpdateTaskSortOrders with empty map should not error: %v", err)
	}

	tasks, _ := db.ListTasks()
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestTask_CRUD_CompleteLifecycle(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	task := Task{
		Name:         "lifecycle-test",
		Mode:         "sync",
		SourceRemote: "remoteA",
		SourcePath:   "/source",
		TargetRemote: "remoteB",
		TargetPath:   "/target",
		Options:      json.RawMessage(`{"dry_run": true}`),
	}

	created, err := db.AddTask(task)
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	tasks, _ := db.ListTasks()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task after add, got %d", len(tasks))
	}

	created.Name = "lifecycle-updated"
	created.Mode = "copy"
	if err := db.UpdateTask(created.ID, created); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	got, _ := db.GetTask(created.ID)
	if got.Name != "lifecycle-updated" {
		t.Errorf("expected Name lifecycle-updated, got %s", got.Name)
	}
	if got.Mode != "copy" {
		t.Errorf("expected Mode copy, got %s", got.Mode)
	}

	if err := db.DeleteTask(created.ID); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	tasks, _ = db.ListTasks()
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after delete, got %d", len(tasks))
	}
}

func TestListTasks_Empty(t *testing.T) {
	db := openTasksDB(t)
	defer db.Close()

	tasks, err := db.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks in empty DB, got %d", len(tasks))
	}
}