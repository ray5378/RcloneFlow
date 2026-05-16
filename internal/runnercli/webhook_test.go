package runnercli

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "os"
  "sync"
  "testing"
  "time"

  "rcloneflow/internal/store"
)

func TestHasTransferEvidence(t *testing.T) {
  tests := []struct {
    name string
    sum map[string]any
    want bool
  }{
    {name: "transferred bytes", sum: map[string]any{"transferredBytes": float64(1), "completedCount": float64(0)}, want: true},
    {name: "completed count only", sum: map[string]any{"transferredBytes": float64(0), "completedCount": float64(1)}, want: true},
    {name: "no transfer", sum: map[string]any{"transferredBytes": float64(0), "completedCount": float64(0)}, want: false},
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      if got := hasTransferEvidence(tt.sum); got != tt.want {
        t.Fatalf("hasTransferEvidence()=%v want %v", got, tt.want)
      }
    })
  }
}

func TestPostWebhookIfNeeded_HasTransferFilter(t *testing.T) {
  tests := []struct {
    name string
    notifyStatus map[string]any
    finalSummary map[string]any
    wantPosts int
  }{
    {
      name: "sends when transferredBytes positive",
      notifyStatus: map[string]any{"success": true, "failed": true, "hasTransfer": true},
      finalSummary: map[string]any{
        "counts": map[string]any{"copied": float64(0), "failed": float64(0), "skipped": float64(0), "total": float64(1)},
        "transferredBytes": float64(1024),
        "totalBytes": float64(1024),
        "avgSpeedBps": float64(100),
      },
      wantPosts: 1,
    },
    {
      name: "sends when completedCount positive and bytes zero",
      notifyStatus: map[string]any{"success": true, "failed": true, "hasTransfer": true},
      finalSummary: map[string]any{
        "counts": map[string]any{"copied": float64(1), "failed": float64(0), "skipped": float64(0), "total": float64(1)},
        "transferredBytes": float64(0),
        "totalBytes": float64(0),
        "avgSpeedBps": float64(0),
      },
      wantPosts: 1,
    },
    {
      name: "skips when no transfer evidence",
      notifyStatus: map[string]any{"success": true, "failed": true, "hasTransfer": true},
      finalSummary: map[string]any{
        "counts": map[string]any{"copied": float64(0), "failed": float64(0), "skipped": float64(1), "total": float64(1)},
        "transferredBytes": float64(0),
        "totalBytes": float64(0),
        "avgSpeedBps": float64(0),
      },
      wantPosts: 0,
    },
    {
      name: "legacy config still sends without hasTransfer field",
      notifyStatus: map[string]any{"success": true, "failed": true},
      finalSummary: map[string]any{
        "counts": map[string]any{"copied": float64(0), "failed": float64(0), "skipped": float64(1), "total": float64(1)},
        "transferredBytes": float64(0),
        "totalBytes": float64(0),
        "avgSpeedBps": float64(0),
      },
      wantPosts: 1,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      tmpDir, err := os.MkdirTemp("", "rcloneflow-webhook-*")
      if err != nil {
        t.Fatalf("MkdirTemp() error = %v", err)
      }
      defer os.RemoveAll(tmpDir)

      db, err := store.Open(tmpDir)
      if err != nil {
        t.Fatalf("store.Open() error = %v", err)
      }
      defer db.Close()

      var mu sync.Mutex
      hits := 0
      done := make(chan struct{}, 2)
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        mu.Lock()
        hits++
        mu.Unlock()
        w.WriteHeader(http.StatusOK)
        done <- struct{}{}
      }))
      defer srv.Close()

      opts, _ := json.Marshal(map[string]any{
        "webhookPostUrl": srv.URL,
        "webhookNotifyOn": map[string]any{"manual": true, "schedule": true, "webhook": true},
        "webhookNotifyStatus": tt.notifyStatus,
      })
      task, err := db.AddTask(store.Task{Name: "webhook-task", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b", Options: opts})
      if err != nil {
        t.Fatalf("AddTask() error = %v", err)
      }
      run, err := db.AddRun(store.Run{
        TaskID: task.ID,
        Status: "finished",
        Trigger: "manual",
        TaskName: task.Name,
        TaskMode: task.Mode,
        Summary: map[string]any{
          "finalSummary": tt.finalSummary,
        },
      })
      if err != nil {
        t.Fatalf("AddRun() error = %v", err)
      }

      r := New(db)
      r.postWebhookIfNeeded(run.ID)

      if tt.wantPosts > 0 {
        select {
        case <-done:
        case <-time.After(2 * time.Second):
          t.Fatalf("timed out waiting for webhook post")
        }
      } else {
        time.Sleep(300 * time.Millisecond)
      }

      mu.Lock()
      got := hits
      mu.Unlock()
      if got != tt.wantPosts {
        t.Fatalf("webhook posts=%d want %d", got, tt.wantPosts)
      }
    })
  }
}
