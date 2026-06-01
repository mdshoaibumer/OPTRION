package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/optrion/optrion/internal/alert/domain/alert"
	"github.com/optrion/optrion/internal/alert/port/repository"
	"github.com/optrion/optrion/internal/platform/server"
)

// AlertHandler handles HTTP requests for alerts.
type AlertHandler struct {
	alerts repository.AlertRepository
	logger *slog.Logger
}

// NewAlertHandler creates a new alert handler with its dependencies.
func NewAlertHandler(alerts repository.AlertRepository, logger *slog.Logger) *AlertHandler {
	return &AlertHandler{
		alerts: alerts,
		logger: logger,
	}
}

type alertResponse struct {
	ID         string `json:"id"`
	TenantID   string `json:"tenant_id"`
	RuleID     string `json:"rule_id"`
	IncidentID string `json:"incident_id"`
	Severity   string `json:"severity"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func toAlertResponse(a *alert.Alert) alertResponse {
	return alertResponse{
		ID:         a.ID,
		TenantID:   a.TenantID,
		RuleID:     a.RuleID,
		IncidentID: a.IncidentID,
		Severity:   a.Severity,
		Status:     string(a.Status),
		Message:    a.Message,
		CreatedAt:  a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  a.UpdatedAt.Format(time.RFC3339),
	}
}

// GetAlerts handles GET /api/v1/alerts
func (h *AlertHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	alerts, err := h.alerts.ListByTenant(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to list alerts", "error", err, "tenant_id", tenantID)
		server.WriteError(w, http.StatusInternalServerError, "failed to retrieve alerts")
		return
	}

	resp := make([]alertResponse, 0, len(alerts))
	for _, a := range alerts {
		resp = append(resp, toAlertResponse(a))
	}

	server.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  resp,
		"total": len(resp),
	})
}

// GetAlertByID handles GET /api/v1/alerts/{id}
func (h *AlertHandler) GetAlertByID(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("id")
	if alertID == "" {
		server.WriteError(w, http.StatusBadRequest, "missing alert id")
		return
	}

	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	a, err := h.alerts.FindByID(r.Context(), alertID)
	if err != nil {
		h.logger.Error("failed to get alert", "error", err, "alert_id", alertID)
		server.WriteError(w, http.StatusInternalServerError, "failed to retrieve alert")
		return
	}

	if a == nil {
		server.WriteError(w, http.StatusNotFound, "alert not found")
		return
	}

	// Enforce tenant isolation
	if a.TenantID != tenantID {
		server.WriteError(w, http.StatusNotFound, "alert not found")
		return
	}

	server.WriteJSON(w, http.StatusOK, toAlertResponse(a))
}

// UpdateAlertStatus handles PATCH /api/v1/alerts/{id}
func (h *AlertHandler) UpdateAlertStatus(w http.ResponseWriter, r *http.Request) {
	alertID := r.PathValue("id")
	if alertID == "" {
		server.WriteError(w, http.StatusBadRequest, "missing alert id")
		return
	}

	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		string(alert.AlertStatusPending):    true,
		string(alert.AlertStatusSent):       true,
		string(alert.AlertStatusDelivered):  true,
		string(alert.AlertStatusFailed):     true,
		string(alert.AlertStatusSuppressed): true,
	}
	if !validStatuses[req.Status] {
		server.WriteError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	a, err := h.alerts.FindByID(r.Context(), alertID)
	if err != nil {
		h.logger.Error("failed to get alert", "error", err, "alert_id", alertID)
		server.WriteError(w, http.StatusInternalServerError, "failed to retrieve alert")
		return
	}

	if a == nil || a.TenantID != tenantID {
		server.WriteError(w, http.StatusNotFound, "alert not found")
		return
	}

	a.Status = alert.AlertStatus(req.Status)
	a.UpdatedAt = time.Now().UTC()

	if err := h.alerts.Update(r.Context(), a); err != nil {
		h.logger.Error("failed to update alert", "error", err, "alert_id", alertID)
		server.WriteError(w, http.StatusInternalServerError, "failed to update alert")
		return
	}

	server.WriteJSON(w, http.StatusOK, toAlertResponse(a))
}
