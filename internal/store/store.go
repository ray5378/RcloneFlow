package store

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sync"
    "time"
)

type Task struct {
    ID           int64     `json:"id"`
    Name         string    `json:"name"`
    Mode         string    `json:"mode"`
    SourceRemote string    `json:"sourceRemote"`
    SourcePath   string    `json:"sourcePath"`
    TargetRemote string    `json:"targetRemote"`
    TargetPath   string    `json:"targetPath"`
    CreatedAt    time.Time `json:"createdAt"`
}

type Schedule struct {
    ID        int64     `json:"id"`
    TaskID    int64     `json:"taskId"`
    Spec      string    `json:"spec"`
    Enabled   bool      `json:"enabled"`
    CreatedAt time.Time `json:"createdAt"`
}

type Run struct {
    ID        int64                  `json:"id"`
    TaskID     int64                 `json:"taskId"`
    RcJobID   int64                  `json:"rcJobId"`
    Status    string                 `json:"status"`
    Trigger   string                 `json:"trigger"`
    Summary   map[string]any         `json:"summary,omitempty"`
    Error     string                 `json:"error,omitempty"`
    CreatedAt time.Time              `json:"createdAt"`
    UpdatedAt time.Time              `json:"updatedAt"`
}

type DB struct {
    mu        sync.Mutex
    dir       string
    Tasks     []Task     `json:"tasks"`
    Schedules []Schedule `json:"schedules"`
    Runs      []Run      `json:"runs"`
    NextTaskID int64     `json:"nextTaskId"`
    NextScheduleID int64 `json:"nextScheduleId"`
    NextRunID int64      `json:"nextRunId"`
}

func Open(dir string) (*DB, error) {
    _ = os.MkdirAll(dir, 0o755)
    db := &DB{dir: dir, NextTaskID: 1, NextScheduleID: 1, NextRunID: 1}
    path := filepath.Join(dir, "app.json")
    bs, err := os.ReadFile(path)
    if err == nil && len(bs) > 0 {
        if err := json.Unmarshal(bs, db); err != nil {
            return nil, err
        }
    }
    return db, nil
}

func (db *DB) saveLocked() error {
    bs, err := json.MarshalIndent(db, "", "  ")
    if err != nil { return err }
    return os.WriteFile(filepath.Join(db.dir, "app.json"), bs, 0o644)
}

func (db *DB) ListTasks() []Task { db.mu.Lock(); defer db.mu.Unlock(); out := append([]Task(nil), db.Tasks...); return out }
func (db *DB) ListSchedules() []Schedule { db.mu.Lock(); defer db.mu.Unlock(); out := append([]Schedule(nil), db.Schedules...); return out }
func (db *DB) ListRuns() []Run { db.mu.Lock(); defer db.mu.Unlock(); out := append([]Run(nil), db.Runs...); return out }

func (db *DB) AddTask(t Task) (Task, error) {
    db.mu.Lock(); defer db.mu.Unlock()
    t.ID = db.NextTaskID; db.NextTaskID++; t.CreatedAt = time.Now()
    db.Tasks = append([]Task{t}, db.Tasks...)
    return t, db.saveLocked()
}

func (db *DB) AddSchedule(s Schedule) (Schedule, error) {
    db.mu.Lock(); defer db.mu.Unlock()
    s.ID = db.NextScheduleID; db.NextScheduleID++; s.CreatedAt = time.Now()
    db.Schedules = append([]Schedule{s}, db.Schedules...)
    return s, db.saveLocked()
}

func (db *DB) AddRun(r Run) (Run, error) {
    db.mu.Lock(); defer db.mu.Unlock()
    r.ID = db.NextRunID; db.NextRunID++; now := time.Now(); r.CreatedAt = now; r.UpdatedAt = now
    db.Runs = append([]Run{r}, db.Runs...)
    return r, db.saveLocked()
}

func (db *DB) UpdateRun(id int64, fn func(*Run)) error {
    db.mu.Lock(); defer db.mu.Unlock()
    for i := range db.Runs {
        if db.Runs[i].ID == id {
            fn(&db.Runs[i])
            db.Runs[i].UpdatedAt = time.Now()
            return db.saveLocked()
        }
    }
    return nil
}

func (db *DB) GetTask(id int64) (Task, bool) {
    db.mu.Lock(); defer db.mu.Unlock()
    for _, t := range db.Tasks { if t.ID == id { return t, true } }
    return Task{}, false
}
