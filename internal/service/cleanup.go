package service

import (
	"context"
	"time"

	"rcloneflow/internal/logger"
	"go.uber.org/zap"
)

// CleanupService 自动清理服务
type CleanupService struct {
	runSvc    *RunService
	interval  time.Duration
	retention int // 保留天数（任务运行记录）
	// 预留：日志文件/事件表等的独立保留天数，可从环境变量读取，默认与 retention 一致
	// logRetention int
	stopCh    chan struct{}
}

// NewCleanupService 创建自动清理服务
func NewCleanupService(runSvc *RunService, interval time.Duration, retentionDays int) *CleanupService {
	return &CleanupService{
		runSvc:    runSvc,
		interval:  interval,
		retention: retentionDays,
		stopCh:    make(chan struct{}),
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
func (s *CleanupService) Stop() {
	close(s.stopCh)
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
