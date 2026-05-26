package store

import (
	"os"
	"testing"
	"time"
)

func openSchedulesDB(t *testing.T) *DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_schedules_test_*")
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

func addTestTaskForSchedules(t *testing.T, db *DB, name string) Task {
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

func TestListSchedules_Empty(t *testing.T) {
	db := openSchedulesDB(t)
	defer db.Close()

	schedules, err := db.ListSchedules()
	if err != nil {
		t.Fatalf("ListSchedules() error = %v", err)
	}
	if len(schedules) != 0 {
		t.Errorf("expected 0 schedules in empty DB, got %d", len(schedules))
	}
}

func TestAddSchedule_NonExistentTask(t *testing.T) {
	db := openSchedulesDB(t)
	defer db.Close()

	_, err := db.AddSchedule(Schedule{
		TaskID:  99999,
		Spec:    "@every 1h",
		Enabled: true,
	})
	if err == nil {
		t.Error("expected error when adding schedule for non-existent task (FK constraint)")
	}
}

func TestGetSchedule_NotExist(t *testing.T) {
	db := openSchedulesDB(t)
	defer db.Close()

	_, ok := db.GetSchedule(99999)
	if ok {
		t.Error("expected false for non-existent schedule")
	}
}

func TestDeleteSchedule_NotExist(t *testing.T) {
	db := openSchedulesDB(t)
	defer db.Close()

	err := db.DeleteSchedule(99999)
	if err != nil {
		t.Fatalf("DeleteSchedule on non-existent should not fail: %v", err)
	}
}

func TestSetScheduleEnabled_NotExist(t *testing.T) {
	db := openSchedulesDB(t)
	defer db.Close()

	err := db.SetScheduleEnabled(99999, true)
	if err != nil {
		t.Fatalf("SetScheduleEnabled on non-existent should not fail: %v", err)
	}
}

func TestUpdateScheduleSpec_NotExist(t *testing.T) {
	db := openSchedulesDB(t)
	defer db.Close()

	err := db.UpdateScheduleSpec(99999, "0 0 * * *")
	if err != nil {
		t.Fatalf("UpdateScheduleSpec on non-existent should not fail: %v", err)
	}
}

func TestUpdateScheduleNextRunTime_NotExist(t *testing.T) {
	db := openSchedulesDB(t)
	defer db.Close()

	err := db.UpdateScheduleNextRunTime(99999, time.Now())
	if err != nil {
		t.Fatalf("UpdateScheduleNextRunTime on non-existent should not fail: %v", err)
	}
}

func TestSchedule_FullLifecycle(t *testing.T) {
	db := openSchedulesDB(t)
	defer db.Close()

	task := addTestTaskForSchedules(t, db, "sched-lifecycle")

	sched, err := db.AddSchedule(Schedule{
		TaskID:  task.ID,
		Spec:    "0 6 * * *",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("AddSchedule() error = %v", err)
	}
	if sched.ID == 0 {
		t.Error("expected non-zero schedule ID")
	}
	if sched.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	got, ok := db.GetSchedule(sched.ID)
	if !ok {
		t.Fatal("expected GetSchedule to return true")
	}
	if got.Spec != "0 6 * * *" {
		t.Errorf("expected spec '0 6 * * *', got %s", got.Spec)
	}

	if err := db.SetScheduleEnabled(sched.ID, false); err != nil {
		t.Fatalf("SetScheduleEnabled() error = %v", err)
	}
	got, _ = db.GetSchedule(sched.ID)
	if got.Enabled {
		t.Error("expected schedule to be disabled")
	}

	if err := db.SetScheduleEnabled(sched.ID, true); err != nil {
		t.Fatalf("SetScheduleEnabled back to true: %v", err)
	}

	if err := db.UpdateScheduleSpec(sched.ID, "0 12 * * *"); err != nil {
		t.Fatalf("UpdateScheduleSpec() error = %v", err)
	}
	got, _ = db.GetSchedule(sched.ID)
	if got.Spec != "0 12 * * *" {
		t.Errorf("expected spec '0 12 * * *', got %s", got.Spec)
	}

	nextTime := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	if err := db.UpdateScheduleNextRunTime(sched.ID, nextTime); err != nil {
		t.Fatalf("UpdateScheduleNextRunTime() error = %v", err)
	}

	schedules, _ := db.ListSchedules()
	if len(schedules) != 1 {
		t.Errorf("expected 1 schedule, got %d", len(schedules))
	}

	if err := db.DeleteSchedule(sched.ID); err != nil {
		t.Fatalf("DeleteSchedule() error = %v", err)
	}

	schedules, _ = db.ListSchedules()
	if len(schedules) != 0 {
		t.Errorf("expected 0 schedules after delete, got %d", len(schedules))
	}
}

func TestSchedule_EnabledDefaults(t *testing.T) {
	db := openSchedulesDB(t)
	defer db.Close()

	task := addTestTaskForSchedules(t, db, "enabled-default")

	sched, err := db.AddSchedule(Schedule{
		TaskID:  task.ID,
		Spec:    "@weekly",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("AddSchedule() error = %v", err)
	}

	got, _ := db.GetSchedule(sched.ID)
	if !got.Enabled {
		t.Error("expected schedule to be enabled when explicitly set to true")
	}
}