package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"rcloneflow/internal/service"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	WriteJSON(w, 200, map[string]any{"key": "value"})

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), `"key":"value"`) {
		t.Errorf("expected body to contain key:value, got %s", w.Body.String())
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestWriteJSONWithError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteJSON(w, 500, map[string]any{"error": "something went wrong"})

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestDecodeRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":"test","value":123}`))

	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	err := DecodeRequest(httptest.NewRecorder(), req, &result)
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}

	if result.Name != "test" {
		t.Errorf("expected Name test, got %s", result.Name)
	}

	if result.Value != 123 {
		t.Errorf("expected Value 123, got %d", result.Value)
	}
}

func TestDecodeRequestInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{invalid json}`))

	var result struct {
		Name string `json:"name"`
	}

	err := DecodeRequest(httptest.NewRecorder(), req, &result)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMapServiceError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantMsg  string
	}{
		{"task not found", service.ErrTaskNotFound, 404, "task not found"},
		{"task name exists", service.ErrTaskNameExists, 409, "task name already exists"},
		{"schedule not found", service.ErrScheduleNotFound, 404, "schedule not found"},
		{"run not found", service.ErrRunNotFound, 404, "run not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, code := mapServiceError(tt.err)
			if msg != tt.wantMsg {
				t.Errorf("msg = %q, want %q", msg, tt.wantMsg)
			}
			if code != tt.wantCode {
				t.Errorf("code = %d, want %d", code, tt.wantCode)
			}
		})
	}
}

func TestMapServiceError_Default(t *testing.T) {
	msg, code := mapServiceError(errors.New("unknown error"))
	if msg != "internal error" {
		t.Errorf("msg = %q, want \"internal error\"", msg)
	}
	if code != 500 {
		t.Errorf("code = %d, want 500", code)
	}
}
