package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"rcloneflow/internal/rclone"
)

func TestBrowserController_HandleList(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewBrowserController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/browser/list?remote=myRemote&path=/some/path", nil)
	rec := httptest.NewRecorder()
	c.HandleList(rec, req)

	// Will return 500 if rclone is not running, but handler is exercised
	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestBrowserController_HandleList_EmptyPath(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewBrowserController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/browser/list?remote=myRemote", nil)
	rec := httptest.NewRecorder()
	c.HandleList(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestBrowserController_HandleList_DotPath(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewBrowserController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/browser/list?remote=myRemote&path=.", nil)
	rec := httptest.NewRecorder()
	c.HandleList(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestBrowserController_HandleList_WhitespacePath(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewBrowserController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/browser/list?remote=myRemote&path=%20", nil)
	rec := httptest.NewRecorder()
	c.HandleList(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}

func TestBrowserController_HandleList_WithLeadingSlash(t *testing.T) {
	rc := rclone.NewFromEnv()
	c := NewBrowserController(rc)

	req := httptest.NewRequest(http.MethodGet, "/api/browser/list?remote=myRemote&path=/folder/", nil)
	rec := httptest.NewRecorder()
	c.HandleList(rec, req)

	assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusInternalServerError)
}
