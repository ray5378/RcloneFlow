package store

import (
	"os"
	"testing"
	"time"
)

func TestUserOperations(t *testing.T) {
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

	u, err := db.CreateUser("admin", "hashed_pw")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if u.ID == 0 {
		t.Error("expected non-zero user ID")
	}
	if u.Username != "admin" {
		t.Errorf("expected username admin, got %s", u.Username)
	}

	got, ok := db.GetUserByUsername("admin")
	if !ok {
		t.Fatal("expected GetUserByUsername to return true")
	}
	if got.Password != "hashed_pw" {
		t.Errorf("expected password hashed_pw, got %s", got.Password)
	}

	got2, ok := db.GetUserByID(u.ID)
	if !ok {
		t.Fatal("expected GetUserByID to return true")
	}
	if got2.Username != "admin" {
		t.Errorf("expected username admin, got %s", got2.Username)
	}

	_, ok = db.GetUserByUsername("nonexistent")
	if ok {
		t.Error("expected GetUserByUsername to return false for nonexistent user")
	}

	_, ok = db.GetUserByID(999)
	if ok {
		t.Error("expected GetUserByID to return false for nonexistent ID")
	}

	if err := db.UpdatePassword(u.ID, "new_hashed_pw"); err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}
	got3, _ := db.GetUserByID(u.ID)
	if got3.Password != "new_hashed_pw" {
		t.Errorf("expected password new_hashed_pw, got %s", got3.Password)
	}

	if err := db.UpdateUsername(u.ID, "newadmin"); err != nil {
		t.Fatalf("UpdateUsername() error = %v", err)
	}
	got4, _ := db.GetUserByID(u.ID)
	if got4.Username != "newadmin" {
		t.Errorf("expected username newadmin, got %s", got4.Username)
	}

	users, err := db.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestDeleteRunOperations(t *testing.T) {
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

	task, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	r1, _ := db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})
	r2, _ := db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "schedule"})

	if err := db.DeleteRun(r1.ID); err != nil {
		t.Fatalf("DeleteRun() error = %v", err)
	}
	runs, total, _ := db.ListRuns(1, 50)
	if total != 1 {
		t.Errorf("expected 1 run after delete, got %d", total)
	}
	if runs[0].ID != r2.ID {
		t.Errorf("expected run ID %d, got %d", r2.ID, runs[0].ID)
	}

	if err := db.DeleteAllRuns(); err != nil {
		t.Fatalf("DeleteAllRuns() error = %v", err)
	}
	_, total2, _ := db.ListRuns(1, 50)
	if total2 != 0 {
		t.Errorf("expected 0 runs after delete all, got %d", total2)
	}
}

