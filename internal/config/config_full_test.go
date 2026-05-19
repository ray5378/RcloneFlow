package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadWithConfigFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_config_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
rclone:
  rc_url: "http://custom:5572"
  timeout: "60s"
server:
  addr: ":8080"
  static_dir: "/custom/web"
storage:
  data_dir: "/custom/data"
log:
  level: "debug"
  output: "stderr"
  retention: 14
sync:
  pool_interval: 60
  schedule_interval: 5
  cleanup_interval: 12
  cleanup_retention: 7
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Rclone.RCURL != "http://custom:5572" {
		t.Errorf("expected RCURL http://custom:5572, got %s", cfg.Rclone.RCURL)
	}
	if cfg.Rclone.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %s", cfg.Rclone.Timeout)
	}
	if cfg.Server.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %s", cfg.Server.Addr)
	}
	if cfg.Server.StaticDir != "/custom/web" {
		t.Errorf("expected static_dir /custom/web, got %s", cfg.Server.StaticDir)
	}
	if cfg.Storage.DataDir != "/custom/data" {
		t.Errorf("expected data_dir /custom/data, got %s", cfg.Storage.DataDir)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.Log.Level)
	}
	if cfg.Log.Output != "stderr" {
		t.Errorf("expected log output stderr, got %s", cfg.Log.Output)
	}
	if cfg.Log.Retention != 14 {
		t.Errorf("expected log retention 14, got %d", cfg.Log.Retention)
	}
	if cfg.Sync.PoolInterval != 60 {
		t.Errorf("expected pool_interval 60, got %d", cfg.Sync.PoolInterval)
	}
	if cfg.Sync.ScheduleInterval != 5 {
		t.Errorf("expected schedule_interval 5, got %d", cfg.Sync.ScheduleInterval)
	}
	if cfg.Sync.CleanupInterval != 12 {
		t.Errorf("expected cleanup_interval 12, got %d", cfg.Sync.CleanupInterval)
	}
	if cfg.Sync.CleanupRetention != 7 {
		t.Errorf("expected cleanup_retention 7, got %d", cfg.Sync.CleanupRetention)
	}
}

func TestLoadWithInvalidFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent config file")
	}
}

func TestLoadWithBadYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_config_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("{{invalid yaml:::"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err = Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	os.Setenv("RCLONE_RC_URL", "http://env:5572")
	os.Setenv("RCLONE_RC_USER", "envuser")
	os.Setenv("RCLONE_RC_PASS", "envpass")
	os.Setenv("RCLONE_RC_TIMEOUT", "30s")
	os.Setenv("APP_ADDR", ":9999")
	os.Setenv("APP_STATIC_DIR", "/env/web")
	os.Setenv("APP_DATA_DIR", "/env/data")
	os.Setenv("LOG_LEVEL", "warn")
	os.Setenv("LOG_OUTPUT", "file.log")
	os.Setenv("LOG_RETENTION", "30")
	os.Setenv("CLEANUP_INTERVAL_HOURS", "48")
	os.Setenv("CLEANUP_RETENTION_DAYS", "14")
	defer func() {
		os.Unsetenv("RCLONE_RC_URL")
		os.Unsetenv("RCLONE_RC_USER")
		os.Unsetenv("RCLONE_RC_PASS")
		os.Unsetenv("RCLONE_RC_TIMEOUT")
		os.Unsetenv("APP_ADDR")
		os.Unsetenv("APP_STATIC_DIR")
		os.Unsetenv("APP_DATA_DIR")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_OUTPUT")
		os.Unsetenv("LOG_RETENTION")
		os.Unsetenv("CLEANUP_INTERVAL_HOURS")
		os.Unsetenv("CLEANUP_RETENTION_DAYS")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Rclone.RCURL != "http://env:5572" {
		t.Errorf("expected RCURL from env, got %s", cfg.Rclone.RCURL)
	}
	if cfg.Rclone.RCUser != "envuser" {
		t.Errorf("expected RCUser from env, got %s", cfg.Rclone.RCUser)
	}
	if cfg.Rclone.RCPass != "envpass" {
		t.Errorf("expected RCPass from env, got %s", cfg.Rclone.RCPass)
	}
	if cfg.Rclone.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s from env, got %s", cfg.Rclone.Timeout)
	}
	if cfg.Server.Addr != ":9999" {
		t.Errorf("expected addr :9999 from env, got %s", cfg.Server.Addr)
	}
	if cfg.Server.StaticDir != "/env/web" {
		t.Errorf("expected static_dir from env, got %s", cfg.Server.StaticDir)
	}
	if cfg.Storage.DataDir != "/env/data" {
		t.Errorf("expected data_dir from env, got %s", cfg.Storage.DataDir)
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("expected log level warn from env, got %s", cfg.Log.Level)
	}
	if cfg.Log.Output != "file.log" {
		t.Errorf("expected log output file.log from env, got %s", cfg.Log.Output)
	}
	if cfg.Log.Retention != 30 {
		t.Errorf("expected log retention 30 from env, got %d", cfg.Log.Retention)
	}
	if cfg.Sync.CleanupInterval != 48 {
		t.Errorf("expected cleanup_interval 48 from env, got %d", cfg.Sync.CleanupInterval)
	}
	if cfg.Sync.CleanupRetention != 14 {
		t.Errorf("expected cleanup_retention 14 from env, got %d", cfg.Sync.CleanupRetention)
	}
}

