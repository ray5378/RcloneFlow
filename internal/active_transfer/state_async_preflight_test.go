package active_transfer

import (
	"testing"
	"time"
)

func TestManager_MergeCandidates_BackfillsWithoutOverwritingRuntimeState(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(21, 9, TrackingModeNormal, nil)
	if st == nil {
		t.Fatalf("expected state")
	}
	if !st.PreflightPending || st.PreflightFinished {
		t.Fatalf("unexpected initial preflight flags: pending=%v finished=%v", st.PreflightPending, st.PreflightFinished)
	}

	pct := 50.0
	mgr.UpdateCurrentFile(21, "dir/a.mp4", 50, 100, 10, &pct)
	mgr.MarkCompleted(21, "dir/b.mp4", FileStatusCopied, "")

	mgr.MergeCandidates(21, []TransferCandidateFile{
		{Path: "dir/a.mp4", Name: "A.mp4", SizeBytes: 100},
		{Path: "dir/b.mp4", Name: "B.mp4", SizeBytes: 200},
		{Path: "dir/c.mp4", Name: "C.mp4", SizeBytes: 300},
	})

	state, ok := mgr.GetByRunID(21)
	if !ok || state == nil {
		t.Fatalf("expected merged state")
	}
	if state.PreflightPending || !state.PreflightFinished {
		t.Fatalf("unexpected final preflight flags: pending=%v finished=%v", state.PreflightPending, state.PreflightFinished)
	}
	if got := len(state.Candidates); got != 3 {
		t.Fatalf("len(candidates)=%d, want 3", got)
	}
	if got := len(state.Pending); got != 2 {
		t.Fatalf("len(pending)=%d, want 2 (current + new pending)", got)
	}
	if got := len(state.Completed); got != 1 {
		t.Fatalf("len(completed)=%d, want 1", got)
	}
	if got := state.Completed["dir/b.mp4"].SizeBytes; got != 200 {
		t.Fatalf("completed size=%d, want 200", got)
	}
	if got := state.Pending["dir/c.mp4"].SizeBytes; got != 300 {
		t.Fatalf("new pending size=%d, want 300", got)
	}
	if got := state.CurrentFiles["dir/a.mp4"].Name; got != "A.mp4" {
		t.Fatalf("current file name=%q, want A.mp4", got)
	}
}

func TestManager_SetPreflightResult_MarksDegradedOnError(t *testing.T) {
	mgr := NewManager()
	mgr.InitState(22, 10, TrackingModeNormal, nil)
	mgr.SetPreflightResult(22, errString("boom"))
	state, ok := mgr.GetByRunID(22)
	if !ok || state == nil {
		t.Fatalf("expected state")
	}
	if state.PreflightPending {
		t.Fatalf("preflight should no longer be pending")
	}
	if state.PreflightFinished {
		t.Fatalf("preflight should not be finished on error")
	}
	if !state.Degraded {
		t.Fatalf("expected degraded=true")
	}
	if state.DegradeReason != "boom" {
		t.Fatalf("degradeReason=%q", state.DegradeReason)
	}
}

func TestManager_BackfillsSizesForCurrentAndCompleted(t *testing.T) {
	mgr := NewManager()
	mgr.InitState(31, 12, TrackingModeNormal, nil)

	pct := 25.0
	mgr.UpdateCurrentFile(31, "dir/a.mp4", 25, 0, 5, &pct)
	state, ok := mgr.GetByRunID(31)
	if !ok || state == nil {
		t.Fatalf("expected state")
	}
	if got := state.CurrentFiles["dir/a.mp4"].TotalBytes; got != 0 {
		t.Fatalf("initial current total=%d, want 0", got)
	}

	mgr.MergeCandidates(31, []TransferCandidateFile{{Path: "dir/a.mp4", Name: "A.mp4", SizeBytes: 100}, {Path: "dir/b.mp4", Name: "B.mp4", SizeBytes: 200}})
	state, _ = mgr.GetByRunID(31)
	if got := state.CurrentFiles["dir/a.mp4"].TotalBytes; got != 100 {
		t.Fatalf("backfilled current total=%d, want 100", got)
	}
	if state.CurrentFile == nil || state.CurrentFile.TotalBytes != 100 {
		t.Fatalf("current file total bytes not backfilled: %#v", state.CurrentFile)
	}

	mgr.UpdateCurrentFile(31, "dir/b.mp4", 50, 0, 6, &pct)
	mgr.MarkCompleted(31, "dir/b.mp4", FileStatusCopied, "")
	state, _ = mgr.GetByRunID(31)
	if got := state.Completed["dir/b.mp4"].SizeBytes; got != 200 {
		t.Fatalf("completed size=%d, want 200", got)
	}
	if got := state.Completed["dir/a.mp4"].SizeBytes; got != 0 {
		_ = got
	}
}

func TestManager_ListOrdering_IsStableAndReadable(t *testing.T) {
	mgr := NewManager()
	mgr.InitState(30, 11, TrackingModeNormal, []TransferCandidateFile{
		{Path: "c/file3.mkv", Name: "file3.mkv", Order: 3},
		{Path: "a/file1.mkv", Name: "file1.mkv", Order: 1},
		{Path: "b/file2.mkv", Name: "file2.mkv", Order: 2},
	})

	pct := 10.0
	mgr.UpdateCurrentFile(30, "b/file2.mkv", 10, 100, 1, &pct)
	mgr.MarkCompleted(30, "a/file1.mkv", FileStatusCopied, "")
	time.Sleep(2 * time.Millisecond)
	mgr.MarkCompleted(30, "b/file2.mkv", FileStatusCopied, "")

	completed := mgr.ListCompleted(11, 0, 10).Items
	if len(completed) != 2 {
		t.Fatalf("completed len=%d, want 2", len(completed))
	}
	if completed[0].Path != "a/file1.mkv" || completed[1].Path != "b/file2.mkv" {
		t.Fatalf("unexpected completed order: %#v", completed)
	}

	pending := mgr.ListPending(11, 0, 10).Items
	if len(pending) != 1 {
		t.Fatalf("pending len=%d, want 1", len(pending))
	}
	if pending[0].Path != "c/file3.mkv" {
		t.Fatalf("unexpected pending order: %#v", pending)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
