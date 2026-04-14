package auth

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("rcloneflow-secret-key-change-in-production")

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Claims JWT声明
type Claims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func parseTTL(envKey string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(envKey))
	if v == "" {
		return def
	}
	// 支持标准 time.ParseDuration，如 "24h"；也支持 "90d"（天）
	if strings.HasSuffix(v, "d") {
		num := strings.TrimSuffix(v, "d")
		if n, err := strconv.Atoi(num); err == nil && n > 0 {
			return time.Duration(n) * 24 * time.Hour
		}
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	return def
}

func accessTTL() time.Duration  { return parseTTL("ACCESS_TOKEN_TTL", 24*time.Hour) }
func refreshTTL() time.Duration { return parseTTL("REFRESH_TOKEN_TTL", 90*24*time.Hour) }

// GenerateTokenPair 生成访问令牌和刷新令牌
func GenerateTokenPair(userID int64, username string) (*TokenPair, error) {
	// 访问令牌（默认 24h）
	accessClaims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTTL())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	// 刷新令牌（默认 90d）
	refreshClaims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTTL())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
	}, nil
}

// ValidateToken 验证JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshTokens 使用刷新令牌获取新的令牌对
func RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("refresh token invalid")
	}
	return GenerateTokenPair(claims.UserID, claims.Username)
}
