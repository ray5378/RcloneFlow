package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/adapter"
)

func TestRemoteController_RunTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		assert.Equal(t, "remotesrc:", body["srcFs"])
		assert.Equal(t, "remotedst:", body["dstFs"])
		assert.Equal(t, true, body["_async"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jobId": 42})
	}))
	defer ts.Close()

	rc := adapter.NewRcloneClient(&adapter.RcloneConfig{
		BaseURL: ts.URL,
	})
	ctrl := NewRemoteController(rc)

	jobID, err := ctrl.RunTask(context.Background(), 1, "copy", "remotesrc", "/", "remotedst", "/", "manual", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(42), jobID)
}

func TestRemoteController_RunTask_Sync(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jobId": 99})
	}))
	defer ts.Close()

	rc := adapter.NewRcloneClient(&adapter.RcloneConfig{
		BaseURL: ts.URL,
	})
	ctrl := NewRemoteController(rc)

	jobID, err := ctrl.RunTask(context.Background(), 2, "sync", "src", "/a", "dst", "/b", "scheduled", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(99), jobID)
}

func TestRemoteController_RunTask_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rclone error", 500)
	}))
	defer ts.Close()

	rc := adapter.NewRcloneClient(&adapter.RcloneConfig{
		BaseURL: ts.URL,
	})
	ctrl := NewRemoteController(rc)

	_, err := ctrl.RunTask(context.Background(), 3, "copy", "src", "/", "dst", "/", "manual", nil)
	assert.Error(t, err)
}
