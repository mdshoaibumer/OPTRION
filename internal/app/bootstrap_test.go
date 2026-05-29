package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optrion/optrion/internal/platform/config"
	"github.com/optrion/optrion/internal/platform/logger"
	"github.com/optrion/optrion/internal/platform/server"
)

// TestBootstrap_RouterSetup verifies the middleware chain and routing work correctly.
func TestBootstrap_RouterSetup(t *testing.T) {
	cfg := config.LogConfig{Level: "info", Format: "json"}
	log := logger.New(cfg)

	router := server.NewRouter(log)
	router.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		server.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	handler := router.Handler()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Verify middleware headers are set
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header")
	}
	if rec.Header().Get("X-Correlation-ID") == "" {
		t.Error("expected X-Correlation-ID header")
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("expected X-Content-Type-Options: nosniff")
	}
}

// TestBootstrap_MiddlewareChainOrder verifies middleware executes in correct order.
func TestBootstrap_MiddlewareChainOrder(t *testing.T) {
	cfg := config.LogConfig{Level: "error", Format: "json"}
	log := logger.New(cfg)

	router := server.NewRouter(log)
	router.HandleFunc("GET /ctx", func(w http.ResponseWriter, r *http.Request) {
		// Verify context has IDs set by middleware
		reqID := logger.RequestID(r.Context())
		corrID := logger.CorrelationID(r.Context())

		server.WriteJSON(w, http.StatusOK, map[string]string{
			"request_id":     reqID,
			"correlation_id": corrID,
		})
	})

	handler := router.Handler()

	req := httptest.NewRequest(http.MethodGet, "/ctx", nil)
	req.Header.Set("X-Request-ID", "test-req-123")
	req.Header.Set("X-Correlation-ID", "test-corr-456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["request_id"] != "test-req-123" {
		t.Errorf("expected request_id 'test-req-123', got %q", body["request_id"])
	}
	if body["correlation_id"] != "test-corr-456" {
		t.Errorf("expected correlation_id 'test-corr-456', got %q", body["correlation_id"])
	}
}

// TestBootstrap_PanicRecovery verifies panics are caught and return 500.
func TestBootstrap_PanicRecovery(t *testing.T) {
	cfg := config.LogConfig{Level: "error", Format: "json"}
	log := logger.New(cfg)

	router := server.NewRouter(log)
	router.HandleFunc("GET /panic", func(w http.ResponseWriter, r *http.Request) {
		panic("intentional test panic")
	})

	handler := router.Handler()

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["error"] != "internal server error" {
		t.Errorf("expected error 'internal server error', got %q", body["error"])
	}
}

// TestBootstrap_ConfigLoad verifies configuration loads with defaults.
func TestBootstrap_ConfigLoad(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.App.Name != "optrion" {
		t.Errorf("expected app name 'optrion', got %q", cfg.App.Name)
	}
	if cfg.HTTP.Port < 1 {
		t.Errorf("expected valid HTTP port, got %d", cfg.HTTP.Port)
	}
}

// TestBootstrap_LoggerCreation verifies logger initializes without error.
func TestBootstrap_LoggerCreation(t *testing.T) {
	cfg := config.LogConfig{Level: "debug", Format: "json"}
	log := logger.New(cfg)
	if log == nil {
		t.Fatal("expected non-nil logger")
	}

	// Should not panic
	log.InfoContext(context.Background(), "bootstrap test", "phase", "init")
}

// TestBootstrap_CORSPreflight verifies CORS preflight requests work.
func TestBootstrap_CORSPreflight(t *testing.T) {
	cfg := config.LogConfig{Level: "error", Format: "json"}
	log := logger.New(cfg)

	router := server.NewRouter(log)
	router.HandleFunc("POST /api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := router.Handler()

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS Allow-Origin header")
	}
}
