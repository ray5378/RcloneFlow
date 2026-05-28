package active_transfer

import (
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"
)

type PersistFunc func(runID int64, snap ActiveTransferSnapshot)
type PersistErrorFunc func(runID int64, snap ActiveTransferSnapshot, err error)

const (
	degradeCandidateThreshold = 2000
	retainedCompletedLimit    = 200
	retainedPendingLimit      = 200
)

type Manager struct {
	mu      sync.RWMutex
	byRunID map[int64]*ActiveTransferState
	byTask  map[int64]*ActiveTransferState
	persist PersistFunc
	onError PersistErrorFunc

	persistThrottle time.Duration
	pendingPersist  map[int64]ActiveTransferSnapshot
	persistTimers   map[int64]*time.Timer
}

func NewManager() *Manager {
	return &Manager{
		byRunID:          map[int64]*ActiveTransferState{},
		byTask:           map[int64]*ActiveTransferState{},
		persistThrottle:  250 * time.Millisecond,
		pendingPersist:   map[int64]ActiveTransferSnapshot{},
		persistTimers:    map[int64]*time.Timer{},
	}
}

func (m *Manager) SetPersistFunc(fn PersistFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.persist = fn
}

func (m *Manager) SetPersistErrorFunc(fn PersistErrorFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onError = fn
}

func (m *Manager) SetPersistThrottle(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.persistThrottle = d
}

func (m *Manager) emitPersist(runID int64, snap ActiveTransferSnapshot) {
	m.mu.RLock()
	persist := m.persist
	onError := m.onError
	m.mu.RUnlock()
	if persist == nil {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic in persist", zap.Any("panic", r))
				if onError != nil {
					onError(runID, snap, &PersistPanicError{Panic: r})
				}
			}
		}()
		persist(runID, snap)
	}()
}

func (m *Manager) flushPendingPersist(runID int64) {
	m.mu.Lock()
	snap, ok := m.pendingPersist[runID]
	if !ok {
		delete(m.persistTimers, runID)
		m.mu.Unlock()
		return
	}
	delete(m.pendingPersist, runID)
	delete(m.persistTimers, runID)
	m.mu.Unlock()
	m.emitPersist(runID, snap)
}

type PersistPanicError struct {
	Panic interface{}
}

func (e *PersistPanicError) Error() string {
	return "persist panic"
}

func (m *Manager) persistSnapshotLocked(st *ActiveTransferState) {
	m.persistSnapshotLockedMode(st, false)
}

func (m *Manager) persistSnapshotLockedImmediate(st *ActiveTransferState) {
	m.persistSnapshotLockedMode(st, true)
}

func (m *Manager) persistSnapshotLockedMode(st *ActiveTransferState, immediate bool) {
	if m == nil || st == nil || m.persist == nil {
		return
	}
	snap := st.Snapshot()
	runID := st.RunID
	if immediate || m.persistThrottle <= 0 {
		if timer, ok := m.persistTimers[runID]; ok {
			timer.Stop()
			delete(m.persistTimers, runID)
		}
		delete(m.pendingPersist, runID)
		go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine panic in persist", zap.Any("panic", r))
				m.mu.RLock()
				onError := m.onError
				m.mu.RUnlock()
				if onError != nil {
					onError(runID, snap, &PersistPanicError{Panic: r})
				}
			}
		}()
		m.persist(runID, snap)
	}()
		return
	}
	m.pendingPersist[runID] = snap
	if _, exists := m.persistTimers[runID]; exists {
		return
	}
	m.persistTimers[runID] = time.AfterFunc(m.persistThrottle, func() {
		m.flushPendingPersist(runID)
	})
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
		RunID:              runID,
		TaskID:             taskID,
		TrackingMode:       mode,
		Candidates:         map[string]TransferCandidateFile{},
		CurrentFiles:       map[string]TransferCurrentFile{},
		Completed:          map[string]TransferCompletedFile{},
		Pending:            map[string]TransferPendingFile{},
		TotalCount:         len(candidates),
		CompletedCount:     0,
		PendingCount:       len(candidates),
		TransferSlots:      1,
		PreflightPending:   true,
		PreflightFinished:  len(candidates) > 0,
		NextOrder:          1,
		NextCompletedOrder: 1,
		StartedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if len(candidates) > degradeCandidateThreshold {
		st.Degraded = true
		st.DegradeReason = "large transfer set; retaining only recent completed/pending items in memory"
	}
	if len(candidates) == 0 {
		st.PreflightFinished = false
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
		if c.Order <= 0 {
			c.Order = st.NextOrder
			st.NextOrder++
		} else if c.Order >= st.NextOrder {
			st.NextOrder = c.Order + 1
		}
		st.Candidates[key] = c
		st.Pending[key] = TransferPendingFile{Path: key, Name: c.Name, SizeBytes: c.SizeBytes, Status: FileStatusPending, Order: c.Order}
		trimPendingRetainedLocked(st)
	}
	m.byRunID[runID] = st
	m.byTask[taskID] = st
	m.persistSnapshotLockedImmediate(st)
	return st
}

