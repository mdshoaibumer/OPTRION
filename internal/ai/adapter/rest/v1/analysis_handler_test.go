package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/optrion/optrion/internal/ai/adapter/rest/v1"
)

func TestAnalysisHandler_GetAnalysis_MissingIncidentID(t *testing.T) {
	handler := v1.NewAnalysisHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis", nil)
	w := httptest.NewRecorder()

	handler.GetAnalysis(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAnalysisHandler_GetAnalysis_InvalidUUID(t *testing.T) {
	handler := v1.NewAnalysisHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analysis?incident_id=not-a-uuid", nil)
	w := httptest.NewRecorder()

	handler.GetAnalysis(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAnalysisHandler_PostIncidentAnalyze_InvalidUUID(t *testing.T) {
	handler := v1.NewAnalysisHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/bad-uuid/analyze", nil)
	req.SetPathValue("id", "bad-uuid")
	w := httptest.NewRecorder()

	handler.PostIncidentAnalyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAnalysisHandler_PostIncidentAnalyze_NilService(t *testing.T) {
	handler := v1.NewAnalysisHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/550e8400-e29b-41d4-a716-446655440000/analyze", nil)
	req.SetPathValue("id", "550e8400-e29b-41d4-a716-446655440000")
	w := httptest.NewRecorder()

	handler.PostIncidentAnalyze(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestAnalysisHandler_GetIncidentAnalysis_InvalidUUID(t *testing.T) {
	handler := v1.NewAnalysisHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/bad/analysis", nil)
	req.SetPathValue("id", "bad")
	w := httptest.NewRecorder()

	handler.GetIncidentAnalysis(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
