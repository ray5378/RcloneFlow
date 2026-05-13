package service

import (
	"context"
	"os"
	"testing"
	"time"

	"rcloneflow/internal/store"
)

type cleanupRunSvcMock struct {
	listRunsFn   func(page, pageSize int) ([]RunRecord, int, error)
	deleteRunIDs []int64
}

func (m *cleanupRunSvcMock) ListRuns(page, pageSize int) ([]RunRecord, int, error) {
	if m.listRunsFn != nil {
		return m.listRunsFn(page, pageSize)
	}
	return nil, 0, nil
}
func (m *cleanupRunSvcMock) ListRunsByTask(taskId int64) ([]RunRecord, error)     { return nil, nil }
func (m *cleanupRunSvcMock) ListActiveRuns() ([]RunRecord, error)                 { return nil, nil }
func (m *cleanupRunSvcMock) GetActiveRunByTaskID(taskID int64) (RunRecord, error) { return RunRecord{}, nil }
func (m *cleanupRunSvcMock) GetRun(id int64) (RunRecord, error)                   { return RunRecord{}, nil }
func (m *cleanupRunSvcMock) UpdateRun(id int64, updateFn func(*RunRecord))        {}
func (m *cleanupRunSvcMock) DeleteRun(id int64) error {
	m.deleteRunIDs = append(m.deleteRunIDs, id)
	return nil
}
func (m *cleanupRunSvcMock) DeleteAllRuns() error                  { return nil }
func (m *cleanupRunSvcMock) DeleteRunsByTask(taskId int64) error   { return nil }
func (m *cleanupRunSvcMock) CleanOldRuns(days int) (int64, error)  { return 0, nil }

func TestCleanupService_ReplanAndCleanup(t *testing.T) {
	runMock := &cleanupRunSvcMock{
		listRunsFn: func(page, pageSize int) ([]RunRecord, int, error) {
			return []RunRecord{{ID: 1, StartedAt: time.Now().AddDate(0, 0, -8).Format(time.RFC3339)}}, 1, nil
		},
	}
	svc := NewCleanupService(NewRunService(runMock), time.Hour, 7)
	if svc.interval != time.Hour || svc.retention != 7 {
		t.Fatalf("unexpected initial cleanup service state: interval=%v retention=%d", svc.interval, svc.retention)
	}

	svc.cleanup()
	if len(runMock.deleteRunIDs) != 1 || runMock.deleteRunIDs[0] != 1 {
		t.Fatalf("unexpected deleted run ids after cleanup: %#v", runMock.deleteRunIDs)
	}

	svc.Replan(2, 0)
	if svc.interval != 2*time.Hour || svc.retention != 0 {
		t.Fatalf("unexpected replanned values: interval=%v retention=%d", svc.interval, svc.retention)
	}
	if len(svc.resetCh) != 1 {
		t.Fatalf("expected one reset notification, got %d", len(svc.resetCh))
	}
	svc.Replan(3, 9)
	if len(svc.resetCh) != 1 {
		t.Fatalf("expected reset channel to remain coalesced, got %d", len(svc.resetCh))
	}

	svc.retention = 5
	svc.cleanup()
	if len(runMock.deleteRunIDs) != 2 || runMock.deleteRunIDs[1] != 1 {
		t.Fatalf("unexpected deleted run ids after second cleanup: %#v", runMock.deleteRunIDs)
	}

	svc.retention = 0
	svc.cleanup()
	if len(runMock.deleteRunIDs) != 2 {
		t.Fatalf("cleanup should skip non-positive retention, got %#v", runMock.deleteRunIDs)
	}
}

func TestCleanupService_StartStopsOnContextAndStop(t *testing.T) {
	t.Run("context done exits after initial cleanup", func(t *testing.T) {
		runMock := &cleanupRunSvcMock{
			listRunsFn: func(page, pageSize int) ([]RunRecord, int, error) {
				return []RunRecord{{ID: 1, StartedAt: time.Now().AddDate(0, 0, -4).Format(time.RFC3339)}}, 1, nil
			},
		}
		svc := NewCleanupService(NewRunService(runMock), 24*time.Hour, 3)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		done := make(chan struct{})
		go func() {
			svc.Start(ctx)
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Start() did not stop after canceled context")
		}
		if len(runMock.deleteRunIDs) != 1 || runMock.deleteRunIDs[0] != 1 {
			t.Fatalf("unexpected deleted run ids: %#v", runMock.deleteRunIDs)
		}
	})

	t.Run("Stop exits loop", func(t *testing.T) {
		runMock := &cleanupRunSvcMock{
			listRunsFn: func(page, pageSize int) ([]RunRecord, int, error) {
				return []RunRecord{{ID: 2, StartedAt: time.Now().AddDate(0, 0, -5).Format(time.RFC3339)}}, 1, nil
			},
		}
		svc := NewCleanupService(NewRunService(runMock), 24*time.Hour, 4)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan struct{})
		go func() {
			svc.Start(ctx)
			close(done)
		}()
		for i := 0; i < 50 && len(runMock.deleteRunIDs) == 0; i++ {
			time.Sleep(10 * time.Millisecond)
		}
		svc.Stop()
		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Start() did not stop after Stop()")
		}
		if len(runMock.deleteRunIDs) == 0 {
			t.Fatal("expected initial cleanup before stop")
		}
	})
}

func TestScheduleService_DelegatesToStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-schedulesvc-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	createdTask, err := db.AddTask(store.Task{
		Name:         "task-1",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/from",
		TargetRemote: "dst",
		TargetPath:   "/to",
	})
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	svc := NewScheduleService(db)
	created, err := svc.CreateSchedule(createdTask.ID, "10|*|*|*|*", true)
	if err != nil {
		t.Fatalf("CreateSchedule() error = %v", err)
	}
	if created.TaskID != createdTask.ID || created.Spec != "10|*|*|*|*" || !created.Enabled {
		t.Fatalf("unexpected created schedule: %#v", created)
	}

	schedules, err := svc.ListSchedules()
	if err != nil || len(schedules) != 1 || schedules[0].ID != created.ID {
		t.Fatalf("ListSchedules() = (%v, %v)", schedules, err)
	}
	if err := svc.UpdateSpec(created.ID, "15|1|*|*|*"); err != nil {
		t.Fatalf("UpdateSpec() error = %v", err)
	}
	if err := svc.SetScheduleEnabled(created.ID, false); err != nil {
		t.Fatalf("SetScheduleEnabled() error = %v", err)
	}

	updated, err := svc.ListSchedules()
	if err != nil {
		t.Fatalf("ListSchedules() after update error = %v", err)
	}
	if updated[0].Spec != "15|1|*|*|*" || updated[0].Enabled != false {
		t.Fatalf("unexpected updated schedule: %#v", updated[0])
	}
	if err := svc.DeleteSchedule(created.ID); err != nil {
		t.Fatalf("DeleteSchedule() error = %v", err)
	}
	remaining, err := svc.ListSchedules()
	if err != nil {
		t.Fatalf("ListSchedules() after delete error = %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected no schedules after delete, got %#v", remaining)
	}
}
