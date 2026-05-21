package service

import (
	"fmt"

	"rcloneflow/internal/auth"
	"rcloneflow/internal/store"

	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct {
	db *store.DB
}

// NewAuthService 创建认证服务
func NewAuthService(db *store.DB) *AuthService {
	return &AuthService{db: db}
}

// HasUsers 检查是否存在用户
func (s *AuthService) HasUsers() (bool, error) {
	users, err := s.db.ListUsers()
	if err != nil {
		return false, err
	}
	return len(users) > 0, nil
}

// Register 注册用户
func (s *AuthService) Register(username, password string) (*store.User, *auth.TokenPair, error) {
	if username == "" || password == "" {
		return nil, nil, fmt.Errorf("username and password required")
	}
	if len(password) < 6 {
		return nil, nil, fmt.Errorf("password must be at least 6 characters")
	}

	if _, exists := s.db.GetUserByUsername(username); exists {
		return nil, nil, fmt.Errorf("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("internal error")
	}

	user, err := s.db.CreateUser(username, string(hashedPassword))
	if err != nil {
		return nil, nil, fmt.Errorf("internal error")
	}

	tokens, err := auth.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, nil, fmt.Errorf("internal error")
	}

	return &user, tokens, nil
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (*store.User, *auth.TokenPair, error) {
	if username == "" || password == "" {
		return nil, nil, fmt.Errorf("username and password required")
	}

	user, exists := s.db.GetUserByUsername(username)
	if !exists {
		return nil, nil, fmt.Errorf("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, nil, fmt.Errorf("用户名或密码错误")
	}

	tokens, err := auth.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, nil, fmt.Errorf("internal error")
	}

	return &user, tokens, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(refreshToken string) (*auth.TokenPair, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refreshToken required")
	}

	tokens, err := auth.RefreshTokens(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refreshToken")
	}

	return tokens, nil
}

// ChangeProfile 修改密码和用户名
func (s *AuthService) ChangeProfile(userID int64, oldPassword, newPassword, newUsername string) (*store.User, error) {
	user, exists := s.db.GetUserByID(userID)
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	if newPassword != "" {
		if oldPassword == "" {
			return nil, fmt.Errorf("请提供旧密码")
		}
		if len(newPassword) < 6 {
			return nil, fmt.Errorf("password must be at least 6 characters")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
			return nil, fmt.Errorf("旧密码错误")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("internal error")
		}

		if err := s.db.UpdatePassword(user.ID, string(hashedPassword)); err != nil {
			return nil, fmt.Errorf("internal error")
		}

		user, _ = s.db.GetUserByID(user.ID)
	}

	if newUsername != "" && newUsername != user.Username {
		if existingUser, exists := s.db.GetUserByUsername(newUsername); exists && existingUser.ID != user.ID {
			return nil, fmt.Errorf("用户名已被占用")
		}

		if err := s.db.UpdateUsername(user.ID, newUsername); err != nil {
			return nil, fmt.Errorf("internal error")
		}

		user, _ = s.db.GetUserByID(user.ID)
	}

	return &user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *AuthService) GetUserByID(id int64) (*store.User, bool) {
	user, exists := s.db.GetUserByID(id)
	if !exists {
		return nil, false
	}
	return &user, true
}

// GetUserByUsername 根据用户名获取用户
func (s *AuthService) GetUserByUsername(username string) (*store.User, bool) {
	user, exists := s.db.GetUserByUsername(username)
	if !exists {
		return nil, false
	}
	return &user, true
}
