package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"rcloneflow/internal/auth"
	"rcloneflow/internal/store"

	"golang.org/x/crypto/bcrypt"
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
	db *store.DB
}

// NewAuthController 创建认证控制器
func NewAuthController(db *store.DB) *AuthController {
	return &AuthController{db: db}
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
	users, err := c.db.ListUsers()
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": "internal error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"exists": len(users) > 0,
	})
}

// Register 注册用户
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid request"})
		return
	}

	if req.Username == "" || req.Password == "" {
		WriteJSON(w, 400, map[string]any{"error": "username and password required"})
		return
	}

	if len(req.Password) < 6 {
		WriteJSON(w, 400, map[string]any{"error": "password must be at least 6 characters"})
		return
	}

	// 检查用户是否已存在
	if _, exists := c.db.GetUserByUsername(req.Username); exists {
		WriteJSON(w, 409, map[string]any{"error": "用户名已存在"})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": "internal error"})
		return
	}

	// 创建用户
	user, err := c.db.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": "internal error"})
		return
	}

	// 生成token对
	tokens, err := auth.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": "internal error"})
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

	if req.Username == "" || req.Password == "" {
		WriteJSON(w, 400, map[string]any{"error": "username and password required"})
		return
	}

	// 查找用户
	user, exists := c.db.GetUserByUsername(req.Username)
	if !exists {
		WriteJSON(w, 401, map[string]any{"error": "用户名或密码错误"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		WriteJSON(w, 401, map[string]any{"error": "用户名或密码错误"})
		return
	}

	// 生成token对
	tokens, err := auth.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		WriteJSON(w, 500, map[string]any{"error": "internal error"})
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

	if req.RefreshToken == "" {
		WriteJSON(w, 400, map[string]any{"error": "refreshToken required"})
		return
	}

	// 使用刷新令牌获取新的令牌对
	tokens, err := auth.RefreshTokens(req.RefreshToken)
	if err != nil {
		WriteJSON(w, 401, map[string]any{"error": "invalid or expired refreshToken"})
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
	user, exists := c.db.GetUserByUsername(claims.Username)
	if !exists {
		WriteJSON(w, 404, map[string]any{"error": "user not found"})
		return
	}

	// 如果提供了新密码，则验证旧密码并更新
	if req.NewPassword != "" {
		if req.OldPassword == "" {
			WriteJSON(w, 400, map[string]any{"error": "请提供旧密码"})
			return
		}
		if len(req.NewPassword) < 6 {
			WriteJSON(w, 400, map[string]any{"error": "password must be at least 6 characters"})
			return
		}

		// 验证旧密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
			WriteJSON(w, 401, map[string]any{"error": "旧密码错误"})
			return
		}

		// 加密新密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			WriteJSON(w, 500, map[string]any{"error": "internal error"})
			return
		}

		// 更新密码
		if err := c.db.UpdatePassword(user.ID, string(hashedPassword)); err != nil {
			WriteJSON(w, 500, map[string]any{"error": "internal error"})
			return
		}

		// 刷新用户信息
		user, _ = c.db.GetUserByID(user.ID)
	}

	// 如果提供了新用户名，则更新用户名
	if req.Username != "" && req.Username != user.Username {
		// 检查新用户名是否已被占用
		if existingUser, exists := c.db.GetUserByUsername(req.Username); exists && existingUser.ID != user.ID {
			WriteJSON(w, 409, map[string]any{"error": "用户名已被占用"})
			return
		}

		if err := c.db.UpdateUsername(user.ID, req.Username); err != nil {
			WriteJSON(w, 500, map[string]any{"error": "internal error"})
			return
		}

		// 刷新用户信息
		user, _ = c.db.GetUserByID(user.ID)
	}

	w.Header().Set("Content-Type", "application/json")

	finalUsername := user.Username

	json.NewEncoder(w).Encode(map[string]any{
		"message": "修改成功",
		"user": map[string]any{
			"id":       user.ID,
			"username": finalUsername,
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

	user, exists := c.db.GetUserByUsername(claims.Username)
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
