package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"rcloneflow/internal/store"
)

func setupAuthService(t *testing.T) (*AuthService, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	svc := NewAuthService(db)
	return svc, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestAuthService_HasUsers_Empty(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	exists, err := svc.HasUsers()
	if err != nil {
		t.Fatalf("HasUsers() error = %v", err)
	}
	if exists {
		t.Error("HasUsers() should return false for empty db")
	}
}

func TestAuthService_HasUsers_WithUsers(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("testuser", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	exists, err := svc.HasUsers()
	if err != nil {
		t.Fatalf("HasUsers() error = %v", err)
	}
	if !exists {
		t.Error("HasUsers() should return true when users exist")
	}
}

func TestAuthService_Register_Success(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	user, tokens, err := svc.Register("newuser", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user == nil {
		t.Fatal("Register() returned nil user")
	}
	if user.Username != "newuser" {
		t.Errorf("Username = %q, want %q", user.Username, "newuser")
	}
	if tokens == nil {
		t.Fatal("Register() returned nil tokens")
	}
	if tokens.AccessToken == "" {
		t.Error("AccessToken is empty")
	}
	if tokens.RefreshToken == "" {
		t.Error("RefreshToken is empty")
	}
}

func TestAuthService_Register_Duplicate(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("dupuser", "password123")
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	_, _, err = svc.Register("dupuser", "password456")
	if err == nil {
		t.Fatal("second Register() should fail for duplicate username")
	}
	if !errors.Is(err, ErrAuthUserExists) {
		t.Errorf("Error = %v, want ErrAuthUserExists", err)
	}
}

func TestAuthService_Register_ShortPassword(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("shortpw", "123")
	if err == nil {
		t.Fatal("Register() should fail for short password")
	}
	if !errors.Is(err, ErrAuthPasswordTooShort) {
		t.Errorf("Error = %v, want ErrAuthPasswordTooShort", err)
	}
}

func TestAuthService_Register_EmptyFields(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("", "password123")
	if err == nil {
		t.Fatal("Register() should fail for empty username")
	}

	_, _, err = svc.Register("user", "")
	if err == nil {
		t.Fatal("Register() should fail for empty password")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("loginuser", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, tokens, err := svc.Login("loginuser", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user == nil {
		t.Fatal("Login() returned nil user")
	}
	if user.Username != "loginuser" {
		t.Errorf("Username = %q, want %q", user.Username, "loginuser")
	}
	if tokens == nil {
		t.Fatal("Login() returned nil tokens")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("wrongpw", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, _, err = svc.Login("wrongpw", "wrongpassword")
	if err == nil {
		t.Fatal("Login() should fail for wrong password")
	}
	if !errors.Is(err, ErrAuthInvalidCredential) {
		t.Errorf("Error = %v, want ErrAuthInvalidCredential", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Login("nonexistent", "password123")
	if err == nil {
		t.Fatal("Login() should fail for nonexistent user")
	}
	if !errors.Is(err, ErrAuthInvalidCredential) {
		t.Errorf("Error = %v, want ErrAuthInvalidCredential", err)
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, tokens, err := svc.Register("refreshuser", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	newTokens, err := svc.RefreshToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if newTokens == nil {
		t.Fatal("RefreshToken() returned nil tokens")
	}
	if newTokens.AccessToken == "" {
		t.Error("New AccessToken is empty")
	}
	if newTokens.RefreshToken == "" {
		t.Error("New RefreshToken is empty")
	}
}

func TestAuthService_RefreshToken_Invalid(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, err := svc.RefreshToken("invalid-token")
	if err == nil {
		t.Fatal("RefreshToken() should fail for invalid token")
	}
	if !errors.Is(err, ErrAuthRefreshInvalid) {
		t.Errorf("Error = %v, want ErrAuthRefreshInvalid", err)
	}
}

func TestAuthService_ChangeProfile_PasswordOnly(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("changepw", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, _ := svc.GetUserByUsername("changepw")
	if user == nil {
		t.Fatal("User not found after register")
	}

	updatedUser, err := svc.ChangeProfile(user.ID, "password123", "newpassword456", "")
	if err != nil {
		t.Fatalf("ChangeProfile() error = %v", err)
	}
	if updatedUser == nil {
		t.Fatal("ChangeProfile() returned nil user")
	}

	// 验证新密码可以登录
	_, _, err = svc.Login("changepw", "newpassword456")
	if err != nil {
		t.Fatalf("Login() with new password failed: %v", err)
	}
}

func TestAuthService_ChangeProfile_UsernameOnly(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("oldname", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, _ := svc.GetUserByUsername("oldname")
	if user == nil {
		t.Fatal("User not found after register")
	}

	updatedUser, err := svc.ChangeProfile(user.ID, "", "", "newname")
	if err != nil {
		t.Fatalf("ChangeProfile() error = %v", err)
	}
	if updatedUser.Username != "newname" {
		t.Errorf("Username = %q, want %q", updatedUser.Username, "newname")
	}

	// 验证新用户名可以登录
	_, _, err = svc.Login("newname", "password123")
	if err != nil {
		t.Fatalf("Login() with new username failed: %v", err)
	}
}

func TestAuthService_ChangeProfile_Both(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("bothold", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, _ := svc.GetUserByUsername("bothold")
	if user == nil {
		t.Fatal("User not found after register")
	}

	updatedUser, err := svc.ChangeProfile(user.ID, "password123", "newpassword456", "bothnew")
	if err != nil {
		t.Fatalf("ChangeProfile() error = %v", err)
	}
	if updatedUser.Username != "bothnew" {
		t.Errorf("Username = %q, want %q", updatedUser.Username, "bothnew")
	}

	// 验证新用户名和新密码可以登录
	_, _, err = svc.Login("bothnew", "newpassword456")
	if err != nil {
		t.Fatalf("Login() with new credentials failed: %v", err)
	}
}

func TestAuthService_ChangeProfile_WrongOldPassword(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("wrongold", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, _ := svc.GetUserByUsername("wrongold")
	if user == nil {
		t.Fatal("User not found after register")
	}

	_, err = svc.ChangeProfile(user.ID, "wrongpassword", "newpassword456", "")
	if err == nil {
		t.Fatal("ChangeProfile() should fail for wrong old password")
	}
	if !errors.Is(err, ErrAuthOldPasswordWrong) {
		t.Errorf("Error = %v, want ErrAuthOldPasswordWrong", err)
	}
}

func TestAuthService_ChangeProfile_MissingOldPassword(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("missingold", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, _ := svc.GetUserByUsername("missingold")
	if user == nil {
		t.Fatal("User not found after register")
	}

	_, err = svc.ChangeProfile(user.ID, "", "newpassword456", "")
	if err == nil {
		t.Fatal("ChangeProfile() should fail when old password is missing")
	}
	if !errors.Is(err, ErrAuthOldPasswordEmpty) {
		t.Errorf("Error = %v, want ErrAuthOldPasswordEmpty", err)
	}
}

func TestAuthService_ChangeProfile_DuplicateUsername(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, _, err := svc.Register("user1", "password123")
	if err != nil {
		t.Fatalf("Register() user1 error = %v", err)
	}
	_, _, err = svc.Register("user2", "password456")
	if err != nil {
		t.Fatalf("Register() user2 error = %v", err)
	}

	user2, _ := svc.GetUserByUsername("user2")
	if user2 == nil {
		t.Fatal("User2 not found")
	}

	_, err = svc.ChangeProfile(user2.ID, "", "", "user1")
	if err == nil {
		t.Fatal("ChangeProfile() should fail for duplicate username")
	}
	if !errors.Is(err, ErrAuthUsernameTaken) {
		t.Errorf("Error = %v, want ErrAuthUsernameTaken", err)
	}
}

func TestAuthService_GetUserByID_NotFound(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, exists := svc.GetUserByID(999)
	if exists {
		t.Error("GetUserByID() should return false for non-existent ID")
	}
}

func TestAuthService_GetUserByUsername_NotFound(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, exists := svc.GetUserByUsername("nonexistent")
	if exists {
		t.Error("GetUserByUsername() should return false for non-existent username")
	}
}

func TestAuthService_RefreshToken_Empty(t *testing.T) {
	svc, cleanup := setupAuthService(t)
	defer cleanup()

	_, err := svc.RefreshToken("")
	if err == nil {
		t.Fatal("RefreshToken() should fail for empty token")
	}
	if !errors.Is(err, ErrAuthRefreshRequired) {
		t.Errorf("Error = %v, want ErrAuthRefreshRequired", err)
	}
}
