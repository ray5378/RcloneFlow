package controller

import (
	"encoding/json"
	"net/http"
)

// ResponseWriter 封装HTTP响应
type ResponseWriter struct {
	http.ResponseWriter
}

// WriteJSON 统一JSON响应
func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// DecodeRequest 解码请求体
func DecodeRequest(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}