func TestLoadEnvInvalidDuration(t *testing.T) {
	os.Setenv("RCLONE_RC_TIMEOUT", "not-a-duration")
	defer os.Unsetenv("RCLONE_RC_TIMEOUT")

	cfg := DefaultConfig()
	loadFromEnv(cfg)

	if cfg.Rclone.Timeout != 120*time.Second {
		t.Errorf("expected default timeout when env is invalid, got %s", cfg.Rclone.Timeout)
	}
}

func TestLoadEnvInvalidInt(t *testing.T) {
	os.Setenv("LOG_RETENTION", "not-a-number")
	os.Setenv("LOG_RETENTION_DAYS", "also-not-a-number")
	os.Setenv("CLEANUP_INTERVAL_HOURS", "bad")
	os.Setenv("CLEANUP_RETENTION_DAYS", "bad")
	defer func() {
		os.Unsetenv("LOG_RETENTION")
		os.Unsetenv("LOG_RETENTION_DAYS")
		os.Unsetenv("CLEANUP_INTERVAL_HOURS")
		os.Unsetenv("CLEANUP_RETENTION_DAYS")
	}()

	cfg := DefaultConfig()
	loadFromEnv(cfg)

	if cfg.Log.Retention != 7 {
		t.Errorf("expected default retention when env is invalid, got %d", cfg.Log.Retention)
	}
}

func TestLoadEnvFinalSummaryRetentionPriority(t *testing.T) {
	os.Setenv("FINAL_SUMMARY_RETENTION_DAYS", "30")
	os.Setenv("CLEANUP_RETENTION_DAYS", "10")
	defer func() {
		os.Unsetenv("FINAL_SUMMARY_RETENTION_DAYS")
		os.Unsetenv("CLEANUP_RETENTION_DAYS")
	}()

	cfg := DefaultConfig()
	loadFromEnv(cfg)

	if cfg.Sync.CleanupRetention != 30 {
		t.Errorf("expected FINAL_SUMMARY_RETENTION_DAYS to take priority, got %d", cfg.Sync.CleanupRetention)
	}
}

func TestLoadEnvZeroRetention(t *testing.T) {
	os.Setenv("CLEANUP_RETENTION_DAYS", "0")
	defer os.Unsetenv("CLEANUP_RETENTION_DAYS")

	cfg := DefaultConfig()
	loadFromEnv(cfg)

	if cfg.Sync.CleanupRetention != 0 {
		t.Errorf("expected 0 retention, got %d", cfg.Sync.CleanupRetention)
	}
}

func TestFindConfigFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_config_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir)

	if found := findConfigFile(); found != "" {
		t.Errorf("expected empty when no config exists, got %s", found)
	}

	if err := os.WriteFile("config.yaml", []byte("rclone:\n  rc_url: test"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	if found := findConfigFile(); found != "./config.yaml" {
		t.Errorf("expected ./config.yaml, got %s", found)
	}
}

func TestFindConfigFileRcloneflowYaml(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rcloneflow_config_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.WriteFile("rcloneflow.yaml", []byte("rclone:\n  rc_url: test"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	if found := findConfigFile(); found != "./rcloneflow.yaml" {
		t.Errorf("expected ./rcloneflow.yaml, got %s", found)
	}
}

func TestGetters(t *testing.T) {
	cfg := &Config{
		Rclone: RcloneConfig{
			RCURL:   "http://test:5572",
			RCUser:  "user",
			RCPass:  "pass",
			Timeout: 90 * time.Second,
		},
		Server: ServerConfig{
			Addr:      ":8080",
			StaticDir: "/web",
		},
		Storage: StorageConfig{
			DataDir: "/data",
		},
		Log: LogConfig{
			Level:     "debug",
			Output:    "stderr",
			Retention: 14,
		},
		Sync: SyncConfig{
			PoolInterval:     60,
			ScheduleInterval: 5,
			CleanupInterval:  12,
			CleanupRetention: 7,
		},
	}

	if cfg.GetRcloneAddr() != "http://test:5572" {
		t.Errorf("GetRcloneAddr() = %s", cfg.GetRcloneAddr())
	}
	if cfg.GetRcloneUser() != "user" {
		t.Errorf("GetRcloneUser() = %s", cfg.GetRcloneUser())
	}
	if cfg.GetRclonePass() != "pass" {
		t.Errorf("GetRclonePass() = %s", cfg.GetRclonePass())
	}
	if cfg.GetRcloneTimeout() != 90*time.Second {
		t.Errorf("GetRcloneTimeout() = %s", cfg.GetRcloneTimeout())
	}
	if cfg.GetServerAddr() != ":8080" {
		t.Errorf("GetServerAddr() = %s", cfg.GetServerAddr())
	}
	if cfg.GetStaticDir() != "/web" {
		t.Errorf("GetStaticDir() = %s", cfg.GetStaticDir())
	}
	if cfg.GetDataDir() != "/data" {
		t.Errorf("GetDataDir() = %s", cfg.GetDataDir())
	}
	if cfg.GetLogLevel() != "debug" {
		t.Errorf("GetLogLevel() = %s", cfg.GetLogLevel())
	}
	if cfg.GetLogOutput() != "stderr" {
		t.Errorf("GetLogOutput() = %s", cfg.GetLogOutput())
	}
	if cfg.GetPoolInterval() != 60 {
		t.Errorf("GetPoolInterval() = %d", cfg.GetPoolInterval())
	}
	if cfg.GetScheduleInterval() != 5 {
		t.Errorf("GetScheduleInterval() = %d", cfg.GetScheduleInterval())
	}
	if cfg.GetCleanupInterval() != 12 {
		t.Errorf("GetCleanupInterval() = %d", cfg.GetCleanupInterval())
	}
	if cfg.GetCleanupRetention() != 7 {
		t.Errorf("GetCleanupRetention() = %d", cfg.GetCleanupRetention())
	}
	if cfg.GetLogRetention() != 14 {
		t.Errorf("GetLogRetention() = %d", cfg.GetLogRetention())
	}
}

func TestToEnvMapEmptyCredentials(t *testing.T) {
	cfg := &Config{
		Rclone: RcloneConfig{
			RCURL: "http://test:5572",
		},
		Server: ServerConfig{
			Addr: ":8080",
		},
		Storage: StorageConfig{
			DataDir: "/data",
		},
		Log: LogConfig{
			Level:  "info",
			Output: "stdout",
		},
	}

	envMap := cfg.ToEnvMap()
	if _, ok := envMap["RCLONE_RC_USER"]; ok {
		t.Error("expected no RCLONE_RC_USER when empty")
	}
	if _, ok := envMap["RCLONE_RC_PASS"]; ok {
		t.Error("expected no RCLONE_RC_PASS when empty")
	}
}

func TestConfigString(t *testing.T) {
	cfg := &Config{
		Rclone: RcloneConfig{
			RCURL:   "http://test:5572",
			RCUser:  "user",
			Timeout: 60 * time.Second,
		},
		Server: ServerConfig{
			Addr:      ":8080",
			StaticDir: "/web",
		},
		Storage: StorageConfig{
			DataDir: "/data",
		},
		Log: LogConfig{
			Level:  "debug",
			Output: "stderr",
		},
	}

	s := cfg.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
	if !contains(s, "http://test:5572") {
		t.Error("expected rc_url in string output")
	}
	if !contains(s, ":8080") {
		t.Error("expected addr in string output")
	}
	if !contains(s, "/data") {
		t.Error("expected data_dir in string output")
	}
	if !contains(s, "debug") {
		t.Error("expected log level in string output")
	}
	if !contains(s, "user") {
		t.Error("expected rc_user in string output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
