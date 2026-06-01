package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testCapture struct {
	params     map[string]any
	remoteName string
	remoteType string
}

func setupCapturingServer(t *testing.T) (*httptest.Server, *testCapture) {
	t.Helper()
	capture := &testCapture{params: make(map[string]any)}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateRemoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		capture.remoteName = req.Name
		capture.remoteType = req.Type
		for k, v := range req.Parameters {
			capture.params[k] = v
		}
		w.WriteHeader(200)
	}))
	return server, capture
}

func createRemoteHelper(t *testing.T, server *httptest.Server, name, typ string, params map[string]any) {
	t.Helper()
	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)
	err := client.CreateRemote(context.Background(), &CreateRemoteRequest{
		Name:       name,
		Type:       typ,
		Parameters: params,
	})
	if err != nil {
		t.Fatalf("CreateRemote(%s) error = %v", typ, err)
	}
}

// ============================================================================
// Wrapping Backends (remote 字段校验)
// ============================================================================

func TestProviderMapping_WrappingBackends(t *testing.T) {
	t.Run("crypt", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mycrypt", "crypt", map[string]any{
			"remote":                    "gdrive:encrypted",
			"password":                  "pass12345678",
			"password2":                 "salt87654321",
			"filename_encryption":       "standard",
			"directory_name_encryption": "true",
			"filename_encoding":         "base64",
		})

		assertHasKey(t, cap, "remote")
		assertHasKey(t, cap, "password")
		assertHasKey(t, cap, "password2")
		assertHasKey(t, cap, "filename_encryption")
		assertHasKey(t, cap, "directory_name_encryption")
	})

	t.Run("alias", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myalias", "alias", map[string]any{
			"remote": "gdrive:backup",
		})

		assertHasKey(t, cap, "remote")
		assertEq(t, "gdrive:backup", cap.params["remote"])
	})

	t.Run("chunker", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mychunker", "chunker", map[string]any{
			"remote":      "gdrive:chunked",
			"chunk_size":  "100M",
			"name_format": "*.rcc",
		})

		assertHasKey(t, cap, "remote")
		assertHasKey(t, cap, "chunk_size")
	})

	t.Run("hasher", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myhasher", "hasher", map[string]any{
			"remote": "gdrive:hashed",
			"hashes": "MD5,SHA1",
		})

		assertHasKey(t, cap, "remote")
	})

	t.Run("compress", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mycompress", "compress", map[string]any{
			"remote":              "gdrive:compressed",
			"mode":                "gzip",
			"compression-level":   "6",
		})

		assertHasKey(t, cap, "remote")
	})

	t.Run("union", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myunion", "union", map[string]any{
			"upstreams": "gdrive:dir1 gdrive:dir2",
			"action_policy": "epall",
		})

		assertHasKey(t, cap, "upstreams")
		assertMissingKey(t, cap, "remote")
	})
}

// ============================================================================
// Auth Backends (token / OAuth)
// ============================================================================

func TestProviderMapping_AuthBackends(t *testing.T) {
	t.Run("gdrive", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mygdrive", "drive", map[string]any{
			"client_id":     "my-client-id",
			"client_secret": "my-client-secret",
			"token":         "{\"access_token\":\"xxx\"}",
			"scope":         "drive",
			"root_folder_id": "root",
		})

		assertHasKey(t, cap, "client_id")
		assertHasKey(t, cap, "client_secret")
		assertHasKey(t, cap, "token")
		assertHasKey(t, cap, "scope")
	})

	t.Run("onedrive", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myonedrive", "onedrive", map[string]any{
			"client_id":     "my-client-id",
			"client_secret": "my-client-secret",
			"token":         "{\"access_token\":\"xxx\"}",
			"drive_type":    "personal",
			"drive_id":      "drive123",
		})

		assertHasKey(t, cap, "client_id")
		assertHasKey(t, cap, "client_secret")
		assertHasKey(t, cap, "token")
		assertHasKey(t, cap, "drive_type")
	})

	t.Run("dropbox", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mydropbox", "dropbox", map[string]any{
			"client_id":     "my-client-id",
			"client_secret": "my-client-secret",
			"token":         "{\"access_token\":\"xxx\"}",
		})

		assertHasKey(t, cap, "client_id")
		assertHasKey(t, cap, "client_secret")
		assertHasKey(t, cap, "token")
	})

	t.Run("pcloud", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mypcloud", "pcloud", map[string]any{
			"client_id":     "my-client-id",
			"client_secret": "my-client-secret",
			"token":         "{\"access_token\":\"xxx\"}",
		})

		assertHasKey(t, cap, "client_id")
		assertHasKey(t, cap, "client_secret")
		assertHasKey(t, cap, "token")
	})
}

