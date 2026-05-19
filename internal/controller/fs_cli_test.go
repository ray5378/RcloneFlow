package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }

func TestFsController_Wrap_MethodNotAllowed(t *testing.T) {
	c := NewFsController(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/fs/mkdir", nil)
	rec := httptest.NewRecorder()
	c.HandleMkdir(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestFsController_AllHandlers_MethodCheck(t *testing.T) {
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
			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name    string
		fs      string
		remote  string
		wantFs  string
		wantRem string
	}{
		{"basic", "myRemote", "/path/to/file", "myRemote:", "path/to/file"},
		{"already colon", "myRemote:", "/path", "myRemote:", "path"},
		{"no leading slash", "myRemote", "path/to/file", "myRemote:", "path/to/file"},
		{"empty fs", "", "/path", "", "path"},
		{"whitespace", "  myRemote  ", "  /path  ", "myRemote:", "path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, rem := normalize(tt.fs, tt.remote)
			assert.Equal(t, tt.wantFs, fs)
			assert.Equal(t, tt.wantRem, rem)
		})
	}
}

func TestSplitFsRemote(t *testing.T) {
	tests := []struct {
		name    string
		fs      string
		remote  string
		wantFs  string
		wantRem string
	}{
		{"with remote", "myRemote:", "/path", "myRemote:", "path"},
		{"fs with path", "myRemote:/some/path", "", "myRemote:", "some/path"},
		{"both empty", "", "", "", ""},
		{"remote takes precedence", "myRemote:", "explicit/path", "myRemote:", "explicit/path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, rem := splitFsRemote(tt.fs, tt.remote)
			assert.Equal(t, tt.wantFs, fs)
			assert.Equal(t, tt.wantRem, rem)
		})
	}
}

func TestTryStripFirstSegment(t *testing.T) {
	assert.Equal(t, "bar/baz", tryStripFirstSegment("foo/bar/baz"))
	assert.Equal(t, "baz", tryStripFirstSegment("foo/baz"))
	assert.Equal(t, "single", tryStripFirstSegment("single"))
	assert.Equal(t, "", tryStripFirstSegment(""))
}

func TestParentDir(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/path/to/file", "/path/to"},
		{"/path/to/dir/", "/path/to"},
		{"/file", "/"},
		{"", ""},
		{"/", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, parentDir(tt.input))
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal/path", "normal/path"},
		{"path/with:colon:", "path/with:colon"},
		{"a:/b:/c:", "a/b/c"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizePath(tt.input))
		})
	}
}

func TestIsWebdavMoveError(t *testing.T) {
	assert.True(t, isWebdavMoveError(&testErr{"DirMove: not supported"}))
	assert.True(t, isWebdavMoveError(&testErr{"move call failed"}))
	assert.True(t, isWebdavMoveError(&testErr{"500 Internal Server Error"}))
	assert.False(t, isWebdavMoveError(&testErr{"connection refused"}))
	assert.False(t, isWebdavMoveError(&testErr{"network name not found"}))
}

func TestWriteJSON_FsCli(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, 200, map[string]any{"ok": true})
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), `"ok":true`)
}

func TestRcConfigPath(t *testing.T) {
	t.Setenv("APP_DATA_DIR", "/tmp/testdata")
	p := rcConfigPath()
	assert.Contains(t, p, "rclone.conf")
	assert.Contains(t, p, "/tmp/testdata")
}

func TestRcConfigPath_Default(t *testing.T) {
	t.Setenv("APP_DATA_DIR", "")
	p := rcConfigPath()
	assert.Contains(t, p, "rclone.conf")
}

func TestRemoteNameMap_NoConfig(t *testing.T) {
	t.Setenv("APP_DATA_DIR", "/nonexistent/path/for/test")
	m := remoteNameMap()
	assert.Empty(t, m)
}

func TestCanonicalRemoteName_NoConfig(t *testing.T) {
	t.Setenv("APP_DATA_DIR", "/nonexistent/path/for/test")
	assert.Equal(t, "myRemote", canonicalRemoteName("myRemote"))
	assert.Equal(t, "", canonicalRemoteName(""))
}
