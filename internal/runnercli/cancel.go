package runnercli

import (
	"context"
	"sync"
)

var runCancelFns sync.Map

func registerCancelFn(runID int64, cancel context.CancelFunc) {
	runCancelFns.Store(runID, cancel)
}

func unregisterCancelFn(runID int64) {
	runCancelFns.Delete(runID)
}

func CancelRun(runID int64) bool {
	if v, ok := runCancelFns.Load(runID); ok {
		v.(context.CancelFunc)()
		return true
	}
	return false
}
