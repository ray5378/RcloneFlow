package cli

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"rcloneflow/internal/adapter"
)

// TaskRunnerAdapter 实现 adapter.TaskRunner 接口，但以 CLI 子进程直控 rclone。
// 返回的 int64 为 CLI 运行标识（runID），用于 /runs/active 与 /jobs/{status,stop} 等接口。
type TaskRunnerAdapter struct{}

func NewTaskRunnerAdapter() *TaskRunnerAdapter { return &TaskRunnerAdapter{} }

var runSeq int64

func nextRunID() int64 { return atomic.AddInt64(&runSeq, 1) }

// RunTask 将 task 定义映射为 CLI 参数并启动子进程。
func (a *TaskRunnerAdapter) RunTask(ctx context.Context, taskID int64, mode, srcRemote, srcPath, dstRemote, dstPath, trigger string, opts *adapter.TaskOptions) (int64, error) {
	// 生成源/目标
	src := srcRemote + ":" + strings.TrimPrefix(srcPath, "/")
	dst := dstRemote + ":" + strings.TrimPrefix(dstPath, "/")

	// 构造 CLI 选项（最小子集，逐步完善映射）
	cli := TaskCLIOptions{
		Src: src,
		Dst: dst,
		StatsInterval: "5s",
		LogLevel:      "NOTICE",
	}
	if opts != nil {
		if opts.Transfers > 0 { cli.Transfers = opts.Transfers }
		if opts.MultiThreadStreams > 0 { cli.MultiThreadStreams = opts.MultiThreadStreams }
		if opts.MultiThreadCutoff > 0 { cli.MultiThreadCutoff = fmt.Sprintf("%dM", opts.MultiThreadCutoff) }
		if opts.BufferSize > 0 { cli.BufferSize = fmt.Sprintf("%dM", opts.BufferSize) }
		if opts.UseServerModtime { cli.UseServerModtime = true }
		if opts.SizeOnly { cli.SizeOnly = true }
		if opts.Timeout > 0 { cli.Timeout = fmt.Sprintf("%ds", opts.Timeout) }
		if opts.ConnTimeout > 0 { cli.ConnTimeout = fmt.Sprintf("%ds", opts.ConnTimeout) }
		if opts.ExpectContinueTimeout > 0 { cli.ExpectContTimeout = fmt.Sprintf("%ds", opts.ExpectContinueTimeout) }
	}

	// 目前统一按 copy 执行（与原有默认一致）；后续基于 mode 切换 copy/sync/move
	runID := nextRunID()
	_, err := StartRun(StartOptions{RunID: runID, WorkDir: "", CLI: cli})
	if err != nil { return 0, err }
	return runID, nil
}
