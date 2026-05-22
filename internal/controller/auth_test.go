package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

func setupAuthTestDB(t *testing.T) *store.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_auth_test_*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := store.Open(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestAuthController_Register_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	body := `{"username":"testuser","password":"testpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	ctrl.Register(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp["accessToken"])
	assert.NotEmpty(t, resp["refreshToken"])
	user := resp["user"].(map[string]any)
	assert.Equal(t, "testuser", user["username"])
}

func TestAuthController_Register_Duplicate(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	body := `{"username":"dupuser","password":"pass123"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(body)))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	ctrl.Register(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(body)))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	ctrl.Register(rec2, req2)
	assert.Equal(t, http.StatusConflict, rec2.Code)
	assert.Contains(t, rec2.Body.String(), "用户名已存在")
}

func TestAuthController_Register_EmptyFields(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	tests := []struct {
		name string
		body string
	}{
		{"empty username", `{"username":"","password":"pass"}`},
		{"empty password", `{"username":"user","password":""}`},
		{"both empty", `{"username":"","password":""}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctrl.Register(rec, req)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func TestAuthController_Register_InvalidJSON(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Register(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthController_Login_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	regBody := `{"username":"loginuser","password":"loginpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(regBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Register(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	loginBody := `{"username":"loginuser","password":"loginpass123"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(loginBody)))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	ctrl.Login(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var resp map[string]any
	err := json.Unmarshal(rec2.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp["accessToken"])
	assert.NotEmpty(t, resp["refreshToken"])
}

func TestAuthController_Login_WrongPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	regBody := `{"username":"pwduser","password":"correctpass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(regBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Register(rec, req)

	loginBody := `{"username":"pwduser","password":"wrongpass"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(loginBody)))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	ctrl.Login(rec2, req2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)
	assert.Contains(t, rec2.Body.String(), "用户名或密码错误")
}

func TestAuthController_Login_UserNotFound(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	loginBody := `{"username":"nonexistent","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(loginBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Login(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthController_Login_EmptyFields(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	tests := []struct {
		name string
		body string
	}{
		{"empty username", `{"username":"","password":"pass"}`},
		{"empty password", `{"username":"user","password":""}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctrl.Login(rec, req)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func TestAuthController_Refresh_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	regBody := `{"username":"refreshuser","password":"refreshpass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(regBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Register(rec, req)

	var regResp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &regResp)
	require.NoError(t, err)
	refreshToken := regResp["refreshToken"].(string)

	refBody := `{"refreshToken":"` + refreshToken + `"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader([]byte(refBody)))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	ctrl.Refresh(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var refResp map[string]any
	err = json.Unmarshal(rec2.Body.Bytes(), &refResp)
	require.NoError(t, err)
	assert.NotEmpty(t, refResp["accessToken"])
	assert.NotEmpty(t, refResp["refreshToken"])
}

func TestAuthController_Refresh_EmptyToken(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	refBody := `{"refreshToken":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader([]byte(refBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Refresh(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthController_Refresh_InvalidToken(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	refBody := `{"refreshToken":"invalid.token.here"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader([]byte(refBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Refresh(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthController_ChangePassword_Success(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	regBody := `{"username":"chgpass","password":"oldpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(regBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Register(rec, req)

	var regResp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &regResp)
	require.NoError(t, err)
	accessToken := regResp["accessToken"].(string)

	chgBody := `{"oldPassword":"oldpass123","newPassword":"newpass456"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader([]byte(chgBody)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+accessToken)
	rec2 := httptest.NewRecorder()
	ctrl.ChangePassword(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var chgResp map[string]any
	err = json.Unmarshal(rec2.Body.Bytes(), &chgResp)
	require.NoError(t, err)
	assert.Equal(t, "修改成功", chgResp["message"])
}

func TestAuthController_ChangePassword_WrongOldPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	regBody := `{"username":"wrongold","password":"correctold"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(regBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Register(rec, req)

	var regResp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &regResp)
	require.NoError(t, err)
	accessToken := regResp["accessToken"].(string)

	chgBody := `{"oldPassword":"wrongpassword","newPassword":"newpass"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader([]byte(chgBody)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+accessToken)
	rec2 := httptest.NewRecorder()
	ctrl.ChangePassword(rec2, req2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)
	assert.Contains(t, rec2.Body.String(), "旧密码错误")
}

func TestAuthController_ChangePassword_MissingOldPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	regBody := `{"username":"missold","password":"pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(regBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Register(rec, req)

	var regResp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &regResp)
	require.NoError(t, err)
	accessToken := regResp["accessToken"].(string)

	chgBody := `{"oldPassword":"","newPassword":"newpass"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader([]byte(chgBody)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+accessToken)
	rec2 := httptest.NewRecorder()
	ctrl.ChangePassword(rec2, req2)
	assert.Equal(t, http.StatusBadRequest, rec2.Code)
	assert.Contains(t, rec2.Body.String(), "请提供旧密码")
}

func TestAuthController_ChangePassword_NoAuthHeader(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	chgBody := `{"oldPassword":"old","newPassword":"new"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader([]byte(chgBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.ChangePassword(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthController_ChangePassword_InvalidToken(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	chgBody := `{"oldPassword":"old","newPassword":"new"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader([]byte(chgBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid.token")
	rec := httptest.NewRecorder()
	ctrl.ChangePassword(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthController_ChangePassword_UpdateUsername(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	regBody := `{"username":"oldname","password":"pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(regBody)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.Register(rec, req)

	var regResp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &regResp)
	require.NoError(t, err)
	accessToken := regResp["accessToken"].(string)

	chgBody := `{"username":"newname"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader([]byte(chgBody)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+accessToken)
	rec2 := httptest.NewRecorder()
	ctrl.ChangePassword(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)

	var chgResp map[string]any
	err = json.Unmarshal(rec2.Body.Bytes(), &chgResp)
	require.NoError(t, err)
	user := chgResp["user"].(map[string]any)
	assert.Equal(t, "newname", user["username"])
}

func TestMapAuthError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   int
		wantMsg    string
	}{
		{"empty credentials", service.ErrAuthEmptyCredentials, 400, "username and password required"},
		{"password too short", service.ErrAuthPasswordTooShort, 400, "password must be at least 6 characters"},
		{"user exists", service.ErrAuthUserExists, 409, "用户名已存在"},
		{"invalid credential", service.ErrAuthInvalidCredential, 401, "用户名或密码错误"},
		{"refresh required", service.ErrAuthRefreshRequired, 400, "refreshToken required"},
		{"refresh invalid", service.ErrAuthRefreshInvalid, 401, "invalid or expired refreshToken"},
		{"old password empty", service.ErrAuthOldPasswordEmpty, 400, "请提供旧密码"},
		{"old password wrong", service.ErrAuthOldPasswordWrong, 401, "旧密码错误"},
		{"username taken", service.ErrAuthUsernameTaken, 409, "用户名已被占用"},
		{"user not found", service.ErrAuthUserNotFound, 404, "user not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg := mapAuthError(tt.err)
			assert.Equal(t, tt.wantCode, code)
			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}

func TestMapAuthError_Default(t *testing.T) {
	code, msg := mapAuthError(assert.AnError)
	assert.Equal(t, 500, code)
	assert.Equal(t, "internal error", msg)
}

func TestAuthController_ChangePassword_UsernameTaken(t *testing.T) {
	db := setupAuthTestDB(t)
	ctrl := NewAuthController(service.NewAuthService(db))

	regBody1 := `{"username":"user1","password":"pass123"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(regBody1)))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	ctrl.Register(rec1, req1)

	regBody2 := `{"username":"user2","password":"pass123"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(regBody2)))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	ctrl.Register(rec2, req2)

	var regResp map[string]any
	err := json.Unmarshal(rec2.Body.Bytes(), &regResp)
	require.NoError(t, err)
	accessToken := regResp["accessToken"].(string)

	chgBody := `{"username":"user1"}`
	req3 := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader([]byte(chgBody)))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+accessToken)
	rec3 := httptest.NewRecorder()
	ctrl.ChangePassword(rec3, req3)
	assert.Equal(t, http.StatusConflict, rec3.Code)
	assert.Contains(t, rec3.Body.String(), "用户名已被占用")
}
