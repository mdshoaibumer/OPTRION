package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/optrion/optrion/internal/alert/domain/alertrule"
	"github.com/optrion/optrion/internal/alert/port/repository"
	"github.com/optrion/optrion/internal/platform/server"
	"github.com/optrion/optrion/internal/shared/id"
)

// AlertRuleHandler handles HTTP requests for alert rules.
type AlertRuleHandler struct {
	rules  repository.AlertRuleRepository
	logger *slog.Logger
}

// NewAlertRuleHandler creates a new alert rule handler.
func NewAlertRuleHandler(rules repository.AlertRuleRepository, logger *slog.Logger) *AlertRuleHandler {
	return &AlertRuleHandler{
		rules:  rules,
		logger: logger,
	}
}

type alertRuleResponse struct {
	ID                 string   `json:"id"`
	TenantID           string   `json:"tenant_id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Severity           string   `json:"severity"`
	Enabled            bool     `json:"enabled"`
	Channels           []string `json:"channels"`
	EscalationPolicyID string   `json:"escalation_policy_id,omitempty"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

func toAlertRuleResponse(r *alertrule.AlertRule) alertRuleResponse {
	return alertRuleResponse{
		ID:                 r.ID,
		TenantID:           r.TenantID,
		Name:               r.Name,
		Description:        r.Description,
		Severity:           r.Severity,
		Enabled:            r.Enabled,
		Channels:           r.Channels,
		EscalationPolicyID: r.EscalationPolicyID,
		CreatedAt:          r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          r.UpdatedAt.Format(time.RFC3339),
	}
}

// GetAlertRules handles GET /api/v1/alert-rules
func (h *AlertRuleHandler) GetAlertRules(w http.ResponseWriter, r *http.Request) {
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	rules, err := h.rules.ListByTenant(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to list alert rules", "error", err, "tenant_id", tenantID)
		server.WriteError(w, http.StatusInternalServerError, "failed to retrieve alert rules")
		return
	}

	resp := make([]alertRuleResponse, 0, len(rules))
	for _, rule := range rules {
		resp = append(resp, toAlertRuleResponse(rule))
	}

	server.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  resp,
		"total": len(resp),
	})
}

// PostAlertRule handles POST /api/v1/alert-rules
func (h *AlertRuleHandler) PostAlertRule(w http.ResponseWriter, r *http.Request) {
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req struct {
		Name               string   `json:"name"`
		Description        string   `json:"description"`
		Severity           string   `json:"severity"`
		Enabled            bool     `json:"enabled"`
		Channels           []string `json:"channels"`
		EscalationPolicyID string   `json:"escalation_policy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		server.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Severity == "" {
		server.WriteError(w, http.StatusBadRequest, "severity is required")
		return
	}

	now := time.Now().UTC()
	rule := &alertrule.AlertRule{
		ID:                 id.New(),
		TenantID:           tenantID,
		Name:               req.Name,
		Description:        req.Description,
		Severity:           req.Severity,
		Enabled:            req.Enabled,
		Channels:           req.Channels,
		EscalationPolicyID: req.EscalationPolicyID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := h.rules.Create(r.Context(), rule); err != nil {
		h.logger.Error("failed to create alert rule", "error", err, "tenant_id", tenantID)
		server.WriteError(w, http.StatusInternalServerError, "failed to create alert rule")
		return
	}

	server.WriteJSON(w, http.StatusCreated, toAlertRuleResponse(rule))
}

// PatchAlertRule handles PATCH /api/v1/alert-rules/{id}
func (h *AlertRuleHandler) PatchAlertRule(w http.ResponseWriter, r *http.Request) {
	ruleID := r.PathValue("id")
	if ruleID == "" {
		server.WriteError(w, http.StatusBadRequest, "missing rule id")
		return
	}

	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	rule, err := h.rules.FindByID(r.Context(), ruleID)
	if err != nil {
		h.logger.Error("failed to get alert rule", "error", err, "rule_id", ruleID)
		server.WriteError(w, http.StatusInternalServerError, "failed to retrieve alert rule")
		return
	}

	if rule == nil || rule.TenantID != tenantID {
		server.WriteError(w, http.StatusNotFound, "alert rule not found")
		return
	}

	var req struct {
		Name               *string  `json:"name"`
		Description        *string  `json:"description"`
		Severity           *string  `json:"severity"`
		Enabled            *bool    `json:"enabled"`
		Channels           []string `json:"channels"`
		EscalationPolicyID *string  `json:"escalation_policy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.Severity != nil {
		rule.Severity = *req.Severity
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Channels != nil {
		rule.Channels = req.Channels
	}
	if req.EscalationPolicyID != nil {
		rule.EscalationPolicyID = *req.EscalationPolicyID
	}
	rule.UpdatedAt = time.Now().UTC()

	if err := h.rules.Update(r.Context(), rule); err != nil {
		h.logger.Error("failed to update alert rule", "error", err, "rule_id", ruleID)
		server.WriteError(w, http.StatusInternalServerError, "failed to update alert rule")
		return
	}

	server.WriteJSON(w, http.StatusOK, toAlertRuleResponse(rule))
}
