package scheduler

import (
	"context"
	"errors"
	"strconv"
	"strings"

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
	cron  *cron.Cron
	db    *store.DB
	r     Runner
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

// ParseSpecToCron 将前端格式转为标准cron表达式(5段:分 时 日 月 周)
// 前端格式: year,month,week,day,hour,minute
// 例如: "*,1,3,5,*,00,30" -> "00 00 1,3,5 1 *"
// 特殊值: * = 每年/每月/每时/每分, 空 = 不设置(用于日/周)
func ParseSpecToCron(year, month, week, day, hour, minute string) (string, bool) {
	// 处理分钟
	min := minute
	if min == "" || min == "*" {
		min = "*"
	}

	// 处理小时
	h := hour
	if h == "" || h == "*" {
		h = "*"
	}

	// 处理日
	d := day
	if d == "" {
		d = "*"
	}

	// 处理月
	m := month
	if m == "" || m == "*" {
		m = "*"
	}

	// 处理周
	w := week
	if w == "" {
		w = "*"
	}

	// 验证分钟有效(0-59)
	if min != "*" {
		for _, part := range strings.Split(min, ",") {
			val, err := strconv.Atoi(part)
			if err != nil {
				return "", false
			}
			if val < 0 || val > 59 {
				return "", false
			}
		}
	}

	// 验证小时有效(0-23)
	if h != "*" {
		for _, part := range strings.Split(h, ",") {
			val, err := strconv.Atoi(part)
			if err != nil {
				return "", false
			}
			if val < 0 || val > 23 {
				return "", false
			}
		}
	}

	// 验证日有效(1-31)
	if d != "*" {
		for _, part := range strings.Split(d, ",") {
			val, err := strconv.Atoi(part)
			if err != nil {
				return "", false
			}
			if val < 1 || val > 31 {
				return "", false
			}
		}
	}

	// 验证月有效(1-12)
	if m != "*" {
		for _, part := range strings.Split(m, ",") {
			val, err := strconv.Atoi(part)
			if err != nil {
				return "", false
			}
			if val < 1 || val > 12 {
				return "", false
			}
		}
	}

	// 验证周有效(0-6, 0=周日)
	if w != "*" {
		for _, part := range strings.Split(w, ",") {
			val, err := strconv.Atoi(part)
			if err != nil {
				return "", false
			}
			if val < 0 || val > 6 {
				return "", false
			}
		}
	}

	// 返回标准cron格式(秒 分 时 日 月 周)
	return "0 " + min + " " + h + " " + d + " " + m + " " + w, true
}

func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid number")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
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
		// spec格式: "year,month,week,day,hour,minute"
		parts := strings.Split(item.Spec, ",")
		if len(parts) != 6 {
			logger.Warn("跳过无效的定时规格",
				zap.Int64("schedule_id", item.ID),
				zap.String("spec", item.Spec),
				zap.String("reason", "格式应为 year,month,week,day,hour,minute"))
			continue
		}
		cronSpec, ok := ParseSpecToCron(parts[0], parts[1], parts[2], parts[3], parts[4], parts[5])
		if !ok {
			logger.Warn("跳过不支持的定时规格",
				zap.Int64("schedule_id", item.ID),
				zap.String("spec", item.Spec),
				zap.String("reason", "cron表达式解析失败"))
			continue
		}
		taskID := item.TaskID
		_, err := s.cron.AddFunc(cronSpec, func() {
			if err := s.r.RunTask(context.Background(), taskID, "schedule"); err != nil {
				logger.Error("定时任务执行失败",
					zap.Int64("task_id", taskID),
					zap.String("spec", cronSpec),
					zap.Error(err))
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