// ============================================================================
// Object Storage Backends
// ============================================================================

func TestProviderMapping_ObjectStorageBackends(t *testing.T) {
	t.Run("s3", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mys3", "s3", map[string]any{
			"provider":          "AWS",
			"env_auth":          "false",
			"access_key_id":     "AKIAXXXX",
			"secret_access_key": "secret123",
			"region":            "us-east-1",
			"endpoint":          "",
			"bucket":            "my-bucket",
			"acl":               "private",
		})

		assertHasKey(t, cap, "provider")
		assertHasKey(t, cap, "access_key_id")
		assertHasKey(t, cap, "secret_access_key")
		assertHasKey(t, cap, "region")
		assertHasKey(t, cap, "bucket")
		assertMissingKey(t, cap, "endpoint") // empty filtered
	})

	t.Run("b2", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myb2", "b2", map[string]any{
			"account": "my-account-id",
			"key":     "my-app-key",
			"bucket":  "my-bucket",
		})

		assertHasKey(t, cap, "account")
		assertHasKey(t, cap, "key")
	})

	t.Run("azureblob", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myazure", "azureblob", map[string]any{
			"account":   "myaccount",
			"key":       "base64-key",
			"container": "my-container",
		})

		assertHasKey(t, cap, "account")
		assertHasKey(t, cap, "key")
		assertHasKey(t, cap, "container")
	})

	t.Run("swift", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myswift", "swift", map[string]any{
			"user":      "myuser",
			"key":       "mykey",
			"auth":      "https://auth.example.com/v2.0",
			"container": "mycontainer",
		})

		assertHasKey(t, cap, "user")
		assertHasKey(t, cap, "key")
		assertHasKey(t, cap, "auth")
	})
}

// ============================================================================
// File Transfer Backends (SFTP / FTP / WebDAV / SMB)
// ============================================================================

func TestProviderMapping_FileTransferBackends(t *testing.T) {
	t.Run("sftp", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mysftp", "sftp", map[string]any{
			"host": "sftp.example.com",
			"port": "22",
			"user": "myuser",
			"pass": "mypassword",
		})

		assertHasKey(t, cap, "host")
		assertHasKey(t, cap, "port")
		assertHasKey(t, cap, "user")
		assertHasKey(t, cap, "pass")
	})

	t.Run("ftp", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myftp", "ftp", map[string]any{
			"host": "ftp.example.com",
			"port": "21",
			"user": "myuser",
			"pass": "mypassword",
		})

		assertHasKey(t, cap, "host")
		assertHasKey(t, cap, "user")
		assertHasKey(t, cap, "pass")
	})

	t.Run("webdav", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mywebdav", "webdav", map[string]any{
			"url":  "https://dav.example.com/remote.php/dav",
			"user": "myuser",
			"pass": "mypassword",
		})

		assertHasKey(t, cap, "url")
		assertHasKey(t, cap, "user")
		assertHasKey(t, cap, "pass")
	})

	t.Run("smb", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mysmb", "smb", map[string]any{
			"host":   "192.168.1.100",
			"user":   "myuser",
			"pass":   "mypassword",
			"domain": "WORKGROUP",
		})

		assertHasKey(t, cap, "host")
		assertHasKey(t, cap, "user")
		assertHasKey(t, cap, "pass")
		assertHasKey(t, cap, "domain")
	})
}

