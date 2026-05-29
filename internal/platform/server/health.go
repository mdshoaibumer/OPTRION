package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/optrion/optrion/internal/platform/cache"
	"github.com/optrion/optrion/internal/platform/database"
)

// HealthResponse represents the health check response payload.
type HealthResponse struct {
	Status    string                    `json:"status"`
	Timestamp string                    `json:"timestamp"`
	Version   string                    `json:"version"`
	Checks    map[string]ComponentCheck `json:"checks"`
}

// ComponentCheck represents a single component's health status.
type ComponentCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthHandler handles health check requests.
type HealthHandler struct {
	db      *database.PostgreSQL
	redis   *cache.Redis
	version string
	logger  *slog.Logger
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db *database.PostgreSQL, redis *cache.Redis, version string, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:      db,
		redis:   redis,
		version: version,
		logger:  logger,
	}
}

// Liveness is a simple liveness probe (always returns 200 if the process is running).
func (h *HealthHandler) Liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "alive",
		})
	}
}

// Readiness checks all critical dependencies.
func (h *HealthHandler) Readiness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]ComponentCheck)
		overallStatus := "healthy"

		// Check PostgreSQL
		if err := h.db.Health(ctx); err != nil {
			checks["postgresql"] = ComponentCheck{Status: "unhealthy", Message: err.Error()}
			overallStatus = "unhealthy"
			h.logger.WarnContext(ctx, "health check failed: postgresql", "error", err)
		} else {
			checks["postgresql"] = ComponentCheck{Status: "healthy"}
		}

		// Check Redis
		if err := h.redis.Health(ctx); err != nil {
			checks["redis"] = ComponentCheck{Status: "unhealthy", Message: err.Error()}
			overallStatus = "unhealthy"
			h.logger.WarnContext(ctx, "health check failed: redis", "error", err)
		} else {
			checks["redis"] = ComponentCheck{Status: "healthy"}
		}

		resp := HealthResponse{
			Status:    overallStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   h.version,
			Checks:    checks,
		}

		status := http.StatusOK
		if overallStatus != "healthy" {
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
