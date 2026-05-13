package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func withAppDataDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	old := os.Getenv("APP_DATA_DIR")
	if err := os.Setenv("APP_DATA_DIR", tmpDir); err != nil {
		t.Fatalf("Setenv(APP_DATA_DIR) error = %v", err)
	}
	t.Cleanup(func() {
		if old == "" {
			_ = os.Unsetenv("APP_DATA_DIR")
		} else {
			_ = os.Setenv("APP_DATA_DIR", old)
		}
	})
	return tmpDir
}

func TestSettingsController_HandleSettings_MethodDispatch(t *testing.T) {
	withAppDataDir(t)
	ctrl := NewSettingsController()

	t.Run("GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleSettings(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PUT", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewBufferString(`{"values":{"LOG_LEVEL":"debug"}}`))
		rec := httptest.NewRecorder()
		ctrl.HandleSettings(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/settings", nil)
		rec := httptest.NewRecorder()
		ctrl.HandleSettings(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d, want 405", rec.Code)
		}
	})
}

func TestSettingsController_HandleGet_UsesDefaultsOverridesAndEnv(t *testing.T) {
	appDir := withAppDataDir(t)
	ctrl := NewSettingsController()

	overrides := map[string]string{
		"LOG_LEVEL":                "warn",
		"CLEANUP_INTERVAL_HOURS":   "12",
		"WEBHOOK_MAX_FILES":        "9",
		"RUNNING_HINT_DEBUG_ENABLED": "true",
	}
	b, _ := json.Marshal(overrides)
	if err := os.WriteFile(filepath.Join(appDir, "settings.json"), b, 0o644); err != nil {
		t.Fatalf("WriteFile(settings.json) error = %v", err)
	}

	oldLogOutput := os.Getenv("LOG_OUTPUT")
	if err := os.Setenv("LOG_OUTPUT", "stderr"); err != nil {
		t.Fatalf("Setenv(LOG_OUTPUT) error = %v", err)
	}
	defer func() {
		if oldLogOutput == "" {
			_ = os.Unsetenv("LOG_OUTPUT")
		} else {
			_ = os.Setenv("LOG_OUTPUT", oldLogOutput)
		}
	}()

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	logSection, _ := body["log"].(map[string]any)
	if logSection == nil {
		t.Fatalf("expected log section, got %#v", body["log"])
	}
	logLevel, _ := logSection["LOG_LEVEL"].(map[string]any)
	logOutput, _ := logSection["LOG_OUTPUT"].(map[string]any)
	if logLevel["effective"] != "warn" {
		t.Fatalf("LOG_LEVEL effective=%#v, want warn", logLevel["effective"])
	}
	if logLevel["default"] != "info" {
		t.Fatalf("LOG_LEVEL default=%#v, want info", logLevel["default"])
	}
	if logOutput["effective"] != "stderr" {
		t.Fatalf("LOG_OUTPUT effective=%#v, want stderr env override", logOutput["effective"])
	}

	historySection, _ := body["history"].(map[string]any)
	cleanup, _ := historySection["CLEANUP_INTERVAL_HOURS"].(map[string]any)
	if cleanup["effective"] != "12" {
		t.Fatalf("CLEANUP_INTERVAL_HOURS effective=%#v, want 12", cleanup["effective"])
	}

	stored, _ := body["stored"].(map[string]any)
	if stored["WEBHOOK_MAX_FILES"] != "9" {
		t.Fatalf("stored WEBHOOK_MAX_FILES=%#v, want 9", stored["WEBHOOK_MAX_FILES"])
	}
	meta, _ := body["meta"].(map[string]any)
	if meta["APP_DATA_DIR"] != appDir {
		t.Fatalf("meta APP_DATA_DIR=%#v, want %q", meta["APP_DATA_DIR"], appDir)
	}
	if meta["settingsPath"] != filepath.Join(appDir, "settings.json") {
		t.Fatalf("meta settingsPath=%#v", meta["settingsPath"])
	}
}

