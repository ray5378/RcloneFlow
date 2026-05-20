package middleware

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func readSettingsCorsOrigins() string {
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		return v
	}
	dataDir := os.Getenv("APP_DATA_DIR")
	if dataDir == "" {
		dataDir = "."
	}
	fp := filepath.Join(dataDir, "settings.json")
	b, err := os.ReadFile(fp)
	if err != nil || len(b) == 0 {
		return ""
	}
	var m map[string]string
	if json.Unmarshal(b, &m) != nil {
		return ""
	}
	return strings.TrimSpace(m["CORS_ALLOWED_ORIGINS"])
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := readSettingsCorsOrigins()
		origin := r.Header.Get("Origin")
		if allowedOrigins != "" && origin != "" {
			for _, o := range strings.Split(allowedOrigins, ",") {
				if strings.TrimSpace(o) == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
			if w.Header().Get("Access-Control-Allow-Origin") == "" {
				w.Header().Set("Access-Control-Allow-Origin", strings.Split(allowedOrigins, ",")[0])
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}
