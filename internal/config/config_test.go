package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.Rclone.RCURL != "http://127.0.0.1:5572" {
		t.Errorf("expected default RCURL http://127.0.0.1:5572, got %s", cfg.Rclone.RCURL)
	}
	
	if cfg.Server.Addr != ":17870" {
		t.Errorf("expected default addr :17870, got %s", cfg.Server.Addr)
	}
	
	if cfg.Storage.DataDir != "./data" {
		t.Errorf("expected default data_dir ./data, got %s", cfg.Storage.DataDir)
	}
	
	if cfg.Log.Level != "info" {
		t.Errorf("expected default log level info, got %s", cfg.Log.Level)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("RCLONE_RC_URL", "http://localhost:8080")
	os.Setenv("APP_ADDR", ":9000")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("RCLONE_RC_URL")
		os.Unsetenv("APP_ADDR")
		os.Unsetenv("LOG_LEVEL")
	}()
	
	cfg := DefaultConfig()
	loadFromEnv(cfg)
	
	if cfg.Rclone.RCURL != "http://localhost:8080" {
		t.Errorf("expected RCURL http://localhost:8080 from env, got %s", cfg.Rclone.RCURL)
	}
	
	if cfg.Server.Addr != ":9000" {
		t.Errorf("expected addr :9000 from env, got %s", cfg.Server.Addr)
	}
	
	if cfg.Log.Level != "debug" {
		t.Errorf("expected log level debug from env, got %s", cfg.Log.Level)
	}
}

func TestGetRcloneAddr(t *testing.T) {
	cfg := &Config{
		Rclone: RcloneConfig{
			RCURL: "http://test:9090",
		},
	}
	
	if cfg.GetRcloneAddr() != "http://test:9090" {
		t.Errorf("expected http://test:9090, got %s", cfg.GetRcloneAddr())
	}
}

func TestGetDataDir(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			DataDir: "/tmp/data",
		},
	}
	
	if cfg.GetDataDir() != "/tmp/data" {
		t.Errorf("expected /tmp/data, got %s", cfg.GetDataDir())
	}
}

func TestToEnvMap(t *testing.T) {
	cfg := &Config{
		Rclone: RcloneConfig{
			RCURL:  "http://test:9090",
			RCUser: "user",
			RCPass: "pass",
		},
		Server: ServerConfig{
			Addr: ":9000",
		},
		Storage: StorageConfig{
			DataDir: "/data",
		},
		Log: LogConfig{
			Level:  "debug",
			Output: "stdout",
		},
	}
	
	envMap := cfg.ToEnvMap()
	
	if envMap["RCLONE_RC_URL"] != "http://test:9090" {
		t.Errorf("expected RCLONE_RC_URL http://test:9090, got %s", envMap["RCLONE_RC_URL"])
	}
	
	if envMap["APP_ADDR"] != ":9000" {
		t.Errorf("expected APP_ADDR :9000, got %s", envMap["APP_ADDR"])
	}
	
	if envMap["LOG_LEVEL"] != "debug" {
		t.Errorf("expected LOG_LEVEL debug, got %s", envMap["LOG_LEVEL"])
	}
}
