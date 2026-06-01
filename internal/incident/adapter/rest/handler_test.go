package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/optrion/optrion/internal/incident/adapter/rest"
)

func TestHandler_ListIncidents_MissingTenantID(t *testing.T) {
	handler := rest.NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	w := httptest.NewRecorder()

	handler.ListIncidents(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "tenant_id") {
		t.Error("expected error mentioning tenant_id")
	}
}

func TestHandler_GetIncident_MissingID(t *testing.T) {
	handler := rest.NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handler.GetIncident(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Acknowledge_InvalidBody(t *testing.T) {
	handler := rest.NewHandler(nil, nil)

	body := strings.NewReader(`{invalid`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/abc/acknowledge", body)
	req.SetPathValue("id", "abc")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Acknowledge(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Resolve_InvalidBody(t *testing.T) {
	handler := rest.NewHandler(nil, nil)

	body := strings.NewReader(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/abc/resolve", body)
	req.SetPathValue("id", "abc")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Resolve(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_AddComment_EmptyContent(t *testing.T) {
	handler := rest.NewHandler(nil, nil)

	body := strings.NewReader(`{"tenant_id":"t1","author_id":"a1","content":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/abc/comments", body)
	req.SetPathValue("id", "abc")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.AddComment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_GetStats_MissingTenantID(t *testing.T) {
	handler := rest.NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_GetTenantTimeline_MissingTenantID(t *testing.T) {
	handler := rest.NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/timeline", nil)
	w := httptest.NewRecorder()

	handler.GetTenantTimeline(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
