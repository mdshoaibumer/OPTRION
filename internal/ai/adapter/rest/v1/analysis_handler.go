package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/optrion/optrion/internal/ai/app/service"
	"github.com/optrion/optrion/internal/platform/server"
)

// AnalysisHandler handles HTTP requests for AI analysis.
type AnalysisHandler struct {
	service *service.RootCauseService
}

// NewAnalysisHandler creates a new handler with the AI service.
func NewAnalysisHandler(svc *service.RootCauseService) *AnalysisHandler {
	return &AnalysisHandler{service: svc}
}

func (h *AnalysisHandler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	incidentIDStr := r.URL.Query().Get("incident_id")
	if incidentIDStr == "" {
		server.WriteError(w, http.StatusBadRequest, "incident_id query parameter is required")
		return
	}

	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid incident_id format")
		return
	}

	analyses, err := h.service.GetAnalysesByIncident(r.Context(), incidentID)
	if err != nil {
		server.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  analyses,
		"count": len(analyses),
	})
}

func (h *AnalysisHandler) GetIncidentAnalysis(w http.ResponseWriter, r *http.Request) {
	incidentIDStr := r.PathValue("id")
	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid incident ID format")
		return
	}

	reports, err := h.service.GetReportsByIncident(r.Context(), incidentID)
	if err != nil {
		server.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if len(reports) == 0 {
		server.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"data":    []interface{}{},
			"count":   0,
			"message": "no analysis available for this incident",
		})
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  reports,
		"count": len(reports),
	})
}

func (h *AnalysisHandler) PostIncidentAnalyze(w http.ResponseWriter, r *http.Request) {
	incidentIDStr := r.PathValue("id")
	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid incident ID format")
		return
	}

	if h.service == nil {
		server.WriteError(w, http.StatusServiceUnavailable, "AI analysis is not configured")
		return
	}

	if err := h.service.Analyze(r.Context(), incidentID); err != nil {
		server.WriteError(w, http.StatusInternalServerError, "analysis failed")
		return
	}

	server.WriteJSON(w, http.StatusAccepted, map[string]string{
		"status":  "analysis_triggered",
		"message": "root cause analysis has been initiated",
	})
}
