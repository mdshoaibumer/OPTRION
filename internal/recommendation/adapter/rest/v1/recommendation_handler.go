package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/optrion/optrion/internal/platform/server"
	"github.com/optrion/optrion/internal/recommendation/app/service"
)

// RecommendationHandler handles HTTP requests for recommendations.
type RecommendationHandler struct {
	service *service.RecommendationService
}

// NewRecommendationHandler creates a new recommendation handler.
func NewRecommendationHandler(svc *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{service: svc}
}

func (h *RecommendationHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
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

	recs, err := h.service.GetRecommendationsByIncident(r.Context(), incidentID)
	if err != nil {
		server.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  recs,
		"count": len(recs),
	})
}

func (h *RecommendationHandler) GetIncidentRecommendations(w http.ResponseWriter, r *http.Request) {
	incidentIDStr := r.PathValue("id")
	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid incident ID format")
		return
	}

	recs, err := h.service.GetRecommendationsByIncident(r.Context(), incidentID)
	if err != nil {
		server.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  recs,
		"count": len(recs),
	})
}

func (h *RecommendationHandler) PostIncidentRecommend(w http.ResponseWriter, r *http.Request) {
	incidentIDStr := r.PathValue("id")
	incidentID, err := uuid.Parse(incidentIDStr)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid incident ID format")
		return
	}

	if h.service == nil {
		server.WriteError(w, http.StatusServiceUnavailable, "recommendation service is not configured")
		return
	}

	if err := h.service.Recommend(r.Context(), incidentID); err != nil {
		server.WriteError(w, http.StatusInternalServerError, "recommendation generation failed")
		return
	}

	server.WriteJSON(w, http.StatusAccepted, map[string]string{
		"status":  "recommendation_triggered",
		"message": "recommendations are being generated",
	})
}
