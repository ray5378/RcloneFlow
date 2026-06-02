package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"rcloneflow/internal/adapter"
)

func setupRemoteController(t *testing.T) *RemoteController {
	t.Helper()
	rc := adapter.NewRcloneClient(nil)
	return NewRemoteController(rc)
}

func TestRemoteController_Healthz(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	ctrl.Healthz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRemoteController_HandleRemotes_Get(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/remotes", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRemotes(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRemoteController_HandleRemotes_Post(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/remotes", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRemotes(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRemoteController_HandleRemotes_Put(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodPut, "/api/remotes", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRemotes(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRemoteController_HandleRemotes_MethodNotAllowed(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/remotes", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRemotes(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRemoteController_HandleRemoteConfig_Get(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/remotes/config/test", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRemoteConfig(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRemoteController_HandleRemoteConfig_MissingName(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/remotes/config/", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRemoteConfig(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRemoteController_HandleRemoteConfig_MethodNotAllowed(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/remotes/config/test", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRemoteConfig(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRemoteController_HandleRemoteTest(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/remotes/test", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleRemoteTest(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRemoteController_HandleProviders(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/providers", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleProviders(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRemoteController_HandleProviders_MethodNotAllowed(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/providers", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleProviders(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRemoteController_HandleConfigDump(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleConfigDump(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRemoteController_HandleConfigDump_MethodNotAllowed(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/config", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleConfigDump(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRemoteController_HandleConfigActions_Get(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/config/test", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleConfigActions(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRemoteController_HandleConfigActions_Delete(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/config/test", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleConfigActions(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRemoteController_HandleConfigActions_MethodNotAllowed(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/config/test", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleConfigActions(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRemoteController_HandleUsage(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/usage/test:", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleUsage(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRemoteController_HandleUsage_MissingFs(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/usage/", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleUsage(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRemoteController_HandleUsage_MethodNotAllowed(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/usage/test:", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleUsage(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRemoteController_HandleFsInfo(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/fsinfo/test:", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleFsInfo(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRemoteController_HandleFsInfo_MissingFs(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodGet, "/api/fsinfo/", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleFsInfo(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRemoteController_HandleFsInfo_MethodNotAllowed(t *testing.T) {
	ctrl := setupRemoteController(t)

	req := httptest.NewRequest(http.MethodPost, "/api/fsinfo/test:", nil)
	rec := httptest.NewRecorder()
	ctrl.HandleFsInfo(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRemoteController_RcloneClient(t *testing.T) {
	ctrl := setupRemoteController(t)

	rc := ctrl.RcloneClient()
	assert.NotNil(t, rc)
}

func TestRemoteController_RunTask(t *testing.T) {
	ctrl := setupRemoteController(t)

	_, err := ctrl.RunTask(nil, 1, "copy", "src", "/path", "dst", "/path", "manual", nil)
	assert.Error(t, err)
}
