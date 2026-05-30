package app_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/health/anomaly"
	"github.com/optrion/optrion/internal/health/app"
	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
	"github.com/optrion/optrion/internal/health/scoring"
)

// --- Mocks ---

type mockMetricRepo struct {
	metrics []*domain.HealthMetric
}

func (m *mockMetricRepo) Create(_ context.Context, metric *domain.HealthMetric) error {
	m.metrics = append(m.metrics, metric)
	return nil
}

func (m *mockMetricRepo) GetByID(_ context.Context, id string) (*domain.HealthMetric, error) {
	for _, met := range m.metrics {
		if met.ID == id {
			return met, nil
		}
	}
	return nil, nil
}

func (m *mockMetricRepo) ListByComponent(_ context.Context, componentID string) ([]*domain.HealthMetric, error) {
	var result []*domain.HealthMetric
	for _, met := range m.metrics {
		if met.ComponentID == componentID {
			result = append(result, met)
		}
	}
	return result, nil
}

func (m *mockMetricRepo) ListByTenant(_ context.Context, tenantID string) ([]*domain.HealthMetric, error) {
	var result []*domain.HealthMetric
	for _, met := range m.metrics {
		if met.TenantID == tenantID {
			result = append(result, met)
		}
	}
	return result, nil
}

func (m *mockMetricRepo) ListEnabled(_ context.Context, tenantID string) ([]*domain.HealthMetric, error) {
	var result []*domain.HealthMetric
	for _, met := range m.metrics {
		if met.TenantID == tenantID && met.Enabled {
			result = append(result, met)
		}
	}
	return result, nil
}

func (m *mockMetricRepo) Update(_ context.Context, metric *domain.HealthMetric) error {
	for i, met := range m.metrics {
		if met.ID == metric.ID {
			m.metrics[i] = metric
			return nil
		}
	}
	return nil
}

func (m *mockMetricRepo) Upsert(_ context.Context, metric *domain.HealthMetric) error {
	for i, met := range m.metrics {
		if met.ID == metric.ID {
			m.metrics[i] = metric
			return nil
		}
	}
	m.metrics = append(m.metrics, metric)
	return nil
}

type mockSnapshotRepo struct {
	snapshots []*domain.MetricSnapshot
}

func (m *mockSnapshotRepo) Create(_ context.Context, s *domain.MetricSnapshot) error {
	m.snapshots = append(m.snapshots, s)
	return nil
}

func (m *mockSnapshotRepo) CreateBatch(_ context.Context, snapshots []*domain.MetricSnapshot) error {
	m.snapshots = append(m.snapshots, snapshots...)
	return nil
}

func (m *mockSnapshotRepo) ListByMetric(_ context.Context, _ string, _, _ time.Time, _ int) ([]*domain.MetricSnapshot, error) {
	return m.snapshots, nil
}

func (m *mockSnapshotRepo) ListByTenant(_ context.Context, _ string, _, _ time.Time, _ int) ([]*domain.MetricSnapshot, error) {
	return m.snapshots, nil
}

func (m *mockSnapshotRepo) GetLatestByMetric(_ context.Context, metricID string) (*domain.MetricSnapshot, error) {
	for i := len(m.snapshots) - 1; i >= 0; i-- {
		if m.snapshots[i].MetricID == metricID {
			return m.snapshots[i], nil
		}
	}
	return nil, nil
}

type mockScoreRepo struct {
	scores []*domain.HealthScore
}

func (m *mockScoreRepo) Create(_ context.Context, s *domain.HealthScore) error {
	m.scores = append(m.scores, s)
	return nil
}

func (m *mockScoreRepo) GetLatestByComponent(_ context.Context, componentID string) (*domain.HealthScore, error) {
	for i := len(m.scores) - 1; i >= 0; i-- {
		if m.scores[i].ComponentID == componentID {
			return m.scores[i], nil
		}
	}
	return nil, nil
}

