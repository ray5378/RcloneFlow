package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"rcloneflow/internal/auth"
	"rcloneflow/internal/service"
)

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

var loginLimiter = &rateLimiter{
	attempts: make(map[string][]time.Time),
	max:      5,
	window:   5 * time.Minute,
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)
	var kept []time.Time
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= rl.max {
		rl.attempts[key] = kept
		return false
	}
	rl.attempts[key] = append(kept, now)
	return true
}

// AuthController 认证控制器
type AuthController struct {
	authSvc *service.AuthService
}

// NewAuthController 创建认证控制器
func NewAuthController(authSvc *service.AuthService) *AuthController {
	return &AuthController{authSvc: authSvc}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HasUsers 检查是否存在用户
func (c *AuthController) HasUsers(w http.ResponseWriter, r *http.Request) {
	exists, err := c.authSvc.HasUsers()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": "internal error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"exists": exists,
	})
}

// Register 注册用户
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid request"})
		return
	}

	user, tokens, err := c.authSvc.Register(req.Username, req.Password)
	if err != nil {
		switch err.Error() {
		case "username and password required":
			WriteJSON(w, 400, map[string]any{"error": "username and password required"})
		case "password must be at least 6 characters":
			WriteJSON(w, 400, map[string]any{"error": "password must be at least 6 characters"})
		case "用户名已存在":
			WriteJSON(w, 409, map[string]any{"error": "用户名已存在"})
		default:
			WriteJSON(w, 500, map[string]any{"error": "internal error"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

// Login 登录
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr
	if !loginLimiter.allow(ip) {
		WriteJSON(w, 429, map[string]any{"error": "尝试次数过多，请稍后再试"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid request"})
		return
	}

	user, tokens, err := c.authSvc.Login(req.Username, req.Password)
	if err != nil {
		switch err.Error() {
		case "username and password required":
			WriteJSON(w, 400, map[string]any{"error": "username and password required"})
		case "用户名或密码错误":
			WriteJSON(w, 401, map[string]any{"error": "用户名或密码错误"})
		default:
			WriteJSON(w, 500, map[string]any{"error": "internal error"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// Refresh 刷新令牌
func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid request"})
		return
	}

	tokens, err := c.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		switch err.Error() {
		case "refreshToken required":
			WriteJSON(w, 400, map[string]any{"error": "refreshToken required"})
		case "invalid or expired refreshToken":
			WriteJSON(w, 401, map[string]any{"error": "invalid or expired refreshToken"})
		default:
			WriteJSON(w, 500, map[string]any{"error": "internal error"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
	Username    string `json:"username,omitempty"`
}

// ChangePassword 修改密码和用户名
func (c *AuthController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid request"})
		return
	}

	// 从Authorization header获取当前用户
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		WriteJSON(w, 401, map[string]any{"error": "authentication required"})
		return
	}

	// 解析token获取用户名
	claims, err := auth.ValidateToken(strings.TrimPrefix(authHeader, "Bearer "))
	if err != nil {
		WriteJSON(w, 401, map[string]any{"error": "invalid token"})
		return
	}

	// 获取用户信息
	user, exists := c.authSvc.GetUserByUsername(claims.Username)
	if !exists {
		WriteJSON(w, 404, map[string]any{"error": "user not found"})
		return
	}

	// 调用服务层修改配置
	updatedUser, err := c.authSvc.ChangeProfile(user.ID, req.OldPassword, req.NewPassword, req.Username)
	if err != nil {
		switch err.Error() {
		case "请提供旧密码":
			WriteJSON(w, 400, map[string]any{"error": "请提供旧密码"})
		case "password must be at least 6 characters":
			WriteJSON(w, 400, map[string]any{"error": "password must be at least 6 characters"})
		case "旧密码错误":
			WriteJSON(w, 401, map[string]any{"error": "旧密码错误"})
		case "用户名已被占用":
			WriteJSON(w, 409, map[string]any{"error": "用户名已被占用"})
		case "user not found":
			WriteJSON(w, 404, map[string]any{"error": "user not found"})
		default:
			WriteJSON(w, 500, map[string]any{"error": "internal error"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "修改成功",
		"user": map[string]any{
			"id":       updatedUser.ID,
			"username": updatedUser.Username,
		},
	})
}

// Me 验证当前用户token并返回用户信息
func (c *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		WriteJSON(w, 401, map[string]any{"error": "authentication required"})
		return
	}

	claims, err := auth.ValidateToken(strings.TrimPrefix(authHeader, "Bearer "))
	if err != nil {
		WriteJSON(w, 401, map[string]any{"error": "invalid token"})
		return
	}

	user, exists := c.authSvc.GetUserByUsername(claims.Username)
	if !exists {
		WriteJSON(w, 404, map[string]any{"error": "user not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}
