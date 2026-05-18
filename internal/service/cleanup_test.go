package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunService_CleanOldRuns_DeletesExpiredRecords(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-cleanup-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock runs: 3 old (10 days ago), 2 recent (1 day ago)
	now := time.Now()
	oldDate := now.AddDate(0, 0, -10).Format(time.RFC3339)
	recentDate := now.AddDate(0, 0, -1).Format(time.RFC3339)

	mock := &runServiceDBMock{
		listRunsPages: map[int][]RunRecord{
			1: {
				{ID: 1, TaskID: 1, Status: "finished", StartedAt: oldDate},
				{ID: 2, TaskID: 1, Status: "finished", StartedAt: oldDate},
				{ID: 3, TaskID: 1, Status: "finished", StartedAt: oldDate},
				{ID: 4, TaskID: 1, Status: "finished", StartedAt: recentDate},
				{ID: 5, TaskID: 1, Status: "finished", StartedAt: recentDate},
			},
		},
		listRunsTotal: 5,
	}

	svc := NewRunService(mock)
	deleted, err := svc.CleanOldRuns(7)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected 3 deleted, got %d", deleted)
	}

	// Verify only old runs were deleted
	if len(mock.deletedRun) != 3 {
		t.Fatalf("expected 3 delete calls, got %d", len(mock.deletedRun))
	}
	deletedSet := map[int64]bool{}
	for _, id := range mock.deletedRun {
		deletedSet[id] = true
	}
	for _, id := range []int64{1, 2, 3} {
		if !deletedSet[id] {
			t.Fatalf("expected run %d to be deleted", id)
		}
	}
	for _, id := range []int64{4, 5} {
		if deletedSet[id] {
			t.Fatalf("expected run %d to NOT be deleted", id)
		}
	}
}

