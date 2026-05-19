package rclone

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

func TestNewFromEnv(t *testing.T) {
	c := NewFromEnv()
	require.NotNil(t, c)
	require.NotNil(t, c.cli)
}

func TestNewWithConfig(t *testing.T) {
	cfg := &adapter.RcloneConfig{
		BaseURL: "http://localhost:5572",
	}
	c := NewWithConfig(cfg)
	require.NotNil(t, c)
	require.NotNil(t, c.cli)
}

func TestNewWithConfig_Nil(t *testing.T) {
	c := NewWithConfig(nil)
	require.NotNil(t, c)
	require.NotNil(t, c.cli)
}

func TestClient_Structure(t *testing.T) {
	c := NewFromEnv()

	assert.NotNil(t, c.ListPath)
	assert.NotNil(t, c.Version)
	assert.NotNil(t, c.ListRemotes)
	assert.NotNil(t, c.CreateRemote)
	assert.NotNil(t, c.GetConfig)
	assert.NotNil(t, c.DeleteRemote)
	assert.NotNil(t, c.DumpConfig)
	assert.NotNil(t, c.GetProviders)
	assert.NotNil(t, c.GetUsage)
	assert.NotNil(t, c.GetFsInfo)
	assert.NotNil(t, c.RunTask)
	assert.NotNil(t, c.CoreStats)
	assert.NotNil(t, c.CoreStatsGroup)
}

func TestClient_ListPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/operations/list", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"list": []map[string]any{
				{"Name": "file.txt", "Path": "file.txt", "IsDir": false, "Size": 100},
			},
		})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	result, err := c.ListPath(context.Background(), "local", "/test")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "file.txt", result[0]["Name"])
}

func TestClient_Version(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/core/version", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{"version": "v1.65.0"})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	v, err := c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "v1.65.0", v)
}

func TestClient_ListRemotes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/config/listremotes", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{"remotes": []string{"remote1", "remote2"}})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	remotes, err := c.ListRemotes(context.Background())
	require.NoError(t, err)
	assert.Len(t, remotes, 2)
}

func TestClient_CreateRemote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/config/create", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	err := c.CreateRemote(context.Background(), "myremote", "s3", map[string]any{"provider": "AWS"})
	require.NoError(t, err)
}

func TestClient_GetConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/config/get", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{"type": "s3", "provider": "AWS"})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	cfg, err := c.GetConfig(context.Background(), "myremote")
	require.NoError(t, err)
	assert.Equal(t, "s3", cfg["type"])
}

func TestClient_DeleteRemote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/config/delete", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	err := c.DeleteRemote(context.Background(), "myremote")
	require.NoError(t, err)
}

func TestClient_DumpConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/config/dump", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"remote1": map[string]any{"type": "s3"},
		})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	dump, err := c.DumpConfig(context.Background())
	require.NoError(t, err)
	assert.Contains(t, dump, "remote1")
}

func TestClient_GetProviders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/config/providers", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"providers": []map[string]any{
				{"Name": "s3", "Description": "Amazon S3"},
			},
		})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	providers, err := c.GetProviders(context.Background())
	require.NoError(t, err)
	assert.Len(t, providers, 1)
	assert.Equal(t, "s3", providers[0]["Name"])
}

func TestClient_GetUsage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/operations/about", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"used": float64(1000),
			"free": float64(4000),
		})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	usage, err := c.GetUsage(context.Background(), "remote:")
	require.NoError(t, err)
	assert.Equal(t, int64(1000), usage["used"])
	assert.Equal(t, int64(5000), usage["total"])
}

func TestClient_GetFsInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/operations/fsinfo", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"name":      "local",
			"precision": 1,
			"root":      "/test",
		})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	info, err := c.GetFsInfo(context.Background(), "local:/test")
	require.NoError(t, err)
	assert.Equal(t, "local", info["name"])
}

func TestClient_CoreStats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/core/stats", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"transferring": []map[string]any{},
			"checks":       0,
		})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	stats, err := c.CoreStats(context.Background())
	require.NoError(t, err)
	assert.Contains(t, stats, "transferring")
}

func TestClient_CoreStatsGroup(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/core/stats", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"group":        "mygroup",
			"transferring": []map[string]any{},
		})
	}))
	defer srv.Close()

	c := NewWithConfig(&adapter.RcloneConfig{BaseURL: srv.URL})
	stats, err := c.CoreStatsGroup(context.Background(), "mygroup")
	require.NoError(t, err)
	assert.Equal(t, "mygroup", stats["group"])
}
