package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/version"
)

func TestVersionController_HandleVersion_Success(t *testing.T) {
	version.CommitHash = "abc1234"

	mux := http.NewServeMux()
	rc := adapter.NewRcloneClient(&adapter.RcloneConfig{BaseURL: "http://127.0.0.1:5572"})
	ctrl := NewVersionController(rc)
	mux.HandleFunc("/api/version", ctrl.HandleVersion)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Without a real RC server, the rclone version will be empty
	resp, err := http.Get(srv.URL + "/api/version")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	if body["commitHash"] != "abc1234" {
		t.Fatalf("expected commitHash=abc1234, got %v", body["commitHash"])
	}
	// rcloneVersion will be "" since no real RC daemon
	if body["rcloneVersion"] != "" {
		t.Fatalf("expected rcloneVersion=empty, got %v", body["rcloneVersion"])
	}
}

func TestVersionController_HandleVersion_WithRcloneVersion(t *testing.T) {
	version.CommitHash = "def5678"

	rcSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"version":    "v1.65.0",
			"decomposed": []int{1, 65, 0},
			"isGit":      false,
			"isBeta":     false,
		})
	}))
	defer rcSrv.Close()

	rc := adapter.NewRcloneClient(&adapter.RcloneConfig{BaseURL: rcSrv.URL})
	ctrl := NewVersionController(rc)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/version", ctrl.HandleVersion)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/version")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	if body["commitHash"] != "def5678" {
		t.Fatalf("expected commitHash=def5678, got %v", body["commitHash"])
	}
	if body["rcloneVersion"] != "v1.65.0" {
		t.Fatalf("expected rcloneVersion=v1.65.0, got %v", body["rcloneVersion"])
	}
}

func TestNewVersionController(t *testing.T) {
	rc := adapter.NewRcloneClient(nil)
	ctrl := NewVersionController(rc)
	if ctrl == nil {
		t.Fatal("expected non-nil controller")
	}
}
