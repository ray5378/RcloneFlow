package service

import (
	"context"
	"encoding/json"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/store"
)

// TaskService 任务服务层
type TaskService struct {
	db  *store.DB
	runner adapter.TaskRunner
}

// NewTaskService 创建任务服务
func NewTaskService(db *store.DB, runner adapter.TaskRunner) *TaskService {
	return &TaskService{db: db, runner: runner}
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
		return ErrTaskNotFound
	}

	// 解析任务选项
	var opts *adapter.TaskOptions
	if len(t.Options) > 0 {
		var taskOpts adapter.TaskOptions
		if err := json.Unmarshal(t.Options, &taskOpts); err == nil {
			opts = &taskOpts
		}
	}

	// 所有任务默认使用流式/大文件友好的传输配置，用户显式选项覆盖默认值
	opts = adapter.MergeTaskOptions(opts)

	// 通过runner启动任务
	jobID, err := s.runner.RunTask(ctx, t.ID, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath, trigger, opts)
	if err != nil {
		return err
	}

	// 记录运行
	_, err = s.db.AddRun(store.Run{
		TaskID:       taskID,
		RcJobID:      jobID,
		Status:       "running",
		Trigger:     trigger,
		TaskName:    t.Name,
		TaskMode:    t.Mode,
		SourceRemote: t.SourceRemote,
		SourcePath:  t.SourcePath,
		TargetRemote: t.TargetRemote,
		TargetPath:  t.TargetPath,
	})
	return err
}

// GetTask 获取单个任务
func (s *TaskService) GetTask(id int64) (store.Task, bool) {
	return s.db.GetTask(id)
}
