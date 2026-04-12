package service

import (
	"context"
	"time"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"
)

// CleanupService 自动清理服务
type CleanupService struct {
	runSvc    *RunService
	interval  time.Duration
	retention int // 保留天数
	stopCh    chan struct{}
	resetCh   chan struct{}
}

// NewCleanupService 创建自动清理服务
func NewCleanupService(runSvc *RunService, interval time.Duration, retentionDays int) *CleanupService {
	return &CleanupService{
		runSvc:    runSvc,
		interval:  interval,
		retention: retentionDays,
		stopCh:    make(chan struct{}),
		resetCh:   make(chan struct{}, 1),
	}
}

// Start 启动清理服务
func (s *CleanupService) Start(ctx context.Context) {
	logger.Info("启动历史记录清理服务",
		zap.Int("interval", int(s.interval.Seconds())),
		zap.Int("retention_days", s.retention))

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 启动时立即执行一次清理
	s.cleanup()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.resetCh:
			// 重排：立刻重置 ticker，并以最新 interval/retention 继续
			ticker.Stop()
			ticker = time.NewTicker(s.interval)
			logger.Info("历史记录清理计划已重排",
				zap.Int("interval_hours", int(s.interval.Hours())),
				zap.Int("retention_days", s.retention))
			// 立即执行一次
			s.cleanup()
		case <-s.stopCh:
			logger.Info("停止历史记录清理服务")
			return
		case <-ctx.Done():
			logger.Info("历史记录清理服务上下文已关闭")
			return
		}
	}
}

// Stop 停止清理服务
func (s *CleanupService) Stop() { close(s.stopCh) }

// Replan 重新设置 interval/retention，并触发重排
func (s *CleanupService) Replan(intervalHours int, retentionDays int) {
	if intervalHours > 0 {
		s.interval = time.Duration(intervalHours) * time.Hour
	}
	if retentionDays >= 0 {
		s.retention = retentionDays
	}
	select {
	case s.resetCh <- struct{}{}:
	default:
	}
}

// cleanup 执行清理
func (s *CleanupService) cleanup() {
	if s.retention <= 0 {
		return
	}

	deleted, err := s.runSvc.CleanOldRuns(s.retention)
	if err != nil {
		logger.Error("清理历史记录失败", zap.Error(err))
		return
	}

	if deleted > 0 {
		logger.Info("清理历史记录成功",
			zap.Int64("deleted", deleted),
			zap.Int("retention_days", s.retention))
	}
}
