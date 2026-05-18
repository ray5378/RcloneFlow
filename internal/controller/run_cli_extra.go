package controller

import (
	"net/http"
	"strconv"
	"strings"
)

// HandleRunStopCLI POST /api/runs/{id}/stop
func (c *RunController) HandleRunStopCLI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	idStr = strings.TrimSuffix(idStr, "/stop")
	id, _ := strconv.ParseInt(strings.Trim(idStr, "/"), 10, 64)
	if id <= 0 {
		WriteJSON(w, 400, map[string]any{"error": "invalid id"})
		return
	}
	// Runner 在 service 层异步创建，这里简化为标记停止：交由后端 Runner 处理（下一步可扩展为全局 Runner）
	// 兼容先期：直接更新状态为 stopped（如需硬停，后续接 Global Runner 实例）
	c.runSvc.UpdateRunStatus(id, map[string]any{"finished": true, "success": false, "error": "stopped by user"})
	WriteJSON(w, 200, map[string]any{"stopped": true})
}

// (deprecated) 旧版 /api/runs/{id}/log 接口已移至 run.go，仅保留占位避免编译错误
// func (c *RunController) HandleRunLog(w http.ResponseWriter, r *http.Request) {}
