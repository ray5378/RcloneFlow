package app

import "net/http"

// attachCLIRoutes 迁移完成后不再挂载任何临时路由（保留空实现以兼容调用处）。
func attachCLIRoutes(mux *http.ServeMux) {}
