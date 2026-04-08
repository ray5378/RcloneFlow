package cli

import (
	"fmt"
	"path/filepath"
)

// LogPaths 返回默认的 stdout/stderr 路径（基于 Start 中的默认 workDir 规则）。
func LogPaths(runID int64) (string, string) {
	workDir := filepath.Join("/tmp", fmt.Sprintf("rcloneflow-run-%d", runID))
	return filepath.Join(workDir, "stdout.log"), filepath.Join(workDir, "stderr.log")
}
