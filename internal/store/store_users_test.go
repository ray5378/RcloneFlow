package store

import (
	"os"
	"testing"
)

func openUsersDB(t *testing.T) *DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rcloneflow_users_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })
	db, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return db
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	_, err := db.CreateUser("admin", "password123")
	if err != nil {
		t.Fatalf("first CreateUser() error = %v", err)
	}

	_, err = db.CreateUser("admin", "different_password")
	if err == nil {
		t.Fatal("expected error on duplicate username")
	}
}

func TestGetUserByUsername_NotExist(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	_, ok := db.GetUserByUsername("nobody")
	if ok {
		t.Error("expected false for non-existent user")
	}
}

func TestGetUserByID_NotExist(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	_, ok := db.GetUserByID(99999)
	if ok {
		t.Error("expected false for non-existent user ID")
	}
}

func TestUpdatePassword_UserNotExist(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	err := db.UpdatePassword(99999, "newpassword")
	if err != nil {
		t.Fatalf("UpdatePassword on non-existent user should not fail: %v", err)
	}
}

func TestUpdateUsername_UserNotExist(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	err := db.UpdateUsername(99999, "newname")
	if err != nil {
		t.Fatalf("UpdateUsername on non-existent user should not fail: %v", err)
	}
}

func TestUpdateUsername_DuplicateUsername(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	_, _ = db.CreateUser("user1", "pw1")
	u2, _ := db.CreateUser("user2", "pw2")

	err := db.UpdateUsername(u2.ID, "user1")
	if err == nil {
		t.Fatal("expected error when updating username to existing name")
	}

	got, _ := db.GetUserByID(u2.ID)
	if got.Username != "user2" {
		t.Errorf("expected username to remain 'user2', got %s", got.Username)
	}
}

func TestListUsers_Empty(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	users, err := db.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users in empty DB, got %d", len(users))
	}
}

func TestUser_PasswordChangedDefault(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	u, err := db.CreateUser("testuser", "hashedpw")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if u.PasswordChanged {
		t.Error("new user should have PasswordChanged=false")
	}

	got, _ := db.GetUserByUsername("testuser")
	if got.PasswordChanged {
		t.Error("PasswordChanged should be false in DB")
	}
}

func TestUser_UpdatePasswordSetsChanged(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	u, _ := db.CreateUser("changepw", "oldpw")

	if err := db.UpdatePassword(u.ID, "newhashedpw"); err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}

	got, _ := db.GetUserByID(u.ID)
	if !got.PasswordChanged {
		t.Error("PasswordChanged should be true after password update")
	}
	if got.Password != "newhashedpw" {
		t.Errorf("expected password 'newhashedpw', got %s", got.Password)
	}
}

func TestListUsers_Multiple(t *testing.T) {
	db := openUsersDB(t)
	defer db.Close()

	db.CreateUser("alice", "pw_a")
	db.CreateUser("bob", "pw_b")
	db.CreateUser("charlie", "pw_c")

	users, err := db.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}

	names := make(map[string]bool)
	for _, u := range users {
		names[u.Username] = true
	}
	for _, expected := range []string{"alice", "bob", "charlie"} {
		if !names[expected] {
			t.Errorf("expected user %s in list", expected)
		}
	}
}