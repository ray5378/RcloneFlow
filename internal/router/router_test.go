package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rcloneflow/internal/adapter"
	"rcloneflow/internal/auth"
	"rcloneflow/internal/controller"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
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

func TestRouter_New(t *testing.T) {
	r := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "/static")
	assert.NotNil(t, r)
	assert.Equal(t, "/static", r.staticDir)
}

func setupRouterTest(t *testing.T) *Router {
	t.Helper()
	tmpDir := t.TempDir()
	db, err := store.Open(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	rc := adapter.NewRcloneClient(nil)
	taskSvc := service.NewTaskService(db, nil)
	scheduleSvc := service.NewScheduleService(db)
	runSvc := service.NewRunService(service.NewStoreRunAdapter(db))
	authSvc := service.NewAuthService(db)

	remoteCtrl := controller.NewRemoteController(rc)
	taskCtrl := controller.NewTaskController(taskSvc, scheduleSvc, runSvc, rc)
	browserCtrl := controller.NewBrowserController(rc)
	scheduleCtrl := controller.NewScheduleController(scheduleSvc, nil)
	runCtrl := controller.NewRunController(runSvc, rc)
	fsCtrl := controller.NewFsController(rc)
	authCtrl := controller.NewAuthController(authSvc)
	activeTransferCtrl := controller.NewActiveTransferController(nil, runSvc)
	tagCtrl := controller.NewTagController(service.NewTagService(db))
	versionCtrl := controller.NewVersionController(rc)

	staticDir := filepath.Join(tmpDir, "web")
	os.MkdirAll(staticDir, 0755)
	os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("RcloneFlow UI"), 0644)

	return New(remoteCtrl, taskCtrl, browserCtrl, scheduleCtrl, runCtrl, fsCtrl, authCtrl, activeTransferCtrl, tagCtrl, versionCtrl, staticDir)
}

func TestRouter_Setup_Healthz(t *testing.T) {
	r := setupRouterTest(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, true, body["ok"])
}

func TestRouter_Setup_HasUsers(t *testing.T) {
	r := setupRouterTest(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/auth/has-users")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, false, body["exists"])
}

func TestRouter_Setup_StaticRoot(t *testing.T) {
	r := setupRouterTest(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	buf := make([]byte, 256)
	n, _ := resp.Body.Read(buf)
	assert.Contains(t, string(buf[:n]), "RcloneFlow UI")
}

func TestRouter_Setup_BisyncEndpoint_Returns404(t *testing.T) {
	r := setupRouterTest(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	pair, err := auth.GenerateTokenPair(1, "admin")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/tasks/999/bisync/lst-files", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRouter_Setup_BisyncResolveConflict(t *testing.T) {
	r := setupRouterTest(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	pair, err := auth.GenerateTokenPair(1, "admin")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/tasks/1/bisync/resolve-conflict", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}
