package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRunMain_DBError(t *testing.T) {
	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "badfile")
	os.WriteFile(badPath, []byte("not a directory"), 0644)

	t.Setenv("EMBED_RC", "false")
	t.Setenv("APP_DATA_DIR", badPath)
	t.Setenv("LOG_LEVEL", "fatal")
	t.Setenv("APP_ADDR", ":17876")

	err := RunMain()
	if err == nil {
		t.Fatal("expected error for bad data dir path")
	}
	if !strings.Contains(err.Error(), "启动失败") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunMain_SuccessAndShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	staticDir := filepath.Join(tmpDir, "static")
	os.MkdirAll(staticDir, 0755)
	os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("RcloneFlow UI"), 0644)
	dataDir := filepath.Join(tmpDir, "data")

	t.Setenv("EMBED_RC", "false")
	t.Setenv("APP_DATA_DIR", dataDir)
	t.Setenv("APP_STATIC_DIR", staticDir)
	t.Setenv("APP_ADDR", ":17875")
	t.Setenv("LOG_LEVEL", "fatal")

	errCh := make(chan error, 1)
	go func() {
		errCh <- RunMain()
	}()

	baseURL := "http://127.0.0.1:17875"
	ready := false
	for i := 0; i < 30; i++ {
		resp, err := http.Get(baseURL + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				ready = true
				break
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server not ready")
	}

	resp, err := http.Get(baseURL + "/api/auth/has-users")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("RunMain returned error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown timeout")
	}
}
