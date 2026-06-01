package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/optrion/optrion/internal/recommendation/adapter/rest/v1"
)

func TestRecommendationHandler_GetRecommendations_MissingIncidentID(t *testing.T) {
	handler := v1.NewRecommendationHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations", nil)
	w := httptest.NewRecorder()

	handler.GetRecommendations(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRecommendationHandler_GetRecommendations_InvalidUUID(t *testing.T) {
	handler := v1.NewRecommendationHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations?incident_id=not-a-uuid", nil)
	w := httptest.NewRecorder()

	handler.GetRecommendations(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRecommendationHandler_GetIncidentRecommendations_InvalidUUID(t *testing.T) {
	handler := v1.NewRecommendationHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/xyz/recommendations", nil)
	req.SetPathValue("id", "xyz")
	w := httptest.NewRecorder()

	handler.GetIncidentRecommendations(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRecommendationHandler_PostIncidentRecommend_InvalidUUID(t *testing.T) {
	handler := v1.NewRecommendationHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/invalid/recommend", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	handler.PostIncidentRecommend(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRecommendationHandler_PostIncidentRecommend_NilService(t *testing.T) {
	handler := v1.NewRecommendationHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/550e8400-e29b-41d4-a716-446655440000/recommend", nil)
	req.SetPathValue("id", "550e8400-e29b-41d4-a716-446655440000")
	w := httptest.NewRecorder()

	handler.PostIncidentRecommend(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