func TestRunService_CleanOldRuns_RemovesLogFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-cleanup-logs-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create log files for old runs
	logDir1 := filepath.Join(tmpDir, "logs", "task-a-0101")
	logDir2 := filepath.Join(tmpDir, "logs", "task-b-0101")
	if err := os.MkdirAll(logDir1, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(logDir2, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	logPath1 := filepath.Join(logDir1, "0001.log")
	logPath2 := filepath.Join(logDir2, "0002.log")
	if err := os.WriteFile(logPath1, []byte("old log 1"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(logPath2, []byte("old log 2"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Create a recent run log that should NOT be deleted
	recentLogDir := filepath.Join(tmpDir, "logs", "task-c-0110")
	if err := os.MkdirAll(recentLogDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	recentLogPath := filepath.Join(recentLogDir, "0003.log")
	if err := os.WriteFile(recentLogPath, []byte("recent log"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	oldDate := time.Now().AddDate(0, 0, -10).Format(time.RFC3339)
	recentDate := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)

	sum1, _ := json.Marshal(map[string]any{"stderrFile": logPath1})
	sum2, _ := json.Marshal(map[string]any{"stderrFile": logPath2})
	sum3, _ := json.Marshal(map[string]any{"stderrFile": recentLogPath})

	rec1 := RunRecord{ID: 1, TaskID: 1, Status: "finished", StartedAt: oldDate, Summary: string(sum1)}
	rec2 := RunRecord{ID: 2, TaskID: 1, Status: "finished", StartedAt: oldDate, Summary: string(sum2)}
	rec3 := RunRecord{ID: 3, TaskID: 1, Status: "finished", StartedAt: recentDate, Summary: string(sum3)}

	mock := &runServiceDBMock{
		runsByID: map[int64]RunRecord{
			1: rec1,
			2: rec2,
			3: rec3,
		},
		listRunsPages: map[int][]RunRecord{
			1: {rec1, rec2, rec3},
		},
		listRunsTotal: 3,
	}

	svc := NewRunService(mock)
	deleted, err := svc.CleanOldRuns(7)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if deleted != 2 {
		t.Fatalf("expected 2 deleted, got %d", deleted)
	}

	// Verify old log files were removed
	if _, err := os.Stat(logPath1); !os.IsNotExist(err) {
		t.Fatal("expected old log1 to be removed")
	}
	if _, err := os.Stat(logPath2); !os.IsNotExist(err) {
		t.Fatal("expected old log2 to be removed")
	}
	if _, err := os.Stat(logDir1); !os.IsNotExist(err) {
		t.Fatal("expected old log dir1 to be removed")
	}
	if _, err := os.Stat(logDir2); !os.IsNotExist(err) {
		t.Fatal("expected old log dir2 to be removed")
	}

	// Verify recent log file was NOT removed
	if _, err := os.Stat(recentLogPath); os.IsNotExist(err) {
		t.Fatal("expected recent log to NOT be removed")
	}
}

func TestRunService_CleanOldRuns_ZeroRetention(t *testing.T) {
	mock := &runServiceDBMock{
		listRunsPages: map[int][]RunRecord{
			1: {
				{ID: 1, TaskID: 1, Status: "finished", StartedAt: time.Now().AddDate(0, 0, -30).Format(time.RFC3339)},
			},
		},
		listRunsTotal: 1,
	}

	svc := NewRunService(mock)
	deleted, err := svc.CleanOldRuns(0)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if deleted != 0 {
		t.Fatalf("expected 0 deleted when retention=0, got %d", deleted)
	}
	if len(mock.deletedRun) > 0 {
		t.Fatalf("expected no delete calls when retention=0, got %d", len(mock.deletedRun))
	}
}

func TestRunService_CleanOldRuns_Pagination(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow-cleanup-paging-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create 600 old runs (more than one page of 500)
	oldDate := time.Now().AddDate(0, 0, -10).Format(time.RFC3339)
	page1 := make([]RunRecord, 500)
	page2 := make([]RunRecord, 100)
	for i := 0; i < 500; i++ {
		page1[i] = RunRecord{ID: int64(i + 1), TaskID: 1, Status: "finished", StartedAt: oldDate}
	}
	for i := 0; i < 100; i++ {
		page2[i] = RunRecord{ID: int64(501 + i), TaskID: 1, Status: "finished", StartedAt: oldDate}
	}

	mock := &runServiceDBMock{
		listRunsPages: map[int][]RunRecord{
			1: page1,
			2: page2,
		},
		listRunsTotal: 600,
	}

	svc := NewRunService(mock)
	deleted, err := svc.CleanOldRuns(7)
	if err != nil {
		t.Fatalf("CleanOldRuns() error = %v", err)
	}
	if deleted != 600 {
		t.Fatalf("expected 600 deleted, got %d", deleted)
	}
	if len(mock.deletedRun) != 600 {
		t.Fatalf("expected 600 delete calls, got %d", len(mock.deletedRun))
	}
}

func TestCleanupService_Cleanup_RespectsRetention(t *testing.T) {
	oldDate := time.Now().AddDate(0, 0, -10).Format(time.RFC3339)
	recentDate := time.Now().AddDate(0, 0, -2).Format(time.RFC3339)

	mock := &runServiceDBMock{
		listRunsPages: map[int][]RunRecord{
			1: {
				{ID: 1, TaskID: 1, Status: "finished", StartedAt: oldDate},
				{ID: 2, TaskID: 1, Status: "finished", StartedAt: recentDate},
			},
		},
		listRunsTotal: 2,
	}

	svc := NewRunService(mock)
	cleanupSvc := NewCleanupService(svc, 24*time.Hour, 7)
	cleanupSvc.cleanup()

	// Only the old run should be deleted
	if len(mock.deletedRun) != 1 || mock.deletedRun[0] != 1 {
		t.Fatalf("expected only run 1 deleted, got %#v", mock.deletedRun)
	}
}

func TestCleanupService_Cleanup_SkipsWhenRetentionZero(t *testing.T) {
	oldDate := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)

	mock := &runServiceDBMock{
		listRunsPages: map[int][]RunRecord{
			1: {
				{ID: 1, TaskID: 1, Status: "finished", StartedAt: oldDate},
			},
		},
		listRunsTotal: 1,
	}

	svc := NewRunService(mock)
	cleanupSvc := NewCleanupService(svc, 24*time.Hour, 0)
	cleanupSvc.cleanup()

	if len(mock.deletedRun) > 0 {
		t.Fatalf("expected no deletes when retention=0, got %#v", mock.deletedRun)
	}
}
