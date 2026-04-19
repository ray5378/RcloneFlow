package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rcloneflow/internal/logger"
	"rcloneflow/internal/rclone"
	"rcloneflow/internal/store"
	"rcloneflow/internal/websocket"

	"go.uber.org/zap"
)

// JobSyncService 任务状态同步服务
// 定期从 rclone job API 同步任务状态到数据库
type JobSyncService struct {
	db      JobStatusProvider
	rc      *rclone.Client
	poolGap time.Duration
	stop    chan struct{}
}

// JobStatusProvider 任务状态提供者接口
type JobStatusProvider interface {
	ListRunningRuns() ([]store.JobStatus, error)
	UpdateRunStatus(id int64, status, errorMsg string, summary map[string]any) error
	UpdateRunProgress(id int64, bytesTransferred int64, speed string) error
}

// NewJobSyncService 创建任务同步服务
func NewJobSyncService(db JobStatusProvider, rc *rclone.Client, poolIntervalSec int) *JobSyncService {
	return &JobSyncService{
		db:      db,
		rc:      rc,
		poolGap: time.Duration(poolIntervalSec) * time.Second,
		stop:    make(chan struct{}),
	}
}

// Start 启动同步服务
// 后台goroutine定期从rclone job API获取任务状态并更新数据库
func (s *JobSyncService) Start(ctx context.Context) {
	logger.Info("启动任务状态同步服务", zap.Duration("interval", s.poolGap))

	ticker := time.NewTicker(s.poolGap)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("任务状态同步服务已停止")
			return
		case <-s.stop:
			logger.Info("任务状态同步服务已停止")
			return
		case <-ticker.C:
			s.syncRunningJobs()
		}
	}
}

// Stop 停止同步服务
func (s *JobSyncService) Stop() {
	close(s.stop)
}

// syncRunningJobs 同步运行中的任务状态
// 从数据库获取所有 running 状态的任务，调用 rclone job API 获取最新状态并更新
func (s *JobSyncService) syncRunningJobs() {
	runs, err := s.db.ListRunningRuns()
	if err != nil {
		logger.Error("获取运行中的任务失败", zap.Error(err))
		return
	}

	if len(runs) == 0 {
		return
	}

	logger.Debug("发现运行中的任务", zap.Int("count", len(runs)))

	for _, run := range runs {
		if run.RcJobID <= 0 {
			continue
		}

		// 调用 rclone job API 获取任务状态
		// 注意：进度数据来自日志解析（consume 函数），不通过 RC 获取
		status, err := s.rc.JobStatus(context.Background(), run.RcJobID)
		if err != nil {
			// 如果job不存在于rclone中(可能已清理)，标记为finished
			errStr := err.Error()
			if strings.Contains(errStr, "job not found") || strings.Contains(errStr, "not found") {
				logger.Info("任务已从rclone中移除，标记为已完成",
					zap.Int64("run_id", run.ID),
					zap.Int64("job_id", run.RcJobID))
				if updErr := s.db.UpdateRunStatus(run.ID, "finished", "", nil); updErr != nil {
					logger.Error("更新任务状态失败", zap.Int64("run_id", run.ID), zap.Error(updErr))
				}
				continue
			}
			logger.Warn("查询任务状态失败",
				zap.Int64("run_id", run.ID),
				zap.Int64("job_id", run.RcJobID),
				zap.Error(err))
			continue
		}

		// 解析 rclone 返回的状态
		newStatus := "running"
		errorMsg := ""

		finished, _ := status["finished"].(bool)
		success, _ := status["success"].(bool)
		if finished {
			if success {
				newStatus = "finished"
			} else {
				newStatus = "failed"
				if errStr, ok := status["error"].(string); ok {
					errorMsg = errStr
				}
			}
		}

		// 如果状态有变化，更新数据库
		if newStatus != run.Status {
			if err := s.db.UpdateRunStatus(run.ID, newStatus, errorMsg, status); err != nil {
				logger.Error("更新任务状态失败",
					zap.Int64("run_id", run.ID),
					zap.String("new_status", newStatus),
					zap.Error(err))
			} else {
				logger.Info("任务状态已同步",
					zap.Int64("run_id", run.ID),
					zap.Int64("job_id", run.RcJobID),
					zap.String("status", newStatus))
				// 广播 WebSocket 通知
				websocket.Broadcast("run_status", map[string]interface{}{
					"run_id": run.ID,
					"status": newStatus,
				})
			}
		}
	}
}

// formatSpeed 格式化速度为可读字符串
func formatSpeed(bytesPerSec int64) string {
	if bytesPerSec <= 0 {
		return "0 B/s"
	}
	const unit = 1024
	if bytesPerSec < unit {
		return fmt.Sprintf("%d B/s", bytesPerSec)
	}
	div, exp := int64(unit), 0
	for n := bytesPerSec / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB/s", float64(bytesPerSec)/float64(div), "KMGTPE"[exp])
}
