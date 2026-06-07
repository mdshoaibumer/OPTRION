package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/optrion/optrion/internal/health/anomaly"
	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
	"github.com/optrion/optrion/internal/health/scoring"
)

// HealthService orchestrates health monitoring use cases.
type HealthService struct {
	metrics      port.HealthMetricRepository
	snapshots    port.MetricSnapshotRepository
	scores       port.HealthScoreRepository
	components   port.ComponentHealthRepository
	anomalies    port.AnomalyRepository
	checkConfigs port.HealthCheckConfigRepository
	engine       *scoring.Engine
	detector     *anomaly.Detector
	logger       *slog.Logger
}

// NewHealthService creates a new HealthService with all dependencies.
func NewHealthService(
	metrics port.HealthMetricRepository,
	snapshots port.MetricSnapshotRepository,
	scores port.HealthScoreRepository,
	components port.ComponentHealthRepository,
	anomalies port.AnomalyRepository,
	engine *scoring.Engine,
	detector *anomaly.Detector,
	logger *slog.Logger,
) *HealthService {
	return &HealthService{
		metrics:    metrics,
		snapshots:  snapshots,
		scores:     scores,
		components: components,
		anomalies:  anomalies,
		engine:     engine,
		detector:   detector,
		logger:     logger,
	}
}

// WithCheckConfigs sets the health check configuration repository.
func (s *HealthService) WithCheckConfigs(repo port.HealthCheckConfigRepository) *HealthService {
	s.checkConfigs = repo
	return s
}

