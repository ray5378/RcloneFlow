// Package cli 提供基于命令行（CLI）的 rclone 运行器：启动/监控/停止 rclone 子进程。
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	rclone "rcloneflow/internal/rclone"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Runner 负责：
// - 安全地以参数数组方式启动 rclone 子进程（避免命令注入）
// - 将标准输出/错误写入独立日志文件（后续接入滚动/保留策略）
// - 解析 --stats 的进度行（或 JSON 日志）并上报给上层（UpdateProgress）
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

// GetHandle 通过 RunID 获取句柄。
func (r *Runner) GetHandle(runID int64) (*RunHandle, bool) {
	r.mu.Lock()
	h, ok := r.procs[runID]
	r.mu.Unlock()
	return h, ok
}

// Start 启动 rclone 子进程（默认 copy）。
// 说明：
// - 后续会根据任务类型选择 copy/sync 等；此处先用 copy 做最小可用实现
// - 进度解析：解析 --stats-one-line 或 JSONL（若启用 --use-json-log）
func (r *Runner) Start(opts StartOptions) (*RunHandle, error) {
	if opts.RunID == 0 { return nil, errors.New("RunID 不能为空") }
	if opts.CLI.Src == "" || opts.CLI.Dst == "" { return nil, errors.New("源/目标不能为空") }

	args := BuildCopyArgs(opts.CLI)
	// 建议上层确保 StatsInterval 或 JSONLog 被启用，便于进度解析
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, rclone.RclonePath(), args...)

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

	// 为了实时解析，使用管道而不是仅文件句柄
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil { cancel(); _ = stdoutFile.Close(); _ = stderrFile.Close(); return nil, err }
	stderrPipe, err := cmd.StderrPipe()
	if err != nil { cancel(); _ = stdoutFile.Close(); _ = stderrFile.Close(); return nil, err }

	// 并发控制
	acquire()
	// 启动
	if err := cmd.Start(); err != nil {
		release()
		_ = stdoutFile.Close(); _ = stderrFile.Close(); cancel()
		return nil, err
	}

	h := &RunHandle{RunID: opts.RunID, PID: cmd.Process.Pid, Cmd: cmd, Cancel: cancel, Stdout: stdoutPath, Stderr: stderrPath, Started: time.Now()}

	// 后台：复制到日志文件 + 解析进度
	go r.consumeAndParse(opts.RunID, stdoutPipe, stdoutFile)
	go r.consumeOnly(stderrPipe, stderrFile)

	// 后台：等待退出并清理资源
	go func(runID int64) {
		_ = cmd.Wait()
		release()
		stdoutFile.Close()
		stderrFile.Close()
		RemoveRun(runID)
		r.mu.Lock()
		delete(r.procs, runID)
		r.mu.Unlock()
	}(opts.RunID)

	r.mu.Lock()
	r.procs[opts.RunID] = h
	r.mu.Unlock()
	return h, nil
}

// 消费 stdout，写入文件并解析进度。
func (r *Runner) consumeAndParse(runID int64, pipe io.Reader, outFile *os.File) {
	reader := bufio.NewReader(pipe)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			// 写入日志
			_, _ = outFile.WriteString(line)
			// 解析进度
			if p, ok := ParseProgressLine(line); ok {
				UpdateProgress(runID, p)
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) { time.Sleep(100 * time.Millisecond); continue }
			return
		}
	}
}

// 仅消费 stderr 到文件。
func (r *Runner) consumeOnly(pipe io.Reader, outFile *os.File) {
	reader := bufio.NewReader(pipe)
	for {
		buf := make([]byte, 4096)
		n, err := reader.Read(buf)
		if n > 0 { _, _ = outFile.Write(buf[:n]) }
		if err != nil {
			if errors.Is(err, io.EOF) { time.Sleep(100 * time.Millisecond); continue }
			return
		}
	}
}

// Stop 优雅停止（INT→TERM→KILL）。
func (r *Runner) Stop(h *RunHandle) error {
	if h == nil || h.Cmd == nil || h.Cmd.Process == nil { return errors.New("无效句柄") }

	// 防止父进程收到 Ctrl+C 导致一并退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	_ = h.Cmd.Process.Signal(syscall.SIGINT)
	if waitExited(h.Cmd, 10*time.Second) { return nil }
	_ = h.Cmd.Process.Signal(syscall.SIGTERM)
	if waitExited(h.Cmd, 10*time.Second) { return nil }
	_ = h.Cmd.Process.Kill()
	return nil
}

// StopByRunID 通过 RunID 查句柄并停止。
func (r *Runner) StopByRunID(runID int64) error {
	h, ok := r.GetHandle(runID)
	if !ok { return errors.New("未找到运行句柄") }
	return r.Stop(h)
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
