package active_transfer

import (
	"testing"
	"time"
)

func TestManager_OnFileProgress(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})

	pct := 10.0
	mgr.OnFileProgress(st.RunID, "a.bin", 10, 100, 1, &pct)

	got, ok := mgr.GetByRunID(st.RunID)
	if !ok {
		t.Fatalf("expected run state")
	}
	if got.CurrentFile == nil {
		t.Fatalf("expected current file")
	}
}

func TestManager_OnFileCopied(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})

	mgr.OnFileCopied(st.RunID, "a.bin")

	got, ok := mgr.GetByRunID(st.RunID)
	if !ok {
		t.Fatalf("expected run state")
	}
	if got.CompletedCount != 1 {
		t.Fatalf("expected 1 completed, got %d", got.CompletedCount)
	}
}

func TestManager_OnFileCASMatched(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})

	mgr.OnFileCASMatched(st.RunID, "a.bin")

	got, ok := mgr.GetByRunID(st.RunID)
	if !ok {
		t.Fatalf("expected run state")
	}
	if got.CompletedCount != 1 {
		t.Fatalf("expected 1 completed, got %d", got.CompletedCount)
	}
}

func TestManager_OnFileFailed(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})

	mgr.OnFileFailed(st.RunID, "a.bin", "test error")

	got, ok := mgr.GetByRunID(st.RunID)
	if !ok {
		t.Fatalf("expected run state")
	}
	if got.CompletedCount != 1 {
		t.Fatalf("expected 1 completed, got %d", got.CompletedCount)
	}
}

func TestManager_OnFileSkipped(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})

	mgr.OnFileSkipped(st.RunID, "a.bin", "test skipped")

	got, ok := mgr.GetByRunID(st.RunID)
	if !ok {
		t.Fatalf("expected run state")
	}
	if got.CompletedCount != 1 {
		t.Fatalf("expected 1 completed, got %d", got.CompletedCount)
	}
}

func TestManager_OnFileDeleted(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})

	mgr.OnFileDeleted(st.RunID, "a.bin")

	got, ok := mgr.GetByRunID(st.RunID)
	if !ok {
		t.Fatalf("expected run state")
	}
	if got.CompletedCount != 1 {
		t.Fatalf("expected 1 completed, got %d", got.CompletedCount)
	}
}

func TestManager_GetByTaskID(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})

	got, ok := mgr.GetByTaskID(st.TaskID)
	if !ok {
		t.Fatalf("expected to get task state")
	}
	if got.RunID != st.RunID {
		t.Fatalf("expected run id %d, got %d", st.RunID, got.RunID)
	}
}

func TestManager_RemoveState(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})

	mgr.RemoveState(st.RunID)

	_, ok := mgr.GetByRunID(st.RunID)
	if ok {
		t.Fatalf("expected run state removed")
	}
}

func TestManager_SetTransferSlots(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})

	mgr.SetTransferSlots(st.RunID, 8)

	got, ok := mgr.GetByRunID(st.RunID)
	if !ok {
		t.Fatalf("expected to get task state")
	}
	if got.TransferSlots != 8 {
		t.Fatalf("expected transfer slots 8, got %d", got.TransferSlots)
	}
}

func TestManager_SetPersistThrottle(t *testing.T) {
	mgr := NewManager()
	mgr.SetPersistThrottle(100 * time.Millisecond)
}

func TestActiveTransferState_Snapshot(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{
		{Path: "a.bin", Name: "a.bin", SizeBytes: 100},
		{Path: "b.txt", Name: "b.txt", SizeBytes: 200, Order: 1},
	})
	pct := 10.0
	mgr.UpdateCurrentFile(st.RunID, "a.bin", 10, 100, 1, &pct)
	mgr.MarkCompleted(st.RunID, "b.txt", FileStatusCopied, "")
	mgr.MarkCompleted(st.RunID, "a.bin", FileStatusFailed, "test")

	snap := st.Snapshot()
	if snap.RunID != st.RunID {
		t.Fatalf("RunID mismatch")
	}
	if snap.TaskID != st.TaskID {
		t.Fatalf("TaskID mismatch")
	}
	if snap.TotalCount != st.TotalCount {
		t.Fatalf("TotalCount mismatch")
	}
	if len(snap.Completed) != 2 {
		t.Fatalf("Expected 2 completed, got %d", len(snap.Completed))
	}
	if len(snap.Pending) != 0 {
		t.Fatalf("Expected 0 pending, got %d", len(snap.Pending))
	}
}

func TestRestoreStateFromSnapshot(t *testing.T) {
	// Test case for RestoreStateFromSnapshot
	snap := ActiveTransferSnapshot{
		RunID:            202,
		TaskID:           101,
		TrackingMode:     TrackingModeNormal,
		TotalCount:       2,
		CompletedCount:   1,
		PendingCount:    1,
		TransferSlots:  8,
		CurrentFile:     &TransferCurrentFile{Path: "a.bin", Name: "a.bin", Bytes: 10, TotalBytes: 100, Status: FileStatusInProgress},
		CurrentFiles:    []TransferCurrentFile{{Path: "a.bin", Name: "a.bin", Bytes: 10, TotalBytes: 100, Status: FileStatusInProgress}},
		Completed:       []TransferCompletedFile{{Path: "b.txt", Name: "b.txt", SizeBytes: 200, Status: FileStatusCopied, Order: 1}},
		Pending:         []TransferPendingFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100, Status: FileStatusInProgress}},
		Degraded:         false,
		DegradeReason:   "",
		PreflightPending: true,
		PreflightFinished: false,
	}

	restored := RestoreStateFromSnapshot(snap)
	if restored == nil {
		t.Fatalf("Restored state is nil")
	}
	if restored.RunID != 202 {
		t.Fatalf("Restored RunID mismatch")
	}
	if restored.TotalCount != 2 {
		t.Fatalf("TotalCount mismatch")
	}
	if len(restored.Completed) != 1 {
		t.Fatalf("Completed count mismatch")
	}
}

func TestActiveTransferState_SnapshotNil(t *testing.T) {
	var st *ActiveTransferState = nil
	snap := st.Snapshot()
	if snap.RunID != 0 {
		t.Fatalf("Expected RunID should be 0 for nil snapshot")
	}
}

func TestManager_RestoreState(t *testing.T) {
	mgr := NewManager()
	snap := ActiveTransferSnapshot{
		RunID:            202,
		TaskID:           101,
		TrackingMode:     TrackingModeNormal,
		TotalCount:       2,
		CompletedCount:   1,
		PendingCount:    1,
	}
	st := RestoreStateFromSnapshot(snap)
	mgr.RestoreState(st)

	got, ok := mgr.GetByRunID(202)
	if !ok {
		t.Fatalf("Expected to retrieve restored state by run id")
	}
	if got.RunID != 202 {
		t.Fatalf("Unexpected run id mismatch")
	}
}

func TestManager_EmitPersist(t *testing.T) {
	mgr := NewManager()
	mgr.SetPersistFunc(func(runID int64, snap ActiveTransferSnapshot) {
	})
	st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})
	pct := 10.0
	mgr.UpdateCurrentFile(st.RunID, "a.bin", 10, 100, 1, &pct)
}
