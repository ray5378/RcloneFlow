package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsPathTraversal_DotDot(t *testing.T) {
	assert.True(t, containsPathTraversal("../etc/passwd"))
	assert.True(t, containsPathTraversal("foo/../../../bar"))
	assert.True(t, containsPathTraversal(".."))
	assert.True(t, containsPathTraversal("foo/.."))
}

func TestContainsPathTraversal_Backslash(t *testing.T) {
	assert.True(t, containsPathTraversal("..\\etc\\passwd"))
	assert.True(t, containsPathTraversal("foo\\..\\bar"))
}

func TestContainsPathTraversal_Encoded(t *testing.T) {
	assert.True(t, containsPathTraversal("%2e%2e/etc/passwd"))
	assert.True(t, containsPathTraversal("%2e./etc/passwd"))
}

func TestContainsPathTraversal_NullByte(t *testing.T) {
	assert.True(t, containsPathTraversal("file\x00.txt"))
	assert.True(t, containsPathTraversal("path\x00/../../etc"))
}

func TestContainsPathTraversal_SafePaths(t *testing.T) {
	assert.False(t, containsPathTraversal("/"))
	assert.False(t, containsPathTraversal("/my/remote/path"))
	assert.False(t, containsPathTraversal("bucket/folder/file.txt"))
	assert.False(t, containsPathTraversal(".hidden"))
	assert.False(t, containsPathTraversal("."))
	assert.False(t, containsPathTraversal("./"))
	assert.False(t, containsPathTraversal("/path/.gitignore"))
}

func TestContainsPathTraversal_TrailingDot(t *testing.T) {
	// The implementation only checks trailing dot for paths starting with "."
	// "/path/." does NOT trigger traversal detection in the current implementation
	assert.False(t, containsPathTraversal("/path/."))
	assert.False(t, containsPathTraversal("path/to/."))
	// But dot-dot-slash at end does
	assert.True(t, containsPathTraversal("/path/.."))
}

func TestValidatePath_Safe(t *testing.T) {
	assert.True(t, ValidatePath("/safe/path"))
	assert.True(t, ValidatePath("bucket/file.txt"))
}

func TestValidatePath_Unsafe(t *testing.T) {
	assert.False(t, ValidatePath("../etc/passwd"))
	assert.False(t, ValidatePath("%2e%2e/secret"))
}

func TestSanitizePath_RemovesNullBytes(t *testing.T) {
	result := SanitizePath("file\x00.txt")
	assert.NotContains(t, result, "\x00")
}

func TestSanitizePath_RemovesDotDotBackslash(t *testing.T) {
	result := SanitizePath("foo\\..\\bar")
	assert.NotContains(t, result, "\\..")
}

func TestSanitizePath_RemovesDotDotSlash(t *testing.T) {
	result := SanitizePath("foo/../bar")
	assert.NotContains(t, result, "/..")
}

func TestSanitizePath_SafePath(t *testing.T) {
	input := "/my/remote/path/file.txt"
	result := SanitizePath(input)
	assert.Equal(t, input, result)
}

func TestPathSecurityMiddleware_NoPathParameter(t *testing.T) {
	handler := PathSecurityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/fs/list", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPathSecurityMiddleware_SafePath(t *testing.T) {
	handler := PathSecurityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/fs/list?path=/safe/path", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPathSecurityMiddleware_TraversalPath(t *testing.T) {
	handler := PathSecurityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/fs/list?path=../etc/passwd", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "路径包含非法字符")
}

func TestPathSecurityMiddleware_EncodedTraversal(t *testing.T) {
	handler := PathSecurityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/fs/list?path=%2e%2e/secret", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
