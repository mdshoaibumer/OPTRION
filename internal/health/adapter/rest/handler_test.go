package rest_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/health/adapter/rest"
	"github.com/optrion/optrion/internal/health/anomaly"
	"github.com/optrion/optrion/internal/health/app"
	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
	"github.com/optrion/optrion/internal/health/scoring"
)

// --- Mocks (simplified for handler tests) ---

type mockMetricRepo struct{}

func (m *mockMetricRepo) Create(_ context.Context, _ *domain.HealthMetric) error { return nil }
func (m *mockMetricRepo) GetByID(_ context.Context, _ string) (*domain.HealthMetric, error) {
	return nil, nil
}
func (m *mockMetricRepo) ListByComponent(_ context.Context, _ string) ([]*domain.HealthMetric, error) {
	return nil, nil
}
func (m *mockMetricRepo) ListByTenant(_ context.Context, _ string) ([]*domain.HealthMetric, error) {
	return nil, nil
}
func (m *mockMetricRepo) ListEnabled(_ context.Context, _ string) ([]*domain.HealthMetric, error) {
	return nil, nil
}
func (m *mockMetricRepo) Update(_ context.Context, _ *domain.HealthMetric) error { return nil }
func (m *mockMetricRepo) Upsert(_ context.Context, _ *domain.HealthMetric) error { return nil }

type mockSnapshotRepo struct{}

func (m *mockSnapshotRepo) Create(_ context.Context, _ *domain.MetricSnapshot) error { return nil }
func (m *mockSnapshotRepo) CreateBatch(_ context.Context, _ []*domain.MetricSnapshot) error {
	return nil
}
func (m *mockSnapshotRepo) ListByMetric(_ context.Context, _ string, _, _ time.Time, _ int) ([]*domain.MetricSnapshot, error) {
	return nil, nil
}
func (m *mockSnapshotRepo) ListByTenant(_ context.Context, _ string, _, _ time.Time, _ int) ([]*domain.MetricSnapshot, error) {
	return nil, nil
}
func (m *mockSnapshotRepo) GetLatestByMetric(_ context.Context, _ string) (*domain.MetricSnapshot, error) {
	return nil, nil
}

type mockScoreRepo struct {
	scores []*domain.HealthScore
}

func (m *mockScoreRepo) Create(_ context.Context, _ *domain.HealthScore) error { return nil }
func (m *mockScoreRepo) GetLatestByComponent(_ context.Context, _ string) (*domain.HealthScore, error) {
	return nil, nil
}
func (m *mockScoreRepo) ListByTenant(_ context.Context, _ string, _, _ time.Time, _ int) ([]*domain.HealthScore, error) {
	return m.scores, nil
}
func (m *mockScoreRepo) ListByComponent(_ context.Context, _ string, _, _ time.Time, _ int) ([]*domain.HealthScore, error) {
	return nil, nil
}

type mockComponentHealthRepo struct {
	statuses []*domain.ComponentHealth
}

func (m *mockComponentHealthRepo) Upsert(_ context.Context, _ *domain.ComponentHealth) error {
	return nil
}
func (m *mockComponentHealthRepo) GetByComponent(_ context.Context, _ string) (*domain.ComponentHealth, error) {
	return nil, nil
}
func (m *mockComponentHealthRepo) ListByTenant(_ context.Context, _ string) ([]*domain.ComponentHealth, error) {
	return m.statuses, nil
}

type mockAnomalyRepo struct {
	anomalies []*domain.Anomaly
}

func (m *mockAnomalyRepo) Create(_ context.Context, _ *domain.Anomaly) error { return nil }
func (m *mockAnomalyRepo) GetByID(_ context.Context, _ string) (*domain.Anomaly, error) {
	return nil, nil
}
func (m *mockAnomalyRepo) ListByTenant(_ context.Context, _ string, _ port.AnomalyFilter) ([]*domain.Anomaly, error) {
	return m.anomalies, nil
}
func (m *mockAnomalyRepo) ListUnresolved(_ context.Context, _ string) ([]*domain.Anomaly, error) {
	return nil, nil
}
func (m *mockAnomalyRepo) Resolve(_ context.Context, _ string) error { return nil }

// --- Handler Setup ---

