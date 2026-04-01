package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskRunnerImpl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync/copy" {
			t.Errorf("expected /sync/copy, got %s", r.URL.Path)
		}
		
		// 验证请求体
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		
		if req["srcFs"] != "local:/src" {
			t.Errorf("expected srcFs 'local:/src', got %v", req["srcFs"])
		}
		if req["dstFs"] != "gdrive:/dst" {
			t.Errorf("expected dstFs 'gdrive:/dst', got %v", req["dstFs"])
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jobid": 456})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	runner := NewTaskRunner(client)
	
	jobID, err := runner.RunTask(context.Background(), 1, "copy", "local", "/src", "gdrive", "/dst", "manual")
	if err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}
	
	if jobID != 456 {
		t.Errorf("expected jobID 456, got %d", jobID)
	}
}

func TestTaskRunnerImpl_Sync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync/sync" {
			t.Errorf("expected /sync/sync, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jobid": 789})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	runner := NewTaskRunner(client)
	
	jobID, err := runner.RunTask(context.Background(), 2, "sync", "local", "/src", "gdrive", "/dst", "schedule")
	if err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}
	
	if jobID != 789 {
		t.Errorf("expected jobID 789, got %d", jobID)
	}
}

func TestTaskRunnerImpl_Move(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sync/move" {
			t.Errorf("expected /sync/move, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jobid": 321})
	}))
	defer server.Close()
	
	client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
	runner := NewTaskRunner(client)
	
	jobID, err := runner.RunTask(context.Background(), 3, "move", "local", "/src", "gdrive", "/dst", "manual")
	if err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}
	
	if jobID != 321 {
		t.Errorf("expected jobID 321, got %d", jobID)
	}
}

func TestTaskRunnerImpl_PathHandling(t *testing.T) {
	// 测试路径处理
	tests := []struct {
		name    string
		srcPath string
		dstPath string
		wantSrc string
		wantDst string
	}{
		{"no leading slash", "src", "dst", "local:src", "gdrive:dst"},
		{"leading slash", "/src", "/dst", "local:/src", "gdrive:/dst"},
		{"nested path", "/a/b/c", "/x/y/z", "local:/a/b/c", "gdrive:/x/y/z"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req map[string]any
				json.NewDecoder(r.Body).Decode(&req)
				
				if req["srcFs"] != tt.wantSrc {
					t.Errorf("expected srcFs '%s', got %v", tt.wantSrc, req["srcFs"])
				}
				if req["dstFs"] != tt.wantDst {
					t.Errorf("expected dstFs '%s', got %v", tt.wantDst, req["dstFs"])
				}
				
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"jobid": 1})
			}))
			defer server.Close()
			
			client := NewRcloneClient(&RcloneConfig{BaseURL: server.URL})
			runner := NewTaskRunner(client)
			
			_, err := runner.RunTask(context.Background(), 1, "copy", "local", tt.srcPath, "gdrive", tt.dstPath, "manual")
			if err != nil {
				t.Fatalf("RunTask() error = %v", err)
			}
		})
	}
}
