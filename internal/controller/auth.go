package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"rcloneflow/internal/auth"
	"rcloneflow/internal/store"

	"golang.org/x/crypto/bcrypt"
)

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

// Register 注册用户
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"无效的请求格式"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"用户名和密码不能为空"}`, http.StatusBadRequest)
		return
	}

	// 检查用户是否已存在
	if _, exists := c.db.GetUserByUsername(req.Username); exists {
		http.Error(w, `{"error":"用户名已存在"}`, http.StatusConflict)
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"密码加密失败"}`, http.StatusInternalServerError)
		return
	}

	// 创建用户
	user, err := c.db.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, `{"error":"创建用户失败"}`, http.StatusInternalServerError)
		return
	}

	// 生成token对
	tokens, err := auth.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		http.Error(w, `{"error":"生成token失败"}`, http.StatusInternalServerError)
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
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"无效的请求格式"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"用户名和密码不能为空"}`, http.StatusBadRequest)
		return
	}

	// 查找用户
	user, exists := c.db.GetUserByUsername(req.Username)
	if !exists {
		http.Error(w, `{"error":"用户名或密码错误"}`, http.StatusUnauthorized)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"用户名或密码错误"}`, http.StatusUnauthorized)
		return
	}

	// 生成token对
	tokens, err := auth.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		http.Error(w, `{"error":"生成token失败"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"error":"无效的请求格式"}`, http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, `{"error":"refreshToken不能为空"}`, http.StatusBadRequest)
		return
	}

	// 使用刷新令牌获取新的令牌对
	tokens, err := auth.RefreshTokens(req.RefreshToken)
	if err != nil {
		http.Error(w, `{"error":"refreshToken无效或已过期"}`, http.StatusUnauthorized)
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
		http.Error(w, `{"error":"无效的请求格式"}`, http.StatusBadRequest)
		return
	}

	// 从Authorization header获取当前用户
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error":"未提供认证token"}`, http.StatusUnauthorized)
		return
	}

	// 解析token获取用户名
	claims, err := auth.ValidateToken(strings.TrimPrefix(authHeader, "Bearer "))
	if err != nil {
		http.Error(w, `{"error":"token无效"}`, http.StatusUnauthorized)
		return
	}

	// 获取用户信息
	user, exists := c.db.GetUserByUsername(claims.Username)
	if !exists {
		http.Error(w, `{"error":"用户不存在"}`, http.StatusNotFound)
		return
	}

	// 如果提供了新密码，则验证旧密码并更新
	if req.NewPassword != "" {
		if req.OldPassword == "" {
			http.Error(w, `{"error":"请提供旧密码"}`, http.StatusBadRequest)
			return
		}

		// 验证旧密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
			http.Error(w, `{"error":"旧密码错误"}`, http.StatusUnauthorized)
			return
		}

		// 加密新密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error":"密码加密失败"}`, http.StatusInternalServerError)
			return
		}

		// 更新密码
		if err := c.db.UpdatePassword(user.ID, string(hashedPassword)); err != nil {
			http.Error(w, `{"error":"更新密码失败"}`, http.StatusInternalServerError)
			return
		}
	}

	// 如果提供了新用户名，则更新用户名
	if req.Username != "" && req.Username != user.Username {
		// 检查新用户名是否已被占用
		if existingUser, exists := c.db.GetUserByUsername(req.Username); exists && existingUser.ID != user.ID {
			http.Error(w, `{"error":"用户名已被占用"}`, http.StatusConflict)
			return
		}

		if err := c.db.UpdateUsername(user.ID, req.Username); err != nil {
			http.Error(w, `{"error":"更新用户名失败"}`, http.StatusInternalServerError)
			return
		}

		// 更新localStorage中的用户信息（前端通过重新登录处理）
	}

	w.Header().Set("Content-Type", "application/json")

	finalUsername := req.Username
	if finalUsername == "" {
		finalUsername = user.Username
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "修改成功",
		"user": map[string]any{
			"id":       user.ID,
			"username": finalUsername,
		},
	})
}
