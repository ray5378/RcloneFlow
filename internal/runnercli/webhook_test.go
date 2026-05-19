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
    {name: "transferred bytes only, no copied files", sum: map[string]any{"transferredBytes": float64(1), "completedCount": float64(0)}, want: false},
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
      name: "skips when only transferredBytes positive but no copied files",
      notifyStatus: map[string]any{"success": true, "failed": true, "hasTransfer": true},
      finalSummary: map[string]any{
        "counts": map[string]any{"copied": float64(0), "failed": float64(0), "skipped": float64(0), "total": float64(1)},
        "transferredBytes": float64(1024),
        "totalBytes": float64(1024),
        "avgSpeedBps": float64(100),
      },
      wantPosts: 0,
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
      name: "hasTransfer overrides success=false failed=false when copied files exist",
      notifyStatus: map[string]any{"success": false, "failed": false, "hasTransfer": true},
      finalSummary: map[string]any{
        "counts": map[string]any{"copied": float64(3), "failed": float64(0), "skipped": float64(0), "total": float64(3)},
        "transferredBytes": float64(0),
        "totalBytes": float64(0),
        "avgSpeedBps": float64(0),
      },
      wantPosts: 1,
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

func TestToInt64(t *testing.T) {
  tests := []struct {
    name string
    input any
    want  int64
  }{
    {"int", 42, 42},
    {"int64", int64(100), 100},
    {"float64", float64(3.14), 3},
    {"json.Number int", json.Number("50"), 50},
    {"json.Number float", json.Number("3.7"), 3},
    {"string int", "123", 123},
    {"string float", "3.14", 3},
    {"string invalid", "abc", 0},
    {"nil", nil, 0},
    {"bool", true, 0},
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      if got := toInt64(tt.input); got != tt.want {
        t.Errorf("toInt64(%v) = %d, want %d", tt.input, got, tt.want)
      }
    })
  }
}

func TestToFloat64(t *testing.T) {
  tests := []struct {
    name string
    input any
    want  float64
  }{
    {"float64", 3.14, 3.14},
    {"int", 42, 42.0},
    {"int64", int64(100), 100.0},
    {"json.Number", json.Number("2.5"), 2.5},
    {"string", "3.14", 3.14},
    {"string invalid", "abc", 0},
    {"nil", nil, 0},
    {"bool", true, 0},
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      if got := toFloat64(tt.input); got != tt.want {
        t.Errorf("toFloat64(%v) = %f, want %f", tt.input, got, tt.want)
      }
    })
  }
}

func TestFormatBytes(t *testing.T) {
  tests := []struct {
    input int64
    want  string
  }{
    {0, "0B"},
    {500, "500B"},
    {1024, "1KB"},
    {1536, "1.5KB"},
    {1048576, "1MB"},
    {1073741824, "1GB"},
    {1099511627776, "1TB"},
    {1125899906842624, "1PB"},
  }
  for _, tt := range tests {
    t.Run(tt.want, func(t *testing.T) {
      if got := formatBytes(tt.input); got != tt.want {
        t.Errorf("formatBytes(%d) = %q, want %q", tt.input, got, tt.want)
      }
    })
  }
}

func TestFormatBps(t *testing.T) {
  if got := formatBps(1024); got != "1KB/s" {
    t.Errorf("formatBps(1024) = %q, want %q", got, "1KB/s")
  }
  if got := formatBps(1536); got != "1.5KB/s" {
    t.Errorf("formatBps(1536) = %q, want %q", got, "1.5KB/s")
  }
}

func TestBaseName(t *testing.T) {
  tests := []struct {
    input string
    want  string
  }{
    {"", ""},
    {"file.txt", "file.txt"},
    {"/path/to/file.txt", "file.txt"},
    {"C:\\Windows\\file.txt", "file.txt"},
    {"/", "/"},
    {"\\", "\\"},
  }
  for _, tt := range tests {
    t.Run(tt.input, func(t *testing.T) {
      if got := baseName(tt.input); got != tt.want {
        t.Errorf("baseName(%q) = %q, want %q", tt.input, got, tt.want)
      }
    })
  }
}

func TestNested(t *testing.T) {
  tests := []struct {
    name string
    m    map[string]any
    path string
    want any
  }{
    {"nil map", nil, "a", nil},
    {"missing key", map[string]any{"a": 1}, "b", nil},
    {"single level", map[string]any{"a": 1}, "a", 1},
    {"nested", map[string]any{"a": map[string]any{"b": 2}}, "a.b", 2},
    {"deep nested", map[string]any{"a": map[string]any{"b": map[string]any{"c": 3}}}, "a.b.c", 3},
    {"partial missing", map[string]any{"a": map[string]any{"b": 2}}, "a.c", nil},
    {"non-map intermediate", map[string]any{"a": "not-a-map"}, "a.b", nil},
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      got := nested(tt.m, tt.path)
      if got != tt.want {
        t.Errorf("nested(%v, %q) = %v, want %v", tt.m, tt.path, got, tt.want)
      }
    })
  }
}

