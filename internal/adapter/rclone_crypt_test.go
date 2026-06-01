package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateRemote_Crypt_ParameterMapping(t *testing.T) {
	capturedParams := make(map[string]any)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/create" {
			t.Errorf("expected /config/create, got %s", r.URL.Path)
		}
		var req CreateRemoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		for k, v := range req.Parameters {
			capturedParams[k] = v
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.CreateRemote(context.Background(), &CreateRemoteRequest{
		Name: "secret",
		Type: "crypt",
		Parameters: map[string]any{
			"remote":                     "gdrive:encrypted",
			"password":                   "my-secret-password-123",
			"password2":                  "my-salt-password-456",
			"filename_encryption":        "standard",
			"directory_name_encryption":  "true",
			"filename_encoding":          "base64",
		},
	})
	if err != nil {
		t.Fatalf("CreateRemote() error = %v", err)
	}

	// remote 字段验证：必须包含远程名和路径
	t.Run("remote field format", func(t *testing.T) {
		remote, ok := capturedParams["remote"].(string)
		if !ok {
			t.Fatal("remote parameter missing or not a string")
		}
		if !strings.Contains(remote, ":") {
			t.Errorf("remote must contain ':', got %q", remote)
		}
		parts := strings.SplitN(remote, ":", 2)
		if parts[0] == "" {
			t.Error("remote name must not be empty (must specify a remote before colon)")
		}
	})

	// password 字段验证：必须是明文密码
	t.Run("password field is plain text", func(t *testing.T) {
		password, ok := capturedParams["password"].(string)
		if !ok {
			t.Fatal("password parameter missing or not a string")
		}
		if password != "my-secret-password-123" {
			t.Errorf("password should be plain text, got %q", password)
		}
		if len(password) < 8 {
			t.Logf("WARNING: password length %d is less than recommended 8 characters", len(password))
		}
	})

	// password2 字段验证：可选的盐密码
	t.Run("password2 salt field", func(t *testing.T) {
		password2, ok := capturedParams["password2"].(string)
		if !ok {
			t.Fatal("password2 parameter missing or not a string")
		}
		if password2 != "my-salt-password-456" {
			t.Errorf("expected salt password 'my-salt-password-456', got %q", password2)
		}
		if password2 == capturedParams["password"] {
			t.Error("password2 should be different from password for security")
		}
	})

	// filename_encryption 验证：必须是 standard/obfuscate/off 之一
	t.Run("filename_encryption valid options", func(t *testing.T) {
		fe, ok := capturedParams["filename_encryption"].(string)
		if !ok {
			t.Fatal("filename_encryption parameter missing or not a string")
		}
		valid := map[string]bool{"standard": true, "obfuscate": true, "off": true}
		if !valid[fe] {
			t.Errorf("filename_encryption must be one of standard/obfuscate/off, got %q", fe)
		}
	})

	// directory_name_encryption 验证：布尔类型
	t.Run("directory_name_encryption boolean", func(t *testing.T) {
		v := capturedParams["directory_name_encryption"]
		// rclone accepts both string "true"/"false" and boolean true/false
		isTruthy := false
		switch val := v.(type) {
		case string:
			isTruthy = val == "true" || val == "1"
		case bool:
			isTruthy = val
		default:
			t.Errorf("directory_name_encryption type %T unexpected", v)
		}
		if !isTruthy {
			t.Errorf("directory_name_encryption should be truthy, got %v", v)
		}
	})
}

func TestCreateRemote_Crypt_MinimalRequired(t *testing.T) {
	capturedParams := make(map[string]any)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/create" {
			t.Errorf("expected /config/create, got %s", r.URL.Path)
		}
		var req CreateRemoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		for k, v := range req.Parameters {
			capturedParams[k] = v
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.CreateRemote(context.Background(), &CreateRemoteRequest{
		Name: "secret",
		Type: "crypt",
		Parameters: map[string]any{
			"remote":   "gdrive:",
			"password": "my-password",
		},
	})
	if err != nil {
		t.Fatalf("CreateRemote() error = %v", err)
	}

	// 只需 remote + password 即可创建 crypt 存储
	if _, ok := capturedParams["remote"]; !ok {
		t.Error("remote is required for crypt")
	}
	if _, ok := capturedParams["password"]; !ok {
		t.Error("password is required for crypt")
	}

	// rclone会在没有password2时使用内置盐值，不会报错
	_, hasPassword2 := capturedParams["password2"]
	_ = hasPassword2
}

