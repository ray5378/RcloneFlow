package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	maxRetries        = 3
	baseBackoff       = 1 * time.Second
	maxBackoff        = 5 * time.Minute
	scheduleNextDelay = 5 * time.Minute
)

func init() {
	_ = baseBackoff
	_ = maxBackoff
}

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

type BackoffScheduler struct {
	*Scheduler
}

func (s *BackoffScheduler) runTaskWithRetry(ctx context.Context, taskID int64, trigger string, scheduleID int64, cronSpec string) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := baseBackoff * time.Duration(1<<(attempt-1))
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			logger.Info("任务执行失败，准备重试",
				zap.Int64("task_id", taskID),
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff))
			time.Sleep(backoff)
		}

		err := s.r.RunTask(ctx, taskID, trigger)
		if err == nil {
			if attempt > 0 {
				logger.Info("任务重试成功",
					zap.Int64("task_id", taskID),
					zap.Int("attempts", attempt+1))
			}
			return
		}
		lastErr = err
		logger.Warn("任务执行失败",
			zap.Int64("task_id", taskID),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", maxRetries),
			zap.Error(err))
	}

	s.db.UpdateScheduleNextRunTime(scheduleID, time.Now().Add(scheduleNextDelay))
	logger.Error("任务重试次数已用尽",
		zap.Int64("task_id", taskID),
		zap.Int64("schedule_id", scheduleID),
		zap.String("cron_spec", cronSpec),
		zap.Error(lastErr))
}

func (s *BackoffScheduler) AddScheduleWithRetry(schedule store.Schedule) error {
	if !schedule.Enabled {
		return nil
	}
	cronSpec, ok := ParseSpecToCron(schedule.Spec)
	if !ok {
		return fmt.Errorf("invalid cron spec")
	}
	taskID := schedule.TaskID
	scheduleID := schedule.ID

	s.mu.Lock()
	if eid, ok := s.entries[scheduleID]; ok {
		s.cron.Remove(eid)
		delete(s.entries, scheduleID)
	}
	s.mu.Unlock()

	nextTime, err := CalcNextRun(cronSpec)
	if err == nil {
		s.db.UpdateScheduleNextRunTime(scheduleID, nextTime)
	}

	eid, err := s.cron.AddFunc(cronSpec, func() {
		s.runTaskWithRetry(context.Background(), taskID, "schedule", scheduleID, cronSpec)
		nextTime, err := CalcNextRun(cronSpec)
		if err == nil {
			s.db.UpdateScheduleNextRunTime(scheduleID, nextTime)
		}
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.entries[scheduleID] = eid
	s.mu.Unlock()
	logger.Info("定时任务已添加(运行时)",
		zap.Int64("schedule_id", scheduleID),
		zap.Int64("task_id", taskID),
		zap.String("cron_spec", cronSpec))
	return nil
}

func NewWithBackoff(db *store.DB, runner Runner) *BackoffScheduler {
	return &BackoffScheduler{
		Scheduler: &Scheduler{
			cron:    cron.New(cron.WithSeconds()),
			db:      db,
			r:       runner,
			entries: map[int64]cron.EntryID{},
		},
	}
}

func (s *BackoffScheduler) Start() error {
	schedules, err := s.db.ListSchedules()
	if err != nil {
		return err
	}
	for _, item := range schedules {
		if !item.Enabled {
			continue
		}
		if err := s.AddScheduleWithRetry(item); err != nil {
			logger.Warn("添加定时任务失败",
				zap.Int64("schedule_id", item.ID),
				zap.String("spec", item.Spec),
				zap.Error(err))
			continue
		}
		logger.Info("定时任务已启动",
			zap.Int64("schedule_id", item.ID),
			zap.Int64("task_id", item.TaskID))
	}
	s.cron.Start()
	return nil
}

// Runner 任务运行器接口
type Runner interface {
	RunTask(ctx context.Context, taskID int64, trigger string) error
}

// Scheduler 定时任务调度器
type Scheduler struct {
	cron *cron.Cron
	db   *store.DB
	r    Runner
	mu   sync.Mutex
	// 映射 scheduleID -> cron EntryID，便于删除/重建
	entries map[int64]cron.EntryID
}

// DB 返回底层存储（仅用于控制器在更新启用状态后读取最新 schedule）
func (s *Scheduler) DB() *store.DB { return s.db }

// RemoveSchedule 移除一个运行时调度项（若存在）
func (s *Scheduler) RemoveSchedule(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if eid, ok := s.entries[id]; ok {
		s.cron.Remove(eid)
		delete(s.entries, id)
	}
}

// New 创建调度器（默认使用 RC 运行器）。
// 生产环境应使用 NewWithRunner 以通过 CLI Runner 产出 stderr 日志文件。
func New(db *store.DB, rc *adapter.RcloneClient) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		db:      db,
		r:       &taskRunner{db: db, rc: rc},
		entries: map[int64]cron.EntryID{},
	}
}

