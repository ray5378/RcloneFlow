package active_transfer

import "testing"

func TestManager_TransferSlotsAppearInOverviewAndSnapshot(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(101, 202, TrackingModeNormal, nil)
	st.TransferSlots = 4

	overview, ok := mgr.BuildSummary(202, 0, 0, 0, 0, 0)
	if !ok {
		t.Fatal("BuildSummary ok=false")
	}
	if overview.TransferSlots != 4 {
		t.Fatalf("overview transferSlots=%d want 4", overview.TransferSlots)
	}
	if overview.Summary.TransferSlots != 4 {
		t.Fatalf("summary transferSlots=%d want 4", overview.Summary.TransferSlots)
	}

	snap := st.Snapshot()
	if snap.TransferSlots != 4 {
		t.Fatalf("snapshot transferSlots=%d want 4", snap.TransferSlots)
	}
}

func TestManager_TransferSlotsDefaultsToOne(t *testing.T) {
	mgr := NewManager()
	st := mgr.InitState(102, 203, TrackingModeNormal, nil)

	overview, ok := mgr.BuildSummary(203, 0, 0, 0, 0, 0)
	if !ok {
		t.Fatal("BuildSummary ok=false")
	}
	if overview.TransferSlots != 1 {
		t.Fatalf("overview transferSlots=%d want 1", overview.TransferSlots)
	}
	if st.Snapshot().TransferSlots != 1 {
		t.Fatalf("snapshot transferSlots=%d want 1", st.Snapshot().TransferSlots)
	}
}
