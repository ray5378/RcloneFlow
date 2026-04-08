package service

import (
	"context"
	"fmt"
	"time"

	"rcloneflow/internal/logger"
	"rcloneflow/internal/store"
	"go.uber.org/zap"
)

// CleanupService 自动清理服务
type CleanupService struct {
	runSvc    *RunService
	dataDB    *store.DB
	interval  time.Duration
	retention int // 保留天数（任务运行记录 + 事件采样）
	// 预留：日志文件/事件表等的独立保留天数，可从环境变量读取，默认与 retention 一致
	// logRetention int
	stopCh    chan struct{}
}

// NewCleanupService 创建自动清理服务
func NewCleanupService(runSvc *RunService, dataDB *store.DB, interval time.Duration, retentionDays int) *CleanupService {
	return &CleanupService{
		runSvc:    runSvc,
		dataDB:   dataDB,
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
	// 清理历史 run_events（若表存在）
	s.cleanupEvents()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
			s.cleanupEvents()
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
