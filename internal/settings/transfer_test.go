package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	return dir
}

func TestDefaults(t *testing.T) {
	d := defaults()
	assert.True(t, d.PostVerifyEnabled)
	assert.Equal(t, "mount", d.PostVerifyMode)
	assert.Equal(t, "5s", d.PostVerifyInterval)
	assert.Equal(t, "30m", d.PostVerifyTimeout)
	assert.Equal(t, "60s", d.PostVerifyMtimeGrace)
	assert.Equal(t, "size", d.PostVerifyMatch)
	assert.Equal(t, "30m", d.MinRerunInterval)
}

func TestDataDir_EnvSet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	got := dataDir()
	assert.Equal(t, dir, got)
}

func TestDataDir_Default(t *testing.T) {
	t.Setenv("APP_DATA_DIR", "")
	got := dataDir()
	assert.Equal(t, "./data", got)
}

func TestPath(t *testing.T) {
	dir := setupTestDir(t)
	p := path()
	assert.Equal(t, filepath.Join(dir, "transfer_settings.json"), p)
}

func TestLoad_FileNotExists(t *testing.T) {
	setupTestDir(t)
	ts, err := Load()
	require.NoError(t, err)
	def := defaults()
	assert.Equal(t, def, ts)
}

func TestLoad_ValidFile(t *testing.T) {
	dir := setupTestDir(t)
	content := `{
		"postVerifyEnabled": false,
		"postVerifyMode": "remote",
		"postVerifyInterval": "10s",
		"postVerifyTimeout": "1h",
		"postVerifyMtimeGrace": "120s",
		"postVerifyMatch": "size,modtime",
		"minRerunInterval": "1h"
	}`
	err := os.WriteFile(filepath.Join(dir, "transfer_settings.json"), []byte(content), 0o644)
	require.NoError(t, err)

	ts, err := Load()
	require.NoError(t, err)
	assert.False(t, ts.PostVerifyEnabled)
	assert.Equal(t, "remote", ts.PostVerifyMode)
	assert.Equal(t, "10s", ts.PostVerifyInterval)
	assert.Equal(t, "1h", ts.PostVerifyTimeout)
	assert.Equal(t, "120s", ts.PostVerifyMtimeGrace)
	assert.Equal(t, "size,modtime", ts.PostVerifyMatch)
	assert.Equal(t, "1h", ts.MinRerunInterval)
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := setupTestDir(t)
	err := os.WriteFile(filepath.Join(dir, "transfer_settings.json"), []byte("{invalid}"), 0o644)
	require.NoError(t, err)

	_, err = Load()
	assert.Error(t, err)
}

func TestLoad_PartialJSON_FillsDefaults(t *testing.T) {
	dir := setupTestDir(t)
	content := `{"postVerifyEnabled": false}`
	err := os.WriteFile(filepath.Join(dir, "transfer_settings.json"), []byte(content), 0o644)
	require.NoError(t, err)

	ts, err := Load()
	require.NoError(t, err)
	assert.False(t, ts.PostVerifyEnabled)
	def := defaults()
	assert.Equal(t, def.PostVerifyMode, ts.PostVerifyMode)
	assert.Equal(t, def.PostVerifyInterval, ts.PostVerifyInterval)
	assert.Equal(t, def.PostVerifyTimeout, ts.PostVerifyTimeout)
	assert.Equal(t, def.PostVerifyMtimeGrace, ts.PostVerifyMtimeGrace)
	assert.Equal(t, def.PostVerifyMatch, ts.PostVerifyMatch)
	assert.Equal(t, def.MinRerunInterval, ts.MinRerunInterval)
}

func TestSave_AndReload(t *testing.T) {
	setupTestDir(t)
	ts := TransferSettings{
		PostVerifyEnabled:    false,
		PostVerifyMode:       "remote",
		PostVerifyInterval:   "15s",
		PostVerifyTimeout:    "45m",
		PostVerifyMtimeGrace: "90s",
		PostVerifyMatch:      "modtime",
		MinRerunInterval:     "2h",
	}
	err := Save(ts)
	require.NoError(t, err)

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ts, loaded)
}

func TestSave_CreatesFile(t *testing.T) {
	setupTestDir(t)
	ts := defaults()
	err := Save(ts)
	require.NoError(t, err)

	p := path()
	_, err = os.Stat(p)
	assert.NoError(t, err)
}

func TestSave_ReadBackJSON(t *testing.T) {
	dir := setupTestDir(t)
	ts := TransferSettings{
		PostVerifyEnabled:  true,
		PostVerifyMode:     "mount",
		PostVerifyInterval: "5s",
	}
	err := Save(ts)
	require.NoError(t, err)

	b, err := os.ReadFile(filepath.Join(dir, "transfer_settings.json"))
	require.NoError(t, err)
	assert.Contains(t, string(b), `"postVerifyEnabled": true`)
	assert.Contains(t, string(b), `"postVerifyMode": "mount"`)
}
