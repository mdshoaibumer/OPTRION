package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/optrion/optrion/internal/alert/domain/escalationpolicy"
	"github.com/optrion/optrion/internal/alert/port/repository"
	"github.com/optrion/optrion/internal/platform/server"
	"github.com/optrion/optrion/internal/shared/id"
)

// EscalationPolicyHandler handles HTTP requests for escalation policies.
type EscalationPolicyHandler struct {
	policies repository.EscalationPolicyRepository
	logger   *slog.Logger
}

// NewEscalationPolicyHandler creates a new escalation policy handler.
func NewEscalationPolicyHandler(policies repository.EscalationPolicyRepository, logger *slog.Logger) *EscalationPolicyHandler {
	return &EscalationPolicyHandler{
		policies: policies,
		logger:   logger,
	}
}

type escalationStepResponse struct {
	DelayMinutes int      `json:"delay_minutes"`
	ChannelIDs   []string `json:"channel_ids"`
	Reminder     bool     `json:"reminder"`
}

type escalationPolicyResponse struct {
	ID          string                   `json:"id"`
	TenantID    string                   `json:"tenant_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Steps       []escalationStepResponse `json:"steps"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
}

func toEscalationPolicyResponse(p *escalationpolicy.EscalationPolicy) escalationPolicyResponse {
	steps := make([]escalationStepResponse, 0, len(p.Steps))
	for _, s := range p.Steps {
		steps = append(steps, escalationStepResponse{
			DelayMinutes: s.DelayMinutes,
			ChannelIDs:   s.ChannelIDs,
			Reminder:     s.Reminder,
		})
	}
	return escalationPolicyResponse{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Name:        p.Name,
		Description: p.Description,
		Steps:       steps,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

// GetEscalationPolicies handles GET /api/v1/escalation-policies
func (h *EscalationPolicyHandler) GetEscalationPolicies(w http.ResponseWriter, r *http.Request) {
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	policies, err := h.policies.ListByTenant(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to list escalation policies", "error", err, "tenant_id", tenantID)
		server.WriteError(w, http.StatusInternalServerError, "failed to retrieve escalation policies")
		return
	}

	resp := make([]escalationPolicyResponse, 0, len(policies))
	for _, p := range policies {
		resp = append(resp, toEscalationPolicyResponse(p))
	}

	server.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data":  resp,
		"total": len(resp),
	})
}

// PostEscalationPolicy handles POST /api/v1/escalation-policies
func (h *EscalationPolicyHandler) PostEscalationPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Steps       []struct {
			DelayMinutes int      `json:"delay_minutes"`
			ChannelIDs   []string `json:"channel_ids"`
			Reminder     bool     `json:"reminder"`
		} `json:"steps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		server.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Steps) == 0 {
		server.WriteError(w, http.StatusBadRequest, "at least one escalation step is required")
		return
	}

	steps := make([]escalationpolicy.EscalationStep, 0, len(req.Steps))
	for _, s := range req.Steps {
		steps = append(steps, escalationpolicy.EscalationStep{
			DelayMinutes: s.DelayMinutes,
			ChannelIDs:   s.ChannelIDs,
			Reminder:     s.Reminder,
		})
	}

	now := time.Now().UTC()
	policy := &escalationpolicy.EscalationPolicy{
		ID:          id.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Steps:       steps,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.policies.Create(r.Context(), policy); err != nil {
		h.logger.Error("failed to create escalation policy", "error", err, "tenant_id", tenantID)
		server.WriteError(w, http.StatusInternalServerError, "failed to create escalation policy")
		return
	}

	server.WriteJSON(w, http.StatusCreated, toEscalationPolicyResponse(policy))
}

// PatchEscalationPolicy handles PATCH /api/v1/escalation-policies/{id}
func (h *EscalationPolicyHandler) PatchEscalationPolicy(w http.ResponseWriter, r *http.Request) {
	policyID := r.PathValue("id")
	if policyID == "" {
		server.WriteError(w, http.StatusBadRequest, "missing policy id")
		return
	}

	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	policy, err := h.policies.FindByID(r.Context(), policyID)
	if err != nil {
		h.logger.Error("failed to get escalation policy", "error", err, "policy_id", policyID)
		server.WriteError(w, http.StatusInternalServerError, "failed to retrieve escalation policy")
		return
	}

	if policy == nil || policy.TenantID != tenantID {
		server.WriteError(w, http.StatusNotFound, "escalation policy not found")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Steps       []struct {
			DelayMinutes int      `json:"delay_minutes"`
			ChannelIDs   []string `json:"channel_ids"`
			Reminder     bool     `json:"reminder"`
		} `json:"steps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		policy.Name = *req.Name
	}
	if req.Description != nil {
		policy.Description = *req.Description
	}
	if req.Steps != nil {
		steps := make([]escalationpolicy.EscalationStep, 0, len(req.Steps))
		for _, s := range req.Steps {
			steps = append(steps, escalationpolicy.EscalationStep{
				DelayMinutes: s.DelayMinutes,
				ChannelIDs:   s.ChannelIDs,
				Reminder:     s.Reminder,
			})
		}
		policy.Steps = steps
	}
	policy.UpdatedAt = time.Now().UTC()

	if err := h.policies.Update(r.Context(), policy); err != nil {
		h.logger.Error("failed to update escalation policy", "error", err, "policy_id", policyID)
		server.WriteError(w, http.StatusInternalServerError, "failed to update escalation policy")
		return
	}

	server.WriteJSON(w, http.StatusOK, toEscalationPolicyResponse(policy))
}
