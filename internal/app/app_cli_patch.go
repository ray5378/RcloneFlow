package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/rcloneflow/internal/router"
)

// attachCLIRoutes 在原有 ServeMux 之上，挂载 CLI 运行器的临时路由。
func attachCLIRoutes(mux *http.ServeMux) {
	engine := gin.New()
	engine.Use(gin.Recovery())
	router.SetupCLI(mux, engine)
}
