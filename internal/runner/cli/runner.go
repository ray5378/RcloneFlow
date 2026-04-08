// Package cli 提供基于命令行（CLI）的 rclone 运行器：启动/监控/停止 rclone 子进程。
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Runner 负责：
// - 安全地以参数数组方式启动 rclone 子进程（避免命令注入）
// - 将标准输出/错误写入独立日志文件（后续接入滚动/保留策略）
// - 解析 --stats 的进度行（或 JSON 日志）并上报给上层（后续接入）
// - 提供优雅停止：INT→TERM→KILL 的信号梯度
// - 管理运行中的进程句柄
type Runner struct {
	mu    sync.Mutex
	procs map[int64]*RunHandle // key: RunID
}

// StartOptions（精简版），与 TaskCLIOptions 搭配：
type StartOptions struct {
	RunID   int64
	WorkDir string          // 任务工作目录（日志/临时文件）
	CLI     TaskCLIOptions  // CLI 参数映射
}

// RunHandle 表示一次运行态。
type RunHandle struct {
	RunID   int64
	PID     int
	Cmd     *exec.Cmd
	Cancel  context.CancelFunc
	Stdout  string // 标准输出日志路径
	Stderr  string // 标准错误日志路径
	Started time.Time
}

// NewRunner 创建运行器。
func NewRunner() *Runner { return &Runner{procs: make(map[int64]*RunHandle)} }

// Start 启动 rclone 子进程（默认 copy）。
// 说明：
// - 后续会根据任务类型选择 copy/sync 等；此处先用 copy 做最小可用实现
// - 进度解析：先解析 --stats-one-line；后续补充 --use-json-log JSONL
func (r *Runner) Start(opts StartOptions) (*RunHandle, error) {
	if opts.RunID == 0 { return nil, errors.New("RunID 不能为空") }
	if opts.CLI.Src == "" || opts.CLI.Dst == "" { return nil, errors.New("源/目标不能为空") }

	args := BuildCopyArgs(opts.CLI)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "rclone", args...)

	// 工作目录
	if opts.WorkDir == "" { opts.WorkDir = filepath.Join(os.TempDir(), fmt.Sprintf("rcloneflow-run-%d", opts.RunID)) }
	_ = os.MkdirAll(opts.WorkDir, 0o755)

	// 日志文件
	stdoutPath := filepath.Join(opts.WorkDir, "stdout.log")
	stderrPath := filepath.Join(opts.WorkDir, "stderr.log")
	stdoutFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil { cancel(); return nil, err }
	stderrFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil { cancel(); _ = stdoutFile.Close(); return nil, err }

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// 启动
	if err := cmd.Start(); err != nil {
		_ = stdoutFile.Close(); _ = stderrFile.Close(); cancel()
		return nil, err
	}

	h := &RunHandle{RunID: opts.RunID, PID: cmd.Process.Pid, Cmd: cmd, Cancel: cancel, Stdout: stdoutPath, Stderr: stderrPath, Started: time.Now()}

	// 后台协程：等待退出并清理文件句柄
	go func() {
		defer stdoutFile.Close()
		defer stderrFile.Close()
		_ = cmd.Wait()
	}()

	// 后台协程：简单解析单行 stats（当启用 --stats-one-line 时），后续接入事件上报
	go func() {
		// 仅当启用了单行 stats 时有意义；这里先尝试读取 stdout 并作占位解析
		f, err := os.Open(stdoutPath)
		if err != nil { return }
		defer f.Close()
		reader := bufio.NewReader(f)
		for {
			line, err := reader.ReadString('\n')
			if err != nil { time.Sleep(500 * time.Millisecond); continue }
			_ = line // TODO: 调用解析器并上报 DerivedProgress
		}
	}()

	r.mu.Lock()
	r.procs[opts.RunID] = h
	r.mu.Unlock()
	return h, nil
}

// Stop 优雅停止（INT→TERM→KILL）。
func (r *Runner) Stop(h *RunHandle) error {
	if h == nil || h.Cmd == nil || h.Cmd.Process == nil { return errors.New("无效句柄") }

	// 监听外部中断（可选），避免中断信号把父进程一起杀掉
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	// 梯度发送信号
	_ = h.Cmd.Process.Signal(syscall.SIGINT)
	if waitExited(h.Cmd, 10*time.Second) { return nil }
	_ = h.Cmd.Process.Signal(syscall.SIGTERM)
	if waitExited(h.Cmd, 10*time.Second) { return nil }
	_ = h.Cmd.Process.Kill()
	return nil
}

func waitExited(cmd *exec.Cmd, d time.Duration) bool {
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	select {
	case <-done:
		return true
	case <-time.After(d):
		return false
	}
}
