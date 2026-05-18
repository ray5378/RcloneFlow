package adapter

import (
	"testing"
	"time"
)

func TestRcloneConfigDefault(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.BaseURL != "http://127.0.0.1:5572" {
		t.Errorf("expected default BaseURL http://127.0.0.1:5572, got %s", cfg.BaseURL)
	}

	if cfg.Timeout != 120*time.Second {
		t.Errorf("expected default Timeout 120s, got %v", cfg.Timeout)
	}
}

func TestRcloneConfigEnvOverride(t *testing.T) {
	t.Setenv("RCLONE_RC_URL", "http://custom:8080")
	t.Setenv("RCLONE_RC_USER", "testuser")
	t.Setenv("RCLONE_RC_PASS", "testpass")
	t.Setenv("RCLONE_RC_TIMEOUT", "60s")

	cfg := DefaultConfig()

	if cfg.BaseURL != "http://custom:8080" {
		t.Errorf("expected BaseURL http://custom:8080, got %s", cfg.BaseURL)
	}

	if cfg.User != "testuser" {
		t.Errorf("expected User testuser, got %s", cfg.User)
	}

	if cfg.Pass != "testpass" {
		t.Errorf("expected Pass testpass, got %s", cfg.Pass)
	}

	if cfg.Timeout != 60*time.Second {
		t.Errorf("expected Timeout 60s, got %v", cfg.Timeout)
	}
}

func TestNewRcloneClient(t *testing.T) {
	cfg := &RcloneConfig{
		BaseURL: "http://127.0.0.1:5572",
		Timeout: 30 * time.Second,
	}

	client := NewRcloneClient(cfg)

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.config.BaseURL != "http://127.0.0.1:5572" {
		t.Errorf("expected BaseURL http://127.0.0.1:5572, got %s", client.config.BaseURL)
	}
}

func TestNewRcloneClientNilConfig(t *testing.T) {
	// 应该使用默认配置而不是panic
	client := NewRcloneClient(nil)

	if client == nil {
		t.Fatal("expected non-nil client with nil config")
	}
}

func TestVersionResponse(t *testing.T) {
	resp := &VersionResponse{
		Version:    "v1.60.0",
		Decomposed: []int{1, 60, 0},
		IsGit:      true,
		IsBeta:     false,
		Os:         "linux",
		OsKernel:   "5.15.0",
		GoVersion:  "go1.19",
	}

	if resp.Version != "v1.60.0" {
		t.Errorf("expected Version v1.60.0, got %s", resp.Version)
	}

	if len(resp.Decomposed) != 3 {
		t.Errorf("expected Decomposed length 3, got %d", len(resp.Decomposed))
	}
}

func TestPathInfo(t *testing.T) {
	info := &PathInfo{
		Name:    "test.txt",
		Path:    "/path/to/test.txt",
		IsDir:   false,
		Size:    1024,
		ModTime: "2024-01-01T12:00:00Z",
	}

	if info.Name != "test.txt" {
		t.Errorf("expected Name test.txt, got %s", info.Name)
	}

	if info.IsDir {
		t.Error("expected IsDir false")
	}

	if info.Size != 1024 {
		t.Errorf("expected Size 1024, got %d", info.Size)
	}
}

func TestCreateRemoteRequest(t *testing.T) {
	req := &CreateRemoteRequest{
		Name: "myremote",
		Type: "s3",
		Parameters: map[string]any{
			"provider": "AWS",
			"env_auth": true,
		},
	}

	if req.Name != "myremote" {
		t.Errorf("expected Name myremote, got %s", req.Name)
	}

	if req.Type != "s3" {
		t.Errorf("expected Type s3, got %s", req.Type)
	}

	if req.Parameters["provider"] != "AWS" {
		t.Errorf("expected Parameters[provider] AWS, got %v", req.Parameters["provider"])
	}
}
