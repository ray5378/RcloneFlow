package service

import (
	"fmt"
)

// cleanupEvents 清理 run_events 表中过期数据（若表存在）
func (s *CleanupService) cleanupEvents() {
	if s.dataDB == nil || s.retention <= 0 { return }
	// 检查表是否存在
	var name string
	err := s.dataDB.Raw().QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='run_events'").Scan(&name)
	if err != nil || name != "run_events" { return }
	// 删除过期事件
	_, _ = s.dataDB.Raw().Exec("DELETE FROM run_events WHERE created_at < datetime('now', ?)",
		fmt.Sprintf("-%d days", s.retention))
}
