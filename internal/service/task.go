package service

import (
	"context"
	"fmt"

	"rcloneflow/internal/rclone"
	"rcloneflow/internal/store"
)

// TaskService 任务服务层
type TaskService struct {
	db *store.DB
	rc *rclone.Client
}

// NewTaskService 创建任务服务
func NewTaskService(db *store.DB, rc *rclone.Client) *TaskService {
	return &TaskService{db: db, rc: rc}
}

// ListTasks 获取所有任务
func (s *TaskService) ListTasks() ([]store.Task, error) {
	return s.db.ListTasks()
}

// CreateTask 创建新任务
func (s *TaskService) CreateTask(task store.Task) (store.Task, error) {
	return s.db.AddTask(task)
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(id int64, task store.Task) error {
	return s.db.UpdateTask(id, task)
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(id int64) error {
	return s.db.DeleteTask(id)
}

// RunTask 运行指定任务
func (s *TaskService) RunTask(ctx context.Context, taskID int64, trigger string) error {
	t, ok := s.db.GetTask(taskID)
	if !ok {
		return fmt.Errorf("task not found")
	}

	// 启动rclone任务
	jobID, err := s.rc.RunTask(ctx, t.ID, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, trigger)
	if err != nil {
		return err
	}

	// 记录运行
	_, err = s.db.AddRun(store.Run{
		TaskID:   taskID,
		RcJobID:  jobID,
		Status:   "running",
		Trigger:  trigger,
	})
	return err
}

// GetTask 获取单个任务
func (s *TaskService) GetTask(id int64) (store.Task, bool) {
	return s.db.GetTask(id)
}
