package port

import (
	"context"
	"time"

	"github.com/optrion/optrion/internal/health/domain"
)

// HealthMetricRepository persists health metric definitions.
type HealthMetricRepository interface {
	Create(ctx context.Context, metric *domain.HealthMetric) error
	GetByID(ctx context.Context, id string) (*domain.HealthMetric, error)
	ListByComponent(ctx context.Context, componentID string) ([]*domain.HealthMetric, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.HealthMetric, error)
	ListEnabled(ctx context.Context, tenantID string) ([]*domain.HealthMetric, error)
	Update(ctx context.Context, metric *domain.HealthMetric) error
	Upsert(ctx context.Context, metric *domain.HealthMetric) error
}

// MetricSnapshotRepository persists metric measurements.
type MetricSnapshotRepository interface {
	Create(ctx context.Context, snapshot *domain.MetricSnapshot) error
	CreateBatch(ctx context.Context, snapshots []*domain.MetricSnapshot) error
	ListByMetric(ctx context.Context, metricID string, from, to time.Time, limit int) ([]*domain.MetricSnapshot, error)
	ListByTenant(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.MetricSnapshot, error)
	GetLatestByMetric(ctx context.Context, metricID string) (*domain.MetricSnapshot, error)
}

// HealthScoreRepository persists computed health scores.
type HealthScoreRepository interface {
	Create(ctx context.Context, score *domain.HealthScore) error
	GetLatestByComponent(ctx context.Context, componentID string) (*domain.HealthScore, error)
	ListByTenant(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.HealthScore, error)
	ListByComponent(ctx context.Context, componentID string, from, to time.Time, limit int) ([]*domain.HealthScore, error)
}

// ComponentHealthRepository persists current component health status.
type ComponentHealthRepository interface {
	Upsert(ctx context.Context, status *domain.ComponentHealth) error
	GetByComponent(ctx context.Context, componentID string) (*domain.ComponentHealth, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.ComponentHealth, error)
}

// AnomalyRepository persists detected anomalies.
type AnomalyRepository interface {
	Create(ctx context.Context, anomaly *domain.Anomaly) error
	GetByID(ctx context.Context, id string) (*domain.Anomaly, error)
	ListByTenant(ctx context.Context, tenantID string, filter AnomalyFilter) ([]*domain.Anomaly, error)
	ListUnresolved(ctx context.Context, tenantID string) ([]*domain.Anomaly, error)
	Resolve(ctx context.Context, id string) error
}

// AnomalyFilter defines filtering options for anomaly queries.
type AnomalyFilter struct {
	ComponentID *string
	Severity    *domain.Severity
	Resolved    *bool
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}