// ============================================================================
// Local & Other Backends
// ============================================================================

func TestProviderMapping_LocalAndOther(t *testing.T) {
	t.Run("local", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mylocal", "local", map[string]any{
			"nounc":       "false",
			"copy_links":  "false",
			"links":       "false",
			"skip_links":  "false",
			"one_file_system": "false",
		})

		assertHasKey(t, cap, "nounc")
		assertHasKey(t, cap, "copy_links")
	})

	t.Run("memory", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "mymemory", "memory", map[string]any{})

		assertEq(t, cap.remoteType, "memory")
	})
}

// ============================================================================
// 网关/文件服务后端
// ============================================================================

func TestProviderMapping_GatewayBackends(t *testing.T) {
	t.Run("http", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myhttp", "http", map[string]any{
			"url": "https://files.example.com",
		})

		assertHasKey(t, cap, "url")
	})

	t.Run("hdfs", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "myhdfs", "hdfs", map[string]any{
			"namenode": "hdfs-namenode:8020",
			"user":     "hdfsuser",
		})

		assertHasKey(t, cap, "namenode")
		assertHasKey(t, cap, "user")
	})
}

// ============================================================================
// encoding 字段特殊处理
// ============================================================================

func TestProviderMapping_EncodingDefault(t *testing.T) {
	t.Run("encoding_empty_string_filtered", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "testenc", "s3", map[string]any{
			"provider": "AWS",
			"encoding": "",
		})

		assertMissingKey(t, cap, "encoding")
	})

	t.Run("encoding_non_empty_passed_through", func(t *testing.T) {
		server, cap := setupCapturingServer(t)
		defer server.Close()

		createRemoteHelper(t, server, "testenc2", "s3", map[string]any{
			"provider": "AWS",
			"encoding": "Slash,LtGt",
		})

		assertHasKey(t, cap, "encoding")
		assertEq(t, "Slash,LtGt", cap.params["encoding"])
	})
}

// ============================================================================
// headers 字段被过滤
// ============================================================================

func TestProviderMapping_HeadersFiltered(t *testing.T) {
	server, cap := setupCapturingServer(t)
	defer server.Close()

	createRemoteHelper(t, server, "testhdrs", "webdav", map[string]any{
		"url":     "https://dav.example.com",
		"headers": "Authorization,Bearer xxx",
	})

	assertMissingKey(t, cap, "headers")
}

// ============================================================================
// 空值过滤
// ============================================================================

func TestProviderMapping_EmptyValueFiltering(t *testing.T) {
	server, cap := setupCapturingServer(t)
	defer server.Close()

	createRemoteHelper(t, server, "testempty", "s3", map[string]any{
		"provider":      "AWS",
		"access_key_id": "AKIA123",
		"secret_access_key": "secret",
		"region":        "",
		"endpoint":      "",
		"bucket":        "",
		"acl":           "",
		"storage_class": "",
	})

	assertHasKey(t, cap, "provider")
	assertHasKey(t, cap, "access_key_id")
	assertHasKey(t, cap, "secret_access_key")
	assertMissingKey(t, cap, "region")
	assertMissingKey(t, cap, "endpoint")
	assertMissingKey(t, cap, "bucket")
	assertMissingKey(t, cap, "acl")
	assertMissingKey(t, cap, "storage_class")
}

// ============================================================================
// 辅助断言函数
// ============================================================================

func assertHasKey(t *testing.T, cap *testCapture, key string) {
	t.Helper()
	if _, ok := cap.params[key]; !ok {
		t.Errorf("expected key %q to be present in parameters, got params: %v", key, cap.params)
	}
}

func assertMissingKey(t *testing.T, cap *testCapture, key string) {
	t.Helper()
	if _, ok := cap.params[key]; ok {
		t.Errorf("expected key %q to be absent from parameters, got: %v", key, cap.params[key])
	}
}

func assertEq(t *testing.T, expected, actual any) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}