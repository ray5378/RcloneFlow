package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCopyFile 测试复制文件
func TestCopyFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/copyfile" {
			t.Errorf("expected /operations/copyfile, got %s", r.URL.Path)
		}
		
		var req CopyFileRequest
		json.NewDecoder(r.Body).Decode(&req)
		
		if req.SrcFs != "local:" {
			t.Errorf("expected SrcFs 'local:', got %s", req.SrcFs)
		}
		if req.SrcRemote != "file.txt" {
			t.Errorf("expected SrcRemote 'file.txt', got %s", req.SrcRemote)
		}
		
		w.WriteHeader(200)
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	err := client.CopyFile(context.Background(), "local:", "file.txt", "gdrive:", "file.txt")
	if err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}
}

// TestMoveFile 测试移动文件
func TestMoveFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/movefile" {
			t.Errorf("expected /operations/movefile, got %s", r.URL.Path)
		}
		
		var req MoveFileRequest
		json.NewDecoder(r.Body).Decode(&req)
		
		if req.SrcFs != "local:" {
			t.Errorf("expected SrcFs 'local:', got %s", req.SrcFs)
		}
		if req.DstFs != "gdrive:" {
			t.Errorf("expected DstFs 'gdrive:', got %s", req.DstFs)
		}
		
		w.WriteHeader(200)
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	err := client.MoveFile(context.Background(), "local:", "file.txt", "gdrive:", "file.txt")
	if err != nil {
		t.Fatalf("MoveFile() error = %v", err)
	}
}

// TestPurge 测试删除目录
func TestPurge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/purge" {
			t.Errorf("expected /operations/purge, got %s", r.URL.Path)
		}
		
		var req PurgeRequest
		json.NewDecoder(r.Body).Decode(&req)
		
		if req.Fs != "local:" {
			t.Errorf("expected Fs 'local:', got %s", req.Fs)
		}
		if req.Remote != "dir" {
			t.Errorf("expected Remote 'dir', got %s", req.Remote)
		}
		
		w.WriteHeader(200)
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	err := client.Purge(context.Background(), "local:", "dir")
	if err != nil {
		t.Fatalf("Purge() error = %v", err)
	}
}

// TestPublicLink 测试生成分享链接
func TestPublicLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/publiclink" {
			t.Errorf("expected /operations/publiclink, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"url": "https://example.com/file.pdf",
		})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	url, err := client.PublicLink(context.Background(), "gdrive:", "file.pdf")
	if err != nil {
		t.Fatalf("PublicLink() error = %v", err)
	}
	
	if url != "https://example.com/file.pdf" {
		t.Errorf("expected url 'https://example.com/file.pdf', got %s", url)
	}
}

// TestDumpConfig 测试导出配置
func TestDumpConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/dump" {
			t.Errorf("expected /config/dump, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"local": map[string]any{
				"type": "local",
			},
			"gdrive": map[string]any{
				"type": "drive",
			},
		})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	config, err := client.DumpConfig(context.Background())
	if err != nil {
		t.Fatalf("DumpConfig() error = %v", err)
	}
	
	if len(config) != 2 {
		t.Errorf("expected 2 remotes, got %d", len(config))
	}
	
	if _, ok := config["local"]; !ok {
		t.Error("expected 'local' in config")
	}
	
	if _, ok := config["gdrive"]; !ok {
		t.Error("expected 'gdrive' in config")
	}
}

// TestGetUsage 测试获取使用量
func TestGetUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/about" {
			t.Errorf("expected /operations/about, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"used": 1024000,
			"free": 102400000,
		})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	about, err := client.GetUsage(context.Background(), "gdrive:")
	if err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	
	if about.Used != 1024000 {
		t.Errorf("expected Used 1024000, got %d", about.Used)
	}
	
	if about.Free != 102400000 {
		t.Errorf("expected Free 102400000, got %d", about.Free)
	}
}

// TestGetFsInfo 测试获取文件系统信息
func TestGetFsInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/fsinfo" {
			t.Errorf("expected /operations/fsinfo, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"name":      "local",
			"precision": 1000000000,
			"root":      "/",
		})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	info, err := client.GetFsInfo(context.Background(), "local:")
	if err != nil {
		t.Fatalf("GetFsInfo() error = %v", err)
	}
	
	if info.Name != "local" {
		t.Errorf("expected Name 'local', got %s", info.Name)
	}
	
	if info.Precision != 1000000000 {
		t.Errorf("expected Precision 1000000000, got %d", info.Precision)
	}
}

