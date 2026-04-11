package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey 用户上下文key
type contextKey string

const userIDKey contextKey = "userID"
const usernameKey contextKey = "username"

// JWTMiddleware JWT认证中间件
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"未提供认证token"}`, http.StatusUnauthorized)
			return
		}

		// 检查Bearer token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"无效的认证格式，请使用Bearer token"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// 验证token
		claims, err := ValidateToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"token无效或已过期"}`, http.StatusUnauthorized)
			return
		}

		// 将用户信息存入context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, usernameKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext 从context获取用户ID
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// GetUsernameFromContext 从context获取用户名
func GetUsernameFromContext(ctx context.Context) string {
	username, _ := ctx.Value(usernameKey).(string)
	return username
}