// NewWithRunner 创建调度器（显式指定 Runner，例如 TaskService 以使用 CLI Runner 与 stderr 日志）
func NewWithRunner(db *store.DB, runner Runner) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		db:      db,
		r:       runner,
		entries: map[int64]cron.EntryID{},
	}
}

// taskRunner 任务运行器实现
type taskRunner struct {
	db *store.DB
	rc *adapter.RcloneClient
}

func (r *taskRunner) RunTask(ctx context.Context, taskID int64, trigger string) error {
	t, ok := r.db.GetTask(taskID)
	if !ok {
		return nil
	}

	// 解析任务选项
	var opts *adapter.TaskOptions
	if len(t.Options) > 0 {
		if taskOpts, err := adapter.ParseTaskOptionsCompat(t.Options); err == nil {
			opts = taskOpts
		}
	}

	src := t.SourceRemote + ":" + strings.TrimPrefix(t.SourcePath, "/")
	dst := t.TargetRemote + ":" + strings.TrimPrefix(t.TargetPath, "/")
	_, err := r.rc.StartJob(ctx, t.Mode, src, dst, opts)
	if err != nil {
		r.db.AddRun(store.Run{
			TaskID:       taskID,
			Status:       "failed",
			Trigger:      trigger,
			TaskName:     t.Name,
			TaskMode:     t.Mode,
			SourceRemote: t.SourceRemote,
			SourcePath:   t.SourcePath,
			TargetRemote: t.TargetRemote,
			TargetPath:   t.TargetPath,
		})
		return err
	}

	_, err = r.db.AddRun(store.Run{
		TaskID:       taskID,
		Status:       "running",
		Trigger:      trigger,
		TaskName:     t.Name,
		TaskMode:     t.Mode,
		SourceRemote: t.SourceRemote,
		SourcePath:   t.SourcePath,
		TargetRemote: t.TargetRemote,
		TargetPath:   t.TargetPath,
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

	// 如已有旧 entry，先删除
	s.mu.Lock()
	if eid, ok := s.entries[scheduleID]; ok {
		s.cron.Remove(eid)
		delete(s.entries, scheduleID)
	}
	s.mu.Unlock()

	// 计算下次触发时间
	nextTime, err := CalcNextRun(cronSpec)
	if err == nil {
		s.db.UpdateScheduleNextRunTime(scheduleID, nextTime)
	}

	eid, err := s.cron.AddFunc(cronSpec, func() {
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
	s.mu.Lock()
	s.entries[scheduleID] = eid
	s.mu.Unlock()
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
		if err := s.AddSchedule(item); err != nil {
			logger.Warn("添加定时任务失败",
				zap.Int64("schedule_id", item.ID),
				zap.String("spec", item.Spec),
				zap.Error(err))
			continue
		}
		logger.Info("定时任务已启动",
			zap.Int64("schedule_id", item.ID),
			zap.Int64("task_id", item.TaskID))
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

