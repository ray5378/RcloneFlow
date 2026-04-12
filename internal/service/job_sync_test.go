package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"rcloneflow/internal/rclone"
	"rcloneflow/internal/store"
)

// mockJobDB 模拟数据库
type mockJobDB struct {
	runs []store.JobStatus
	err  error
}

func (m *mockJobDB) ListRunningRuns() ([]store.JobStatus, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.runs, nil
}

func (m *mockJobDB) UpdateRunStatus(id int64, status, errorMsg string, summary map[string]any) error {
	if m.err != nil {
		return m.err
	}
	for i := range m.runs {
		if m.runs[i].ID == id {
			m.runs[i].Status = status
			break
		}
	}
	return nil
}

func TestJobSyncServiceStartStop(t *testing.T) {
	db := &mockJobDB{runs: []store.JobStatus{}}
	rc := rclone.NewFromEnv()

	svc := NewJobSyncService(db, rc)

	// 启动服务
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
		done <- struct{}{}
	}()

	svc.Start(ctx)

	select {
	case <-done:
		// 成功停止
	case <-time.After(1 * time.Second):
		t.Error("service did not stop in time")
	}
}

func TestJobSyncServiceSyncRunningJobs(t *testing.T) {
	db := &mockJobDB{
		runs: []store.JobStatus{
			{ID: 1, RcJobID: 100, Status: "running"},
		},
	}
	rc := rclone.NewFromEnv()

	svc := NewJobSyncService(db, rc)

	// 这个测试主要验证syncRunningJobs不会panic
	// 由于rc是nil，我们无法真正调用JobStatus
	// 但可以测试空场景
	svc.syncRunningJobs()
}

func TestJobSyncServiceWithNoRunningJobs(t *testing.T) {
	db := &mockJobDB{runs: []store.JobStatus{}}
	rc := rclone.NewFromEnv()

	svc := NewJobSyncService(db, rc)

	// 无运行中的任务时不应该出错
	svc.syncRunningJobs()
}

func TestJobSyncServiceMultipleRuns(t *testing.T) {
	db := &mockJobDB{
		runs: []store.JobStatus{
			{ID: 1, RcJobID: 100, Status: "running"},
			{ID: 2, RcJobID: 200, Status: "running"},
			{ID: 3, RcJobID: 0, Status: "running"}, // 无rc_job_id
		},
	}
	rc := rclone.NewFromEnv()

	svc := NewJobSyncService(db, rc)

	// 应该处理多个运行中的任务
	svc.syncRunningJobs()
}

func TestJobSyncServiceStop(t *testing.T) {
	db := &mockJobDB{runs: []store.JobStatus{}}
	rc := rclone.NewFromEnv()

	svc := NewJobSyncService(db, rc)

	// 停止一个未启动的服务不应该出错
	svc.Stop()
}

type mockJobDBForUpdate struct {
	mu      sync.Mutex
	runs    []store.JobStatus
	updated bool
}

func (m *mockJobDBForUpdate) ListRunningRuns() ([]store.JobStatus, error) {
	return m.runs, nil
}

func (m *mockJobDBForUpdate) UpdateRunStatus(id int64, status, errorMsg string, summary map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updated = true
	for i := range m.runs {
		if m.runs[i].ID == id {
			m.runs[i].Status = status
			return nil
		}
	}
	return nil
}

func TestJobSyncServiceUpdateStatus(t *testing.T) {
	// 这个测试验证mock实现正确
	db := &mockJobDBForUpdate{
		runs: []store.JobStatus{
			{ID: 1, RcJobID: 100, Status: "running"},
		},
	}
	rc := rclone.NewFromEnv()

	// 验证mock可以正常工作
	runs, err := db.ListRunningRuns()
	if err != nil {
		t.Fatalf("ListRunningRuns() error = %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 run, got %d", len(runs))
	}

	// 更新状态
	err = db.UpdateRunStatus(1, "finished", "", nil)
	if err != nil {
		t.Fatalf("UpdateRunStatus() error = %v", err)
	}

	if !db.updated {
		t.Error("expected UpdateRunStatus to be called")
	}

	_ = NewJobSyncService(db, rc)
}
