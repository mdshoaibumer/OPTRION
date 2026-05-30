package domain

import (
	"fmt"
	"time"

	"github.com/optrion/optrion/internal/shared/id"
)

// MetricType identifies the category of a health metric.
type MetricType string

const (
	MetricAvailability      MetricType = "availability"
	MetricResponseTime      MetricType = "response_time"
	MetricErrorRate         MetricType = "error_rate"
	MetricThroughput        MetricType = "throughput"
	MetricUptime            MetricType = "uptime"
	MetricConnectionStatus  MetricType = "connection_status"
	MetricQueryLatency      MetricType = "query_latency"
	MetricActiveConnections MetricType = "active_connections"
	MetricSlowQueries       MetricType = "slow_queries"
	MetricDeadlocks         MetricType = "deadlocks"
	MetricIndexUsage        MetricType = "index_usage"
	MetricPoolHealth        MetricType = "pool_health"
	MetricMemoryUsage       MetricType = "memory_usage"
	MetricHitRatio          MetricType = "hit_ratio"
	MetricEvictions         MetricType = "evictions"
	MetricConnectedClients  MetricType = "connected_clients"
	MetricCPU               MetricType = "cpu"
	MetricRAM               MetricType = "ram"
	MetricDisk              MetricType = "disk"
	MetricLoadAverage       MetricType = "load_average"
	MetricNetwork           MetricType = "network"
)

// IsValid checks if the metric type is recognized.
func (m MetricType) IsValid() bool {
	switch m {
	case MetricAvailability, MetricResponseTime, MetricErrorRate, MetricThroughput,
		MetricUptime, MetricConnectionStatus, MetricQueryLatency, MetricActiveConnections,
		MetricSlowQueries, MetricDeadlocks, MetricIndexUsage, MetricPoolHealth,
		MetricMemoryUsage, MetricHitRatio, MetricEvictions, MetricConnectedClients,
		MetricCPU, MetricRAM, MetricDisk, MetricLoadAverage, MetricNetwork:
		return true
	}
	return false
}

// CollectorType identifies the source of metrics.
type CollectorType string

const (
	CollectorBackend  CollectorType = "backend"
	CollectorPostgres CollectorType = "postgres"
	CollectorRedis    CollectorType = "redis"
	CollectorServer   CollectorType = "server"
)

// IsValid checks if the collector type is recognized.
func (c CollectorType) IsValid() bool {
	switch c {
	case CollectorBackend, CollectorPostgres, CollectorRedis, CollectorServer:
		return true
	}
	return false
}

// MetricStatus represents the health status derived from a metric value.
type MetricStatus string

const (
	StatusHealthy  MetricStatus = "healthy"
	StatusDegraded MetricStatus = "degraded"
	StatusCritical MetricStatus = "critical"
	StatusUnknown  MetricStatus = "unknown"
)

