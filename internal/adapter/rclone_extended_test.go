package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCall(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// 返回响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"result": "ok"})
	}))
	defer server.Close()

	// 创建客户端
	cfg := &RcloneConfig{
		BaseURL: server.URL,
	}
	client := NewRcloneClient(cfg)

	// 测试调用
	var resp map[string]any
	err := client.Call(context.Background(), "test/endpoint", map[string]any{"key": "value"}, &resp)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	if resp["result"] != "ok" {
		t.Errorf("expected result ok, got %v", resp["result"])
	}
}

func TestCallWithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	var resp map[string]any
	err := client.Call(context.Background(), "test", nil, &resp)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestListRemotes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/listremotes" {
			t.Errorf("expected /config/listremotes, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"remotes": []string{"local", "gdrive"},
		})
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	remotes, err := client.ListRemotes(context.Background())
	if err != nil {
		t.Fatalf("ListRemotes() error = %v", err)
	}

	if len(remotes) != 2 {
		t.Errorf("expected 2 remotes, got %d", len(remotes))
	}

	if remotes[0] != "local" {
		t.Errorf("expected first remote local, got %s", remotes[0])
	}
}

func TestCreateRemote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/create" {
			t.Errorf("expected /config/create, got %s", r.URL.Path)
		}

		var req CreateRemoteRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Name != "testremote" {
			t.Errorf("expected name testremote, got %s", req.Name)
		}
		if req.Type != "local" {
			t.Errorf("expected type local, got %s", req.Type)
		}

		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.CreateRemote(context.Background(), &CreateRemoteRequest{
		Name:       "testremote",
		Type:       "local",
		Parameters: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CreateRemote() error = %v", err)
	}
}

func TestDeleteRemote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/delete" {
			t.Errorf("expected /config/delete, got %s", r.URL.Path)
		}

		var req DeleteRemoteRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Name != "testremote" {
			t.Errorf("expected name testremote, got %s", req.Name)
		}

		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.DeleteRemote(context.Background(), "testremote")
	if err != nil {
		t.Fatalf("DeleteRemote() error = %v", err)
	}
}

func TestListPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/list" {
			t.Errorf("expected /operations/list, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"fs": "local:",
			"list": []map[string]any{
				{"Name": "file.txt", "Path": "/file.txt", "IsDir": false, "Size": 1024},
				{"Name": "dir", "Path": "/dir", "IsDir": true, "Size": 0},
			},
		})
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	items, err := client.ListPath(context.Background(), "local:", "")
	if err != nil {
		t.Fatalf("ListPath() error = %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	if items[0].Name != "file.txt" {
		t.Errorf("expected first item file.txt, got %s", items[0].Name)
	}

	if items[0].IsDir {
		t.Error("expected first item to be file")
	}

	if !items[1].IsDir {
		t.Error("expected second item to be directory")
	}
}

func TestMkdir(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/mkdir" {
			t.Errorf("expected /operations/mkdir, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.Mkdir(context.Background(), "local:", "/testdir")
	if err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
}

func TestDeleteFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/deletefile" {
			t.Errorf("expected /operations/deletefile, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.DeleteFile(context.Background(), "local:", "/test.txt")
	if err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}
}

func TestJobStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/status" {
			t.Errorf("expected /job/status, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":       123,
			"finished": true,
			"success":  true,
			"duration": 1.5,
		})
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	status, err := client.JobStatus(context.Background(), 123)
	if err != nil {
		t.Fatalf("JobStatus() error = %v", err)
	}

	if !status.Finished {
		t.Error("expected finished to be true")
	}

	if !status.Success {
		t.Error("expected success to be true")
	}

	if status.Duration != 1.5 {
		t.Errorf("expected duration 1.5, got %f", status.Duration)
	}
}

func TestStartJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/sync/copy"
		if r.URL.Path != expectedPath {
			t.Errorf("expected %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jobid": 456})
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	jobID, err := client.StartJob(context.Background(), "copy", "local:/src", "local:/dst")
	if err != nil {
		t.Fatalf("StartJob() error = %v", err)
	}

	if jobID != 456 {
		t.Errorf("expected jobid 456, got %d", jobID)
	}
}

func TestCopyDir(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync/copy" {
			t.Errorf("expected /sync/copy, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.CopyDir(context.Background(), "local:/src", "local:/dst")
	if err != nil {
		t.Fatalf("CopyDir() error = %v", err)
	}
}

func TestMoveDir(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync/move" {
			t.Errorf("expected /sync/move, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.MoveDir(context.Background(), "local:/src", "local:/dst")
	if err != nil {
		t.Fatalf("MoveDir() error = %v", err)
	}
}
