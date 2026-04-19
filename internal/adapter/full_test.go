package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCopyFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/operations/copyfile":
			var req CopyFileRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.SrcFs != "local:" {
				t.Errorf("expected SrcFs 'local:', got %s", req.SrcFs)
			}
			if req.SrcRemote != "file.txt" {
				t.Errorf("expected SrcRemote 'file.txt', got %s", req.SrcRemote)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"jobid": 1})
		case "/job/status":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"finished": true, "success": true})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	if err := client.CopyFile(context.Background(), "local:", "file.txt", "gdrive:", "file.txt"); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}
}

func TestMoveFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/operations/movefile":
			var req MoveFileRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.SrcFs != "local:" {
				t.Errorf("expected SrcFs 'local:', got %s", req.SrcFs)
			}
			if req.DstFs != "gdrive:" {
				t.Errorf("expected DstFs 'gdrive:', got %s", req.DstFs)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"jobid": 1})
		case "/job/status":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"finished": true, "success": true})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	if err := client.MoveFile(context.Background(), "local:", "file.txt", "gdrive:", "file.txt"); err != nil {
		t.Fatalf("MoveFile() error = %v", err)
	}
}

func TestPurge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/purge" {
			t.Fatalf("expected /operations/purge, got %s", r.URL.Path)
		}
		var req PurgeRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Fs != "local:" || req.Remote != "dir" {
			t.Fatalf("unexpected purge request: %+v", req)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	if err := client.Purge(context.Background(), "local:", "dir"); err != nil {
		t.Fatalf("Purge() error = %v", err)
	}
}

func TestPublicLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/publiclink" {
			t.Fatalf("expected /operations/publiclink, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"url": "https://example.com/file.pdf"})
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	url, err := client.PublicLink(context.Background(), "gdrive:", "file.pdf")
	if err != nil {
		t.Fatalf("PublicLink() error = %v", err)
	}
	if url != "https://example.com/file.pdf" {
		t.Errorf("expected url https://example.com/file.pdf, got %s", url)
	}
}

func TestDumpConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/dump" {
			t.Fatalf("expected /config/dump, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"local":  map[string]any{"type": "local"},
			"gdrive": map[string]any{"type": "drive"},
		})
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	cfg, err := client.DumpConfig(context.Background())
	if err != nil {
		t.Fatalf("DumpConfig() error = %v", err)
	}
	if len(cfg) != 2 {
		t.Fatalf("expected 2 remotes, got %d", len(cfg))
	}
}

func TestGetUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/about" {
			t.Fatalf("expected /operations/about, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"used": 1024000, "free": 102400000})
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	about, err := client.GetUsage(context.Background(), "gdrive:")
	if err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	if about.Used != 1024000 || about.Free != 102400000 {
		t.Fatalf("unexpected usage: %+v", about)
	}
}

func TestGetFsInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/fsinfo" {
			t.Fatalf("expected /operations/fsinfo, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "local", "precision": 1000000000, "root": "/"})
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	info, err := client.GetFsInfo(context.Background(), "local:")
	if err != nil {
		t.Fatalf("GetFsInfo() error = %v", err)
	}
	if info.Name != "local" || info.Precision != 1000000000 {
		t.Fatalf("unexpected fs info: %+v", info)
	}
}

func TestSyncCopy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sync/copy":
			var req SyncCopyRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.SrcFs != "local:/src" || req.DstFs != "gdrive:/dst" || !req.CreateEmptySrcDirs {
				t.Fatalf("unexpected sync copy request: %+v", req)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"jobid": 1})
		case "/job/status":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"finished": true, "success": true})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	if err := client.CopyDir(context.Background(), "local:/src", "gdrive:/dst"); err != nil {
		t.Fatalf("CopyDir() error = %v", err)
	}
}

func TestSyncMove(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sync/move":
			var req SyncMoveRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.SrcFs != "local:/src" || !req.DeleteEmptySrcDirs {
				t.Fatalf("unexpected sync move request: %+v", req)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"jobid": 1})
		case "/job/status":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"finished": true, "success": true})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	if err := client.MoveDir(context.Background(), "local:/src", "gdrive:/dst"); err != nil {
		t.Fatalf("MoveDir() error = %v", err)
	}
}

func TestJobStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/stop" {
			t.Fatalf("expected /job/stop, got %s", r.URL.Path)
		}
		var req StopJobRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.JobID != 123 {
			t.Fatalf("expected JobID 123, got %d", req.JobID)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	if err := client.StopJob(context.Background(), 123); err != nil {
		t.Fatalf("StopJob() error = %v", err)
	}
}

func TestJobList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/list" {
			t.Fatalf("expected /job/list, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"jobids": []int64{1, 2, 3}, "running_ids": []int64{1}, "finished_ids": []int64{2, 3}})
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	jobs, err := client.ListJobs(context.Background())
	if err != nil {
		t.Fatalf("ListJobs() error = %v", err)
	}
	if len(jobs.RunningIDs) != 1 || len(jobs.FinishedIDs) != 2 {
		t.Fatalf("unexpected jobs: %+v", jobs)
	}
}

func TestVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/core/version" {
			t.Fatalf("expected /core/version, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"version": "v1.65.0", "decomposed": []int{1, 65, 0}, "isGit": true, "isBeta": false, "os": "linux", "goVersion": "go1.21.0"})
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	resp, err := client.Version(context.Background())
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	if resp.Version != "v1.65.0" || !resp.IsGit || resp.GoVersion != "go1.21.0" {
		t.Fatalf("unexpected version resp: %+v", resp)
	}
}

func TestProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/providers" {
			t.Fatalf("expected /config/providers, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"providers": []map[string]any{{"Name": "local", "Hangul": "本地存储"}, {"Name": "s3", "Hangul": "S3兼容存储"}}})
	}))
	defer server.Close()

	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	providers, err := client.GetProviders(context.Background())
	if err != nil {
		t.Fatalf("GetProviders() error = %v", err)
	}
	if len(providers) != 2 || providers[0].Name != "local" {
		t.Fatalf("unexpected providers: %+v", providers)
	}
}
