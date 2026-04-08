package service

import (
	"fmt"

	clirunner "rcloneflow/internal/runner/cli"
)

// AttachProgressPersistence 挂载 CLI 进度监听器，将 bytes/speed 写回 runs 表（按 RcJobID 匹配）。
func AttachProgressPersistence(db RunServiceInterface) (detach func()) {
	return clirunner.AddProgressListener(func(runID int64, p clirunner.DerivedProgress) {
		// speed 文本化（MiB/s）
		speed := fmt.Sprintf("%.2f MiB/s", p.SpeedBytesPerSec/1024.0/1024.0)
		_ = db.UpdateRunProgressByJobId(runID, int64(p.Bytes), speed)
	})
}
