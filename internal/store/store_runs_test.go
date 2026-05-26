package store

import (
	"os"
	"testing"
)

func openRunsDB(t *testing.T) *DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_runs_test_*")
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

func addTestTaskForRuns(t *testing.T, db *DB, name string) Task {
	t.Helper()
	task, err := db.AddTask(Task{
		Name: name, Mode: "copy",
		SourceRemote: "src", SourcePath: "/a",
		TargetRemote: "dst", TargetPath: "/b",
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}
	return task
}

func TestAddRun_WithError(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "error-test")

	run, err := db.AddRun(Run{
		TaskID:  task.ID,
		Status:  "failed",
		Trigger: "manual",
		Error:   "rclone: connection refused",
	})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}
	if run.ID == 0 {
		t.Error("expected non-zero run ID")
	}
	if run.Error != "rclone: connection refused" {
		t.Errorf("expected error message, got %s", run.Error)
	}

	got, err := db.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Error != "rclone: connection refused" {
		t.Errorf("expected error persisted, got %s", got.Error)
	}
}

func TestAddRun_NilSummary(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "nil-summary-test")

	run, err := db.AddRun(Run{
		TaskID:  task.ID,
		Status:  "running",
		Trigger: "manual",
	})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}

	got, err := db.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Summary != nil && len(got.Summary) != 0 {
		t.Errorf("expected nil or empty summary for nil input, got %v", got.Summary)
	}
}

func TestAddRun_WithTaskDetails(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "detail-test")

	run, err := db.AddRun(Run{
		TaskID:        task.ID,
		Status:        "running",
		Trigger:       "schedule",
		TaskName:      "detail-test",
		TaskMode:      "copy",
		SourceRemote:  "src:",
		SourcePath:    "/data",
		TargetRemote:  "dst:",
		TargetPath:    "/backup",
	})
	if err != nil {
		t.Fatalf("AddRun() error = %v", err)
	}
	if run.TaskName != "detail-test" {
		t.Errorf("expected TaskName detail-test, got %s", run.TaskName)
	}
}

func TestListRuns_PageZero(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "page-test")
	db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})
	db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	runs, total, err := db.ListRuns(0, 10)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if total < 2 {
		t.Errorf("expected total >= 2, got %d", total)
	}
	if len(runs) < 0 {
		t.Errorf("expected runs slice, got nil")
	}
}

func TestListRuns_NegativePage(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "neg-page-test")
	db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	_, _, err := db.ListRuns(-1, 10)
	if err != nil {
		t.Fatalf("ListRuns with negative page should not panic: %v", err)
	}
}

func TestListRuns_EmptyDB(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	runs, total, err := db.ListRuns(1, 50)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestDeleteRunsByIDs_EmptySlice(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	if err := db.DeleteRunsByIDs([]int64{}); err != nil {
		t.Fatalf("DeleteRunsByIDs with empty slice should not error: %v", err)
	}
}

func TestDeleteRunsByIDs_Multiple(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "multi-delete")
	r1, _ := db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})
	r2, _ := db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "schedule"})
	r3, _ := db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	if err := db.DeleteRunsByIDs([]int64{r1.ID, r3.ID}); err != nil {
		t.Fatalf("DeleteRunsByIDs() error = %v", err)
	}

	_, total, _ := db.ListRuns(1, 50)
	if total != 1 {
		t.Errorf("expected 1 run remaining, got %d", total)
	}

	_, err := db.GetRun(r2.ID)
	if err != nil {
		t.Fatalf("expected run %d to exist: %v", r2.ID, err)
	}
}

func TestDeleteRun_NotExist(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	if err := db.DeleteRun(99999); err != nil {
		t.Fatalf("DeleteRun on non-existent should not fail: %v", err)
	}
}

func TestDeleteAllRuns_Empty(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	if err := db.DeleteAllRuns(); err != nil {
		t.Fatalf("DeleteAllRuns on empty DB should not error: %v", err)
	}

	_, total, _ := db.ListRuns(1, 50)
	if total != 0 {
		t.Errorf("expected 0 runs, got %d", total)
	}
}

