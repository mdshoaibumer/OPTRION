package rest

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/optrion/optrion/internal/health/app"
	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
	"github.com/optrion/optrion/internal/platform/server"
)

// Handler handles health monitoring HTTP requests.
type Handler struct {
	service *app.HealthService
	logger  *slog.Logger
}

// NewHandler creates a new health monitoring REST handler.
func NewHandler(service *app.HealthService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all health monitoring routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/health/summary", h.GetSummary)
	mux.HandleFunc("GET /api/v1/health/components", h.GetComponents)
	mux.HandleFunc("GET /api/v1/health/history", h.GetHistory)
	mux.HandleFunc("GET /api/v1/health/anomalies", h.GetAnomalies)
}

// GetSummary handles GET /api/v1/health/summary?tenant_id=...
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "tenant_id query parameter is required"})
		return
	}

	summary, err := h.service.GetSummary(r.Context(), tenantID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get health summary", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	server.WriteJSON(w, http.StatusOK, toSummaryResponse(summary))
}

// GetComponents handles GET /api/v1/health/components?tenant_id=...
func (h *Handler) GetComponents(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "tenant_id query parameter is required"})
		return
	}

	statuses, err := h.service.GetComponentStatuses(r.Context(), tenantID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get component statuses", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	resp := make([]ComponentHealthResponse, 0, len(statuses))
	for _, s := range statuses {
		resp = append(resp, toComponentHealthResponse(s))
	}

	server.WriteJSON(w, http.StatusOK, ListResponse{Data: resp, Count: len(resp)})
}

// GetHistory handles GET /api/v1/health/history?tenant_id=...&from=...&to=...&limit=...
func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "tenant_id query parameter is required"})
		return
	}

	from := parseTime(r.URL.Query().Get("from"), time.Now().Add(-24*time.Hour))
	to := parseTime(r.URL.Query().Get("to"), time.Now())
	limit := parseIntQuery(r, "limit", 100)

	scores, err := h.service.GetHistory(r.Context(), tenantID, from, to, limit)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get health history", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	resp := make([]HealthScoreResponse, 0, len(scores))
	for _, s := range scores {
		resp = append(resp, toHealthScoreResponse(s))
	}

	server.WriteJSON(w, http.StatusOK, ListResponse{Data: resp, Count: len(resp)})
}

// GetAnomalies handles GET /api/v1/health/anomalies?tenant_id=...
func (h *Handler) GetAnomalies(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "tenant_id query parameter is required"})
		return
	}

	filter := port.AnomalyFilter{
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
	}

	if componentID := r.URL.Query().Get("component_id"); componentID != "" {
		filter.ComponentID = &componentID
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		s := domain.Severity(severity)
		filter.Severity = &s
	}
	if resolved := r.URL.Query().Get("resolved"); resolved != "" {
		b := resolved == "true"
		filter.Resolved = &b
	}

	anomalies, err := h.service.GetAnomalies(r.Context(), tenantID, filter)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get anomalies", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	resp := make([]AnomalyResponse, 0, len(anomalies))
	for _, a := range anomalies {
		resp = append(resp, toAnomalyResponse(a))
	}

	server.WriteJSON(w, http.StatusOK, ListResponse{Data: resp, Count: len(resp)})
}

// --- Helpers ---

func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

func parseTime(s string, defaultVal time.Time) time.Time {
	if s == "" {
		return defaultVal
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return defaultVal
	}
	return t
}
