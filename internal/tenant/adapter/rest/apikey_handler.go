package rest

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/optrion/optrion/internal/platform/server"
	tenantpg "github.com/optrion/optrion/internal/tenant/adapter/postgres"
)

// APIKeyHandler handles API key management HTTP requests.
type APIKeyHandler struct {
	repo   *tenantpg.APIKeyRepository
	logger *slog.Logger
}

// NewAPIKeyHandler creates a new API key handler.
func NewAPIKeyHandler(repo *tenantpg.APIKeyRepository, logger *slog.Logger) *APIKeyHandler {
	return &APIKeyHandler{
		repo:   repo,
		logger: logger,
	}
}

// RegisterAuthenticatedRoutes registers API key routes wrapped with auth middleware.
func (h *APIKeyHandler) RegisterAuthenticatedRoutes(mux *http.ServeMux, authWrap func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/api-keys", authWrap(http.HandlerFunc(h.CreateAPIKey)))
	mux.Handle("POST /api/v1/api-keys/{id}/rotate", authWrap(http.HandlerFunc(h.RotateAPIKey)))
	mux.Handle("DELETE /api/v1/api-keys/{id}", authWrap(http.HandlerFunc(h.RevokeAPIKey)))
}

// CreateAPIKeyRequest is the request body for creating an API key.
type CreateAPIKeyRequest struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	ExpiresIn string   `json:"expires_in,omitempty"` // e.g., "720h" for 30 days
}

// CreateAPIKey handles POST /api/v1/api-keys
func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateAPIKeyRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		server.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresIn != "" {
		d, err := time.ParseDuration(req.ExpiresIn)
		if err != nil {
			server.WriteError(w, http.StatusBadRequest, "invalid expires_in duration")
			return
		}
		t := time.Now().UTC().Add(d)
		expiresAt = &t
	}

	rawKey, keyID, err := h.repo.CreateAPIKey(r.Context(), tenantID, req.Name, req.Scopes, expiresAt)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to create api key", "error", err)
		server.WriteError(w, http.StatusInternalServerError, "failed to create api key")
		return
	}

	server.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         keyID,
		"key":        rawKey,
		"name":       req.Name,
		"message":    "Store this key securely. It will not be shown again.",
		"expires_at": expiresAt,
	})
}

// RotateAPIKeyRequest is the request body for rotating an API key.
type RotateAPIKeyRequest struct {
	GracePeriod string `json:"grace_period"` // e.g., "24h", "72h"
}

// RotateAPIKey handles POST /api/v1/api-keys/{id}/rotate
func (h *APIKeyHandler) RotateAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	oldKeyID := r.PathValue("id")
	if oldKeyID == "" {
		server.WriteError(w, http.StatusBadRequest, "key id is required")
		return
	}

	var req RotateAPIKeyRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	gracePeriod := 24 * time.Hour // default 24h grace period
	if req.GracePeriod != "" {
		d, err := time.ParseDuration(req.GracePeriod)
		if err != nil {
			server.WriteError(w, http.StatusBadRequest, "invalid grace_period duration")
			return
		}
		if d < time.Hour {
			server.WriteError(w, http.StatusBadRequest, "grace_period must be at least 1 hour")
			return
		}
		gracePeriod = d
	}

	newRawKey, newKeyID, err := h.repo.RotateAPIKey(r.Context(), oldKeyID, tenantID, "rotated-key", nil, gracePeriod)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to rotate api key", "error", err, "old_key_id", oldKeyID)
		server.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"new_key_id":         newKeyID,
		"new_key":            newRawKey,
		"old_key_id":         oldKeyID,
		"grace_period":       req.GracePeriod,
		"message":            "New key created. Old key remains valid during the grace period.",
		"old_key_expires_at": time.Now().UTC().Add(gracePeriod).Format(time.RFC3339),
	})
}

// RevokeAPIKey handles DELETE /api/v1/api-keys/{id}
func (h *APIKeyHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	keyID := r.PathValue("id")
	if keyID == "" {
		server.WriteError(w, http.StatusBadRequest, "key id is required")
		return
	}

	if err := h.repo.RevokeAPIKey(r.Context(), keyID, tenantID); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to revoke api key", "error", err, "key_id", keyID)
		server.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	server.WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "revoked",
		"message": "API key has been revoked",
	})
}
