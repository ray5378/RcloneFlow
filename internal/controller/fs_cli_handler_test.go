package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFsController_HandleMkdir(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"fs":"local","remote":"/test/dir"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/mkdir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleMkdir(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleMkdir() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandleDeleteFile(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"fs":"local","remote":"/test/file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/delete", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleDeleteFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleDeleteFile() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandleDeleteFile_NotFound(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", &mockError{msg: "not found"}
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"fs":"local","remote":"/missing.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/delete", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleDeleteFile(rec, req)

	// 404-ish errors should be treated as success
	if rec.Code != http.StatusOK {
		t.Fatalf("HandleDeleteFile() code = %d, want %d (not found should be success)", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandlePurge(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"fs":"local","remote":"/test/dir"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/purge", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandlePurge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandlePurge() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandleMove(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"srcFs":"local","srcRemote":"/src/file.txt","dstFs":"local","dstRemote":"/dst/file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/move", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleMove(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleMove() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandleCopy(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"srcFs":"local","srcRemote":"/src/file.txt","dstFs":"local","dstRemote":"/dst/file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/copy", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleCopy(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleCopy() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandleCopyDir(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"srcFs":"local","srcRemote":"/src/dir","dstFs":"local","dstRemote":"/dst/dir"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/copydir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleCopyDir(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleCopyDir() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandleMoveDir(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"srcFs":"local","srcRemote":"/src/dir","dstFs":"local","dstRemote":"/dst/dir"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/movedir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleMoveDir(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleMoveDir() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandlePublicLink(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "http://example.com/link", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"fs":"local","remote":"/test/file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/publiclink", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandlePublicLink(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandlePublicLink() code = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if url, ok := resp["url"].(string); !ok || url != "http://example.com/link" {
		t.Fatalf("HandlePublicLink() url = %v, want http://example.com/link", resp["url"])
	}
}

func TestFsController_HandleMkdir_Error(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "error output", &mockError{msg: "permission denied"}
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"fs":"local","remote":"/protected/dir"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/mkdir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleMkdir(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("HandleMkdir() code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestFsController_HandleCopy_SMBFallback(t *testing.T) {
	callCount := 0
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		callCount++
		// Simulate: mkdir succeeds, copyto fails with SMB error, fallback mkdir + copyto succeed
		if callCount == 1 {
			return "", nil // mkdir
		}
		if callCount == 2 {
			return "", &mockError{msg: "network name not found"} // copyto fails
		}
		return "", nil // fallback operations succeed
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"srcFs":"local","srcRemote":"share/dir/file.txt","dstFs":"local","dstRemote":"dst/file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/copy", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleCopy(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleCopy() code = %d, want %d (callCount=%d)", rec.Code, http.StatusOK, callCount)
	}
}

func TestFsController_HandleMove_WebDAVFallback(t *testing.T) {
	callCount := 0
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			return "", &mockError{msg: "DirMove: not supported"}
		}
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"srcFs":"local","srcRemote":"/src/file.txt","dstFs":"local","dstRemote":"/dst/file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/move", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleMove(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleMove() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandleCopyDir_SMBFallback(t *testing.T) {
	callCount := 0
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			return "", &mockError{msg: "create filesystem error"}
		}
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"srcFs":"local","srcRemote":"share/srcdir","dstFs":"local","dstRemote":"dstdir"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/copydir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleCopyDir(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleCopyDir() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandleMoveDir_WebDAVFallback(t *testing.T) {
	callCount := 0
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			return "", &mockError{msg: "move call failed"}
		}
		return "", nil
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"srcFs":"local","srcRemote":"/srcdir","dstFs":"local","dstRemote":"/dstdir"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/movedir", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandleMoveDir(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HandleMoveDir() code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestFsController_HandlePublicLink_Error(t *testing.T) {
	old := runRclone
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", &mockError{msg: "link failed"}
	}
	defer func() { runRclone = old }()

	c := NewFsController(nil)
	body := `{"fs":"local","remote":"/test/file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/api/fs/publiclink", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c.HandlePublicLink(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("HandlePublicLink() code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}


func TestFsController_Wrap_MethodNotAllowed_All(t *testing.T) {
	c := NewFsController(nil)
	handlers := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
	}{
		{"mkdir", c.HandleMkdir},
		{"delete", c.HandleDeleteFile},
		{"purge", c.HandlePurge},
		{"move", c.HandleMove},
		{"copy", c.HandleCopy},
		{"copyDir", c.HandleCopyDir},
		{"moveDir", c.HandleMoveDir},
		{"publicLink", c.HandlePublicLink},
	}
	for _, h := range handlers {
		t.Run(h.name+"/GET", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/fs/"+h.name, nil)
			rec := httptest.NewRecorder()
			h.fn(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s GET code = %d, want %d", h.name, rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

type mockError struct{ msg string }

func (e *mockError) Error() string { return e.msg }
