package cli

import (
	"sync"
	"time"
)

// 轻量级事件存储（内存）：每个 run 保留最近 N 条进度事件（采样）。

const maxEventsPerRun = 256

var (
	evMu   sync.RWMutex
	events = make(map[int64][]DerivedProgress)
)

// appendEvent 将进度追加到事件列表（环形裁剪）。
func appendEvent(runID int64, p DerivedProgress) {
	evMu.Lock()
	lst := events[runID]
	lst = append(lst, p)
	if len(lst) > maxEventsPerRun {
		lst = lst[len(lst)-maxEventsPerRun:]
	}
	events[runID] = lst
	evMu.Unlock()
}

// ListEvents 返回 run 的进度事件切片（副本）。
func ListEvents(runID int64) []DerivedProgress {
	evMu.RLock()
	lst := events[runID]
	evMu.RUnlock()
	out := make([]DerivedProgress, len(lst))
	copy(out, lst)
	return out
}

// UpdateProgress 除了更新最新进度，也会以 1s 采样写入事件。
var (
	lastSampleMu sync.Mutex
	lastSample   = make(map[int64]time.Time)
)

func UpdateProgress(runID int64, p DerivedProgress) {
	setProgress(runID, p)
	// 采样：至少 1s 才写一条事件
	lastSampleMu.Lock()
	if t, ok := lastSample[runID]; !ok || time.Since(t) >= time.Second {
		appendEvent(runID, p)
		lastSample[runID] = time.Now()
	}
	lastSampleMu.Unlock()
}
