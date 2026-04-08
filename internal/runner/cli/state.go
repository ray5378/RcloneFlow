package cli

import (
	"sync"
)

// 运行时内存状态（进度、存活句柄等）。
var (
	stMu   sync.RWMutex
	stProg = make(map[int64]DerivedProgress) // runID -> progress
)

// 进度监听器（供上层持久化或转发）。
type ProgressListener func(runID int64, p DerivedProgress)

var (
	lstMu     sync.RWMutex
	listeners []ProgressListener
)

func AddProgressListener(l ProgressListener) (remove func()) {
	lstMu.Lock()
	listeners = append(listeners, l)
	idx := len(listeners) - 1
	lstMu.Unlock()
	return func() {
		lstMu.Lock()
		if idx >= 0 && idx < len(listeners) {
			listeners[idx] = nil
		}
		lstMu.Unlock()
	}
}

func notifyListeners(runID int64, p DerivedProgress) {
	lstMu.RLock()
	ls := make([]ProgressListener, len(listeners))
	copy(ls, listeners)
	lstMu.RUnlock()
	for _, l := range ls {
		if l != nil { l(runID, p) }
	}
}

// setProgress 更新某个运行的最新进度（内部）。
func setProgress(runID int64, p DerivedProgress) {
	stMu.Lock()
	stProg[runID] = p
	stMu.Unlock()
}

// GetProgress 读取某个运行的最新进度。
func GetProgress(runID int64) (DerivedProgress, bool) {
	stMu.RLock()
	p, ok := stProg[runID]
	stMu.RUnlock()
	return p, ok
}

// RemoveRun 清除运行相关状态。
func RemoveRun(runID int64) {
	stMu.Lock()
	delete(stProg, runID)
	stMu.Unlock()
}