// TestSyncCopy 测试同步复制
func TestSyncCopy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync/copy" {
			t.Errorf("expected /sync/copy, got %s", r.URL.Path)
		}
		
		var req SyncCopyRequest
		json.NewDecoder(r.Body).Decode(&req)
		
		if req.SrcFs != "local:/src" {
			t.Errorf("expected SrcFs 'local:/src', got %s", req.SrcFs)
		}
		if req.DstFs != "gdrive:/dst" {
			t.Errorf("expected DstFs 'gdrive:/dst', got %s", req.DstFs)
		}
		if !req.CreateEmptySrcDirs {
			t.Error("expected CreateEmptySrcDirs to be true")
		}
		
		w.WriteHeader(200)
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	err := client.CopyDir(context.Background(), "local:/src", "gdrive:/dst")
	if err != nil {
		t.Fatalf("CopyDir() error = %v", err)
	}
}

// TestSyncMove 测试同步移动
func TestSyncMove(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync/move" {
			t.Errorf("expected /sync/move, got %s", r.URL.Path)
		}
		
		var req SyncMoveRequest
		json.NewDecoder(r.Body).Decode(&req)
		
		if req.SrcFs != "local:/src" {
			t.Errorf("expected SrcFs 'local:/src', got %s", req.SrcFs)
		}
		if !req.DeleteEmptySrcDirs {
			t.Error("expected DeleteEmptySrcDirs to be true")
		}
		
		w.WriteHeader(200)
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	err := client.MoveDir(context.Background(), "local:/src", "gdrive:/dst")
	if err != nil {
		t.Fatalf("MoveDir() error = %v", err)
	}
}

// TestJobStop 测试停止任务
func TestJobStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/stop" {
			t.Errorf("expected /job/stop, got %s", r.URL.Path)
		}
		
		var req StopJobRequest
		json.NewDecoder(r.Body).Decode(&req)
		
		if req.JobID != 123 {
			t.Errorf("expected JobID 123, got %d", req.JobID)
		}
		
		w.WriteHeader(200)
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	err := client.StopJob(context.Background(), 123)
	if err != nil {
		t.Fatalf("StopJob() error = %v", err)
	}
}

// TestJobList 测试列出任务
func TestJobList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/list" {
			t.Errorf("expected /job/list, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"jobids":      []int64{1, 2, 3},
			"running_ids":  []int64{1},
			"finished_ids": []int64{2, 3},
		})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	jobs, err := client.ListJobs(context.Background())
	if err != nil {
		t.Fatalf("ListJobs() error = %v", err)
	}
	
	if len(jobs.RunningIDs) != 1 {
		t.Errorf("expected 1 running job, got %d", len(jobs.RunningIDs))
	}
	
	if len(jobs.FinishedIDs) != 2 {
		t.Errorf("expected 2 finished jobs, got %d", len(jobs.FinishedIDs))
	}
}

// TestVersion 测试版本获取
func TestVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/core/version" {
			t.Errorf("expected /core/version, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"version":    "v1.65.0",
			"decomposed":  []int{1, 65, 0},
			"isGit":      true,
			"isBeta":     false,
			"os":         "linux",
			"goVersion":  "go1.21.0",
		})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	resp, err := client.Version(context.Background())
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	
	if resp.Version != "v1.65.0" {
		t.Errorf("expected version 'v1.65.0', got %s", resp.Version)
	}
	
	if !resp.IsGit {
		t.Error("expected IsGit to be true")
	}
	
	if resp.GoVersion != "go1.21.0" {
		t.Errorf("expected GoVersion 'go1.21.0', got %s", resp.GoVersion)
	}
}

// TestProviders 测试获取提供商
func TestProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/providers" {
			t.Errorf("expected /config/providers, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"providers": []map[string]any{
				{"Name": "local", "Hangul": "本地存储"},
				{"Name": "s3", "Hangul": "S3兼容存储"},
			},
		})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	
	providers, err := client.GetProviders(context.Background())
	if err != nil {
		t.Fatalf("GetProviders() error = %v", err)
	}
	
	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}
	
	if providers[0].Name != "local" {
		t.Errorf("expected first provider 'local', got %s", providers[0].Name)
	}
}