func TestCreateRemote_Crypt_RemoteFormatVariations(t *testing.T) {
	tests := []struct {
		name   string
		remote string
		valid  bool
	}{
		{"remote with colon only", "gdrive:", true},
		{"remote with path", "gdrive:backup", true},
		{"remote with nested path", "gdrive:backup/encrypted/data", true},
		{"remote with bucket", "s3:mybucket", true},
		{"local path", "/tmp/encrypted", false},
		{"relative path", "localdir", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedRemote string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req CreateRemoteRequest
				json.NewDecoder(r.Body).Decode(&req)
				if v, ok := req.Parameters["remote"].(string); ok {
					capturedRemote = v
				}
				w.WriteHeader(200)
			}))
			defer server.Close()

			cfg := &RcloneConfig{BaseURL: server.URL}
			client := NewRcloneClient(cfg)

			err := client.CreateRemote(context.Background(), &CreateRemoteRequest{
				Name: "secret", Type: "crypt",
				Parameters: map[string]any{
					"remote":   tt.remote,
					"password": "test",
				},
			})
			if err != nil {
				t.Fatalf("CreateRemote() error = %v", err)
			}

			hasColon := strings.Contains(capturedRemote, ":")
			if tt.valid {
				if !hasColon {
					t.Errorf("expected remote %q to contain colon for valid remote reference", capturedRemote)
				}
			} else {
				if hasColon {
					t.Logf("Note: remote %q contains colon (may be unexpected for local)", capturedRemote)
				}
			}
		})
	}
}

func TestCreateRemote_Crypt_ObscureModeFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateRemoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}

		// Obscure: false means rclone treats passwords as plain text
		if req.Opt == nil {
			t.Error("Opt should be set (with Obscure: false)")
			return
		}
		if req.Opt.Obscure {
			t.Error("Obscure should be false so rclone accepts plain text passwords")
		}

		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.CreateRemote(context.Background(), &CreateRemoteRequest{
		Name: "secret", Type: "crypt",
		Parameters: map[string]any{
			"remote":   "gdrive:",
			"password": "plaintext-password",
		},
	})
	if err != nil {
		t.Fatalf("CreateRemote() error = %v", err)
	}
}

func TestCreateRemote_Crypt_EmptyFieldsFiltered(t *testing.T) {
	capturedParams := make(map[string]any)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateRemoteRequest
		json.NewDecoder(r.Body).Decode(&req)
		for k, v := range req.Parameters {
			capturedParams[k] = v
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	// 模拟前端：空字符串可选字段被过滤
	params := map[string]any{
		"remote":                     "gdrive:backup",
		"password":                   "mypassword",
		"password2":                  "",
		"filename_encryption":        "standard",
		"directory_name_encryption":  "",
		"filename_encoding":          "",
	}

	// CreateRemote 内部的 cleanParams 会过滤空字符串
	err := client.CreateRemote(context.Background(), &CreateRemoteRequest{
		Name:       "secret",
		Type:       "crypt",
		Parameters: params,
	})
	if err != nil {
		t.Fatalf("CreateRemote() error = %v", err)
	}

	// 验证空字符串被过滤
	if _, exists := capturedParams["password2"]; exists {
		t.Error("password2 should be filtered out (empty string)")
	}
	if _, exists := capturedParams["directory_name_encryption"]; exists {
		t.Error("directory_name_encryption should be filtered out (empty string)")
	}
	if _, exists := capturedParams["filename_encoding"]; exists {
		t.Error("filename_encoding should be filtered out (empty string)")
	}

	// 非空值应保留
	if _, exists := capturedParams["remote"]; !exists {
		t.Error("remote should be present (non-empty)")
	}
	if _, exists := capturedParams["password"]; !exists {
		t.Error("password should be present (non-empty)")
	}
}

func TestCreateRemote_Crypt_AllStandardOptions(t *testing.T) {
	capturedParams := make(map[string]any)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateRemoteRequest
		json.NewDecoder(r.Body).Decode(&req)
		for k, v := range req.Parameters {
			capturedParams[k] = v
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	cfg := &RcloneConfig{BaseURL: server.URL}
	client := NewRcloneClient(cfg)

	err := client.CreateRemote(context.Background(), &CreateRemoteRequest{
		Name: "mycrypt",
		Type: "crypt",
		Parameters: map[string]any{
			"remote":                     "gdrive:crypt-data",
			"password":                   "super-secure-password-987",
			"password2":                  "unique-salt-12345",
			"filename_encryption":        "standard",
			"directory_name_encryption":  "true",
			"filename_encoding":          "base64",
			"suffix":                     ".enc",
			"no_data_encryption":         "false",
			"pass_bad_blocks":            "false",
			"strict_names":               "true",
		},
	})
	if err != nil {
		t.Fatalf("CreateRemote() error = %v", err)
	}

	expectedFields := map[string]string{
		"remote":    "must contain colon",
		"password":  "must be non-empty",
		"password2": "optional but provided",
	}

	for field := range expectedFields {
		if _, ok := capturedParams[field]; !ok {
			t.Errorf("expected field %q to be present in parameters", field)
		}
	}
}