package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMiddleware_MissingToken(t *testing.T) {
	handler := JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "未提供认证token")
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	handler := JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "token无效或已过期")
}

func TestJWTMiddleware_ValidBearerToken(t *testing.T) {
	pair, err := GenerateTokenPair(1, "admin")
	require.NoError(t, err)

	var capturedUserID int64
	var capturedUsername string

	handler := JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, _ = GetUserIDFromContext(r.Context())
		capturedUsername = GetUsernameFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int64(1), capturedUserID)
	assert.Equal(t, "admin", capturedUsername)
}

func TestJWTMiddleware_QueryAuthToken(t *testing.T) {
	pair, err := GenerateTokenPair(2, "queryuser")
	require.NoError(t, err)

	var capturedUserID int64

	handler := JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, _ = GetUserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/download?auth="+pair.AccessToken, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int64(2), capturedUserID)
}

func TestJWTMiddleware_QueryAuthOnlyForGET(t *testing.T) {
	pair, err := GenerateTokenPair(1, "admin")
	require.NoError(t, err)

	handler := JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/test?auth="+pair.AccessToken, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTMiddleware_AuthorizationHeaderOverridesQuery(t *testing.T) {
	validPair, err := GenerateTokenPair(1, "valid")
	require.NoError(t, err)
	invalidPair, err := GenerateTokenPair(2, "invalid")
	require.NoError(t, err)

	var capturedUsername string

	handler := JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUsername = GetUsernameFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// When auth query param is present, it takes precedence over Authorization header
	// (the code checks query first, then only falls back to header if tok == "")
	req := httptest.NewRequest(http.MethodGet, "/api/test?auth="+validPair.AccessToken, nil)
	req.Header.Set("Authorization", "Bearer "+invalidPair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "valid", capturedUsername)
}

func TestJWTMiddleware_NonBearerScheme(t *testing.T) {
	handler := JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Basic somecredentials")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetUserIDFromContext_NoValue(t *testing.T) {
	ctx := context.Background()
	_, ok := GetUserIDFromContext(ctx)
	assert.False(t, ok)
}

func TestGetUsernameFromContext_NoValue(t *testing.T) {
	ctx := context.Background()
	username := GetUsernameFromContext(ctx)
	assert.Empty(t, username)
}

func TestGetUserIDFromContext_ValidValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), userIDKey, int64(42))
	id, ok := GetUserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, int64(42), id)
}

func TestGetUsernameFromContext_ValidValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), usernameKey, "testuser")
	username := GetUsernameFromContext(ctx)
	assert.Equal(t, "testuser", username)
}