func TestDeleteRunsByTask(t *testing.T) {
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

	t1, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	t2, _ := db.AddTask(Task{Name: "t2", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(Run{TaskID: t1.ID, Status: "finished", Trigger: "manual"})
	db.AddRun(Run{TaskID: t1.ID, Status: "finished", Trigger: "manual"})
	db.AddRun(Run{TaskID: t2.ID, Status: "finished", Trigger: "manual"})

	if err := db.DeleteRunsByTask(t1.ID); err != nil {
		t.Fatalf("DeleteRunsByTask() error = %v", err)
	}
	_, total, _ := db.ListRuns(1, 50)
	if total != 1 {
		t.Errorf("expected 1 run for task2, got %d", total)
	}
}

func TestCleanOldRuns(t *testing.T) {
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

	task, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	// Manually set created_at to 2 days ago
	db.db.Exec("UPDATE runs SET created_at = datetime('now', '-2 days')")

	affected, err := db.CleanOldRuns(1)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 affected row, got %d", affected)
	}
}

func TestListRunsByTask(t *testing.T) {
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

	t1, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	t2, _ := db.AddTask(Task{Name: "t2", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(Run{TaskID: t1.ID, Status: "finished", Trigger: "manual"})
	db.AddRun(Run{TaskID: t1.ID, Status: "finished", Trigger: "manual"})
	db.AddRun(Run{TaskID: t2.ID, Status: "finished", Trigger: "manual"})

	runs, err := db.ListRunsByTask(t1.ID)
	if err != nil {
		t.Fatalf("ListRunsByTask() error = %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs for task1, got %d", len(runs))
	}
}

func TestListActiveRuns(t *testing.T) {
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

	task, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(Run{TaskID: task.ID, Status: "running", Trigger: "manual"})
	db.AddRun(Run{TaskID: task.ID, Status: "finalizing", Trigger: "manual"})
	db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	runs, err := db.ListActiveRuns()
	if err != nil {
		t.Fatalf("ListActiveRuns() error = %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 active runs, got %d", len(runs))
	}
}

func TestClearAllRunningStatus(t *testing.T) {
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

	task, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(Run{TaskID: task.ID, Status: "running", Trigger: "manual"})
	db.AddRun(Run{TaskID: task.ID, Status: "finalizing", Trigger: "manual"})
	db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	if err := db.ClearAllRunningStatus(); err != nil {
		t.Fatalf("ClearAllRunningStatus() error = %v", err)
	}

	runs, _, _ := db.ListRuns(1, 50)
	for _, r := range runs {
		if r.Status == "running" || r.Status == "finalizing" {
			t.Errorf("expected no running/finalizing runs, got status %s", r.Status)
		}
	}
}

func TestTryAcquireRun(t *testing.T) {
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

	task, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	run := &Run{TaskID: task.ID, Status: "running", Trigger: "manual"}
	acquired, existed, err := db.TryAcquireRun(run)
	if err != nil {
		t.Fatalf("TryAcquireRun() error = %v", err)
	}
	if existed {
		t.Error("expected existed=false on first acquire")
	}
	if acquired == nil {
		t.Fatal("expected non-nil run on acquire")
	}
	if acquired.ID == 0 {
		t.Error("expected non-zero run ID")
	}

	run2 := &Run{TaskID: task.ID, Status: "running", Trigger: "manual"}
	_, existed2, err := db.TryAcquireRun(run2)
	if err != nil {
		t.Fatalf("TryAcquireRun() second error = %v", err)
	}
	if !existed2 {
		t.Error("expected existed=true when task already running")
	}
}

func TestGetActiveRunByTaskID(t *testing.T) {
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

	task, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	db.AddRun(Run{TaskID: task.ID, Status: "running", Trigger: "manual", TaskName: "t1", TaskMode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	r, err := db.GetActiveRunByTaskID(task.ID)
	if err != nil {
		t.Fatalf("GetActiveRunByTaskID() error = %v", err)
	}
	if r.Status != "running" {
		t.Errorf("expected status running, got %s", r.Status)
	}
	if r.TaskName != "t1" {
		t.Errorf("expected taskName t1, got %s", r.TaskName)
	}

	_, err = db.GetActiveRunByTaskID(999)
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestUpdateRunProgress(t *testing.T) {
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

	task, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	run, _ := db.AddRun(Run{TaskID: task.ID, Status: "running", Trigger: "manual"})

	if err := db.UpdateRunProgress(run.ID, 1000, "10MB/s"); err != nil {
		t.Fatalf("UpdateRunProgress() error = %v", err)
	}

	runs, _ := db.ListRunsByTask(task.ID)
	if len(runs) == 0 {
		t.Fatal("expected at least 1 run")
	}
	r := runs[0]
	if r.BytesTransferred != 1000 {
		t.Errorf("expected bytes 1000, got %d", r.BytesTransferred)
	}
	if r.Speed != "10MB/s" {
		t.Errorf("expected speed 10MB/s, got %s", r.Speed)
	}

	if err := db.UpdateRunProgress(run.ID, 500, "5MB/s"); err != nil {
		t.Fatalf("UpdateRunProgress() should not error on lower value")
	}
	runs2, _ := db.ListRunsByTask(task.ID)
	if runs2[0].BytesTransferred != 1000 {
		t.Errorf("expected bytes to remain 1000 (no downgrade), got %d", runs2[0].BytesTransferred)
	}
}

func TestUpdateTaskSortOrders(t *testing.T) {
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

	t1, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	t2, _ := db.AddTask(Task{Name: "t2", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})

	updates := map[int64]int64{t1.ID: 100, t2.ID: 200}
	if err := db.UpdateTaskSortOrders(updates); err != nil {
		t.Fatalf("UpdateTaskSortOrders() error = %v", err)
	}

	got1, _ := db.GetTask(t1.ID)
	if got1.SortOrder != 100 {
		t.Errorf("expected t1 sortOrder 100, got %d", got1.SortOrder)
	}
	got2, _ := db.GetTask(t2.ID)
	if got2.SortOrder != 200 {
		t.Errorf("expected t2 sortOrder 200, got %d", got2.SortOrder)
	}
}

func TestScheduleEnabledAndSpec(t *testing.T) {
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

	task, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	sched, _ := db.AddSchedule(Schedule{TaskID: task.ID, Spec: "0 12 * * *", Enabled: true})

	if err := db.SetScheduleEnabled(sched.ID, false); err != nil {
		t.Fatalf("SetScheduleEnabled() error = %v", err)
	}
	got, _ := db.GetSchedule(sched.ID)
	if got.Enabled {
		t.Error("expected schedule to be disabled")
	}

	if err := db.UpdateScheduleSpec(sched.ID, "0 6 * * *"); err != nil {
		t.Fatalf("UpdateScheduleSpec() error = %v", err)
	}
	got2, _ := db.GetSchedule(sched.ID)
	if got2.Spec != "0 6 * * *" {
		t.Errorf("expected spec '0 6 * * *', got %s", got2.Spec)
	}

	nextTime := time.Now().Add(24 * time.Hour)
	if err := db.UpdateScheduleNextRunTime(sched.ID, nextTime); err != nil {
		t.Fatalf("UpdateScheduleNextRunTime() error = %v", err)
	}
}

func TestGetRun(t *testing.T) {
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

	task, _ := db.AddTask(Task{Name: "t1", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
	run, _ := db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual", Summary: map[string]any{"files": 10}})

	got, err := db.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Status != "finished" {
		t.Errorf("expected status finished, got %s", got.Status)
	}
	if got.Summary["files"] != float64(10) {
		t.Errorf("expected summary files 10, got %v", got.Summary["files"])
	}

	_, err = db.GetRun(999)
	if err == nil {
		t.Error("expected error for non-existent run")
	}
}

func TestNewDB(t *testing.T) {
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

	wrapped := NewDB(db.db)
	if wrapped == nil {
		t.Fatal("expected non-nil DB")
	}
}

func TestClose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
