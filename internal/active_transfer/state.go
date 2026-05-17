package active_transfer

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type PersistFunc func(runID int64, snap ActiveTransferSnapshot)

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

func (m *Manager) SetPersistThrottle(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.persistThrottle = d
}

func (m *Manager) emitPersist(runID int64, snap ActiveTransferSnapshot) {
	m.mu.RLock()
	persist := m.persist
	m.mu.RUnlock()
	if persist == nil {
		return
	}
	go persist(runID, snap)
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
		go m.persist(runID, snap)
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
					st.CurrentFile.Order = c.Order
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
	m.persistSnapshotLockedImmediate(st)
}

func (m *Manager) BuildSummary(taskID int64, bytes, total, speed, eta int64, percentage float64) (ActiveTransferOverviewResponse, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, ok := m.byTask[taskID]
	if !ok {
		return ActiveTransferOverviewResponse{}, false
	}
	currentFiles := make([]TransferCurrentFile, 0, len(st.CurrentFiles))
	for _, v := range st.CurrentFiles {
		currentFiles = append(currentFiles, v)
	}
	sort.SliceStable(currentFiles, func(i, j int) bool {
		if currentFiles[i].Order != currentFiles[j].Order {
			if currentFiles[i].Order == 0 {
				return false
			}
			if currentFiles[j].Order == 0 {
				return true
			}
			return currentFiles[i].Order < currentFiles[j].Order
		}
		return currentFiles[i].Path < currentFiles[j].Path
	})
	return ActiveTransferOverviewResponse{
		TaskID:       st.TaskID,
		RunID:        st.RunID,
		TrackingMode: st.TrackingMode,
		CurrentFile:  cloneCurrent(st.CurrentFile),
		CurrentFiles: currentFiles,
		Summary: ActiveTransferSummary{
			TrackingMode:      st.TrackingMode,
			CompletedCount:    st.CompletedCount,
			PendingCount:      st.PendingCount,
			TotalCount:        st.TotalCount,
			PreflightPending:  st.PreflightPending,
			PreflightFinished: st.PreflightFinished,
			Bytes:             bytes,
			TotalBytes:        total,
			Speed:             speed,
			Eta:               eta,
			Percentage:        percentage,
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
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Order != items[j].Order {
			if items[i].Order == 0 {
				return false
			}
			if items[j].Order == 0 {
				return true
			}
			return items[i].Order < items[j].Order
		}
		if items[i].At != items[j].At {
			return items[i].At < items[j].At
		}
		return items[i].Path < items[j].Path
	})
	resp := paginate(items, offset, limit)
	resp.Total = st.CompletedCount
	return resp
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
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Status != items[j].Status {
			return items[i].Status == FileStatusInProgress
		}
		if items[i].Order != items[j].Order {
			if items[i].Order == 0 {
				return false
			}
			if items[j].Order == 0 {
				return true
			}
			return items[i].Order < items[j].Order
		}
		return items[i].Path < items[j].Path
	})
	resp := paginate(items, offset, limit)
	resp.Total = st.PendingCount
	return resp
}

func trimCompletedRetainedLocked(st *ActiveTransferState) {
	if st == nil || !st.Degraded || len(st.Completed) <= retainedCompletedLimit {
		return
	}
	for len(st.Completed) > retainedCompletedLimit {
		var oldestKey string
		var oldestOrder int
		first := true
		for key, item := range st.Completed {
			if first || item.Order < oldestOrder {
				oldestKey = key
				oldestOrder = item.Order
				first = false
			}
		}
		if oldestKey == "" {
			return
		}
		delete(st.Completed, oldestKey)
	}
}

func trimPendingRetainedLocked(st *ActiveTransferState) {
	if st == nil || !st.Degraded || len(st.Pending) <= retainedPendingLimit {
		return
	}
	type candidate struct {
		key   string
		order int
	}
	removable := make([]candidate, 0, len(st.Pending))
	for key, item := range st.Pending {
		if item.Status == FileStatusInProgress {
			continue
		}
		removable = append(removable, candidate{key: key, order: item.Order})
	}
	sort.SliceStable(removable, func(i, j int) bool {
		if removable[i].order == removable[j].order {
			return removable[i].key > removable[j].key
		}
		return removable[i].order > removable[j].order
	})
	for len(st.Pending) > retainedPendingLimit && len(removable) > 0 {
		victim := removable[0]
		removable = removable[1:]
		delete(st.Pending, victim.key)
	}
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
