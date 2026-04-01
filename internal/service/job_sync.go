package service

import (
	"context"
	"time"

	"rcloneflow/internal/logger"
	"rcloneflow/internal/rclone"
	"rcloneflow/internal/store"

	"go.uber.org/zap"
)

// JobSyncService 任务状态同步服务
type JobSyncService struct {
	db          JobStatusProvider
	rc          *rclone.Client
	poolGap     time.Duration
	stop        chan struct{}
}

// JobStatusProvider 任务状态提供者接口
type JobStatusProvider interface {
	ListRunningRuns() ([]store.JobStatus, error)
	UpdateRunStatus(id int64, status, errorMsg string, summary map[string]any) error
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

		// 查询rclone job状态
		status, err := s.rc.JobStatus(context.Background(), run.RcJobID)
		if err != nil {
			logger.Warn("查询任务状态失败",
				zap.Int64("run_id", run.ID),
				zap.Int64("job_id", run.RcJobID),
				zap.Error(err))
			continue
		}

		// 解析状态
		newStatus := "running"
		errorMsg := ""
		
		if finished, ok := status["finished"].(bool); ok && finished {
			newStatus = "finished"
		}
		if success, ok := status["success"].(bool); ok && !success {
			newStatus = "failed"
			if errStr, ok := status["error"].(string); ok {
				errorMsg = errStr
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
				logger.Info("任务状态已更新",
					zap.Int64("run_id", run.ID),
					zap.Int64("job_id", run.RcJobID),
					zap.String("status", newStatus))
			}
		}
	}
}
