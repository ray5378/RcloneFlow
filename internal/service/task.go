package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"rcloneflow/internal/adapter"
	runnercli "rcloneflow/internal/runnercli"
	"rcloneflow/internal/settings"
	"rcloneflow/internal/store"
)

// TaskService 任务服务层
type TaskService struct {
	db     *store.DB
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

// UpdateTask 更新任务（容忍部分字段未提供；未提供的字段保持不变，避免误清空）
func (s *TaskService) UpdateTask(id int64, task store.Task) error {
	cur, ok := s.db.GetTask(id)
	if !ok {
		return ErrTaskNotFound
	}
	merged := cur
	if strings.TrimSpace(task.Name) != "" {
		merged.Name = task.Name
	}
	if strings.TrimSpace(task.Mode) != "" {
		merged.Mode = task.Mode
	}
	if strings.TrimSpace(task.SourceRemote) != "" {
		merged.SourceRemote = task.SourceRemote
	}
	if strings.TrimSpace(task.SourcePath) != "" {
		merged.SourcePath = task.SourcePath
	}
	if strings.TrimSpace(task.TargetRemote) != "" {
		merged.TargetRemote = task.TargetRemote
	}
	if strings.TrimSpace(task.TargetPath) != "" {
		merged.TargetPath = task.TargetPath
	}
	if len(task.Options) > 0 {
		merged.Options = task.Options
	}
	return s.db.UpdateTask(id, merged)
}

// UpdateTaskOptions 仅更新任务的 Options 字段（用于“传输选项”任务级覆盖）
func (s *TaskService) UpdateTaskOptions(id int64, opts map[string]any) error {
	// 读出任务，合并 Options 再回写（保留已有键，覆盖提交的键）
	t, ok := s.db.GetTask(id)
	if !ok {
		return ErrTaskNotFound
	}
	merged := map[string]any{}
	if len(t.Options) > 0 {
		var cur map[string]any
		if json.Unmarshal(t.Options, &cur) == nil && cur != nil {
			for k, v := range cur {
				merged[k] = v
			}
		}
	}
	for k, v := range opts {
		merged[k] = v
	}
	b, err := json.Marshal(merged)
	if err != nil {
		return err
	}
	t.Options = b
	return s.db.UpdateTask(id, t)
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
	// 合并原始任务 Options（显式配置覆盖默认/推导值，比如 transfers=2 应覆盖默认1）
	if len(t.Options) > 0 {
		var raw map[string]any
		if err := json.Unmarshal(t.Options, &raw); err == nil && raw != nil {
			for k, v := range raw {
				// 总是用任务显式值覆盖（包括 transfers/checkers/bufferSize 等）
				effectiveOptions[k] = v
			}
		}
	}

	streamingEnabled := true
	if v, ok := effectiveOptions["enableStreaming"].(bool); ok {
		streamingEnabled = v
	}

	// 单例模式检查：如果开启了单例模式，使用原子操作确保只有一个任务运行
	singletonMode, isSingleton := effectiveOptions["singletonMode"].(bool)

	// 构建运行记录
	newRun := store.Run{
		TaskID:  taskID,
		Status:  "running",
		Trigger: trigger,
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
	}

	// 单例模式：使用原子操作，如果有任任务在运行则跳过
	if isSingleton && singletonMode {
		// 检查是否有任何任务在运行
		activeRuns, err := s.db.ListActiveRuns()
		if err == nil && len(activeRuns) > 0 {
			return fmt.Errorf("单例模式：有其他任务正在运行，跳过本次执行")
		}
		// 创建运行记录
		run, err := s.db.AddRun(newRun)
		if err != nil {
			return err
		}
		// 成功创建记录，run 已填充
		// 合并全局传输设置
		if ts, err := settings.Load(); err == nil {
			_ = s.db.UpdateRun(run.ID, func(rr *store.Run) {
				if rr.Summary == nil {
					rr.Summary = map[string]any{}
				}
				rr.Summary["transferDefaults"] = ts
			})
		}
		// 异步启动任务
		go func() {
			_ = runnercli.New(s.db).Start(context.Background(), run, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath)
		}()
		return nil
	}

	// 非单例模式：直接创建运行记录
	run, err := s.db.AddRun(newRun)
	if err != nil {
		return err
	}
	// 合并全局传输设置
	if ts, err := settings.Load(); err == nil {
		_ = s.db.UpdateRun(run.ID, func(rr *store.Run) {
			if rr.Summary == nil {
				rr.Summary = map[string]any{}
			}
			rr.Summary["transferDefaults"] = ts
		})
	}
	go func() {
		_ = runnercli.New(s.db).Start(context.Background(), run, t.Mode, t.SourceRemote, t.SourcePath, t.TargetRemote, t.TargetPath)
	}()
	return nil
}

// GetTask 获取单个任务
func (s *TaskService) GetTask(id int64) (store.Task, bool) {
	return s.db.GetTask(id)
}
