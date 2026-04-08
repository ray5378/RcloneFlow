// Package cli 提供基于命令行（CLI）的 rclone 运行器。
// 目的：用本地子进程直接控制 rclone，替换 RC 模式。
// 本文件为脚手架，占位定义，后续逐步补齐实现（不影响现有编译路径）。
package cli

// Runner 负责：
// - 启动 rclone 子进程（安全参数、禁止字符串拼接防注入）
// - 采集标准输出/错误、解析进度（--use-json-log 或 --stats-one-line）
// - 优雅停止：SIGINT → 超时 → SIGTERM → 超时 → SIGKILL
// - 日志滚动与保留策略（文件 + 事件采样入库）
// - 并发上限与队列控制
// 注：本结构先占位，后续提交将逐步补齐。

type Runner struct{}

// StartOptions 用于映射「前端高级选项」到 rclone CLI 旗标。
// 后续会加入：源/目标、过滤、传输/多线程/缓冲、校验/对比、可靠性/超时、带宽/统计等。
type StartOptions struct{
	// TODO: 将 TaskOptions 映射为 rclone 命令行参数
}

// RunHandle 表示一次运行的句柄（用于查询/停止等）。
type RunHandle struct{
	RunID int64 // 对应后端 run 记录
	PID   int   // 子进程 PID
}

// NewRunner 创建 CLI 运行器实例。
func NewRunner() *Runner { return &Runner{} }

// Start 启动一次 rclone 任务（子进程）。后续实现会返回可用句柄或错误。
func (r *Runner) Start(opts StartOptions) (*RunHandle, error) { return nil, nil }

// Stop 优雅停止一次运行（按信号梯度）。后续实现会依据句柄定位并终止子进程。
func (r *Runner) Stop(handle *RunHandle) error { return nil }

// TODO: 进度解析器（JSON 日志 / 单行统计）、日志滚动/保留、DerivedProgress 推送等。
