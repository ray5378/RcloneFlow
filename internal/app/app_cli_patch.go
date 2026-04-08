package app

import (
	"net/http"

	"rcloneflow/internal/router"
)

// attachCLIRoutes 在原有 ServeMux 之上，挂载 CLI 运行器的临时路由。
func attachCLIRoutes(mux *http.ServeMux) {
	router.SetupCLI(mux)
}
