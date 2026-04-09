package service

import (
	"context"
	"encoding/json"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/store"
	runnercli "rcloneflow/internal/runnercli"
	"rcloneflow/internal/settings"
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

	effectiveOptions := map[string]any{}
	if bs, err := json.Marshal(opts); err == nil {
		_ = json.Unmarshal(bs, &effectiveOptions)
	}

	streamingEnabled := true
	if v, ok := effectiveOptions["enableStreaming"].(bool); ok {
		streamingEnabled = v
	}

	// 切换为 CLI：先记录运行，再异步启动（可中断/进度）
	run, err := s.db.AddRun(store.Run{
		TaskID:       taskID,
		Status:       "running",
		Trigger:      trigger,
		Summary: map[string]any{
			"streamingEnabled": streamingEnabled,
			"effectiveOptions": effectiveOptions,
		},
		TaskName:     t.Name,
		TaskMode:     t.Mode,
		SourceRemote: t.SourceRemote,
		SourcePath:   t.SourcePath,
		TargetRemote: t.TargetRemote,
		TargetPath:   t.TargetPath,
	})
	// 合并全局传输设置（用于 Runner 默认值回退）
	if ts, err := settings.Load(); err == nil {
		_ = s.db.UpdateRun(run.ID, func(rr *store.Run){
			if rr.Summary == nil { rr.Summary = map[string]any{} }
			rr.Summary["transferDefaults"] = ts
		})
	}
	if err != nil { return err }
	go func(){ _ = runnercli.New(s.db).Start(context.Background(), run, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath) }()
	return nil
}

// GetTask 获取单个任务
func (s *TaskService) GetTask(id int64) (store.Task, bool) {
	return s.db.GetTask(id)
}
