package rest

import (
	"time"

	"github.com/optrion/optrion/internal/incident/app"
	"github.com/optrion/optrion/internal/incident/domain"
)

// --- Request DTOs ---

// ActionRequest is used for acknowledge/investigate actions.
type ActionRequest struct {
	ActorID string `json:"actor_id"`
}

// ResolveRequest is used for resolving an incident.
type ResolveRequest struct {
	ActorID    string `json:"actor_id"`
	Resolution string `json:"resolution"`
}

// CloseRequest is used for closing an incident.
type CloseRequest struct {
	ActorID string `json:"actor_id"`
	Reason  string `json:"reason"`
}

// CommentRequest is used for adding a comment.
type CommentRequest struct {
	TenantID string `json:"tenant_id"`
	AuthorID string `json:"author_id"`
	Content  string `json:"content"`
}

// --- Response DTOs ---

// ErrorResponse is returned on error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ListResponse is a generic list response with count/total.
type ListResponse struct {
	Data  interface{} `json:"data"`
	Count int         `json:"count"`
	Total int         `json:"total"`
}

// IncidentResponse represents an incident in the REST API.
type IncidentResponse struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	ComponentID    string     `json:"component_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Severity       string     `json:"severity"`
	RuleID         string     `json:"rule_id,omitempty"`
	CorrelationID  string     `json:"correlation_id"`
	OccurredAt     time.Time  `json:"occurred_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	Version        int        `json:"version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CommentResponse represents a comment in the REST API.
type CommentResponse struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	IncidentID string    `json:"incident_id"`
	AuthorID   string    `json:"author_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

// TimelineResponse represents a timeline entry in the REST API.
type TimelineResponse struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	IncidentID string    `json:"incident_id"`
	EntryType  string    `json:"entry_type"`
	Title      string    `json:"title"`
	Details    string    `json:"details"`
	ActorID    string    `json:"actor_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

// StatsResponse represents incident statistics for a tenant.
type StatsResponse struct {
	TotalActive   int            `json:"total_active"`
	TotalOpen     int            `json:"total_open"`
	TotalResolved int            `json:"total_resolved"`
	TotalClosed   int            `json:"total_closed"`
	BySeverity    map[string]int `json:"by_severity"`
	MTTR          string         `json:"mttr"`
}

// --- Converters ---

func toIncidentResponse(inc *domain.Incident) IncidentResponse {
	return IncidentResponse{
		ID:             inc.ID,
		TenantID:       inc.TenantID,
		ComponentID:    inc.ComponentID,
		Title:          inc.Title,
		Description:    inc.Description,
		Status:         string(inc.Status),
		Severity:       string(inc.Severity),
		RuleID:         inc.RuleID,
		CorrelationID:  inc.CorrelationID,
		OccurredAt:     inc.OccurredAt,
		AcknowledgedAt: inc.AcknowledgedAt,
		ResolvedAt:     inc.ResolvedAt,
		ClosedAt:       inc.ClosedAt,
		Version:        inc.Version,
		CreatedAt:      inc.CreatedAt,
		UpdatedAt:      inc.UpdatedAt,
	}
}

func toCommentResponse(c *domain.IncidentComment) CommentResponse {
	return CommentResponse{
		ID:         c.ID,
		TenantID:   c.TenantID,
		IncidentID: c.IncidentID,
		AuthorID:   c.AuthorID,
		Content:    c.Content,
		CreatedAt:  c.CreatedAt,
	}
}

func toTimelineResponse(e *domain.IncidentTimeline) TimelineResponse {
	return TimelineResponse{
		ID:         e.ID,
		TenantID:   e.TenantID,
		IncidentID: e.IncidentID,
		EntryType:  string(e.EntryType),
		Title:      e.Title,
		Details:    e.Details,
		ActorID:    e.ActorID,
		OccurredAt: e.OccurredAt,
	}
}

func toStatsResponse(stats *app.IncidentStats) StatsResponse {
	bySeverity := make(map[string]int)
	for k, v := range stats.BySeverity {
		bySeverity[string(k)] = v
	}
	return StatsResponse{
		TotalActive:   stats.TotalActive,
		TotalOpen:     stats.TotalOpen,
		TotalResolved: stats.TotalResolved,
		TotalClosed:   stats.TotalClosed,
		BySeverity:    bySeverity,
		MTTR:          stats.MTTR.String(),
	}
}