func (m *Manager) MergeCandidates(runID int64, candidates []TransferCandidateFile) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.byRunID[runID]
	if !ok || st == nil {
		return
	}
	if st.Candidates == nil {
		st.Candidates = map[string]TransferCandidateFile{}
	}
	if st.Pending == nil {
		st.Pending = map[string]TransferPendingFile{}
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
		if prev, ok := st.Candidates[key]; ok && prev.Order > 0 && c.Order <= 0 {
			c.Order = prev.Order
		}
		if c.Order <= 0 {
			c.Order = st.NextOrder
			st.NextOrder++
		} else if c.Order >= st.NextOrder {
			st.NextOrder = c.Order + 1
		}
		if _, existed := st.Candidates[key]; !existed {
			st.TotalCount++
			st.PendingCount++
			if st.TotalCount > degradeCandidateThreshold {
				st.Degraded = true
				if strings.TrimSpace(st.DegradeReason) == "" {
					st.DegradeReason = "large transfer set; retaining only recent completed/pending items in memory"
				}
			}
		}
		st.Candidates[key] = c
		if cur, ok := st.CurrentFiles[key]; ok {
			if c.Name != "" {
				cur.Name = c.Name
			}
			if c.Order > 0 {
				cur.Order = c.Order
			}
			if c.SizeBytes > 0 && cur.TotalBytes == 0 {
				cur.TotalBytes = c.SizeBytes
			}
			st.CurrentFiles[key] = cur
			if st.CurrentFile != nil && normalizePath(st.CurrentFile.Path) == key {
				st.CurrentFile.Name = cur.Name
				if c.Order > 0 {
					st.CurrentFile.Order = cur.Order
				}
				if c.SizeBytes > 0 && st.CurrentFile.TotalBytes == 0 {
					st.CurrentFile.TotalBytes = c.SizeBytes
				}
			}
			continue
		}
		if done, ok := st.Completed[key]; ok {
			if c.Name != "" {
				done.Name = c.Name
			}
			if c.SizeBytes > 0 || done.SizeBytes == 0 {
				done.SizeBytes = c.SizeBytes
			}
			st.Completed[key] = done
			continue
		}
		if p, ok := st.Pending[key]; ok {
			if c.Name != "" {
				p.Name = c.Name
			}
			if c.SizeBytes > 0 || p.SizeBytes == 0 {
				p.SizeBytes = c.SizeBytes
			}
			if c.Order > 0 {
				p.Order = c.Order
			}
			st.Pending[key] = p
			continue
		}
		st.Pending[key] = TransferPendingFile{Path: key, Name: c.Name, SizeBytes: c.SizeBytes, Status: FileStatusPending, Order: c.Order}
		trimPendingRetainedLocked(st)
	}
	st.PreflightPending = false
	st.PreflightFinished = true
	st.UpdatedAt = time.Now()
	m.persistSnapshotLockedImmediate(st)
}

func (m *Manager) SetPreflightResult(runID int64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.byRunID[runID]
	if !ok || st == nil {
		return
	}
	st.PreflightPending = false
	st.PreflightFinished = err == nil
	if err != nil {
		st.Degraded = true
		st.DegradeReason = err.Error()
		if strings.TrimSpace(st.DegradeReason) == "" {
			st.DegradeReason = "preflight failed"
		}
	}
	st.UpdatedAt = time.Now()
	m.persistSnapshotLockedImmediate(st)
}

