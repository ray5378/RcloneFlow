package router

import (
	"net/http"

	"rcloneflow/internal/controller"
)

// SetupCLI 在现有 http.ServeMux 上注册 CLI 运行器的最小路由（临时接入）。
func SetupCLI(mux *http.ServeMux) {
	mux.HandleFunc("/api/cli/runs/start/", controller.StartRunCLIHandler)
	mux.HandleFunc("/api/cli/runs/stop/", controller.StopRunCLIHandler)
	mux.HandleFunc("/api/cli/runs/progress/", controller.ProgressRunCLIHandler)
}
