package scheduler

import (
	"context"
	"strings"
	"time"

	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"

	"go.uber.org/zap"
)

// Runner 任务运行器接口
type Runner interface {
	RunTask(ctx context.Context, taskID int64, trigger string) error
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
			logger.Warn("跳过不支持的定时规格", 
				zap.Int64("schedule_id", item.ID),
				zap.String("spec", item.Spec),
				zap.String("reason", "仅支持 @every Xm 或 Xh 格式"))
			continue
		}
		go func(taskID int64, every time.Duration) {
			ticker := time.NewTicker(every)
			defer ticker.Stop()
			for range ticker.C {
				if err := s.r.RunTask(context.Background(), taskID, "schedule"); err != nil {
					logger.Error("定时任务执行失败",
						zap.Int64("task_id", taskID),
						zap.Duration("interval", every),
						zap.Error(err))
				}
			}
		}(item.TaskID, d)
	}
	return nil
}
