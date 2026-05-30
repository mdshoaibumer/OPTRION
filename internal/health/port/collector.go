package port

import (
	"context"

	"github.com/optrion/optrion/internal/health/domain"
)

// CollectorResult holds the metrics collected from a single collection cycle.
type CollectorResult struct {
	ComponentID   string
	CollectorType domain.CollectorType
	Metrics       []MetricReading
	Error         error
	CollectedAt   string
}

// MetricReading is a single metric measurement from a collector.
type MetricReading struct {
	MetricType domain.MetricType
	Value      float64
	Labels     map[string]string
}

// HealthCollector is the interface for all metric collectors.
// Each collector is responsible for gathering metrics from a specific source.
type HealthCollector interface {
	// Type returns the collector type identifier.
	Type() domain.CollectorType

	// Collect gathers metrics from the monitored source.
	Collect(ctx context.Context) (*CollectorResult, error)

	// ComponentID returns the ID of the component being monitored.
	ComponentID() string

	// TenantID returns the ID of the tenant owning the component.
	TenantID() string
}
