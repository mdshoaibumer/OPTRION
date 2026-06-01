package rest

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/optrion/optrion/internal/incident/app"
	"github.com/optrion/optrion/internal/incident/domain"
	"github.com/optrion/optrion/internal/incident/port"
	"github.com/optrion/optrion/internal/platform/server"
)

// Handler handles incident management HTTP requests.
type Handler struct {
	service *app.IncidentService
	logger  *slog.Logger
}

// NewHandler creates a new incident REST handler.
func NewHandler(service *app.IncidentService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all incident management routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/incidents", h.ListIncidents)
	mux.HandleFunc("GET /api/v1/incidents/{id}", h.GetIncident)
	mux.HandleFunc("POST /api/v1/incidents/{id}/acknowledge", h.Acknowledge)
	mux.HandleFunc("POST /api/v1/incidents/{id}/investigate", h.Investigate)
	mux.HandleFunc("POST /api/v1/incidents/{id}/resolve", h.Resolve)
	mux.HandleFunc("POST /api/v1/incidents/{id}/close", h.CloseIncident)
	mux.HandleFunc("POST /api/v1/incidents/{id}/comments", h.AddComment)
	mux.HandleFunc("GET /api/v1/incidents/{id}/timeline", h.GetTimeline)
	mux.HandleFunc("GET /api/v1/incidents/timeline", h.GetTenantTimeline)
	mux.HandleFunc("GET /api/v1/incidents/stats", h.GetStats)
}

// RegisterAuthenticatedRoutes registers incident routes wrapped with authentication middleware.
func (h *Handler) RegisterAuthenticatedRoutes(mux *http.ServeMux, authWrap func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/incidents", authWrap(http.HandlerFunc(h.ListIncidents)))
	mux.Handle("GET /api/v1/incidents/{id}", authWrap(http.HandlerFunc(h.GetIncident)))
	mux.Handle("POST /api/v1/incidents/{id}/acknowledge", authWrap(http.HandlerFunc(h.Acknowledge)))
	mux.Handle("POST /api/v1/incidents/{id}/investigate", authWrap(http.HandlerFunc(h.Investigate)))
	mux.Handle("POST /api/v1/incidents/{id}/resolve", authWrap(http.HandlerFunc(h.Resolve)))
	mux.Handle("POST /api/v1/incidents/{id}/close", authWrap(http.HandlerFunc(h.CloseIncident)))
	mux.Handle("POST /api/v1/incidents/{id}/comments", authWrap(http.HandlerFunc(h.AddComment)))
	mux.Handle("GET /api/v1/incidents/{id}/timeline", authWrap(http.HandlerFunc(h.GetTimeline)))
	mux.Handle("GET /api/v1/incidents/timeline", authWrap(http.HandlerFunc(h.GetTenantTimeline)))
	mux.Handle("GET /api/v1/incidents/stats", authWrap(http.HandlerFunc(h.GetStats)))
}

// ListIncidents handles GET /api/v1/incidents?tenant_id=...
func (h *Handler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "tenant_id query parameter is required"})
		return
	}

	filter := port.IncidentFilter{
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.IncidentStatus(status)
		filter.Status = &s
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		s := domain.IncidentSeverity(severity)
		filter.Severity = &s
	}
	if componentID := r.URL.Query().Get("component_id"); componentID != "" {
		filter.ComponentID = &componentID
	}
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.From = &t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.To = &t
		}
	}

	incidents, total, err := h.service.ListIncidents(r.Context(), tenantID, filter)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to list incidents", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	resp := make([]IncidentResponse, 0, len(incidents))
	for _, inc := range incidents {
		resp = append(resp, toIncidentResponse(inc))
	}

	server.WriteJSON(w, http.StatusOK, ListResponse{Data: resp, Count: len(resp), Total: total})
}

// GetIncident handles GET /api/v1/incidents/{id}
func (h *Handler) GetIncident(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("id")
	if incidentID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "incident ID is required"})
		return
	}

	incident, err := h.service.GetIncident(r.Context(), incidentID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get incident", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	if incident == nil {
		server.WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: "incident not found"})
		return
	}

	server.WriteJSON(w, http.StatusOK, toIncidentResponse(incident))
}

// Acknowledge handles POST /api/v1/incidents/{id}/acknowledge
func (h *Handler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("id")
	var req ActionRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.service.Acknowledge(r.Context(), incidentID, req.ActorID); err != nil {
		h.handleCommandError(w, r, "acknowledge", err)
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

// Investigate handles POST /api/v1/incidents/{id}/investigate
func (h *Handler) Investigate(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("id")
	var req ActionRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.service.Investigate(r.Context(), incidentID, req.ActorID); err != nil {
		h.handleCommandError(w, r, "investigate", err)
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]string{"status": "investigating"})
}

// Resolve handles POST /api/v1/incidents/{id}/resolve
func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("id")
	var req ResolveRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.service.Resolve(r.Context(), incidentID, req.ActorID, req.Resolution); err != nil {
		h.handleCommandError(w, r, "resolve", err)
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

// CloseIncident handles POST /api/v1/incidents/{id}/close
func (h *Handler) CloseIncident(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("id")
	var req CloseRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.service.Close(r.Context(), incidentID, req.ActorID, req.Reason); err != nil {
		h.handleCommandError(w, r, "close", err)
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

// AddComment handles POST /api/v1/incidents/{id}/comments
func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("id")
	var req CommentRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.Content == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "content is required"})
		return
	}

	comment, err := h.service.AddComment(r.Context(), incidentID, req.TenantID, req.AuthorID, req.Content)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to add comment", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	server.WriteJSON(w, http.StatusCreated, toCommentResponse(comment))
}

// GetTimeline handles GET /api/v1/incidents/{id}/timeline
func (h *Handler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("id")
	limit := parseIntQuery(r, "limit", 100)

	entries, err := h.service.GetTimeline(r.Context(), incidentID, limit)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get timeline", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	resp := make([]TimelineResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, toTimelineResponse(e))
	}

	server.WriteJSON(w, http.StatusOK, ListResponse{Data: resp, Count: len(resp), Total: len(resp)})
}

// GetTenantTimeline handles GET /api/v1/incidents/timeline?tenant_id=...
func (h *Handler) GetTenantTimeline(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "tenant_id query parameter is required"})
		return
	}

	from := parseTime(r.URL.Query().Get("from"), time.Now().Add(-24*time.Hour))
	to := parseTime(r.URL.Query().Get("to"), time.Now())
	limit := parseIntQuery(r, "limit", 100)

	entries, err := h.service.GetTenantTimeline(r.Context(), tenantID, from, to, limit)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get tenant timeline", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	resp := make([]TimelineResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, toTimelineResponse(e))
	}

	server.WriteJSON(w, http.StatusOK, ListResponse{Data: resp, Count: len(resp), Total: len(resp)})
}

// GetStats handles GET /api/v1/incidents/stats?tenant_id=...
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "tenant_id query parameter is required"})
		return
	}

	stats, err := h.service.GetStats(r.Context(), tenantID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get stats", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	server.WriteJSON(w, http.StatusOK, toStatsResponse(stats))
}

// handleCommandError maps domain errors to HTTP responses.
func (h *Handler) handleCommandError(w http.ResponseWriter, r *http.Request, action string, err error) {
	if _, ok := err.(domain.ErrInvalidTransition); ok {
		server.WriteJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
		return
	}
	h.logger.ErrorContext(r.Context(), "failed to "+action+" incident", "error", err)
	server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
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
