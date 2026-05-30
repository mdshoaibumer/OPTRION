package port

import (
	"context"
	"time"

	"github.com/optrion/optrion/internal/incident/domain"
)

// IncidentRepository persists incident aggregates.
type IncidentRepository interface {
	Create(ctx context.Context, incident *domain.Incident) error
	GetByID(ctx context.Context, id string) (*domain.Incident, error)
	Update(ctx context.Context, incident *domain.Incident) error
	ListByTenant(ctx context.Context, tenantID string, filter IncidentFilter) ([]*domain.Incident, error)
	CountByTenant(ctx context.Context, tenantID string, filter IncidentFilter) (int, error)
	FindActiveByRuleAndComponent(ctx context.Context, ruleID, componentID string) (*domain.Incident, error)
	FindByCorrelation(ctx context.Context, correlationID string) ([]*domain.Incident, error)
}

// IncidentFilter defines query options for listing incidents.
type IncidentFilter struct {
	Status      *domain.IncidentStatus
	Severity    *domain.IncidentSeverity
	ComponentID *string
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}

// IncidentEventRepository persists incident events (event store).
type IncidentEventRepository interface {
	Append(ctx context.Context, events []domain.IncidentEvent) error
	ListByIncident(ctx context.Context, incidentID string) ([]domain.IncidentEvent, error)
	ListByTenant(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]domain.IncidentEvent, error)
}

// IncidentRuleRepository persists incident rules.
type IncidentRuleRepository interface {
	Create(ctx context.Context, rule *domain.IncidentRule) error
	GetByID(ctx context.Context, id string) (*domain.IncidentRule, error)
	Update(ctx context.Context, rule *domain.IncidentRule) error
	ListEnabled(ctx context.Context, tenantID string) ([]*domain.IncidentRule, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.IncidentRule, error)
	GetLastFiredAt(ctx context.Context, ruleID, componentID string) (*time.Time, error)
	RecordFired(ctx context.Context, ruleID, componentID string, firedAt time.Time) error
}

// IncidentCommentRepository persists incident comments.
type IncidentCommentRepository interface {
	Create(ctx context.Context, comment *domain.IncidentComment) error
	ListByIncident(ctx context.Context, incidentID string) ([]*domain.IncidentComment, error)
}

// IncidentTimelineRepository persists timeline entries.
type IncidentTimelineRepository interface {
	Create(ctx context.Context, entry *domain.IncidentTimeline) error
	ListByIncident(ctx context.Context, incidentID string, limit int) ([]*domain.IncidentTimeline, error)
	ListByTenant(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.IncidentTimeline, error)
}
