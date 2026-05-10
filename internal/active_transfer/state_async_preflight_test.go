package active_transfer

import "testing"

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

type errString string

func (e errString) Error() string { return string(e) }
