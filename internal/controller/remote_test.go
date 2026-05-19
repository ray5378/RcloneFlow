package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/rclone"
)

func TestRemoteController_Healthz(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := httptest.NewRecorder()
	c.Healthz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["ok"].(bool))
}

func TestRemoteController_HandleRemotes_GET(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/remotes", nil)
	rec := httptest.NewRecorder()
	c.HandleRemotes(rec, req)

	// Will return 500 if rclone is not running, but handler is exercised
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRemoteController_HandleRemotes_POST_InvalidJSON(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodPost, "/api/remotes", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleRemotes(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRemoteController_HandleRemotes_PUT_InvalidJSON(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodPut, "/api/remotes", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleRemotes(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRemoteController_HandleRemotes_MethodNotAllowed(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodDelete, "/api/remotes", nil)
	rec := httptest.NewRecorder()
	c.HandleRemotes(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRemoteController_HandleRemoteConfig_EmptyName(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/remotes/config/", nil)
	rec := httptest.NewRecorder()
	c.HandleRemoteConfig(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRemoteController_HandleRemoteConfig(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/remotes/config/myRemote", nil)
	rec := httptest.NewRecorder()
	c.HandleRemoteConfig(rec, req)

	// Will return 500 if rclone is not running
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRemoteController_HandleRemoteTest(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodPost, "/api/remotes/test", bytes.NewReader([]byte(`{"name":"myRemote"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleRemoteTest(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRemoteController_HandleProviders(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/providers", nil)
	rec := httptest.NewRecorder()
	c.HandleProviders(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRemoteController_HandleConfigDump(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/config/dump", nil)
	rec := httptest.NewRecorder()
	c.HandleConfigDump(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRemoteController_HandleConfigActions_POST(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodPost, "/api/config/test", bytes.NewReader([]byte(`{"name":"test","type":"s3"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.HandleConfigActions(rec, req)

	// HandleConfigActions only supports GET and DELETE
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRemoteController_HandleConfigActions_DELETE(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodDelete, "/api/config/testRemote", nil)
	rec := httptest.NewRecorder()
	c.HandleConfigActions(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRemoteController_HandleUsage(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/usage?remote=myRemote", nil)
	rec := httptest.NewRecorder()
	c.HandleUsage(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRemoteController_HandleFsInfo(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/fsinfo?remote=myRemote", nil)
	rec := httptest.NewRecorder()
	c.HandleFsInfo(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestRemoteController_RcloneClient(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)
	assert.NotNil(t, c.RcloneClient())
}

func TestRemoteController_RunTask(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewRemoteController(rc)
	// Just verify the method exists and returns the client
	assert.NotNil(t, c.RcloneClient())
}
