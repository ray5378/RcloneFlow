package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetFinishWaitInterval_Default(t *testing.T) {
	t.Setenv("FINISH_WAIT_INTERVAL", "")
	t.Setenv("APP_DATA_DIR", "/nonexistent")
	assert.Equal(t, 5*time.Second, GetFinishWaitInterval())
}

func TestGetFinishWaitInterval_EnvOverride(t *testing.T) {
	t.Setenv("FINISH_WAIT_INTERVAL", "10s")
	t.Setenv("APP_DATA_DIR", "/nonexistent")
	assert.Equal(t, 10*time.Second, GetFinishWaitInterval())
}

func TestGetFinishWaitInterval_SettingsOverride(t *testing.T) {
	t.Setenv("FINISH_WAIT_INTERVAL", "")
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	data := map[string]string{"FINISH_WAIT_INTERVAL": "15s"}
	b, _ := json.Marshal(data)
	os.WriteFile(filepath.Join(dir, "settings.json"), b, 0o644)

	assert.Equal(t, 15*time.Second, GetFinishWaitInterval())
}

func TestGetFinishWaitTimeout_Default(t *testing.T) {
	t.Setenv("FINISH_WAIT_TIMEOUT", "")
	t.Setenv("APP_DATA_DIR", "/nonexistent")
	assert.Equal(t, 5*time.Hour, GetFinishWaitTimeout())
}

func TestGetFinishWaitTimeout_EnvOverride(t *testing.T) {
	t.Setenv("FINISH_WAIT_TIMEOUT", "1h")
	t.Setenv("APP_DATA_DIR", "/nonexistent")
	assert.Equal(t, 1*time.Hour, GetFinishWaitTimeout())
}

func TestGetWebhookMaxFiles_Default(t *testing.T) {
	t.Setenv("APP_DATA_DIR", "/nonexistent")
	assert.Equal(t, 0, GetWebhookMaxFiles())
}

func TestGetWebhookMaxFiles_SettingsOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	data := map[string]string{"WEBHOOK_MAX_FILES": "100"}
	b, _ := json.Marshal(data)
	os.WriteFile(filepath.Join(dir, "settings.json"), b, 0o644)

	assert.Equal(t, 100, GetWebhookMaxFiles())
}

func TestOverridesPath(t *testing.T) {
	t.Setenv("APP_DATA_DIR", "/tmp/test")
	assert.Equal(t, "/tmp/test/settings.json", overridesPath())
}

func TestOverridesPath_Default(t *testing.T) {
	t.Setenv("APP_DATA_DIR", "")
	assert.Equal(t, "settings.json", overridesPath())
}

func TestReadOverrides_NoFile(t *testing.T) {
	t.Setenv("APP_DATA_DIR", "/nonexistent/path")
	overrides := readOverrides()
	assert.Empty(t, overrides)
}

func TestReadOverrides_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte{}, 0o644)
	overrides := readOverrides()
	assert.Empty(t, overrides)
}

func TestReadOverrides_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte("{invalid}"), 0o644)
	overrides := readOverrides()
	assert.Empty(t, overrides)
}

func TestReadOverrides_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	data := map[string]string{"KEY": "value"}
	b, _ := json.Marshal(data)
	os.WriteFile(filepath.Join(dir, "settings.json"), b, 0o644)
	overrides := readOverrides()
	assert.Equal(t, "value", overrides["KEY"])
}
