package active_transfer

import (
  "sync"
  "testing"
  "time"
)

func TestManager_ThrottlesProgressPersistButFlushesCompletedImmediately(t *testing.T) {
  mgr := NewManager()
  mgr.SetPersistThrottle(40 * time.Millisecond)

  var mu sync.Mutex
  calls := make([]ActiveTransferSnapshot, 0, 8)
  mgr.SetPersistFunc(func(runID int64, snap ActiveTransferSnapshot) {
    mu.Lock()
    calls = append(calls, snap)
    mu.Unlock()
  })

  st := mgr.InitState(101, 202, TrackingModeNormal, []TransferCandidateFile{{Path: "a.bin", Name: "a.bin", SizeBytes: 100}})
  _ = st
  time.Sleep(10 * time.Millisecond)

  pct1 := 10.0
  pct2 := 30.0
  pct3 := 60.0
  mgr.UpdateCurrentFile(101, "a.bin", 10, 100, 1, &pct1)
  mgr.UpdateCurrentFile(101, "a.bin", 30, 100, 1, &pct2)
  mgr.UpdateCurrentFile(101, "a.bin", 60, 100, 1, &pct3)

  time.Sleep(20 * time.Millisecond)
  mu.Lock()
  afterBurst := len(calls)
  mu.Unlock()
  if afterBurst > 2 {
    t.Fatalf("expected throttled progress persists, got %d calls", afterBurst)
  }

  mgr.MarkCompleted(101, "a.bin", FileStatusCopied, "")
  time.Sleep(80 * time.Millisecond)

  mu.Lock()
  defer mu.Unlock()
  if len(calls) < 2 {
    t.Fatalf("expected at least init + completion persist, got %d", len(calls))
  }
  last := calls[len(calls)-1]
  if len(last.Completed) != 1 {
    t.Fatalf("expected completed snapshot flush, got completed=%d", len(last.Completed))
  }
}
