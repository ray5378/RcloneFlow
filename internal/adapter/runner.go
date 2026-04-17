package adapter

import (
	"context"
	"strings"
)

// TaskRunner 任务运行器接口
type TaskRunner interface {
	RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string, opts *TaskOptions) (int64, error)
	HasActiveJobs(ctx context.Context) (bool, error)
}

// TaskRunnerImpl 任务运行器实现
type TaskRunnerImpl struct {
	client *RcloneClient
}

// NewTaskRunner 创建任务运行器
func NewTaskRunner(client *RcloneClient) *TaskRunnerImpl {
	return &TaskRunnerImpl{client: client}
}

// RunTask 运行任务
func (r *TaskRunnerImpl) RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string, opts *TaskOptions) (int64, error) {
	src := srcRemote + ":" + strings.TrimPrefix(srcPath, "/")
	dst := dstRemote + ":" + strings.TrimPrefix(dstPath, "/")

	return r.client.StartJob(ctx, mode, src, dst, opts)
}

// HasActiveJobs 检查是否有正在运行的任务
func (r *TaskRunnerImpl) HasActiveJobs(ctx context.Context) (bool, error) {
	return r.client.HasActiveJobs(ctx)
}
