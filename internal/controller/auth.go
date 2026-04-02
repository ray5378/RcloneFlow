package controller

import (
	"encoding/json"
	"net/http"

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

	// 生成token
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, `{"error":"生成token失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token": token,
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

	// 生成token
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, `{"error":"生成token失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token": token,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}
