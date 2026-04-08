package service

import (
	"database/sql"
	"fmt"

	clirunner "rcloneflow/internal/runner/cli"
	"rcloneflow/internal/store"
)

// AttachProgressPersistence 挂载 CLI 进度监听器，将 bytes/speed 写回 runs 表，并采样写入 run_events（若存在）。
func AttachProgressPersistence(db RunServiceInterface, sdb *store.DB) (detach func()) {
	// 可选：检查 run_events 是否存在（迁移尚未落地时跳过插入）
	hasEventsTable := func() bool {
		if sdb == nil { return false }
		var name string
		err := sdb.DB().QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='run_events'").Scan(&name)
		return err == nil && name == "run_events"
	}()

	return clirunner.AddProgressListener(func(runID int64, p clirunner.DerivedProgress) {
		// speed 文本化（MiB/s）
		speed := fmt.Sprintf("%.2f MiB/s", p.SpeedBytesPerSec/1024.0/1024.0)
		_ = db.UpdateRunProgressByJobId(runID, int64(p.Bytes), speed)

		if hasEventsTable {
			// 将采样写入 run_events（按 runs.id，对应 rc_job_id=runID 的记录）
			// 找到 run 记录 id
			var id int64
			err := sdb.DB().QueryRow("SELECT id FROM runs WHERE rc_job_id = ? ORDER BY id DESC LIMIT 1", runID).Scan(&id)
			if err == nil && id > 0 {
				_, _ = sdb.DB().Exec(
					"INSERT INTO run_events(run_id, bytes, total_bytes, percent, speed_bps, eta_sec) VALUES(?,?,?,?,?,?)",
					id, p.Bytes, p.TotalBytes, p.Percent, p.SpeedBytesPerSec, p.EtaSeconds,
				)
			}
		}
	})
}

// DB 方法暴露（轻量封装）
func (db *store.DB) DB() *sql.DB { return db.Raw() }
