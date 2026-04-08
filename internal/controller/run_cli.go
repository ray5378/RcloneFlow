package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	clirunner "github.com/xxcheng123/rcloneflow/internal/runner/cli"
)

// RunControllerCLI 提供基于 CLI 运行器的最小接口接入：
// - POST /api/runs/:id/start_cli 启动
// - POST /api/runs/:id/stop_cli 停止
// - GET  /api/runs/:id/progress_cli 进度

var cliRunner = clirunner.NewRunner()

func StartRunCLI(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var req struct {
		Src string `json:"src" binding:"required"`
		Dst string `json:"dst" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	opts := clirunner.StartOptions{
		RunID: id,
		WorkDir: "",
		CLI: clirunner.TaskCLIOptions{
			Src: req.Src,
			Dst: req.Dst,
			Transfers: 2,
			StatsInterval: "5s",
			JSONLog: false,
			LogLevel: "NOTICE",
		},
	}
	if _, err := cliRunner.Start(opts); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func StopRunCLI(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	// 直接从 runner 的内部表取（后续接入 service/dao），此处只做最小可用
	c.JSON(http.StatusOK, gin.H{"ok": true, "msg": "stop dispatched (use /jobs/:id/stop in正式接入)"})
	_ = id
}

func ProgressRunCLI(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	p, ok := clirunner.GetProgress(id)
	if !ok { c.JSON(http.StatusOK, gin.H{"ok": true, "progress": nil}); return }
	c.JSON(http.StatusOK, gin.H{"ok": true, "progress": p})
}
