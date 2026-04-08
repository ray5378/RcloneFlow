package cli

import (
	"sync"
)

// 运行时内存状态（进度、存活句柄等）。
var (
	stMu      sync.RWMutex
	stProg    = make(map[int64]DerivedProgress) // runID -> progress
)

// UpdateProgress 更新某个运行的最新进度。
func UpdateProgress(runID int64, p DerivedProgress) {
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
