package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSettingsTestDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
}

func TestSettingsController_HandleGet_Defaults(t *testing.T) {
	setupSettingsTestDir(t)
	ctrl := NewSettingsController()

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp settingsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "24h", resp.Auth["ACCESS_TOKEN_TTL"]["default"])
	assert.Equal(t, "info", resp.Log["LOG_LEVEL"]["default"])
	assert.Equal(t, "info", resp.Log["LOG_LEVEL"]["effective"])
}

func TestSettingsController_HandleGet_WithOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	ctrl := NewSettingsController()

	overrides := map[string]string{
		"LOG_LEVEL": "debug",
	}
	b, _ := json.Marshal(overrides)
	os.WriteFile(filepath.Join(dir, "settings.json"), b, 0o644)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp settingsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "debug", resp.Log["LOG_LEVEL"]["effective"])
	assert.Equal(t, "info", resp.Log["LOG_LEVEL"]["default"])
}

func TestSettingsController_HandleGet_EnvOverridesStored(t *testing.T) {
	setupSettingsTestDir(t)
	t.Setenv("LOG_LEVEL", "warn")
	ctrl := NewSettingsController()

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)

	var resp settingsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "warn", resp.Log["LOG_LEVEL"]["effective"])
}

func TestSettingsController_HandlePut_UpdateValues(t *testing.T) {
	setupSettingsTestDir(t)
	ctrl := NewSettingsController()

	payload := settingsPayload{
		Values: map[string]string{
			"LOG_LEVEL": "debug",
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["ok"].(bool))
}

func TestSettingsController_HandlePut_Reset(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	ctrl := NewSettingsController()

	existing := map[string]string{"LOG_LEVEL": "debug"}
	b, _ := json.Marshal(existing)
	os.WriteFile(filepath.Join(dir, "settings.json"), b, 0o644)

	payload := settingsPayload{Reset: true}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["ok"].(bool))
	assert.True(t, resp["reset"].(bool))
}

func TestSettingsController_HandlePut_IgnoresUnknownKeys(t *testing.T) {
	setupSettingsTestDir(t)
	ctrl := NewSettingsController()

	payload := settingsPayload{
		Values: map[string]string{
			"UNKNOWN_KEY": "value",
			"LOG_LEVEL":   "error",
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	over := readOverrides()
	_, exists := over["UNKNOWN_KEY"]
	assert.False(t, exists)
	assert.Equal(t, "error", over["LOG_LEVEL"])
}

func TestSettingsController_HandlePut_MethodNotAllowed(t *testing.T) {
	setupSettingsTestDir(t)
	ctrl := NewSettingsController()

	req := httptest.NewRequest(http.MethodDelete, "/api/settings", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestSettingsController_HandlePut_WithCleanupHook(t *testing.T) {
	setupSettingsTestDir(t)
	ctrl := NewSettingsController()

	var capturedInterval, capturedRetention int
	ReplanCleanupHook = func(intervalHours, retentionDays int) {
		capturedInterval = intervalHours
		capturedRetention = retentionDays
	}
	defer func() { ReplanCleanupHook = nil }()

	payload := settingsPayload{
		Values: map[string]string{
			"CLEANUP_INTERVAL_HOURS":       "12",
			"FINAL_SUMMARY_RETENTION_DAYS": "30",
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 12, capturedInterval)
	assert.Equal(t, 30, capturedRetention)
}

func TestDefaultsMap(t *testing.T) {
	defs := defaultsMap()
	assert.NotEmpty(t, defs["ACCESS_TOKEN_TTL"])
	assert.NotEmpty(t, defs["REFRESH_TOKEN_TTL"])
	assert.NotEmpty(t, defs["LOG_LEVEL"])
	assert.NotEmpty(t, defs["LOG_OUTPUT"])
}

func TestEffectiveValue_EnvTakesPrecedence(t *testing.T) {
	t.Setenv("TEST_KEY", "env_value")
	defs := map[string]string{"TEST_KEY": "default"}
	over := map[string]string{"TEST_KEY": "stored"}
	assert.Equal(t, "env_value", effectiveValue("TEST_KEY", over, defs))
}

func TestEffectiveValue_StoredOverridesDefault(t *testing.T) {
	t.Setenv("TEST_KEY2", "")
	defs := map[string]string{"TEST_KEY2": "default"}
	over := map[string]string{"TEST_KEY2": "stored"}
	assert.Equal(t, "stored", effectiveValue("TEST_KEY2", over, defs))
}

func TestEffectiveValue_FallsBackToDefault(t *testing.T) {
	t.Setenv("TEST_KEY3", "")
	defs := map[string]string{"TEST_KEY3": "default"}
	over := map[string]string{}
	assert.Equal(t, "default", effectiveValue("TEST_KEY3", over, defs))
}

func TestReadOverrides_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte{}, 0o644)
	over := readOverrides()
	assert.Empty(t, over)
}

func TestReadOverrides_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APP_DATA_DIR", dir)
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte("{invalid}"), 0o644)
	over := readOverrides()
	assert.Empty(t, over)
}

func TestReadOverrides_NoFile(t *testing.T) {
	setupSettingsTestDir(t)
	over := readOverrides()
	assert.Empty(t, over)
}

func TestAtoiDefault(t *testing.T) {
	assert.Equal(t, 42, atoiDefault("42", 0))
	assert.Equal(t, 0, atoiDefault("", 0))
	assert.Equal(t, 10, atoiDefault("", 10))
	assert.Equal(t, 5, atoiDefault("abc", 5))
}