func TestToBool(t *testing.T) {
  tests := []struct {
    name string
    input any
    want  bool
  }{
    {"bool true", true, true},
    {"bool false", false, false},
    {"string true", "true", true},
    {"string false", "false", false},
    {"string yes", "yes", true},
    {"string no", "no", false},
    {"string 1", "1", true},
    {"string 0", "0", false},
    {"int 1", 1, true},
    {"int 0", 0, false},
    {"float64 1", float64(1), true},
    {"float64 0", float64(0), false},
    {"nil", nil, false},
    {"unknown", "maybe", false},
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      if got := toBool(tt.input); got != tt.want {
        t.Errorf("toBool(%v) = %v, want %v", tt.input, got, tt.want)
      }
    })
  }
}

func TestPostWebhookIfNeeded_GetTaskError(t *testing.T) {
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

  // Create a task first, then delete it to simulate GetTask failure after run is created
  task, err := db.AddTask(store.Task{Name: "temp-task", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
  if err != nil {
    t.Fatalf("AddTask() error = %v", err)
  }
  run, err := db.AddRun(store.Run{TaskID: task.ID, Status: "finished", Trigger: "manual", TaskName: task.Name})
  if err != nil {
    t.Fatalf("AddRun() error = %v", err)
  }
  // Delete the task to trigger the "get task failed" path
  err = db.DeleteTask(task.ID)
  if err != nil {
    t.Fatalf("DeleteTask() error = %v", err)
  }

  r := New(db)
  r.postWebhookIfNeeded(run.ID)
}

func TestPostWebhookIfNeeded_NoWebhookURL(t *testing.T) {
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

  task, err := db.AddTask(store.Task{Name: "no-webhook", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b"})
  if err != nil {
    t.Fatalf("AddTask() error = %v", err)
  }
  run, err := db.AddRun(store.Run{TaskID: task.ID, Status: "finished", Trigger: "manual", TaskName: task.Name})
  if err != nil {
    t.Fatalf("AddRun() error = %v", err)
  }

  r := New(db)
  r.postWebhookIfNeeded(run.ID)
}

func TestPostWebhookIfNeeded_WebhookFailure(t *testing.T) {
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

  srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusInternalServerError)
  }))
  defer srv.Close()

  opts, _ := json.Marshal(map[string]any{
    "webhookPostUrl": srv.URL,
    "webhookNotifyOn": map[string]any{"manual": true},
  })
  task, err := db.AddTask(store.Task{Name: "fail-webhook", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b", Options: opts})
  if err != nil {
    t.Fatalf("AddTask() error = %v", err)
  }
  run, err := db.AddRun(store.Run{TaskID: task.ID, Status: "finished", Trigger: "manual", TaskName: task.Name})
  if err != nil {
    t.Fatalf("AddRun() error = %v", err)
  }

  r := New(db)
  r.postWebhookIfNeeded(run.ID)
  time.Sleep(500 * time.Millisecond)
}

func TestPostWebhookIfNeeded_WecomURL(t *testing.T) {
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
  done := make(chan struct{}, 1)
  srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    hits++
    mu.Unlock()
    w.WriteHeader(http.StatusOK)
    done <- struct{}{}
  }))
  defer srv.Close()

  opts, _ := json.Marshal(map[string]any{
    "wecomPostUrl": srv.URL,
    "webhookNotifyOn": map[string]any{"manual": true},
  })
  task, err := db.AddTask(store.Task{Name: "wecom-task", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b", Options: opts})
  if err != nil {
    t.Fatalf("AddTask() error = %v", err)
  }
  run, err := db.AddRun(store.Run{
    TaskID: task.ID, Status: "finished", Trigger: "manual", TaskName: task.Name,
    Summary: map[string]any{"finalSummary": map[string]any{"counts": map[string]any{"copied": float64(1)}, "transferredBytes": float64(1024)}},
  })
  if err != nil {
    t.Fatalf("AddRun() error = %v", err)
  }

  r := New(db)
  r.postWebhookIfNeeded(run.ID)

  select {
  case <-done:
  case <-time.After(2 * time.Second):
    t.Fatal("timed out waiting for wecom webhook")
  }

  mu.Lock()
  if hits != 1 {
    t.Fatalf("wecom hits=%d, want 1", hits)
  }
  mu.Unlock()
}

func TestPostWebhookIfNeeded_SkippedForWrongTrigger(t *testing.T) {
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

  hits := 0
  srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    hits++
    w.WriteHeader(http.StatusOK)
  }))
  defer srv.Close()

  opts, _ := json.Marshal(map[string]any{
    "webhookPostUrl": srv.URL,
    "webhookNotifyOn": map[string]any{"manual": true},
  })
  task, err := db.AddTask(store.Task{Name: "skip-trigger", Mode: "copy", SourceRemote: "src", SourcePath: "/a", TargetRemote: "dst", TargetPath: "/b", Options: opts})
  if err != nil {
    t.Fatalf("AddTask() error = %v", err)
  }
  run, err := db.AddRun(store.Run{
    TaskID: task.ID, Status: "finished", Trigger: "webhook", TaskName: task.Name,
    Summary: map[string]any{"finalSummary": map[string]any{"counts": map[string]any{"copied": float64(1)}, "transferredBytes": float64(1024)}},
  })
  if err != nil {
    t.Fatalf("AddRun() error = %v", err)
  }

  r := New(db)
  r.postWebhookIfNeeded(run.ID)
  time.Sleep(300 * time.Millisecond)

  if hits != 0 {
    t.Fatalf("expected 0 hits for wrong trigger, got %d", hits)
  }
}
