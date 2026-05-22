package controller

import (
	"encoding/json"
	"errors"
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

type AuthController struct {
	authSvc *service.AuthService
}

func NewAuthController(authSvc *service.AuthService) *AuthController {
	return &AuthController{authSvc: authSvc}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

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

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid request"})
		return
	}

	user, tokens, err := c.authSvc.Register(req.Username, req.Password)
	if err != nil {
		code, msg := mapAuthError(err)
		WriteJSON(w, code, map[string]any{"error": msg})
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
		code, msg := mapAuthError(err)
		WriteJSON(w, code, map[string]any{"error": msg})
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

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid request"})
		return
	}

	tokens, err := c.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		code, msg := mapAuthError(err)
		WriteJSON(w, code, map[string]any{"error": msg})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
	Username    string `json:"username,omitempty"`
}

func (c *AuthController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]any{"error": "invalid request"})
		return
	}

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

	updatedUser, err := c.authSvc.ChangeProfile(user.ID, req.OldPassword, req.NewPassword, req.Username)
	if err != nil {
		code, msg := mapAuthError(err)
		WriteJSON(w, code, map[string]any{"error": msg})
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

func mapAuthError(err error) (statusCode int, message string) {
	switch {
	case errors.Is(err, service.ErrAuthEmptyCredentials):
		return 400, "username and password required"
	case errors.Is(err, service.ErrAuthPasswordTooShort):
		return 400, "password must be at least 6 characters"
	case errors.Is(err, service.ErrAuthUserExists):
		return 409, "用户名已存在"
	case errors.Is(err, service.ErrAuthInvalidCredential):
		return 401, "用户名或密码错误"
	case errors.Is(err, service.ErrAuthRefreshRequired):
		return 400, "refreshToken required"
	case errors.Is(err, service.ErrAuthRefreshInvalid):
		return 401, "invalid or expired refreshToken"
	case errors.Is(err, service.ErrAuthOldPasswordEmpty):
		return 400, "请提供旧密码"
	case errors.Is(err, service.ErrAuthOldPasswordWrong):
		return 401, "旧密码错误"
	case errors.Is(err, service.ErrAuthUsernameTaken):
		return 409, "用户名已被占用"
	case errors.Is(err, service.ErrAuthUserNotFound):
		return 404, "user not found"
	default:
		return 500, "internal error"
	}
}