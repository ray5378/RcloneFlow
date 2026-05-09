package active_transfer

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type PersistFunc func(runID int64, snap ActiveTransferSnapshot)

type Manager struct {
	mu      sync.RWMutex
	byRunID map[int64]*ActiveTransferState
	byTask  map[int64]*ActiveTransferState
	persist PersistFunc
}

func NewManager() *Manager {
	return &Manager{byRunID: map[int64]*ActiveTransferState{}, byTask: map[int64]*ActiveTransferState{}}
}

func (m *Manager) SetPersistFunc(fn PersistFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.persist = fn
}

func (m *Manager) persistSnapshotLocked(st *ActiveTransferState) {
	if m == nil || st == nil || m.persist == nil {
		return
	}
	snap := st.Snapshot()
	go m.persist(st.RunID, snap)
}

func (m *Manager) RestoreState(st *ActiveTransferState) {
	if m == nil || st == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byRunID[st.RunID] = st
	m.byTask[st.TaskID] = st
}

func (m *Manager) InitState(runID, taskID int64, mode TrackingMode, candidates []TransferCandidateFile) *ActiveTransferState {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := &ActiveTransferState{
		RunID:        runID,
		TaskID:       taskID,
		TrackingMode: mode,
		Candidates:   map[string]TransferCandidateFile{},
		Completed:    map[string]TransferCompletedFile{},
		Pending:      map[string]TransferPendingFile{},
		StartedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	for _, c := range candidates {
		key := normalizePath(c.Path)
		if key == "" {
			continue
		}
		c.Path = key
		if strings.TrimSpace(c.Name) == "" {
			c.Name = baseName(key)
		}
		st.Candidates[key] = c
		st.Pending[key] = TransferPendingFile{Path: key, Name: c.Name, SizeBytes: c.SizeBytes, Status: FileStatusPending}
	}
	m.byRunID[runID] = st
	m.byTask[taskID] = st
	m.persistSnapshotLocked(st)
	return st
}

func (m *Manager) GetByTaskID(taskID int64) (*ActiveTransferState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, ok := m.byTask[taskID]
	return st, ok
}

func (m *Manager) GetByRunID(runID int64) (*ActiveTransferState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, ok := m.byRunID[runID]
	return st, ok
}

func (m *Manager) RemoveState(runID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.byRunID[runID]
	if !ok {
		return
	}
	delete(m.byRunID, runID)
	delete(m.byTask, st.TaskID)
}

func (m *Manager) UpdateCurrentFile(runID int64, path string, bytes, total, speed int64, pct *float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.byRunID[runID]
	if !ok {
		return
	}
	key := normalizePath(path)
	name := baseName(key)
	if c, ok := st.Candidates[key]; ok && c.Name != "" {
		name = c.Name
	}
	st.CurrentFile = &TransferCurrentFile{Path: key, Name: name, Bytes: bytes, TotalBytes: total, Speed: speed, Percentage: pct, Status: FileStatusInProgress}
	if p, ok := st.Pending[key]; ok {
		p.Status = FileStatusInProgress
		st.Pending[key] = p
	} else {
		st.Pending[key] = TransferPendingFile{Path: key, Name: name, Status: FileStatusInProgress}
	}
	st.UpdatedAt = time.Now()
	m.persistSnapshotLocked(st)
}

func (m *Manager) MarkCompleted(runID int64, path string, status FileStatus, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.byRunID[runID]
	if !ok {
		return
	}
	key := normalizePath(path)
	name := baseName(key)
	size := int64(0)
	if c, ok := st.Candidates[key]; ok {
		if c.Name != "" {
			name = c.Name
		}
		size = c.SizeBytes
	}
	delete(st.Pending, key)
	st.Completed[key] = TransferCompletedFile{Path: key, Name: name, SizeBytes: size, At: time.Now().Format(time.RFC3339), Status: status, Message: message}
	if st.CurrentFile != nil && normalizePath(st.CurrentFile.Path) == key {
		st.CurrentFile = nil
	}
	st.UpdatedAt = time.Now()
	m.persistSnapshotLocked(st)
}

func (m *Manager) BuildSummary(taskID int64, bytes, total, speed, eta int64, percentage float64) (ActiveTransferOverviewResponse, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, ok := m.byTask[taskID]
	if !ok {
		return ActiveTransferOverviewResponse{}, false
	}
	return ActiveTransferOverviewResponse{
		TaskID:       st.TaskID,
		RunID:        st.RunID,
		TrackingMode: st.TrackingMode,
		CurrentFile:  cloneCurrent(st.CurrentFile),
		Summary: ActiveTransferSummary{
			TrackingMode:   st.TrackingMode,
			CompletedCount: len(st.Completed),
			PendingCount:   len(st.Pending),
			TotalCount:     len(st.Candidates),
			Bytes:          bytes,
			TotalBytes:     total,
			Speed:          speed,
			Eta:            eta,
			Percentage:     percentage,
		},
		Degraded:      st.Degraded,
		DegradeReason: st.DegradeReason,
	}, true
}

func (m *Manager) ListCompleted(taskID int64, offset, limit int) ActiveTransferListResponse[TransferCompletedFile] {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st := m.byTask[taskID]
	if st == nil {
		return ActiveTransferListResponse[TransferCompletedFile]{Total: 0, Items: []TransferCompletedFile{}}
	}
	items := make([]TransferCompletedFile, 0, len(st.Completed))
	for _, v := range st.Completed {
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].At > items[j].At })
	return paginate(items, offset, limit)
}

func (m *Manager) ListPending(taskID int64, offset, limit int) ActiveTransferListResponse[TransferPendingFile] {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st := m.byTask[taskID]
	if st == nil {
		return ActiveTransferListResponse[TransferPendingFile]{Total: 0, Items: []TransferPendingFile{}}
	}
	items := make([]TransferPendingFile, 0, len(st.Pending))
	for _, v := range st.Pending {
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return paginate(items, offset, limit)
}

func cloneCurrent(v *TransferCurrentFile) *TransferCurrentFile {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func paginate[T any](items []T, offset, limit int) ActiveTransferListResponse[T] {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	}
	total := len(items)
	if offset >= total {
		return ActiveTransferListResponse[T]{Total: total, Items: []T{}}
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return ActiveTransferListResponse[T]{Total: total, Items: items[offset:end]}
}

func normalizePath(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\\", "/"))
	s = strings.TrimPrefix(s, "./")
	return s
}

func baseName(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "/")
	return parts[len(parts)-1]
}
