package active_transfer

import (
	"sort"
	"strings"
)

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
	transferSlots := normalizedTransferSlots(st.TransferSlots)
	return ActiveTransferOverviewResponse{
		TaskID:        st.TaskID,
		RunID:         st.RunID,
		TrackingMode:  st.TrackingMode,
		TransferSlots: transferSlots,
		CurrentFile:   cloneCurrent(st.CurrentFile),
		CurrentFiles: currentFiles,
		Summary: ActiveTransferSummary{
			TrackingMode:      st.TrackingMode,
			CompletedCount:    st.CompletedCount,
			PendingCount:      st.PendingCount,
			TotalCount:        st.TotalCount,
			TransferSlots:     transferSlots,
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