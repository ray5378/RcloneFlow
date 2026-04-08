package service

import (
	"time"

	clirunner "rcloneflow/internal/runner/cli"
)

// CLIRunAdapter 将 CLI 运行器进度映射为现有的运行记录结构（最小占位）。

type CLIRunAdapter struct{}

func NewCLIRunAdapter() *CLIRunAdapter { return &CLIRunAdapter{} }

// GetDerivedProgress 返回内存态最新进度（用于 /runs/active 等）。
func (a *CLIRunAdapter) GetDerivedProgress(runID int64) map[string]any {
	if p, ok := clirunner.GetProgress(runID); ok {
		m := map[string]any{
			"bytes":        p.Bytes,
			"totalBytes":   p.TotalBytes,
			"percentage":   p.Percent,
			"speed":        p.SpeedBytesPerSec,
			"eta":          p.EtaSeconds,
			"updatedAt":    p.UpdatedAt.Format(time.RFC3339),
		}
		return m
	}
	return nil
}
