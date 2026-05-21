package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureConfigPath_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)

	p := ensureConfigPath()
	if p != filepath.Join(tmpDir, "rclone.conf") {
		t.Errorf("ensureConfigPath() = %q, want %q", p, filepath.Join(tmpDir, "rclone.conf"))
	}

	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Fatal("rclone.conf was not created")
	}

	b, _ := os.ReadFile(p)
	if strings.TrimSpace(string(b)) == "{}" {
		t.Fatal("rclone.conf should not contain JSON '{}'")
	}
}

func TestEnsureConfigPath_FixesJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)

	// Simulate a JSON config (which is wrong for rclone INI format)
	confPath := filepath.Join(tmpDir, "rclone.conf")
	_ = os.WriteFile(confPath, []byte("{}"), 0o644)

	p := ensureConfigPath()
	b, _ := os.ReadFile(p)
	if strings.TrimSpace(string(b)) == "{}" {
		t.Fatal("ensureConfigPath() should have fixed JSON '{}' to empty")
	}
}

func TestEnsureConfigPath_PreservesValidINI(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)

	confPath := filepath.Join(tmpDir, "rclone.conf")
	_ = os.WriteFile(confPath, []byte("[remote]\ntype = drive\n"), 0o644)

	p := ensureConfigPath()
	b, _ := os.ReadFile(p)
	content := strings.TrimSpace(string(b))
	if content == "{}" {
		t.Fatal("ensureConfigPath() should not overwrite valid INI config")
	}
	if !strings.Contains(content, "[remote]") {
		t.Errorf("ensureConfigPath() corrupted valid config: %q", content)
	}
}

func TestEnsureConfigPath_DefaultDataDir(t *testing.T) {
	// Test with no APP_DATA_DIR set (uses ./data)
	oldDir := os.Getenv("APP_DATA_DIR")
	os.Unsetenv("APP_DATA_DIR")
	defer os.Setenv("APP_DATA_DIR", oldDir)

	// This will create ./data/rclone.conf - just verify it doesn't panic
	p := ensureConfigPath()
	if p == "" {
		t.Fatal("ensureConfigPath() returned empty path")
	}
	// Clean up
	_ = os.RemoveAll("./data")
}
