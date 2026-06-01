package v1_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v1 "github.com/optrion/optrion/internal/registration/adapter/rest/v1"
)

func TestRegistrationHandler_Register_InvalidMethod(t *testing.T) {
	handler := v1.NewRegistrationHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/register", nil)
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestRegistrationHandler_Register_InvalidJSON(t *testing.T) {
	handler := v1.NewRegistrationHandler(nil)

	body := strings.NewReader(`{invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRegistrationHandler_Register_EmptyBody(t *testing.T) {
	handler := v1.NewRegistrationHandler(nil)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	// Should fail with 500 because service is nil, or with validation error
	// Either 400 or 500 is acceptable since we can't call nil service
	if w.Code == http.StatusCreated {
		t.Fatal("expected error response, got 201")
	}
}
