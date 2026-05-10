package active_transfer

import (
	"encoding/json"
	"time"
)

type TrackingMode string

const (
	TrackingModeNormal TrackingMode = "normal"
	TrackingModeCAS    TrackingMode = "cas"
)

type FileStatus string

const (
	FileStatusPending    FileStatus = "pending"
	FileStatusInProgress FileStatus = "in_progress"
	FileStatusCopied     FileStatus = "copied"
	FileStatusCASMatched FileStatus = "cas_matched"
	FileStatusSkipped    FileStatus = "skipped"
	FileStatusFailed     FileStatus = "failed"
	FileStatusDeleted    FileStatus = "deleted"
)

type TransferCandidateFile struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	SizeBytes int64  `json:"sizeBytes,omitempty"`
}

type TransferCurrentFile struct {
	Name       string     `json:"name"`
	Path       string     `json:"path,omitempty"`
	Bytes      int64      `json:"bytes,omitempty"`
	TotalBytes int64      `json:"totalBytes,omitempty"`
	Percentage *float64   `json:"percentage,omitempty"`
	Speed      int64      `json:"speed,omitempty"`
	Status     FileStatus `json:"status"`
}

type TransferCompletedFile struct {
	Name      string     `json:"name"`
	Path      string     `json:"path,omitempty"`
	SizeBytes int64      `json:"sizeBytes,omitempty"`
	At        string     `json:"at,omitempty"`
	Status    FileStatus `json:"status"`
	Message   string     `json:"message,omitempty"`
}

type TransferPendingFile struct {
	Name      string     `json:"name"`
	Path      string     `json:"path,omitempty"`
	SizeBytes int64      `json:"sizeBytes,omitempty"`
	Status    FileStatus `json:"status"`
}

type ActiveTransferSummary struct {
	TrackingMode      TrackingMode `json:"trackingMode"`
	CompletedCount    int          `json:"completedCount"`
	PendingCount      int          `json:"pendingCount"`
	TotalCount        int          `json:"totalCount"`
	PreflightPending  bool         `json:"preflightPending,omitempty"`
	PreflightFinished bool         `json:"preflightFinished,omitempty"`
	Percentage        float64      `json:"percentage,omitempty"`
	Bytes             int64        `json:"bytes,omitempty"`
	TotalBytes        int64        `json:"totalBytes,omitempty"`
	Speed             int64        `json:"speed,omitempty"`
	Eta               int64        `json:"eta,omitempty"`
}

type ActiveTransferState struct {
	RunID         int64
	TaskID        int64
	TrackingMode  TrackingMode
	Candidates    map[string]TransferCandidateFile
	CurrentFile   *TransferCurrentFile
	CurrentFiles  map[string]TransferCurrentFile
	Completed     map[string]TransferCompletedFile
	Pending       map[string]TransferPendingFile
	Degraded           bool
	DegradeReason      string
	PreflightPending   bool
	PreflightFinished  bool
	StartedAt          time.Time
	UpdatedAt          time.Time
}

type ActiveTransferOverviewResponse struct {
	TaskID        int64                  `json:"taskId"`
	RunID         int64                  `json:"runId"`
	TrackingMode  TrackingMode           `json:"trackingMode"`
	Summary       ActiveTransferSummary  `json:"summary"`
	CurrentFile   *TransferCurrentFile   `json:"currentFile,omitempty"`
	CurrentFiles  []TransferCurrentFile  `json:"currentFiles,omitempty"`
	Degraded      bool                   `json:"degraded,omitempty"`
	DegradeReason string                 `json:"degradeReason,omitempty"`
}

type ActiveTransferListResponse[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}

type ActiveTransferSnapshot struct {
	RunID         int64                    `json:"runId"`
	TaskID        int64                    `json:"taskId"`
	TrackingMode  TrackingMode             `json:"trackingMode"`
	TotalCount    int                      `json:"totalCount"`
	CurrentFile   *TransferCurrentFile     `json:"currentFile,omitempty"`
	CurrentFiles  []TransferCurrentFile    `json:"currentFiles,omitempty"`
	Completed     []TransferCompletedFile  `json:"completed,omitempty"`
	Pending       []TransferPendingFile    `json:"pending,omitempty"`
	Degraded          bool                     `json:"degraded,omitempty"`
	DegradeReason     string                   `json:"degradeReason,omitempty"`
	PreflightPending  bool                     `json:"preflightPending,omitempty"`
	PreflightFinished bool                     `json:"preflightFinished,omitempty"`
	StartedAt         string                   `json:"startedAt,omitempty"`
	UpdatedAt         string                   `json:"updatedAt,omitempty"`
}