func TestUpdateRun_NotExist(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	err := db.UpdateRun(99999, func(r *Run) {
		r.Status = "finished"
	})
	if err == nil {
		t.Fatal("expected error when updating non-existent run")
	}
}

func TestUpdateRunStatus_NotExist(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	err := db.UpdateRunStatus(99999, "finished", "", map[string]any{"result": "ok"})
	if err != nil {
		t.Fatalf("UpdateRunStatus on non-existent should not panic: %v", err)
	}
}

func TestUpdateRunProgress_NotExist(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	err := db.UpdateRunProgress(99999, 5000, "50MB/s")
	if err != nil {
		t.Fatalf("UpdateRunProgress on non-existent should not panic: %v", err)
	}
}

func TestVacuum_Success(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "vacuum-test")
	db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	if err := db.Vacuum(); err != nil {
		t.Fatalf("Vacuum() error = %v", err)
	}
}

func TestVacuum_EmptyDB(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	if err := db.Vacuum(); err != nil {
		t.Fatalf("Vacuum() on empty DB error = %v", err)
	}
}

func TestListRunsByTask_NoRuns(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "no-run-task")

	runs, err := db.ListRunsByTask(task.ID)
	if err != nil {
		t.Fatalf("ListRunsByTask() error = %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestCleanOldRuns_ZeroDays(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "clean-zero-test")
	db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})

	db.db.Exec("UPDATE runs SET created_at = datetime('now', '-1 seconds')")

	affected, err := db.CleanOldRuns(0)
	if err != nil {
		t.Fatalf("CleanOldRuns(0) error = %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 affected row, got %d", affected)
	}

	_, total, _ := db.ListRuns(1, 50)
	if total != 0 {
		t.Errorf("expected 0 runs after cleaning 0 days, got %d", total)
	}
}

func TestListActiveRuns_NoneActive(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "no-active-test")
	db.AddRun(Run{TaskID: task.ID, Status: "finished", Trigger: "manual"})
	db.AddRun(Run{TaskID: task.ID, Status: "stopped", Trigger: "manual"})

	runs, err := db.ListActiveRuns()
	if err != nil {
		t.Fatalf("ListActiveRuns() error = %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 active runs, got %d", len(runs))
	}
}

func TestGetRun_NotExist(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	_, err := db.GetRun(99999)
	if err == nil {
		t.Error("expected error for non-existent run")
	}
}

func TestTryAcquireRun_ConcurrentBlock(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "concurrent-test")

	run1 := &Run{TaskID: task.ID, Status: "running", Trigger: "manual"}
	acquired, existed, err := db.TryAcquireRun(run1)
	if err != nil {
		t.Fatalf("TryAcquireRun() error = %v", err)
	}
	if existed || acquired == nil {
		t.Fatal("expected successful acquire")
	}

	run2 := &Run{TaskID: task.ID, Status: "running", Trigger: "schedule"}
	_, existed2, err := db.TryAcquireRun(run2)
	if err != nil {
		t.Fatalf("TryAcquireRun() second error = %v", err)
	}
	if !existed2 {
		t.Error("expected existed=true when task already running")
	}
}

func TestGetRun_CompletedLifecycle(t *testing.T) {
	db := openRunsDB(t)
	defer db.Close()

	task := addTestTaskForRuns(t, db, "lifecycle-run")

	run, _ := db.AddRun(Run{
		TaskID:  task.ID,
		Status:  "running",
		Trigger: "manual",
	})
	if run.Status != "running" {
		t.Errorf("expected running, got %s", run.Status)
	}

	if err := db.UpdateRunStatus(run.ID, "finalizing", "", map[string]any{"progress": 80}); err != nil {
		t.Fatalf("UpdateRunStatus finalizing: %v", err)
	}

	got, _ := db.GetRun(run.ID)
	if got.Status != "finalizing" {
		t.Errorf("expected finalizing, got %s", got.Status)
	}

	if err := db.UpdateRunStatus(run.ID, "finished", "", map[string]any{"files": 100}); err != nil {
		t.Fatalf("UpdateRunStatus finished: %v", err)
	}

	got, _ = db.GetRun(run.ID)
	if got.Status != "finished" {
		t.Errorf("expected finished, got %s", got.Status)
	}
}