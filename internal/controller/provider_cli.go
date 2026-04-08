package controller

import (
	"net/http"

	"rcloneflow/internal/rclone"
)

// ProviderListCLIHandler 返回内置的 provider 元数据最小子集。
func ProviderListCLIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	WriteJSON(w, 200, map[string]any{"providers": rclone.StaticProviders()})
}
