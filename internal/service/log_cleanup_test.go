package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLogCleanupService(t *testing.T) {
	svc := NewLogCleanupService("/tmp/logs", 24*time.Hour, 7)
	assert.NotNil(t, svc)
	assert.Equal(t, "/tmp/logs", svc.logsDir)
	assert.Equal(t, 24*time.Hour, svc.interval)
	assert.Equal(t, 7, svc.retentionDays)
}

func TestNewLogCleanupService_Defaults(t *testing.T) {
	svc := NewLogCleanupService("/tmp/logs", 0, 0)
	assert.NotNil(t, svc)
	assert.Equal(t, 24*time.Hour, svc.interval)
	assert.Equal(t, 7, svc.retentionDays)
}

func TestNewLogCleanupService_NegativeValues(t *testing.T) {
	svc := NewLogCleanupService("/tmp/logs", -1*time.Hour, -5)
	assert.NotNil(t, svc)
	assert.Equal(t, 24*time.Hour, svc.interval)
	assert.Equal(t, 7, svc.retentionDays)
}

func TestLogCleanupService_Stop(t *testing.T) {
	svc := NewLogCleanupService("/tmp/logs", 24*time.Hour, 7)
	svc.Stop()
	assert.True(t, isClosed(svc.stopCh))
}

func TestLogCleanupService_Replan(t *testing.T) {
	svc := NewLogCleanupService("/tmp/logs", 24*time.Hour, 7)
	svc.Replan(14)

	svc.mu.RLock()
	assert.Equal(t, 14, svc.retentionDays)
	svc.mu.RUnlock()
}

func TestLogCleanupService_Replan_ZeroDays(t *testing.T) {
	svc := NewLogCleanupService("/tmp/logs", 24*time.Hour, 7)
	svc.Replan(0)

	svc.mu.RLock()
	assert.Equal(t, 7, svc.retentionDays)
	svc.mu.RUnlock()
}

func TestEnvLogRetentionDays(t *testing.T) {
	t.Setenv("LOG_RETENTION_DAYS", "30")
	assert.Equal(t, 30, EnvLogRetentionDays(7))
}

func TestEnvLogRetentionDays_Default(t *testing.T) {
	t.Setenv("LOG_RETENTION_DAYS", "")
	assert.Equal(t, 7, EnvLogRetentionDays(7))
}

func TestEnvLogRetentionDays_Invalid(t *testing.T) {
	t.Setenv("LOG_RETENTION_DAYS", "invalid")
	assert.Equal(t, 7, EnvLogRetentionDays(7))
}

func TestEnvLogCleanupInterval(t *testing.T) {
	t.Setenv("LOG_CLEANUP_INTERVAL_HOURS", "12")
	assert.Equal(t, 12*time.Hour, EnvLogCleanupInterval(24))
}

func TestEnvLogCleanupInterval_Default(t *testing.T) {
	t.Setenv("LOG_CLEANUP_INTERVAL_HOURS", "")
	assert.Equal(t, 24*time.Hour, EnvLogCleanupInterval(24))
}

func TestEnvLogCleanupInterval_Invalid(t *testing.T) {
	t.Setenv("LOG_CLEANUP_INTERVAL_HOURS", "invalid")
	assert.Equal(t, 24*time.Hour, EnvLogCleanupInterval(24))
}

func TestLogCleanupService_Cleanup_NoDir(t *testing.T) {
	svc := NewLogCleanupService("/nonexistent/path/for/test", 24*time.Hour, 7)
	svc.cleanup()
}

func TestLogCleanupService_Cleanup_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	svc := NewLogCleanupService(dir, 24*time.Hour, 7)
	svc.cleanup()
}

func TestLogCleanupService_Cleanup_OldFiles(t *testing.T) {
	dir := t.TempDir()

	taskDir := filepath.Join(dir, "test-task-0101")
	os.MkdirAll(taskDir, 0o755)

	oldFile := filepath.Join(taskDir, "1200.log")
	os.WriteFile(oldFile, []byte("old log"), 0o644)
	oldTime := time.Now().Add(-30 * 24 * time.Hour)
	os.Chtimes(oldFile, oldTime, oldTime)

	svc := NewLogCleanupService(dir, 24*time.Hour, 7)
	svc.cleanup()

	_, err := os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err))
}

func TestLogCleanupService_Cleanup_KeepRecentFiles(t *testing.T) {
	dir := t.TempDir()

	taskDir := filepath.Join(dir, "test-task-0101")
	os.MkdirAll(taskDir, 0o755)

	recentFile := filepath.Join(taskDir, "1200.log")
	os.WriteFile(recentFile, []byte("recent log"), 0o644)

	svc := NewLogCleanupService(dir, 24*time.Hour, 1)
	svc.cleanup()

	_, err := os.Stat(recentFile)
	assert.NoError(t, err)
}

func TestLogCleanupService_getRetentionDays(t *testing.T) {
	svc := NewLogCleanupService("/tmp/logs", 24*time.Hour, 14)
	assert.Equal(t, 14, svc.getRetentionDays())
}

func TestLogCleanupService_getInterval(t *testing.T) {
	svc := NewLogCleanupService("/tmp/logs", 12*time.Hour, 7)
	assert.Equal(t, 12*time.Hour, svc.getInterval())
}

func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