func TestSettingsController_HandlePut_WritesValidKeysAndTriggersHooks(t *testing.T) {
	appDir := withAppDataDir(t)
	ctrl := NewSettingsController()

	oldCleanupHook := ReplanCleanupHook
	oldLogCleanupHook := ReplanLogCleanupHook
	defer func() {
		ReplanCleanupHook = oldCleanupHook
		ReplanLogCleanupHook = oldLogCleanupHook
	}()

	var cleanupCalls []struct{ interval, retention int }
	var logCleanupCalls []int
	ReplanCleanupHook = func(intervalHours int, retentionDays int) {
		cleanupCalls = append(cleanupCalls, struct{ interval, retention int }{intervalHours, retentionDays})
	}
	ReplanLogCleanupHook = func(retentionDays int) {
		logCleanupCalls = append(logCleanupCalls, retentionDays)
	}

	payload := `{"values":{"LOG_LEVEL":" debug ","CLEANUP_INTERVAL_HOURS":" 6 ","FINAL_SUMMARY_RETENTION_DAYS":" 15 ","UNKNOWN_KEY":"x"}}`
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewBufferString(payload))
	rec := httptest.NewRecorder()
	ctrl.HandleSettings(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
	}

	content, err := os.ReadFile(filepath.Join(appDir, "settings.json"))
	if err != nil {
		t.Fatalf("ReadFile(settings.json) error = %v", err)
	}
	var stored map[string]string
	if err := json.Unmarshal(content, &stored); err != nil {
		t.Fatalf("json.Unmarshal(stored) error = %v", err)
	}
	if stored["LOG_LEVEL"] != "debug" {
		t.Fatalf("LOG_LEVEL=%q, want debug", stored["LOG_LEVEL"])
	}
	if stored["CLEANUP_INTERVAL_HOURS"] != "6" || stored["FINAL_SUMMARY_RETENTION_DAYS"] != "15" {
		t.Fatalf("unexpected stored values=%#v", stored)
	}
	if _, ok := stored["UNKNOWN_KEY"]; ok {
		t.Fatalf("UNKNOWN_KEY should be ignored, got %#v", stored)
	}
	if len(cleanupCalls) != 1 || cleanupCalls[0].interval != 6 || cleanupCalls[0].retention != 15 {
		t.Fatalf("unexpected cleanupCalls=%#v", cleanupCalls)
	}
	if len(logCleanupCalls) != 1 || logCleanupCalls[0] != 15 {
		t.Fatalf("unexpected logCleanupCalls=%#v", logCleanupCalls)
	}
}

func TestSettingsController_HandlePut_ResetAndInvalidPayload(t *testing.T) {
	appDir := withAppDataDir(t)
	ctrl := NewSettingsController()

	if err := os.WriteFile(filepath.Join(appDir, "settings.json"), []byte(`{"LOG_LEVEL":"warn"}`), 0o644); err != nil {
		t.Fatalf("WriteFile(settings.json) error = %v", err)
	}

	t.Run("reset clears stored overrides", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewBufferString(`{"reset":true}`))
		rec := httptest.NewRecorder()
		ctrl.HandleSettings(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
		content, err := os.ReadFile(filepath.Join(appDir, "settings.json"))
		if err != nil {
			t.Fatalf("ReadFile(settings.json) error = %v", err)
		}
		if string(bytes.TrimSpace(content)) != "{}" {
			t.Fatalf("expected cleared settings.json, got %q", string(content))
		}
	})

	t.Run("invalid json becomes no-op success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewBufferString(`{"values":`))
		rec := httptest.NewRecorder()
		ctrl.HandleSettings(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestAtoiDefault(t *testing.T) {
	if got := atoiDefault("42", 7); got != 42 {
		t.Fatalf("atoiDefault(valid)=%d, want 42", got)
	}
	if got := atoiDefault("bad", 7); got != 7 {
		t.Fatalf("atoiDefault(invalid)=%d, want 7", got)
	}
	if got := atoiDefault("", 7); got != 7 {
		t.Fatalf("atoiDefault(empty)=%d, want 7", got)
	}
}
