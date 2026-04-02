package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"rcloneflow/internal/logger"
	"rcloneflow/internal/rclone"
	"rcloneflow/internal/store"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Runner 任务运行器接口
type Runner interface {
	RunTask(ctx context.Context, taskID int64, trigger string) error
}

// Scheduler 定时任务调度器
type Scheduler struct {
	cron *cron.Cron
	db   *store.DB
	r    Runner
}

// New 创建调度器
func New(db *store.DB, rc *rclone.Client) *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
		db:   db,
		r:    &taskRunner{db: db, rc: rc},
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

// ParseSpecToCron 将前端格式转为标准cron表达式
// 前端格式: minute|hour|day|month|week (用|分隔,内部值用逗号)
// 例如: "04,03,06|17,19|*|*|*" -> "04,03,06 17,19 * * *"
func ParseSpecToCron(spec string) (string, bool) {
	parts := strings.Split(spec, "|")
	if len(parts) != 5 {
		return "", false
	}
	minute, hour, day, month, week := parts[0], parts[1], parts[2], parts[3], parts[4]

	// 分钟
	if minute == "" || minute == "*" {
		minute = "*"
	} else {
		for _, p := range strings.Split(minute, ",") {
			v, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || v < 0 || v > 59 {
				return "", false
			}
		}
	}

	// 小时
	if hour == "" || hour == "*" {
		hour = "*"
	} else {
		for _, p := range strings.Split(hour, ",") {
			v, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || v < 0 || v > 23 {
				return "", false
			}
		}
	}

	// 日
	if day == "" || day == "*" {
		day = "*"
	} else {
		for _, p := range strings.Split(day, ",") {
			v, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || v < 1 || v > 31 {
				return "", false
			}
		}
	}

	// 月
	if month == "" || month == "*" {
		month = "*"
	} else {
		for _, p := range strings.Split(month, ",") {
			v, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || v < 1 || v > 12 {
				return "", false
			}
		}
	}

	// 周
	if week == "" || week == "*" {
		week = "*"
	} else {
		for _, p := range strings.Split(week, ",") {
			v, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || v < 0 || v > 6 {
				return "", false
			}
		}
	}

	// 标准cron: 秒 分 时 日 月 周
	return "0 " + minute + " " + hour + " " + day + " " + month + " " + week, true
}

// CalcNextRun 计算下次触发时间
func CalcNextRun(spec string) (time.Time, error) {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	cronExpr, err := parser.Parse(spec)
	if err != nil {
		return time.Time{}, err
	}
	return cronExpr.Next(time.Now()), nil
}

// AddSchedule 添加定时任务到调度器(动态添加)
func (s *Scheduler) AddSchedule(schedule store.Schedule) error {
	if !schedule.Enabled {
		return nil
	}
	cronSpec, ok := ParseSpecToCron(schedule.Spec)
	if !ok {
		return fmt.Errorf("invalid cron spec")
	}
	taskID := schedule.TaskID
	scheduleID := schedule.ID

	// 计算下次触发时间
	nextTime, err := CalcNextRun(cronSpec)
	if err == nil {
		s.db.UpdateScheduleNextRunTime(scheduleID, nextTime)
	}

	_, err = s.cron.AddFunc(cronSpec, func() {
		if err := s.r.RunTask(context.Background(), taskID, "schedule"); err != nil {
			logger.Error("定时任务执行失败",
				zap.Int64("task_id", taskID),
				zap.String("spec", cronSpec),
				zap.Error(err))
		}
		// 执行后更新下次触发时间
		nextTime, err := CalcNextRun(cronSpec)
		if err == nil {
			s.db.UpdateScheduleNextRunTime(scheduleID, nextTime)
		}
	})
	if err != nil {
		return err
	}
	logger.Info("定时任务已添加(运行时)",
		zap.Int64("schedule_id", scheduleID),
		zap.Int64("task_id", taskID),
		zap.String("cron_spec", cronSpec))
	return nil
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
		cronSpec, ok := ParseSpecToCron(item.Spec)
		if !ok {
			logger.Warn("跳过不支持的定时规格",
				zap.Int64("schedule_id", item.ID),
				zap.String("spec", item.Spec),
				zap.String("reason", "cron表达式解析失败"))
			continue
		}
		taskID := item.TaskID
		scheduleID := item.ID

		// 计算并存储下次触发时间
		nextTime, err := CalcNextRun(cronSpec)
		if err == nil {
			s.db.UpdateScheduleNextRunTime(scheduleID, nextTime)
		}

		_, err = s.cron.AddFunc(cronSpec, func() {
			if err := s.r.RunTask(context.Background(), taskID, "schedule"); err != nil {
				logger.Error("定时任务执行失败",
					zap.Int64("task_id", taskID),
					zap.String("spec", cronSpec),
					zap.Error(err))
			}
			// 执行后更新下次触发时间
			nextTime, err := CalcNextRun(cronSpec)
			if err == nil {
				s.db.UpdateScheduleNextRunTime(scheduleID, nextTime)
			}
		})
		if err != nil {
			logger.Warn("添加定时任务失败",
				zap.Int64("schedule_id", item.ID),
				zap.String("cron_spec", cronSpec),
				zap.Error(err))
			continue
		}
		logger.Info("定时任务已启动",
			zap.Int64("schedule_id", item.ID),
			zap.Int64("task_id", taskID),
			zap.String("cron_spec", cronSpec))
	}
	s.cron.Start()
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done()
	}
}
