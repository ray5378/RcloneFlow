package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTTL_Default(t *testing.T) {
	d := parseTTL("NONEXISTENT_ENV_VAR_FOR_TEST", 5*time.Minute)
	assert.Equal(t, 5*time.Minute, d)
}

func TestParseTTL_Duration(t *testing.T) {
	os.Setenv("TEST_TTL", "30m")
	defer os.Unsetenv("TEST_TTL")
	d := parseTTL("TEST_TTL", 5*time.Minute)
	assert.Equal(t, 30*time.Minute, d)
}

func TestParseTTL_Days(t *testing.T) {
	os.Setenv("TEST_TTL_DAYS", "7d")
	defer os.Unsetenv("TEST_TTL_DAYS")
	d := parseTTL("TEST_TTL_DAYS", 5*time.Minute)
	assert.Equal(t, 7*24*time.Hour, d)
}

func TestParseTTL_InvalidDays(t *testing.T) {
	os.Setenv("TEST_TTL_BAD_DAYS", "xd")
	defer os.Unsetenv("TEST_TTL_BAD_DAYS")
	d := parseTTL("TEST_TTL_BAD_DAYS", 10*time.Minute)
	assert.Equal(t, 10*time.Minute, d)
}

func TestParseTTL_InvalidDuration(t *testing.T) {
	os.Setenv("TEST_TTL_BAD", "notaduration")
	defer os.Unsetenv("TEST_TTL_BAD")
	d := parseTTL("TEST_TTL_BAD", 15*time.Minute)
	assert.Equal(t, 15*time.Minute, d)
}

func TestAccessTTL_Default(t *testing.T) {
	ttl := accessTTL()
	assert.Equal(t, 24*time.Hour, ttl)
}

func TestRefreshTTL_Default(t *testing.T) {
	ttl := refreshTTL()
	assert.Equal(t, 90*24*time.Hour, ttl)
}

func TestGenerateTokenPair(t *testing.T) {
	pair, err := GenerateTokenPair(1, "admin")
	require.NoError(t, err)
	require.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEqual(t, pair.AccessToken, pair.RefreshToken)
}

func TestGenerateTokenPair_DifferentUsers(t *testing.T) {
	pair1, err := GenerateTokenPair(1, "admin")
	require.NoError(t, err)
	pair2, err := GenerateTokenPair(2, "user")
	require.NoError(t, err)
	assert.NotEqual(t, pair1.AccessToken, pair2.AccessToken)
}

func TestValidateToken_Valid(t *testing.T) {
	pair, err := GenerateTokenPair(42, "testuser")
	require.NoError(t, err)

	claims, err := ValidateToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestValidateToken_Invalid(t *testing.T) {
	_, err := ValidateToken("invalid.token.here")
	assert.Error(t, err)
}

func TestValidateToken_Empty(t *testing.T) {
	_, err := ValidateToken("")
	assert.Error(t, err)
}

func TestValidateToken_Tampered(t *testing.T) {
	pair, err := GenerateTokenPair(1, "admin")
	require.NoError(t, err)

	tampered := pair.AccessToken + "x"
	_, err = ValidateToken(tampered)
	assert.Error(t, err)
}

func TestRefreshTokens_Valid(t *testing.T) {
	pair, err := GenerateTokenPair(1, "admin")
	require.NoError(t, err)

	newPair, err := RefreshTokens(pair.RefreshToken)
	require.NoError(t, err)
	require.NotNil(t, newPair)
	assert.NotEmpty(t, newPair.AccessToken)
	assert.NotEmpty(t, newPair.RefreshToken)

	claims, err := ValidateToken(newPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "admin", claims.Username)
}

func TestRefreshTokens_Invalid(t *testing.T) {
	_, err := RefreshTokens("bad-refresh-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token invalid")
}

func TestRefreshTokens_Empty(t *testing.T) {
	_, err := RefreshTokens("")
	assert.Error(t, err)
}

func TestClaims_Structure(t *testing.T) {
	pair, err := GenerateTokenPair(100, "claimsuser")
	require.NoError(t, err)

	token, _, err := jwt.NewParser().ParseUnverified(pair.AccessToken, &Claims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(*Claims)
	require.True(t, ok)
	assert.Equal(t, int64(100), claims.UserID)
	assert.Equal(t, "claimsuser", claims.Username)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}
