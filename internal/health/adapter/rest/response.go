package rest

import (
	"time"

	"github.com/optrion/optrion/internal/health/app"
	"github.com/optrion/optrion/internal/health/domain"
)

// --- Response DTOs ---

// ErrorResponse wraps error messages.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ListResponse wraps a paginated list of items.
type ListResponse struct {
	Data  interface{} `json:"data"`
	Count int         `json:"count"`
}

// SummaryResponse is the response for GET /api/v1/health/summary.
type SummaryResponse struct {
	TenantID      string   `json:"tenant_id"`
	OverallScore  int      `json:"overall_score"`
	OverallStatus string   `json:"overall_status"`
	Components    int      `json:"components"`
	Healthy       int      `json:"healthy"`
	Degraded      int      `json:"degraded"`
	Critical      int      `json:"critical"`
	Reasons       []string `json:"reasons"`
	LastUpdatedAt string   `json:"last_updated_at"`
}

// ComponentHealthResponse represents a single component's health.
type ComponentHealthResponse struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenant_id"`
	ComponentID   string `json:"component_id"`
	ComponentName string `json:"component_name"`
	CollectorType string `json:"collector_type"`
	Status        string `json:"status"`
	Score         int    `json:"score"`
	LastCheckAt   string `json:"last_check_at"`
	UpdatedAt     string `json:"updated_at"`
}

// HealthScoreResponse represents a historical health score entry.
type HealthScoreResponse struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	ComponentID string   `json:"component_id"`
	Score       int      `json:"score"`
	Status      string   `json:"status"`
	Reasons     []string `json:"reasons"`
	ComputedAt  string   `json:"computed_at"`
}

// AnomalyResponse represents a detected anomaly.
type AnomalyResponse struct {
	ID            string  `json:"id"`
	TenantID      string  `json:"tenant_id"`
	ComponentID   string  `json:"component_id"`
	MetricID      string  `json:"metric_id"`
	MetricType    string  `json:"metric_type"`
	Severity      string  `json:"severity"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	ExpectedValue float64 `json:"expected_value"`
	ActualValue   float64 `json:"actual_value"`
	Resolved      bool    `json:"resolved"`
	DetectedAt    string  `json:"detected_at"`
	ResolvedAt    *string `json:"resolved_at,omitempty"`
}

// --- Mappers ---

func toSummaryResponse(s *app.HealthSummary) SummaryResponse {
	return SummaryResponse{
		TenantID:      s.TenantID,
		OverallScore:  s.OverallScore,
		OverallStatus: string(s.OverallStatus),
		Components:    s.Components,
		Healthy:       s.Healthy,
		Degraded:      s.Degraded,
		Critical:      s.Critical,
		Reasons:       s.Reasons,
		LastUpdatedAt: s.LastUpdatedAt.Format(time.RFC3339),
	}
}

func toComponentHealthResponse(ch *domain.ComponentHealth) ComponentHealthResponse {
	return ComponentHealthResponse{
		ID:            ch.ID,
		TenantID:      ch.TenantID,
		ComponentID:   ch.ComponentID,
		ComponentName: ch.ComponentName,
		CollectorType: string(ch.CollectorType),
		Status:        string(ch.Status),
		Score:         ch.Score,
		LastCheckAt:   ch.LastCheckAt.Format(time.RFC3339),
		UpdatedAt:     ch.UpdatedAt.Format(time.RFC3339),
	}
}

func toHealthScoreResponse(s *domain.HealthScore) HealthScoreResponse {
	return HealthScoreResponse{
		ID:          s.ID,
		TenantID:    s.TenantID,
		ComponentID: s.ComponentID,
		Score:       s.Score,
		Status:      string(s.Status),
		Reasons:     s.Reasons,
		ComputedAt:  s.ComputedAt.Format(time.RFC3339),
	}
}

func toAnomalyResponse(a *domain.Anomaly) AnomalyResponse {
	resp := AnomalyResponse{
		ID:            a.ID,
		TenantID:      a.TenantID,
		ComponentID:   a.ComponentID,
		MetricID:      a.MetricID,
		MetricType:    string(a.MetricType),
		Severity:      string(a.Severity),
		Title:         a.Title,
		Description:   a.Description,
		ExpectedValue: a.ExpectedValue,
		ActualValue:   a.ActualValue,
		Resolved:      a.Resolved,
		DetectedAt:    a.DetectedAt.Format(time.RFC3339),
	}
	if a.ResolvedAt != nil {
		t := a.ResolvedAt.Format(time.RFC3339)
		resp.ResolvedAt = &t
	}
	return resp
}

// CheckConfigResponse represents a health check configuration.
type CheckConfigResponse struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenant_id"`
	ComponentID   string `json:"component_id"`
	CheckInterval string `json:"check_interval"`
	Timeout       string `json:"timeout"`
	Retries       int    `json:"retries"`
	Enabled       bool   `json:"enabled"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// CheckConfigRequest is the request body for creating/updating health check configs.
type CheckConfigRequest struct {
	ComponentID   string `json:"component_id"`
	CheckInterval string `json:"check_interval,omitempty"` // e.g., "30s", "1m"
	Timeout       string `json:"timeout,omitempty"`        // e.g., "10s"
	Retries       *int   `json:"retries,omitempty"`
	Enabled       *bool  `json:"enabled,omitempty"`
}

func toCheckConfigResponse(c *domain.HealthCheckConfig) CheckConfigResponse {
	return CheckConfigResponse{
		ID:            c.ID,
		TenantID:      c.TenantID,
		ComponentID:   c.ComponentID,
		CheckInterval: c.CheckInterval.String(),
		Timeout:       c.Timeout.String(),
		Retries:       c.Retries,
		Enabled:       c.Enabled,
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     c.UpdatedAt.Format(time.RFC3339),
	}
}
