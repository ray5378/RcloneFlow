package service

import (
	"rcloneflow/internal/auth"
	"rcloneflow/internal/store"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *store.DB
}

func NewAuthService(db *store.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) HasUsers() (bool, error) {
	users, err := s.db.ListUsers()
	if err != nil {
		return false, err
	}
	return len(users) > 0, nil
}

func (s *AuthService) Register(username, password string) (*store.User, *auth.TokenPair, error) {
	if username == "" || password == "" {
		return nil, nil, ErrAuthEmptyCredentials
	}
	if len(password) < 6 {
		return nil, nil, ErrAuthPasswordTooShort
	}

	if _, exists := s.db.GetUserByUsername(username); exists {
		return nil, nil, ErrAuthUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.db.CreateUser(username, string(hashedPassword))
	if err != nil {
		return nil, nil, err
	}

	tokens, err := auth.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func (s *AuthService) Login(username, password string) (*store.User, *auth.TokenPair, error) {
	if username == "" || password == "" {
		return nil, nil, ErrAuthEmptyCredentials
	}

	user, exists := s.db.GetUserByUsername(username)
	if !exists {
		return nil, nil, ErrAuthInvalidCredential
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, nil, ErrAuthInvalidCredential
	}

	tokens, err := auth.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*auth.TokenPair, error) {
	if refreshToken == "" {
		return nil, ErrAuthRefreshRequired
	}

	tokens, err := auth.RefreshTokens(refreshToken)
	if err != nil {
		return nil, ErrAuthRefreshInvalid
	}

	return tokens, nil
}

func (s *AuthService) ChangeProfile(userID int64, oldPassword, newPassword, newUsername string) (*store.User, error) {
	user, exists := s.db.GetUserByID(userID)
	if !exists {
		return nil, ErrAuthUserNotFound
	}

	if newPassword != "" {
		if oldPassword == "" {
			return nil, ErrAuthOldPasswordEmpty
		}
		if len(newPassword) < 6 {
			return nil, ErrAuthPasswordTooShort
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
			return nil, ErrAuthOldPasswordWrong
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		if err := s.db.UpdatePassword(user.ID, string(hashedPassword)); err != nil {
			return nil, err
		}

		user, _ = s.db.GetUserByID(user.ID)
	}

	if newUsername != "" && newUsername != user.Username {
		if existingUser, exists := s.db.GetUserByUsername(newUsername); exists && existingUser.ID != user.ID {
			return nil, ErrAuthUsernameTaken
		}

		if err := s.db.UpdateUsername(user.ID, newUsername); err != nil {
			return nil, err
		}

		user, _ = s.db.GetUserByID(user.ID)
	}

	return &user, nil
}

func (s *AuthService) GetUserByID(id int64) (*store.User, bool) {
	user, exists := s.db.GetUserByID(id)
	if !exists {
		return nil, false
	}
	return &user, true
}

func (s *AuthService) GetUserByUsername(username string) (*store.User, bool) {
	user, exists := s.db.GetUserByUsername(username)
	if !exists {
		return nil, false
	}
	return &user, true
}