// Severity represents the severity level of an anomaly.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// HealthMetric defines a metric that can be collected for a component.
// It represents the definition/configuration of what to measure.
type HealthMetric struct {
	ID            string
	TenantID      string
	ComponentID   string
	MetricType    MetricType
	CollectorType CollectorType
	Name          string
	Unit          string
	Thresholds    Thresholds
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Thresholds defines the warning and critical boundaries for a metric.
type Thresholds struct {
	WarningMin  *float64 `json:"warning_min,omitempty"`
	WarningMax  *float64 `json:"warning_max,omitempty"`
	CriticalMin *float64 `json:"critical_min,omitempty"`
	CriticalMax *float64 `json:"critical_max,omitempty"`
}

// Evaluate determines the status for a given metric value.
func (t Thresholds) Evaluate(value float64) MetricStatus {
	// Critical thresholds take priority
	if t.CriticalMax != nil && value > *t.CriticalMax {
		return StatusCritical
	}
	if t.CriticalMin != nil && value < *t.CriticalMin {
		return StatusCritical
	}
	// Warning thresholds
	if t.WarningMax != nil && value > *t.WarningMax {
		return StatusDegraded
	}
	if t.WarningMin != nil && value < *t.WarningMin {
		return StatusDegraded
	}
	return StatusHealthy
}

// NewHealthMetric creates a new health metric definition.
func NewHealthMetric(tenantID, componentID string, metricType MetricType, collectorType CollectorType, name, unit string, thresholds Thresholds) (*HealthMetric, error) {
	if !id.IsValid(tenantID) {
		return nil, fmt.Errorf("invalid tenant ID: %s", tenantID)
	}
	if !id.IsValid(componentID) {
		return nil, fmt.Errorf("invalid component ID: %s", componentID)
	}
	if !metricType.IsValid() {
		return nil, fmt.Errorf("invalid metric type: %s", metricType)
	}
	if !collectorType.IsValid() {
		return nil, fmt.Errorf("invalid collector type: %s", collectorType)
	}
	if name == "" {
		return nil, fmt.Errorf("metric name is required")
	}

	now := time.Now().UTC()
	return &HealthMetric{
		ID:            id.New(),
		TenantID:      tenantID,
		ComponentID:   componentID,
		MetricType:    metricType,
		CollectorType: collectorType,
		Name:          name,
		Unit:          unit,
		Thresholds:    thresholds,
		Enabled:       true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// MetricSnapshot is a single point-in-time measurement of a health metric.
type MetricSnapshot struct {
	ID          string
	TenantID    string
	MetricID    string
	Value       float64
	Status      MetricStatus
	Labels      map[string]string
	CollectedAt time.Time
}

// NewMetricSnapshot creates a new metric snapshot with evaluated status.
func NewMetricSnapshot(tenantID, metricID string, value float64, thresholds Thresholds, labels map[string]string) *MetricSnapshot {
	if labels == nil {
		labels = make(map[string]string)
	}
	return &MetricSnapshot{
		ID:          id.New(),
		TenantID:    tenantID,
		MetricID:    metricID,
		Value:       value,
		Status:      thresholds.Evaluate(value),
		Labels:      labels,
		CollectedAt: time.Now().UTC(),
	}
}

// HealthScore represents the computed health score for a component or tenant.
type HealthScore struct {
	ID          string
	TenantID    string
	ComponentID string
	Score       int
	Status      MetricStatus
	Reasons     []string
	ComputedAt  time.Time
}

// NewHealthScore creates a health score from scoring results.
func NewHealthScore(tenantID, componentID string, score int, reasons []string) *HealthScore {
	status := StatusHealthy
	if score < 70 {
		status = StatusCritical
	} else if score < 90 {
		status = StatusDegraded
	}

	if reasons == nil {
		reasons = []string{}
	}

	// Clamp score between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return &HealthScore{
		ID:          id.New(),
		TenantID:    tenantID,
		ComponentID: componentID,
		Score:       score,
		Status:      status,
		Reasons:     reasons,
		ComputedAt:  time.Now().UTC(),
	}
}

// ComponentHealth represents the current health status of a component.
type ComponentHealth struct {
	ID            string
	TenantID      string
	ComponentID   string
	ComponentName string
	CollectorType CollectorType
	Status        MetricStatus
	Score         int
	LastCheckAt   time.Time
	UpdatedAt     time.Time
}

// NewComponentHealth creates a new component health record.
func NewComponentHealth(tenantID, componentID, componentName string, collectorType CollectorType) *ComponentHealth {
	now := time.Now().UTC()
	return &ComponentHealth{
		ID:            id.New(),
		TenantID:      tenantID,
		ComponentID:   componentID,
		ComponentName: componentName,
		CollectorType: collectorType,
		Status:        StatusUnknown,
		Score:         100,
		LastCheckAt:   now,
		UpdatedAt:     now,
	}
}

// Update sets the current health score and status.
func (ch *ComponentHealth) Update(score int, status MetricStatus) {
	ch.Score = score
	ch.Status = status
	ch.LastCheckAt = time.Now().UTC()
	ch.UpdatedAt = time.Now().UTC()
}

// Anomaly represents a detected deviation from normal behavior.
type Anomaly struct {
	ID            string
	TenantID      string
	ComponentID   string
	MetricID      string
	MetricType    MetricType
	Severity      Severity
	Title         string
	Description   string
	ExpectedValue float64
	ActualValue   float64
	Resolved      bool
	DetectedAt    time.Time
	ResolvedAt    *time.Time
}

// NewAnomaly creates a new anomaly record.
func NewAnomaly(tenantID, componentID, metricID string, metricType MetricType, severity Severity, title, description string, expected, actual float64) *Anomaly {
	return &Anomaly{
		ID:            id.New(),
		TenantID:      tenantID,
		ComponentID:   componentID,
		MetricID:      metricID,
		MetricType:    metricType,
		Severity:      severity,
		Title:         title,
		Description:   description,
		ExpectedValue: expected,
		ActualValue:   actual,
		Resolved:      false,
		DetectedAt:    time.Now().UTC(),
	}
}

// Resolve marks the anomaly as resolved.
func (a *Anomaly) Resolve() {
	now := time.Now().UTC()
	a.Resolved = true
	a.ResolvedAt = &now
}