// ProcessCollectorResult handles results from a collector run.
// This is the core processing pipeline: store metrics → compute score → detect anomalies.
func (s *HealthService) ProcessCollectorResult(ctx context.Context, result *port.CollectorResult) {
	if result == nil || len(result.Metrics) == 0 {
		return
	}

	tenantID := ""
	// Look up metric definitions to get tenant context
	metricsMap := make(map[domain.MetricType]*domain.HealthMetric)
	metricDefs, err := s.metrics.ListByComponent(ctx, result.ComponentID)
	if err != nil {
		s.logger.Error("failed to list metrics for component", "component_id", result.ComponentID, "error", err)
		return
	}

	for _, m := range metricDefs {
		metricsMap[m.MetricType] = m
		tenantID = m.TenantID
	}

	if tenantID == "" {
		s.logger.Warn("no metric definitions found for component", "component_id", result.ComponentID)
		return
	}

	// 1. Store metric snapshots
	snapshots := make([]*domain.MetricSnapshot, 0, len(result.Metrics))
	for _, reading := range result.Metrics {
		metricDef, exists := metricsMap[reading.MetricType]
		if !exists {
			continue
		}

		snapshot := domain.NewMetricSnapshot(
			tenantID,
			metricDef.ID,
			reading.Value,
			metricDef.Thresholds,
			reading.Labels,
		)
		snapshots = append(snapshots, snapshot)
	}

	if len(snapshots) > 0 {
		if err := s.snapshots.CreateBatch(ctx, snapshots); err != nil {
			s.logger.Error("failed to store metric snapshots", "error", err)
		}
	}

	// 2. Compute health score
	scoreResult := s.engine.Compute(result)
	healthScore := domain.NewHealthScore(tenantID, result.ComponentID, scoreResult.Score, scoreResult.Reasons)
	if err := s.scores.Create(ctx, healthScore); err != nil {
		s.logger.Error("failed to store health score", "error", err)
	}

	// 3. Update component status
	componentHealth := &domain.ComponentHealth{
		TenantID:      tenantID,
		ComponentID:   result.ComponentID,
		CollectorType: result.CollectorType,
		Status:        scoreResult.Status,
		Score:         scoreResult.Score,
		LastCheckAt:   time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := s.components.Upsert(ctx, componentHealth); err != nil {
		s.logger.Error("failed to update component status", "error", err)
	}

	// 4. Detect anomalies
	detections := s.detector.Analyze(result)
	for _, detection := range detections {
		if !detection.IsAnomaly {
			continue
		}

		// Find the metric ID for this anomaly
		metricID := ""
		var metricType domain.MetricType
		for _, reading := range result.Metrics {
			if metricDef, exists := metricsMap[reading.MetricType]; exists {
				if reading.Value == detection.ActualValue {
					metricID = metricDef.ID
					metricType = reading.MetricType
					break
				}
			}
		}

		if metricID == "" {
			continue
		}

		anom := domain.NewAnomaly(
			tenantID,
			result.ComponentID,
			metricID,
			metricType,
			detection.Severity,
			detection.Title,
			detection.Description,
			detection.ExpectedValue,
			detection.ActualValue,
		)
		if err := s.anomalies.Create(ctx, anom); err != nil {
			s.logger.Error("failed to store anomaly", "error", err)
		}

		s.logger.Warn("anomaly detected",
			"component_id", result.ComponentID,
			"metric_type", metricType,
			"severity", detection.Severity,
			"title", detection.Title,
		)
	}
}

// --- Queries ---

// HealthSummary represents the overall health summary for a tenant.
type HealthSummary struct {
	TenantID      string              `json:"tenant_id"`
	OverallScore  int                 `json:"overall_score"`
	OverallStatus domain.MetricStatus `json:"overall_status"`
	Components    int                 `json:"components"`
	Healthy       int                 `json:"healthy"`
	Degraded      int                 `json:"degraded"`
	Critical      int                 `json:"critical"`
	Reasons       []string            `json:"reasons"`
	LastUpdatedAt time.Time           `json:"last_updated_at"`
}

// GetSummary returns the overall health summary for a tenant.
func (s *HealthService) GetSummary(ctx context.Context, tenantID string) (*HealthSummary, error) {
	statuses, err := s.components.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	summary := &HealthSummary{
		TenantID:   tenantID,
		Components: len(statuses),
		Reasons:    make([]string, 0),
	}

	if len(statuses) == 0 {
		summary.OverallScore = 100
		summary.OverallStatus = domain.StatusHealthy
		summary.LastUpdatedAt = time.Now().UTC()
		return summary, nil
	}

	totalScore := 0
	var latestUpdate time.Time

	for _, cs := range statuses {
		totalScore += cs.Score
		if cs.UpdatedAt.After(latestUpdate) {
			latestUpdate = cs.UpdatedAt
		}

		switch cs.Status {
		case domain.StatusHealthy:
			summary.Healthy++
		case domain.StatusDegraded:
			summary.Degraded++
		case domain.StatusCritical:
			summary.Critical++
		}
	}

	summary.OverallScore = totalScore / len(statuses)
	summary.LastUpdatedAt = latestUpdate

	if summary.OverallScore < 70 {
		summary.OverallStatus = domain.StatusCritical
	} else if summary.OverallScore < 90 {
		summary.OverallStatus = domain.StatusDegraded
	} else {
		summary.OverallStatus = domain.StatusHealthy
	}

	// Gather reasons from latest scores
	for _, cs := range statuses {
		latest, err := s.scores.GetLatestByComponent(ctx, cs.ComponentID)
		if err != nil || latest == nil {
			continue
		}
		summary.Reasons = append(summary.Reasons, latest.Reasons...)
	}

	return summary, nil
}

// GetComponentStatuses returns the current health status of all components for a tenant.
func (s *HealthService) GetComponentStatuses(ctx context.Context, tenantID string) ([]*domain.ComponentHealth, error) {
	return s.components.ListByTenant(ctx, tenantID)
}

// GetHistory returns historical health scores for a tenant within a time range.
func (s *HealthService) GetHistory(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.HealthScore, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	return s.scores.ListByTenant(ctx, tenantID, from, to, limit)
}

// GetAnomalies returns anomalies for a tenant with optional filtering.
func (s *HealthService) GetAnomalies(ctx context.Context, tenantID string, filter port.AnomalyFilter) ([]*domain.Anomaly, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	return s.anomalies.ListByTenant(ctx, tenantID, filter)
}

// GetCheckConfig retrieves health check configuration for a component.
func (s *HealthService) GetCheckConfig(ctx context.Context, componentID string) (*domain.HealthCheckConfig, error) {
	if s.checkConfigs == nil {
		return nil, nil
	}
	return s.checkConfigs.GetByComponent(ctx, componentID)
}

// ListCheckConfigs retrieves all health check configurations for a tenant.
func (s *HealthService) ListCheckConfigs(ctx context.Context, tenantID string) ([]*domain.HealthCheckConfig, error) {
	if s.checkConfigs == nil {
		return nil, nil
	}
	return s.checkConfigs.ListByTenant(ctx, tenantID)
}

// SaveCheckConfig persists a health check configuration.
func (s *HealthService) SaveCheckConfig(ctx context.Context, config *domain.HealthCheckConfig) error {
	if s.checkConfigs == nil {
		return nil
	}
	return s.checkConfigs.Upsert(ctx, config)
}
