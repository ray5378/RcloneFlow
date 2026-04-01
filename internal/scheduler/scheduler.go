package scheduler

import (
    "context"
    "log"
    "strings"
    "time"

    "rcloneflow/internal/store"
)

type Runner interface {
    RunTask(ctx context.Context, taskID int64, trigger string) error
}

type Scheduler struct {
    db *store.DB
    r  Runner
}

func New(db *store.DB, r Runner) *Scheduler {
    return &Scheduler{db: db, r: r}
}

func parseSpec(spec string) (time.Duration, bool) {
    spec = strings.TrimSpace(spec)
    spec = strings.TrimPrefix(spec, "@every ")
    d, err := time.ParseDuration(spec)
    if err != nil || d <= 0 {
        return 0, false
    }
    return d, true
}

func (s *Scheduler) Start() error {
    schedules, err := s.db.ListSchedules()
    if err != nil {
        return err
    }
    for _, item := range schedules {
        if !item.Enabled {
            continue
        }
        d, ok := parseSpec(item.Spec)
        if !ok {
            log.Printf("skip schedule %d: unsupported spec %q (current MVP supports @every 5m or 5m)", item.ID, item.Spec)
            continue
        }
        go func(taskID int64, every time.Duration) {
            ticker := time.NewTicker(every)
            defer ticker.Stop()
            for range ticker.C {
                if err := s.r.RunTask(context.Background(), taskID, "schedule"); err != nil {
                    log.Printf("schedule run failed for task %d: %v", taskID, err)
                }
            }
        }(item.TaskID, d)
    }
    return nil
}
