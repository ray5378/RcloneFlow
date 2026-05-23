package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticFileHandler_DirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("test content"), 0644)

	handler := staticFileHandler(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/test.txt", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "test content")
}

func TestStaticFileHandler_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	handler := staticFileHandler(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent.txt", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestStaticFileHandler_DirectoryNotExists(t *testing.T) {
	handler := staticFileHandler("/nonexistent/path/12345")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "RcloneFlow")
	assert.Contains(t, rec.Body.String(), "前端构建产物缺失")
	assert.Contains(t, rec.Body.String(), "/nonexistent/path/12345")
	assert.Contains(t, rec.Body.String(), "npm run build")
}

func TestStaticFileHandler_IndexFile(t *testing.T) {
	tmpDir := t.TempDir()
	indexFile := filepath.Join(tmpDir, "index.html")
	os.WriteFile(indexFile, []byte("Index Content"), 0644)

	handler := staticFileHandler(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Index Content")
}

func TestRouter_New(t *testing.T) {
	r := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, "/static")
	assert.NotNil(t, r)
	assert.Equal(t, "/static", r.staticDir)
}

func TestStaticFileHandler_NestedPath(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "assets", "js")
	os.MkdirAll(nestedDir, 0755)
	jsFile := filepath.Join(nestedDir, "app.js")
	os.WriteFile(jsFile, []byte("console.log('test');"), 0644)

	handler := staticFileHandler(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/assets/js/app.js", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "console.log")
}

func TestStaticFileHandler_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	os.WriteFile(emptyFile, []byte(""), 0644)

	handler := staticFileHandler(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/empty.txt", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Body.String())
}

func TestStaticFileHandler_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	largeFile := filepath.Join(tmpDir, "large.bin")
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}
	os.WriteFile(largeFile, content, 0644)

	handler := staticFileHandler(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/large.bin", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int64(len(content)), int64(rec.Body.Len()))
}