func (m *Manager) RemoveState(runID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.byRunID[runID]
	if !ok {
		return
	}
	if timer, ok := m.persistTimers[runID]; ok {
		timer.Stop()
		delete(m.persistTimers, runID)
	}
	delete(m.pendingPersist, runID)
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
	order := 0
	candidateSize := int64(0)
	if c, ok := st.Candidates[key]; ok {
		if c.Name != "" {
			name = c.Name
		}
		order = c.Order
		candidateSize = c.SizeBytes
	}
	if total <= 0 {
		if prev, ok := st.CurrentFiles[key]; ok && prev.TotalBytes > 0 {
			total = prev.TotalBytes
		} else if candidateSize > 0 {
			total = candidateSize
		}
	}
	cur := TransferCurrentFile{Path: key, Name: name, Bytes: bytes, TotalBytes: total, Speed: speed, Percentage: pct, Status: FileStatusInProgress, Order: order}
	st.CurrentFile = &cur
	if st.CurrentFiles == nil {
		st.CurrentFiles = map[string]TransferCurrentFile{}
	}
	st.CurrentFiles[key] = cur
	if p, ok := st.Pending[key]; ok {
		p.Status = FileStatusInProgress
		if order > 0 {
			p.Order = order
		}
		st.Pending[key] = p
	} else {
		st.Pending[key] = TransferPendingFile{Path: key, Name: name, Status: FileStatusInProgress, Order: order}
		trimPendingRetainedLocked(st)
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
	if size == 0 {
		if cur, ok := st.CurrentFiles[key]; ok {
			if cur.TotalBytes > 0 {
				size = cur.TotalBytes
			} else if cur.Bytes > 0 {
				size = cur.Bytes
			}
		}
	}
	if _, existed := st.Pending[key]; existed {
		delete(st.Pending, key)
	}
	if st.PendingCount > 0 {
		st.PendingCount--
	}
	st.CompletedCount++
	completedOrder := st.NextCompletedOrder
	if completedOrder <= 0 {
		completedOrder = 1
	}
	st.NextCompletedOrder = completedOrder + 1
	st.Completed[key] = TransferCompletedFile{Path: key, Name: name, SizeBytes: size, At: time.Now().Format(time.RFC3339), Status: status, Message: message, Order: completedOrder}
	trimCompletedRetainedLocked(st)
	if st.CurrentFiles != nil {
		delete(st.CurrentFiles, key)
	}
	if st.CurrentFile != nil && normalizePath(st.CurrentFile.Path) == key {
		st.CurrentFile = nil
		if len(st.CurrentFiles) > 0 {
			next := make([]TransferCurrentFile, 0, len(st.CurrentFiles))
			for _, v := range st.CurrentFiles {
				next = append(next, v)
			}
			sort.SliceStable(next, func(i, j int) bool {
				if next[i].Order != next[j].Order {
					if next[i].Order == 0 {
						return false
					}
					if next[j].Order == 0 {
						return true
					}
					return next[i].Order < next[j].Order
				}
				return next[i].Path < next[j].Path
			})
			vv := next[0]
			st.CurrentFile = &vv
		}
	}
	st.UpdatedAt = time.Now()
	m.persistSnapshotLocked(st)
}

func (m *Manager) SetTransferSlots(runID int64, slots int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.byRunID[runID]
	if !ok || st == nil {
		return
	}
	st.TransferSlots = normalizedTransferSlots(slots)
	st.UpdatedAt = time.Now()
	m.persistSnapshotLockedImmediate(st)
}

func (m *Manager) OnFileProgress(runID int64, path string, bytes, total, speed int64, pct *float64) {
	m.UpdateCurrentFile(runID, path, bytes, total, speed, pct)
}

func (m *Manager) OnFileCopied(runID int64, path string) {
	m.MarkCompleted(runID, path, FileStatusCopied, "")
}

func (m *Manager) OnFileCASMatched(runID int64, path string) {
	m.MarkCompleted(runID, path, FileStatusCASMatched, "")
}

func (m *Manager) OnFileFailed(runID int64, path string, message string) {
	m.MarkCompleted(runID, path, FileStatusFailed, message)
}

func (m *Manager) OnFileSkipped(runID int64, path string, message string) {
	m.MarkCompleted(runID, path, FileStatusSkipped, message)
}

func (m *Manager) OnFileDeleted(runID int64, path string) {
	m.MarkCompleted(runID, path, FileStatusDeleted, "")
}

