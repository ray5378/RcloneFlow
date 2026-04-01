package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	
	err := DecodeRequest(req, &result)
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
	
	err := DecodeRequest(req, &result)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
