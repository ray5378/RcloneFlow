package app

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"rcloneflow/internal/config"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
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

func TestTaskServiceSchedulerRunner_RunTask(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APP_DATA_DIR", tmpDir)

	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	taskSvc := service.NewTaskService(db, nil)
	runner := taskServiceSchedulerRunner{svc: taskSvc}

	task := store.Task{
		Name:         "runner-test-task",
		Mode:         "copy",
		SourceRemote: "src",
		SourcePath:   "/path",
		TargetRemote: "dst",
		TargetPath:   "/path",
	}
	_, err = db.AddTask(task)
	if err != nil {
		t.Fatal(err)
	}

	err = runner.RunTask(context.Background(), 999, "manual")
	if err == nil {
		t.Fatal("expected error for non-existing task")
	}
}

func TestStartEmbeddedRC_NoRclone(t *testing.T) {
	err := startEmbeddedRC("127.0.0.1:15572", "", "", "/nonexistent/config", 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected error when rclone binary is not found")
	}
}

func TestMaybeStartEmbeddedRC_Disabled(t *testing.T) {
	t.Setenv("EMBED_RC", "false")

	maybeStartEmbeddedRC()
	// Should return without error (just logging)
}

func TestMaybeStartEmbeddedRC_DisabledZero(t *testing.T) {
	t.Setenv("EMBED_RC", "0")

	maybeStartEmbeddedRC()
	// Should return without error (just logging)
}

func TestMaybeStartEmbeddedRC_EmptyDir(t *testing.T) {
	t.Setenv("EMBED_RC", "true")
	t.Setenv("APP_DATA_DIR", t.TempDir())

	maybeStartEmbeddedRC()
	// Should attempt to start embedded RC and fail gracefully
}

func TestRunWithShutdown_StartAndStop(t *testing.T) {
	tmpDir := t.TempDir()
	staticDir := filepath.Join(tmpDir, "static")
	os.MkdirAll(staticDir, 0755)
	os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("RcloneFlow UI"), 0644)

	t.Setenv("EMBED_RC", "false")
	t.Setenv("APP_DATA_DIR", tmpDir)
	t.Setenv("APP_STATIC_DIR", staticDir)

	cfg := config.DefaultConfig()
	cfg.Storage.DataDir = tmpDir
	cfg.Server.StaticDir = staticDir
	cfg.Server.Addr = ":17872"
	cfg.Log.Level = "fatal"
	cfg.Sync.CleanupInterval = 0
	cfg.Sync.CleanupRetention = 0

	stop := make(chan os.Signal, 1)
	errCh := make(chan error, 1)

	go func() {
		errCh <- RunWithShutdown(cfg, stop)
	}()

	baseURL := "http://127.0.0.1:17872"
	var lastErr error
	ready := false
	for i := 0; i < 20; i++ {
		resp, err := http.Get(baseURL + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		lastErr = err
		time.Sleep(200 * time.Millisecond)
	}
	if !ready {
		t.Fatalf("server not ready: %v", lastErr)
	}

	resp, err := http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	resp, err = http.Get(baseURL + "/api/auth/has-users")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["exists"] != false {
		t.Fatalf("expected exists=false, got %v", body["exists"])
	}

	stop <- syscall.SIGINT

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("RunWithShutdown returned error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown timeout")
	}
}

func TestRunWithShutdown_ShutdownByClose(t *testing.T) {
	tmpDir := t.TempDir()
	staticDir := filepath.Join(tmpDir, "static")
	os.MkdirAll(staticDir, 0755)
	os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("RcloneFlow UI"), 0644)

	t.Setenv("EMBED_RC", "false")
	t.Setenv("APP_DATA_DIR", tmpDir)

	cfg := config.DefaultConfig()
	cfg.Storage.DataDir = tmpDir
	cfg.Server.StaticDir = staticDir
	cfg.Server.Addr = ":17874"
	cfg.Log.Level = "fatal"
	cfg.Sync.CleanupInterval = 0
	cfg.Sync.CleanupRetention = 0

	stop := make(chan os.Signal, 1)
	errCh := make(chan error, 1)

	go func() {
		errCh <- RunWithShutdown(cfg, stop)
	}()

	baseURL := "http://127.0.0.1:17874"
	ready := false
	for i := 0; i < 30; i++ {
		resp, err := http.Get(baseURL + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server not ready")
	}

	stop <- syscall.SIGINT

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("RunWithShutdown returned error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown timeout")
	}
}