func (m *mockScoreRepo) ListByTenant(_ context.Context, tenantID string, _, _ time.Time, _ int) ([]*domain.HealthScore, error) {
	var result []*domain.HealthScore
	for _, s := range m.scores {
		if s.TenantID == tenantID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockScoreRepo) ListByComponent(_ context.Context, componentID string, _, _ time.Time, _ int) ([]*domain.HealthScore, error) {
	var result []*domain.HealthScore
	for _, s := range m.scores {
		if s.ComponentID == componentID {
			result = append(result, s)
		}
	}
	return result, nil
}

type mockComponentHealthRepo struct {
	statuses []*domain.ComponentHealth
}

func (m *mockComponentHealthRepo) Upsert(_ context.Context, status *domain.ComponentHealth) error {
	for i, cs := range m.statuses {
		if cs.ComponentID == status.ComponentID {
			m.statuses[i] = status
			return nil
		}
	}
	m.statuses = append(m.statuses, status)
	return nil
}

func (m *mockComponentHealthRepo) GetByComponent(_ context.Context, componentID string) (*domain.ComponentHealth, error) {
	for _, cs := range m.statuses {
		if cs.ComponentID == componentID {
			return cs, nil
		}
	}
	return nil, nil
}

func (m *mockComponentHealthRepo) ListByTenant(_ context.Context, tenantID string) ([]*domain.ComponentHealth, error) {
	var result []*domain.ComponentHealth
	for _, cs := range m.statuses {
		if cs.TenantID == tenantID {
			result = append(result, cs)
		}
	}
	return result, nil
}

type mockAnomalyRepo struct {
	anomalies []*domain.Anomaly
}

func (m *mockAnomalyRepo) Create(_ context.Context, a *domain.Anomaly) error {
	m.anomalies = append(m.anomalies, a)
	return nil
}

func (m *mockAnomalyRepo) GetByID(_ context.Context, id string) (*domain.Anomaly, error) {
	for _, a := range m.anomalies {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAnomalyRepo) ListByTenant(_ context.Context, tenantID string, _ port.AnomalyFilter) ([]*domain.Anomaly, error) {
	var result []*domain.Anomaly
	for _, a := range m.anomalies {
		if a.TenantID == tenantID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAnomalyRepo) ListUnresolved(_ context.Context, tenantID string) ([]*domain.Anomaly, error) {
	var result []*domain.Anomaly
	for _, a := range m.anomalies {
		if a.TenantID == tenantID && !a.Resolved {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAnomalyRepo) Resolve(_ context.Context, id string) error {
	for _, a := range m.anomalies {
		if a.ID == id {
			a.Resolve()
			return nil
		}
	}
	return nil
}

// --- Test Helpers ---

func newTestService(metricRepo *mockMetricRepo) (*app.HealthService, *mockSnapshotRepo, *mockScoreRepo, *mockComponentHealthRepo, *mockAnomalyRepo) {
	snapshotRepo := &mockSnapshotRepo{}
	scoreRepo := &mockScoreRepo{}
	componentRepo := &mockComponentHealthRepo{}
	anomalyRepo := &mockAnomalyRepo{}
	engine := scoring.NewEngine()
	detector := anomaly.NewDetector(3.0, 60, 10)
	logger := slog.Default()

	svc := app.NewHealthService(metricRepo, snapshotRepo, scoreRepo, componentRepo, anomalyRepo, engine, detector, logger)
	return svc, snapshotRepo, scoreRepo, componentRepo, anomalyRepo
}

// --- Service Tests ---

func TestProcessCollectorResult_StoresSnapshots(t *testing.T) {
	metricRepo := &mockMetricRepo{
		metrics: []*domain.HealthMetric{
			{
				ID:            "metric-1",
				TenantID:      "tenant-1",
				ComponentID:   "comp-1",
				MetricType:    domain.MetricAvailability,
				CollectorType: domain.CollectorBackend,
				Enabled:       true,
			},
			{
				ID:            "metric-2",
				TenantID:      "tenant-1",
				ComponentID:   "comp-1",
				MetricType:    domain.MetricResponseTime,
				CollectorType: domain.CollectorBackend,
				Enabled:       true,
			},
		},
	}

	svc, snapshotRepo, _, _, _ := newTestService(metricRepo)

	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 1},
			{MetricType: domain.MetricResponseTime, Value: 250},
		},
	}

	svc.ProcessCollectorResult(context.Background(), result)

	if len(snapshotRepo.snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshotRepo.snapshots))
	}
}

