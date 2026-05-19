package active_transfer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNormalizedTransferSlots(t *testing.T) {
	assert.Equal(t, 1, normalizedTransferSlots(0))
	assert.Equal(t, 1, normalizedTransferSlots(-1))
	assert.Equal(t, 1, normalizedTransferSlots(1))
	assert.Equal(t, 4, normalizedTransferSlots(4))
}

func TestSnapshotEnvelope(t *testing.T) {
	snap := ActiveTransferSnapshot{
		RunID:        1,
		TaskID:       2,
		TrackingMode: TrackingModeNormal,
		TotalCount:   10,
	}
	env := SnapshotEnvelope(snap)
	assert.NotNil(t, env["activeTransfer"])
}

func TestSnapshotFromSummary_Empty(t *testing.T) {
	_, ok := SnapshotFromSummary("")
	assert.False(t, ok)
}

func TestSnapshotFromSummary_InvalidJSON(t *testing.T) {
	_, ok := SnapshotFromSummary("{invalid}")
	assert.False(t, ok)
}

func TestSnapshotFromSummary_NoActiveTransfer(t *testing.T) {
	_, ok := SnapshotFromSummary(`{"other": "data"}`)
	assert.False(t, ok)
}

func TestSnapshotFromSummary_Valid(t *testing.T) {
	summary := `{"activeTransfer":{"runId":1,"taskId":2,"trackingMode":"normal","totalCount":5}}`
	state, ok := SnapshotFromSummary(summary)
	assert.True(t, ok)
	assert.NotNil(t, state)
	assert.Equal(t, int64(1), state.RunID)
	assert.Equal(t, int64(2), state.TaskID)
	assert.Equal(t, TrackingModeNormal, state.TrackingMode)
	assert.Equal(t, 5, state.TotalCount)
}

func TestActiveTransferState_Snapshot_Nil(t *testing.T) {
	var state *ActiveTransferState
	snap := state.Snapshot()
	assert.Equal(t, ActiveTransferSnapshot{}, snap)
}

func TestActiveTransferState_Snapshot_Basic(t *testing.T) {
	now := time.Now()
	state := &ActiveTransferState{
		RunID:          1,
		TaskID:         2,
		TrackingMode:   TrackingModeCAS,
		TotalCount:     10,
		CompletedCount: 3,
		PendingCount:   7,
		TransferSlots:  4,
		StartedAt:      now,
		UpdatedAt:      now,
		CurrentFile: &TransferCurrentFile{
			Name:   "test.txt",
			Status: FileStatusInProgress,
		},
	}
	snap := state.Snapshot()
	assert.Equal(t, int64(1), snap.RunID)
	assert.Equal(t, int64(2), snap.TaskID)
	assert.Equal(t, TrackingModeCAS, snap.TrackingMode)
	assert.Equal(t, 10, snap.TotalCount)
	assert.Equal(t, 3, snap.CompletedCount)
	assert.Equal(t, 7, snap.PendingCount)
	assert.Equal(t, 4, snap.TransferSlots)
	assert.NotNil(t, snap.CurrentFile)
	assert.Equal(t, "test.txt", snap.CurrentFile.Name)
}

func TestRestoreStateFromSnapshot_Basic(t *testing.T) {
	snap := ActiveTransferSnapshot{
		RunID:        1,
		TaskID:       2,
		TrackingMode: TrackingModeNormal,
		TotalCount:   5,
	}
	state := RestoreStateFromSnapshot(snap)
	assert.Equal(t, int64(1), state.RunID)
	assert.Equal(t, int64(2), state.TaskID)
	assert.Equal(t, TrackingModeNormal, state.TrackingMode)
	assert.Equal(t, 5, state.TotalCount)
	assert.NotNil(t, state.Candidates)
	assert.NotNil(t, state.Completed)
	assert.NotNil(t, state.Pending)
	assert.NotNil(t, state.CurrentFiles)
}

func TestRestoreStateFromSnapshot_WithFiles(t *testing.T) {
	snap := ActiveTransferSnapshot{
		RunID:        1,
		TaskID:       2,
		TrackingMode: TrackingModeCAS,
		TotalCount:   3,
		CurrentFiles: []TransferCurrentFile{
			{Name: "file1.txt", Path: "file1.txt", Order: 1},
		},
		Completed: []TransferCompletedFile{
			{Name: "file2.txt", Path: "file2.txt", Order: 2, Status: FileStatusCopied},
		},
		Pending: []TransferPendingFile{
			{Name: "file3.txt", Path: "file3.txt", Order: 3, Status: FileStatusPending},
		},
	}
	state := RestoreStateFromSnapshot(snap)
	assert.Equal(t, 3, state.TotalCount)
	assert.Len(t, state.CurrentFiles, 1)
	assert.Len(t, state.Completed, 1)
	assert.Len(t, state.Pending, 1)
	assert.Equal(t, 4, state.NextOrder)
	assert.Equal(t, 3, state.NextCompletedOrder)
}
