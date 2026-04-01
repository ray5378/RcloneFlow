package scheduler

import (
	"context"
	"log"
	"strings"
	"time"

	"rcloneflow/internal/rclone"
	"rcloneflow/internal/store"
)

// Runner 任务运行器接口
type Runner interface {
	RunTask(ctx context.Context, taskID int64, trigger string) error
}

// Scheduler 定时任务调度器
type Scheduler struct {
	db *store.DB
	r  Runner
}

// New 创建调度器
func New(db *store.DB, rc *rclone.Client) *Scheduler {
	return &Scheduler{
		db: db,
		r:  &taskRunner{db: db, rc: rc},
	}
}

// taskRunner 任务运行器实现
type taskRunner struct {
	db *store.DB
	rc *rclone.Client
}

func (r *taskRunner) RunTask(ctx context.Context, taskID int64, trigger string) error {
	t, ok := r.db.GetTask(taskID)
	if !ok {
		return nil
	}

	jobID, err := r.rc.RunTask(ctx, t.ID, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, trigger)
	if err != nil {
		// 即使启动失败也记录
		r.db.AddRun(store.Run{
			TaskID:  taskID,
			Status:  "failed",
			Trigger: trigger,
		})
		return err
	}

	_, err = r.db.AddRun(store.Run{
		TaskID:  taskID,
		RcJobID: jobID,
		Status:  "running",
		Trigger: trigger,
	})
	return err
}

// parseSpec 解析定时规格
func parseSpec(spec string) (time.Duration, bool) {
	spec = strings.TrimSpace(spec)
	spec = strings.TrimPrefix(spec, "@every ")
	d, err := time.ParseDuration(spec)
	if err != nil || d <= 0 {
		return 0, false
	}
	return d, true
}

// Start 启动调度器
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
