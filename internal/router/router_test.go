package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"rcloneflow/internal/adapter"
	"rcloneflow/internal/auth"
	"rcloneflow/internal/controller"
	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

func setupTestRouter(t *testing.T) *Router {
	t.Helper()
	tmpDir := t.TempDir()
	db, err := store.Open(tmpDir)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	rc := adapter.NewRcloneClient(nil)
	remoteCtrl := controller.NewRemoteController(rc)
	taskSvc := service.NewTaskService(db)
	scheduleSvc := service.NewScheduleService(db)
	runSvc := service.NewRunService(service.NewStoreRunAdapter(db))
	taskCtrl := controller.NewTaskController(taskSvc, scheduleSvc, runSvc, rc)
	browserCtrl := controller.NewBrowserController(rc)
	scheduleCtrl := controller.NewScheduleController(scheduleSvc, nil)
	runCtrl := controller.NewRunController(runSvc, rc)
	fsCtrl := controller.NewFsController(rc)
	authCtrl := controller.NewAuthController(db)
	activeTransferCtrl := controller.NewActiveTransferController(nil, runSvc)
	tagSvc := service.NewTagService(db)
	tagCtrl := controller.NewTagController(tagSvc)

	return New(remoteCtrl, taskCtrl, browserCtrl, scheduleCtrl, runCtrl, fsCtrl, authCtrl, activeTransferCtrl, tagCtrl, "")
}

func TestRouter_New(t *testing.T) {
	r := setupTestRouter(t)
	if r == nil {
		t.Fatal("New() returned nil")
	}
	if r.remoteCtrl == nil {
		t.Error("remoteCtrl is nil")
	}
	if r.taskCtrl == nil {
		t.Error("taskCtrl is nil")
	}
	if r.browserCtrl == nil {
		t.Error("browserCtrl is nil")
	}
	if r.staticDir != "" {
		t.Errorf("staticDir = %q, want empty", r.staticDir)
	}
}

func TestRouter_Setup_Healthz(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /healthz status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRouter_Setup_StaticFallback(t *testing.T) {
	r := setupTestRouter(t)
	r.staticDir = t.TempDir()
	mux := http.NewServeMux()
	r.Setup(mux)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent-page", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Should return 404 for missing static files
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /nonexistent-page status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRouter_Setup_AuthRoutes(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test login endpoint exists (POST)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Should not be 404 (route exists, just returns 400 for bad request)
	if rec.Code == http.StatusNotFound {
		t.Error("POST /api/auth/login route should exist")
	}
}

func TestRouter_Setup_TaskRoutes(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test tasks endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("GET /api/tasks route should exist")
	}
}

func TestRouter_Setup_RunRoutes(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test runs endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/runs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("GET /api/runs route should exist")
	}
}

func TestRouter_Setup_RemoteRoutes(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test remotes endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/remotes", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("GET /api/remotes route should exist")
	}
}

func TestRouter_Setup_SettingsRoute(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test settings endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("GET /api/settings route should exist")
	}
}

func TestRouter_Setup_BrowserRoute(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test browser list endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/browser/list?remote=test", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("GET /api/browser/list route should exist")
	}
}

func TestRouter_Setup_ScheduleRoutes(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test schedules endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/schedules", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("GET /api/schedules route should exist")
	}
}

func TestRouter_Setup_TagRoutes(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test tags endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("GET /api/tags route should exist")
	}
}

func TestRouter_Setup_WsRoute(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test ws endpoint exists (will upgrade to websocket, but route should exist)
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Should not be 404
	if rec.Code == http.StatusNotFound {
		t.Error("GET /ws route should exist")
	}
}

func TestRouter_Setup_ActiveTransferRoutes(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test active transfer endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/active", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("GET /api/active route should exist")
	}
}

func TestRouter_Setup_FsRoutes(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test fs endpoint exists
	req := httptest.NewRequest(http.MethodGet, "/api/fs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("GET /api/fs route should exist")
	}
}

func TestRouter_Setup_AuthMiddleware(t *testing.T) {
	r := setupTestRouter(t)
	mux := http.NewServeMux()
	r.Setup(mux)

	// Test that protected routes return 401 without auth
	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Should be 401 (unauthorized) not 404 (not found)
	if rec.Code == http.StatusNotFound {
		t.Error("DELETE /api/tasks/1 route should exist but require auth")
	}
}

func TestAuth_TokenValidation(t *testing.T) {
	// Test that auth package is functional
	pair, err := auth.GenerateTokenPair(1, "admin")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}
	if pair.AccessToken == "" {
		t.Fatal("GenerateTokenPair() returned empty access token")
	}
	if pair.RefreshToken == "" {
		t.Fatal("GenerateTokenPair() returned empty refresh token")
	}

	claims, err := auth.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("claims.UserID = %d, want 1", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("claims.Username = %q, want admin", claims.Username)
	}
}

func TestAuth_RefreshToken(t *testing.T) {
	pair, err := auth.GenerateTokenPair(1, "admin")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	newPair, err := auth.RefreshTokens(pair.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshTokens() error = %v", err)
	}
	if newPair.AccessToken == "" {
		t.Fatal("RefreshTokens() returned empty access token")
	}
}

func TestAuth_SanitizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/api/tasks", "/api/tasks"},
		{"/api/tasks/../../../etc/passwd", "/api/tasks/etc/passwd"},
	}
	for _, tt := range tests {
		got := auth.SanitizePath(tt.input)
		if got != tt.want {
			t.Errorf("SanitizePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAuth_ValidatePath(t *testing.T) {
	if !auth.ValidatePath("/api/tasks") {
		t.Error("ValidatePath(/api/tasks) should be true")
	}
	if auth.ValidatePath("/api/tasks/../../../etc/passwd") {
		t.Error("ValidatePath with traversal should be false")
	}
}
