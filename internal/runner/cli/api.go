package cli

// 对外导出一组简单函数，便于控制层/服务层调用（后续可替换为依赖注入）。

var defaultRunner = NewRunner()

// StartRun 启动一次运行（封装 defaultRunner）。
func StartRun(opts StartOptions) (*RunHandle, error) { return defaultRunner.Start(opts) }

// StopRun 通过句柄停止（占位，后续可根据 runID 查句柄）。
func StopRun(h *RunHandle) error { return defaultRunner.Stop(h) }