func TestProcessCollectorResult_ComputesScore(t *testing.T) {
	metricRepo := &mockMetricRepo{
		metrics: []*domain.HealthMetric{
			{
				ID:            "metric-1",
				TenantID:      "tenant-1",
				ComponentID:   "comp-1",
				MetricType:    domain.MetricAvailability,
				CollectorType: domain.CollectorBackend,
				Enabled:       true,
			},
		},
	}

	svc, _, scoreRepo, _, _ := newTestService(metricRepo)

	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 1},
		},
	}

	svc.ProcessCollectorResult(context.Background(), result)

	if len(scoreRepo.scores) != 1 {
		t.Fatalf("expected 1 score, got %d", len(scoreRepo.scores))
	}
	if scoreRepo.scores[0].Score != 100 {
		t.Errorf("expected score 100, got %d", scoreRepo.scores[0].Score)
	}
}

func TestProcessCollectorResult_UpdatesComponentStatus(t *testing.T) {
	metricRepo := &mockMetricRepo{
		metrics: []*domain.HealthMetric{
			{
				ID:            "metric-1",
				TenantID:      "tenant-1",
				ComponentID:   "comp-1",
				MetricType:    domain.MetricAvailability,
				CollectorType: domain.CollectorBackend,
				Enabled:       true,
			},
		},
	}

	svc, _, _, componentRepo, _ := newTestService(metricRepo)

	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 0},
		},
	}

	svc.ProcessCollectorResult(context.Background(), result)

	if len(componentRepo.statuses) != 1 {
		t.Fatalf("expected 1 component status, got %d", len(componentRepo.statuses))
	}
	if componentRepo.statuses[0].Status != domain.StatusCritical {
		t.Errorf("expected critical status, got %s", componentRepo.statuses[0].Status)
	}
	if componentRepo.statuses[0].Score != 60 {
		t.Errorf("expected score 60, got %d", componentRepo.statuses[0].Score)
	}
}

func TestProcessCollectorResult_NilResult(t *testing.T) {
	metricRepo := &mockMetricRepo{}
	svc, _, scoreRepo, _, _ := newTestService(metricRepo)

	// Should not panic
	svc.ProcessCollectorResult(context.Background(), nil)

	if len(scoreRepo.scores) != 0 {
		t.Error("expected no scores for nil result")
	}
}

func TestProcessCollectorResult_EmptyMetrics(t *testing.T) {
	metricRepo := &mockMetricRepo{}
	svc, _, scoreRepo, _, _ := newTestService(metricRepo)

	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics:       []port.MetricReading{},
	}

	svc.ProcessCollectorResult(context.Background(), result)

	if len(scoreRepo.scores) != 0 {
		t.Error("expected no scores for empty metrics")
	}
}

func TestProcessCollectorResult_NoMetricDefinitions(t *testing.T) {
	metricRepo := &mockMetricRepo{} // No metric definitions
	svc, _, scoreRepo, _, _ := newTestService(metricRepo)

	result := &port.CollectorResult{
		ComponentID:   "comp-unknown",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 1},
		},
	}

	svc.ProcessCollectorResult(context.Background(), result)

	// No tenant ID found → early return
	if len(scoreRepo.scores) != 0 {
		t.Error("expected no scores when no metric definitions exist")
	}
}

