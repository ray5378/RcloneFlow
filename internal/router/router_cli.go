package router

import (
	"net/http"

	"rcloneflow/internal/controller"
)

// SetupCLI 在现有 http.ServeMux 上注册 CLI 运行器 + providers/config 的临时路由。
func SetupCLI(mux *http.ServeMux) {
	// 运行器
	mux.HandleFunc("/api/cli/runs/start/", controller.StartRunCLIHandler)
	mux.HandleFunc("/api/cli/runs/stop/", controller.StopRunCLIHandler)
	mux.HandleFunc("/api/cli/runs/progress/", controller.ProgressRunCLIHandler)
	// provider/config
	mux.HandleFunc("/api/cli/providers", controller.ProviderListCLIHandler)
	mux.HandleFunc("/api/cli/remotes/create", controller.RemoteCreateCLIHandler)
	mux.HandleFunc("/api/cli/remotes/test", controller.RemoteTestCLIHandler)
	mux.HandleFunc("/api/cli/remotes/dump", controller.RemoteDumpCLIHandler)
}
