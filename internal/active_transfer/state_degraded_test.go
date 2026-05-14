package active_transfer

import "testing"

func TestManager_DegradedRetainsCountsButTrimsMaps(t *testing.T) {
	mgr := NewManager()
	candidates := make([]TransferCandidateFile, 0, degradeCandidateThreshold+50)
	for i := 0; i < degradeCandidateThreshold+50; i++ {
		candidates = append(candidates, TransferCandidateFile{Path: baseNamePath(i), Name: baseNamePath(i)})
	}
	st := mgr.InitState(1, 2, TrackingModeNormal, candidates)
	if !st.Degraded {
		t.Fatalf("expected degraded=true for large candidate set")
	}
	if st.TotalCount != len(candidates) {
		t.Fatalf("totalCount=%d want=%d", st.TotalCount, len(candidates))
	}
	if len(st.Pending) > retainedPendingLimit {
		t.Fatalf("retained pending=%d exceeds limit=%d", len(st.Pending), retainedPendingLimit)
	}

	for i := 0; i < degradeCandidateThreshold+50; i++ {
		mgr.MarkCompleted(1, baseNamePath(i), FileStatusCopied, "")
	}
	got, ok := mgr.GetByRunID(1)
	if !ok {
		t.Fatalf("expected run state")
	}
	if got.CompletedCount != len(candidates) {
		t.Fatalf("completedCount=%d want=%d", got.CompletedCount, len(candidates))
	}
	if got.PendingCount != 0 {
		t.Fatalf("pendingCount=%d want 0", got.PendingCount)
	}
	if len(got.Completed) > retainedCompletedLimit {
		t.Fatalf("retained completed=%d exceeds limit=%d", len(got.Completed), retainedCompletedLimit)
	}

	resp := mgr.ListCompleted(2, 0, 20)
	if resp.Total != len(candidates) {
		t.Fatalf("completed total=%d want=%d", resp.Total, len(candidates))
	}
}

func baseNamePath(i int) string {
	return "dir/file-" + string(rune('a'+(i%26))) + "-" + itoa(i)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	buf := make([]byte, 0, 16)
	for i > 0 {
		buf = append([]byte{byte('0' + (i % 10))}, buf...)
		i /= 10
	}
	return string(buf)
}