func newTestHandler(componentRepo *mockComponentHealthRepo, scoreRepo *mockScoreRepo, anomalyRepo *mockAnomalyRepo) *rest.Handler {
	svc := app.NewHealthService(
		&mockMetricRepo{},
		&mockSnapshotRepo{},
		scoreRepo,
		componentRepo,
		anomalyRepo,
		scoring.NewEngine(),
		anomaly.NewDetector(3.0, 60, 10),
		slog.Default(),
	)
	return rest.NewHandler(svc, slog.Default())
}

// --- Tests ---

func TestGetSummary_RequiresTenantID(t *testing.T) {
	h := newTestHandler(&mockComponentHealthRepo{}, &mockScoreRepo{}, &mockAnomalyRepo{})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/summary", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetSummary_Success(t *testing.T) {
	componentRepo := &mockComponentHealthRepo{
		statuses: []*domain.ComponentHealth{
			{TenantID: "tenant-1", ComponentID: "comp-1", Status: domain.StatusHealthy, Score: 100, UpdatedAt: time.Now()},
			{TenantID: "tenant-1", ComponentID: "comp-2", Status: domain.StatusDegraded, Score: 80, UpdatedAt: time.Now()},
		},
	}
	h := newTestHandler(componentRepo, &mockScoreRepo{}, &mockAnomalyRepo{})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/summary?tenant_id=tenant-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp rest.SummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.OverallScore != 90 {
		t.Errorf("expected overall score 90, got %d", resp.OverallScore)
	}
	if resp.Components != 2 {
		t.Errorf("expected 2 components, got %d", resp.Components)
	}
}

func TestGetComponents_RequiresTenantID(t *testing.T) {
	h := newTestHandler(&mockComponentHealthRepo{}, &mockScoreRepo{}, &mockAnomalyRepo{})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/components", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetComponents_Success(t *testing.T) {
	componentRepo := &mockComponentHealthRepo{
		statuses: []*domain.ComponentHealth{
			{TenantID: "tenant-1", ComponentID: "comp-1", ComponentName: "Backend", CollectorType: domain.CollectorBackend, Status: domain.StatusHealthy, Score: 95},
		},
	}
	h := newTestHandler(componentRepo, &mockScoreRepo{}, &mockAnomalyRepo{})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/components?tenant_id=tenant-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp rest.ListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
}

func TestGetHistory_RequiresTenantID(t *testing.T) {
	h := newTestHandler(&mockComponentHealthRepo{}, &mockScoreRepo{}, &mockAnomalyRepo{})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/history", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetHistory_Success(t *testing.T) {
	scoreRepo := &mockScoreRepo{
		scores: []*domain.HealthScore{
			{TenantID: "tenant-1", ComponentID: "comp-1", Score: 100, Status: domain.StatusHealthy, ComputedAt: time.Now()},
		},
	}
	h := newTestHandler(&mockComponentHealthRepo{}, scoreRepo, &mockAnomalyRepo{})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/history?tenant_id=tenant-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetAnomalies_RequiresTenantID(t *testing.T) {
	h := newTestHandler(&mockComponentHealthRepo{}, &mockScoreRepo{}, &mockAnomalyRepo{})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/anomalies", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetAnomalies_Success(t *testing.T) {
	anomalyRepo := &mockAnomalyRepo{
		anomalies: []*domain.Anomaly{
			{TenantID: "tenant-1", ComponentID: "comp-1", Severity: domain.SeverityHigh, Title: "CPU spike", DetectedAt: time.Now()},
		},
	}
	h := newTestHandler(&mockComponentHealthRepo{}, &mockScoreRepo{}, anomalyRepo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/anomalies?tenant_id=tenant-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetAnomalies_WithFilters(t *testing.T) {
	anomalyRepo := &mockAnomalyRepo{
		anomalies: []*domain.Anomaly{
			{TenantID: "tenant-1", ComponentID: "comp-1", Severity: domain.SeverityHigh, Title: "Spike", DetectedAt: time.Now()},
		},
	}
	h := newTestHandler(&mockComponentHealthRepo{}, &mockScoreRepo{}, anomalyRepo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/anomalies?tenant_id=tenant-1&severity=high&resolved=false&component_id=comp-1&limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