func (s *ActiveTransferState) Snapshot() ActiveTransferSnapshot {
	if s == nil {
		return ActiveTransferSnapshot{}
	}
	completed := make([]TransferCompletedFile, 0, len(s.Completed))
	for _, v := range s.Completed {
		completed = append(completed, v)
	}
	pending := make([]TransferPendingFile, 0, len(s.Pending))
	for _, v := range s.Pending {
		pending = append(pending, v)
	}
	currentFiles := make([]TransferCurrentFile, 0, len(s.CurrentFiles))
	for _, v := range s.CurrentFiles {
		currentFiles = append(currentFiles, v)
	}
	return ActiveTransferSnapshot{
		RunID:         s.RunID,
		TaskID:        s.TaskID,
		TrackingMode:  s.TrackingMode,
		TotalCount:    len(s.Candidates),
		CurrentFile:   cloneCurrent(s.CurrentFile),
		CurrentFiles:  currentFiles,
		Completed:     completed,
		Pending:       pending,
		Degraded:          s.Degraded,
		DegradeReason:     s.DegradeReason,
		PreflightPending:  s.PreflightPending,
		PreflightFinished: s.PreflightFinished,
		StartedAt:         s.StartedAt.Format(time.RFC3339),
		UpdatedAt:         s.UpdatedAt.Format(time.RFC3339),
	}
}

func RestoreStateFromSnapshot(snap ActiveTransferSnapshot) *ActiveTransferState {
	st := &ActiveTransferState{
		RunID:         snap.RunID,
		TaskID:        snap.TaskID,
		TrackingMode:  snap.TrackingMode,
		Candidates:    map[string]TransferCandidateFile{},
		Completed:     map[string]TransferCompletedFile{},
		Pending:       map[string]TransferPendingFile{},
		CurrentFile:   cloneCurrent(snap.CurrentFile),
		CurrentFiles:  map[string]TransferCurrentFile{},
		Degraded:          snap.Degraded,
		DegradeReason:     snap.DegradeReason,
		PreflightPending:  snap.PreflightPending,
		PreflightFinished: snap.PreflightFinished,
	}
	for _, v := range snap.CurrentFiles {
		key := normalizePath(v.Path)
		if key == "" {
			key = normalizePath(v.Name)
		}
		v.Path = key
		st.CurrentFiles[key] = v
	}
	for _, v := range snap.Pending {
		key := normalizePath(v.Path)
		if key == "" {
			key = normalizePath(v.Name)
		}
		v.Path = key
		st.Pending[key] = v
		st.Candidates[key] = TransferCandidateFile{Path: key, Name: v.Name, SizeBytes: v.SizeBytes}
	}
	for _, v := range snap.Completed {
		key := normalizePath(v.Path)
		if key == "" {
			key = normalizePath(v.Name)
		}
		v.Path = key
		st.Completed[key] = v
		if _, ok := st.Candidates[key]; !ok {
			st.Candidates[key] = TransferCandidateFile{Path: key, Name: v.Name, SizeBytes: v.SizeBytes}
		}
	}
	for len(st.Candidates) < snap.TotalCount {
		ghost := "__ghost__/" + time.Now().Format("150405.000000000")
		st.Candidates[ghost] = TransferCandidateFile{Path: ghost, Name: ghost}
	}
	if t, err := time.Parse(time.RFC3339, snap.StartedAt); err == nil {
		st.StartedAt = t
	}
	if t, err := time.Parse(time.RFC3339, snap.UpdatedAt); err == nil {
		st.UpdatedAt = t
	}
	if st.StartedAt.IsZero() {
		st.StartedAt = time.Now()
	}
	if st.UpdatedAt.IsZero() {
		st.UpdatedAt = time.Now()
	}
	return st
}

func SnapshotEnvelope(snap ActiveTransferSnapshot) map[string]any {
	b, _ := json.Marshal(snap)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return map[string]any{"activeTransfer": m}
}

func SnapshotFromSummary(summary string) (*ActiveTransferState, bool) {
	if summary == "" {
		return nil, false
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(summary), &raw); err != nil {
		return nil, false
	}
	at, _ := raw["activeTransfer"]
	if at == nil {
		return nil, false
	}
	b, err := json.Marshal(at)
	if err != nil {
		return nil, false
	}
	var snap ActiveTransferSnapshot
	if err := json.Unmarshal(b, &snap); err != nil {
		return nil, false
	}
	return RestoreStateFromSnapshot(snap), true
}
