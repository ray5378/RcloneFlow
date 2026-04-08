package app

import (
	"net/http"

	"rcloneflow/internal/router"
)

// attachCLIRoutes 在迁移期挂载 CLI 临时路由；即将统一回正式 API 后移除。
func attachCLIRoutes(mux *http.ServeMux) {
	router.SetupCLI(mux)
}
