package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/rcloneflow/internal/controller"
)

// SetupCLI 在现有 http.ServeMux 之外，额外注册用于 CLI 运行器的最小路由（临时接入）。
// 注意：仓库原生使用 net/http + ServeMux，这里用 gin 仅为演示占位，后续将统一风格。
func SetupCLI(mux *http.ServeMux, engine *gin.Engine) {
	// 最小三条：启动/停止/进度
	engine.POST("/api/runs/:id/start_cli", controller.StartRunCLI)
	engine.POST("/api/runs/:id/stop_cli", controller.StopRunCLI)
	engine.GET("/api/runs/:id/progress_cli", controller.ProgressRunCLI)

	// 将 gin 挂到 /api/cli/* 前缀（临时）
	mux.Handle("/api/cli/", http.StripPrefix("/api/cli", engine))
}
