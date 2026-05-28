package active_transfer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSnapshot_SkipsSortWhenNotDirty(t *testing.T) {
	now := time.Now()
	state := &ActiveTransferState{
		RunID:          1,
		TaskID:         2,
		TrackingMode:   TrackingModeNormal,
		TotalCount:     3,
		CompletedCount: 1,
		PendingCount:   2,
		Completed: map[string]TransferCompletedFile{
			"file1.txt": {Path: "file1.txt", Name: "file1.txt", SizeBytes: 100, Order: 2, At: now.Format(time.RFC3339), Status: FileStatusCopied},
		},
		Pending: map[string]TransferPendingFile{
			"file2.txt": {Path: "file2.txt", Name: "file2.txt", SizeBytes: 200, Order: 1, Status: FileStatusPending},
			"file3.txt": {Path: "file3.txt", Name: "file3.txt", SizeBytes: 300, Order: 3, Status: FileStatusPending},
		},
		CurrentFiles: map[string]TransferCurrentFile{
			"file4.txt": {Path: "file4.txt", Name: "file4.txt", Bytes: 50, TotalBytes: 100, Order: 1, Status: FileStatusInProgress},
		},
		dirtySort: false,
	}

	snap1 := state.Snapshot()
	state.dirtySort = false
	snap2 := state.Snapshot()

	assert.Equal(t, snap1.Completed, snap2.Completed)
	assert.Equal(t, snap1.Pending, snap2.Pending)
	assert.Equal(t, snap1.CurrentFiles, snap2.CurrentFiles)
}

func TestSnapshot_SortsWhenDirty(t *testing.T) {
	now := time.Now()
	state := &ActiveTransferState{
		RunID:          1,
		TaskID:         2,
		TrackingMode:   TrackingModeNormal,
		TotalCount:     2,
		CompletedCount: 2,
		Completed: map[string]TransferCompletedFile{
			"file2.txt": {Path: "file2.txt", Name: "file2.txt", SizeBytes: 100, Order: 2, At: now.Format(time.RFC3339), Status: FileStatusCopied},
			"file1.txt": {Path: "file1.txt", Name: "file1.txt", SizeBytes: 50, Order: 1, At: now.Format(time.RFC3339), Status: FileStatusCopied},
		},
		Pending: map[string]TransferPendingFile{
			"file4.txt": {Path: "file4.txt", Name: "file4.txt", SizeBytes: 400, Order: 2, Status: FileStatusPending},
			"file3.txt": {Path: "file3.txt", Name: "file3.txt", SizeBytes: 300, Order: 1, Status: FileStatusInProgress},
		},
		dirtySort: true,
	}

	snap := state.Snapshot()
	assert.Equal(t, 2, len(snap.Completed))
	assert.Equal(t, "file1.txt", snap.Completed[0].Path)
	assert.Equal(t, "file2.txt", snap.Completed[1].Path)

	assert.Equal(t, 2, len(snap.Pending))
	assert.Equal(t, FileStatusInProgress, snap.Pending[0].Status)
	assert.Equal(t, "file3.txt", snap.Pending[0].Path)

	assert.False(t, state.dirtySort, "dirtySort should be reset after sorting")
}

func TestSnapshot_CurrentFilesSortedByOrder(t *testing.T) {
	state := &ActiveTransferState{
		RunID:        1,
		TaskID:       2,
		TrackingMode: TrackingModeNormal,
		CurrentFiles: map[string]TransferCurrentFile{
			"z.txt": {Path: "z.txt", Name: "z.txt", Order: 3},
			"a.txt": {Path: "a.txt", Name: "a.txt", Order: 1},
			"m.txt": {Path: "m.txt", Name: "m.txt", Order: 2},
		},
		dirtySort: true,
	}

	snap := state.Snapshot()
	assert.Equal(t, 3, len(snap.CurrentFiles))
	assert.Equal(t, "a.txt", snap.CurrentFiles[0].Path)
	assert.Equal(t, "m.txt", snap.CurrentFiles[1].Path)
	assert.Equal(t, "z.txt", snap.CurrentFiles[2].Path)
}