func TestGetSummary_NoComponents(t *testing.T) {
	metricRepo := &mockMetricRepo{}
	svc, _, _, _, _ := newTestService(metricRepo)

	summary, err := svc.GetSummary(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.OverallScore != 100 {
		t.Errorf("expected score 100 with no components, got %d", summary.OverallScore)
	}
	if summary.OverallStatus != domain.StatusHealthy {
		t.Errorf("expected healthy, got %s", summary.OverallStatus)
	}
	if summary.Components != 0 {
		t.Errorf("expected 0 components, got %d", summary.Components)
	}
}

func TestGetSummary_MultipleComponents(t *testing.T) {
	metricRepo := &mockMetricRepo{
		metrics: []*domain.HealthMetric{
			{ID: "m1", TenantID: "tenant-1", ComponentID: "comp-1", MetricType: domain.MetricAvailability, CollectorType: domain.CollectorBackend, Enabled: true},
			{ID: "m2", TenantID: "tenant-1", ComponentID: "comp-2", MetricType: domain.MetricConnectionStatus, CollectorType: domain.CollectorPostgres, Enabled: true},
		},
	}

	svc, _, _, componentRepo, _ := newTestService(metricRepo)

	// Simulate two components with different health
	componentRepo.statuses = []*domain.ComponentHealth{
		{TenantID: "tenant-1", ComponentID: "comp-1", Status: domain.StatusHealthy, Score: 100, UpdatedAt: time.Now()},
		{TenantID: "tenant-1", ComponentID: "comp-2", Status: domain.StatusDegraded, Score: 80, UpdatedAt: time.Now()},
	}

	summary, err := svc.GetSummary(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Components != 2 {
		t.Errorf("expected 2 components, got %d", summary.Components)
	}
	if summary.OverallScore != 90 {
		t.Errorf("expected average score 90, got %d", summary.OverallScore)
	}
	if summary.Healthy != 1 {
		t.Errorf("expected 1 healthy, got %d", summary.Healthy)
	}
	if summary.Degraded != 1 {
		t.Errorf("expected 1 degraded, got %d", summary.Degraded)
	}
}

func TestGetComponentStatuses(t *testing.T) {
	metricRepo := &mockMetricRepo{}
	svc, _, _, componentRepo, _ := newTestService(metricRepo)

	componentRepo.statuses = []*domain.ComponentHealth{
		{TenantID: "tenant-1", ComponentID: "comp-1", Status: domain.StatusHealthy, Score: 100},
		{TenantID: "tenant-1", ComponentID: "comp-2", Status: domain.StatusCritical, Score: 40},
		{TenantID: "tenant-2", ComponentID: "comp-3", Status: domain.StatusHealthy, Score: 95},
	}

	statuses, err := svc.GetComponentStatuses(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 2 {
		t.Errorf("expected 2 statuses for tenant-1, got %d", len(statuses))
	}
}

func TestGetHistory_DefaultLimit(t *testing.T) {
	metricRepo := &mockMetricRepo{}
	svc, _, scoreRepo, _, _ := newTestService(metricRepo)

	scoreRepo.scores = []*domain.HealthScore{
		{TenantID: "tenant-1", ComponentID: "comp-1", Score: 100},
		{TenantID: "tenant-1", ComponentID: "comp-1", Score: 95},
	}

	scores, err := svc.GetHistory(context.Background(), "tenant-1", time.Time{}, time.Now(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scores) != 2 {
		t.Errorf("expected 2 scores, got %d", len(scores))
	}
}

func TestGetAnomalies(t *testing.T) {
	metricRepo := &mockMetricRepo{}
	svc, _, _, _, anomalyRepo := newTestService(metricRepo)

	anomalyRepo.anomalies = []*domain.Anomaly{
		{TenantID: "tenant-1", ComponentID: "comp-1", Severity: domain.SeverityHigh, Resolved: false},
		{TenantID: "tenant-1", ComponentID: "comp-2", Severity: domain.SeverityMedium, Resolved: true},
		{TenantID: "tenant-2", ComponentID: "comp-3", Severity: domain.SeverityLow, Resolved: false},
	}

	anomalies, err := svc.GetAnomalies(context.Background(), "tenant-1", port.AnomalyFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(anomalies) != 2 {
		t.Errorf("expected 2 anomalies for tenant-1, got %d", len(anomalies))
	}
}
