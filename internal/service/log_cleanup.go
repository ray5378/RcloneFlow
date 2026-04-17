package service

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.uber.org/zap"
	"rcloneflow/internal/logger"
)

// LogCleanupService 清理 /app/data/logs 下的运行日志文件（run-*-stdout.log / run-*-stderr.log）
// 默认保留 7 天；可用环境变量 LOG_RETENTION_DAYS 覆盖；清理周期默认 24 小时，可用 LOG_CLEANUP_INTERVAL_HOURS 覆盖。
type LogCleanupService struct {
	logsDir       string
	interval      time.Duration
	retentionDays int
	stopCh        chan struct{}
}

func NewLogCleanupService(logsDir string, interval time.Duration, retentionDays int) *LogCleanupService {
	if retentionDays <= 0 {
		retentionDays = 7
	}
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	return &LogCleanupService{logsDir: logsDir, interval: interval, retentionDays: retentionDays, stopCh: make(chan struct{})}
}

func (s *LogCleanupService) Start(ctx context.Context) {
	logger.Info("启动日志清理服务", zap.String("logs_dir", s.logsDir), zap.Int("retention_days", s.retentionDays), zap.Duration("interval", s.interval))
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 启动时先清理一次
	s.cleanup()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopCh:
			logger.Info("停止日志清理服务")
			return
		case <-ctx.Done():
			logger.Info("日志清理服务上下文已关闭")
			return
		}
	}
}

func (s *LogCleanupService) Stop() { close(s.stopCh) }

func (s *LogCleanupService) cleanup() {
	// 删除超出保留期的日志文件和空目录
	cutoff := time.Now().AddDate(0, 0, -s.retentionDays)
	deleted := 0

	entries, err := os.ReadDir(s.logsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(s.logsDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// 检查目录修改时间是否超过保留期
		if info.ModTime().Before(cutoff) {
			if err := os.RemoveAll(subDir); err == nil {
				deleted++
				logger.Debug("删除日志目录", zap.String("dir", entry.Name()))
			}
			continue
		}
		// 目录未过期，清理里面的日志文件
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			continue
		}
		for _, subEntry := range subEntries {
			if subEntry.IsDir() {
				continue
			}
			subInfo, err := subEntry.Info()
			if err != nil {
				continue
			}
			if subInfo.ModTime().Before(cutoff) {
				path := filepath.Join(subDir, subEntry.Name())
				if err := os.Remove(path); err == nil {
					deleted++
				}
			}
		}
		// 检查目录是否变空
		remaining, _ := os.ReadDir(subDir)
		if len(remaining) == 0 {
			os.Remove(subDir)
		}
	}
	if deleted > 0 {
		logger.Info("日志清理完成", zap.Int("deleted", deleted), zap.Int("retention_days", s.retentionDays))
	}
}

// Helpers to read env overrides
func EnvLogRetentionDays(defaultDays int) int {
	if v := os.Getenv("LOG_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultDays
}

func EnvLogCleanupInterval(defaultHours int) time.Duration {
	if v := os.Getenv("LOG_CLEANUP_INTERVAL_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Hour
		}
	}
	return time.Duration(defaultHours) * time.Hour
}
