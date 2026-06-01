package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optrion/optrion/internal/platform/server"
)

func TestHealthHandler_Liveness(t *testing.T) {
	handler := server.NewHealthHandler(nil, nil, "1.0.0-test", nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handler.Liveness().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "alive" {
		t.Errorf("expected status 'alive', got %s", resp["status"])
	}
}

func TestHealthHandler_Liveness_ContentType(t *testing.T) {
	handler := server.NewHealthHandler(nil, nil, "1.0.0", nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handler.Liveness().ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}
