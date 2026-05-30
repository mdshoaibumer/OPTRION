package domain_test

import (
	"testing"

	"github.com/optrion/optrion/internal/health/domain"
)

// --- MetricType Tests ---

func TestMetricType_IsValid(t *testing.T) {
	valid := []domain.MetricType{
		domain.MetricAvailability, domain.MetricResponseTime, domain.MetricErrorRate,
		domain.MetricThroughput, domain.MetricUptime, domain.MetricConnectionStatus,
		domain.MetricQueryLatency, domain.MetricActiveConnections, domain.MetricSlowQueries,
		domain.MetricDeadlocks, domain.MetricIndexUsage, domain.MetricPoolHealth,
		domain.MetricMemoryUsage, domain.MetricHitRatio, domain.MetricEvictions,
		domain.MetricConnectedClients, domain.MetricCPU, domain.MetricRAM,
		domain.MetricDisk, domain.MetricLoadAverage, domain.MetricNetwork,
	}

	for _, mt := range valid {
		if !mt.IsValid() {
			t.Errorf("expected %q to be valid", mt)
		}
	}
}

func TestMetricType_IsValid_Invalid(t *testing.T) {
	invalid := domain.MetricType("nonexistent")
	if invalid.IsValid() {
		t.Error("expected 'nonexistent' to be invalid")
	}
}

// --- CollectorType Tests ---

func TestCollectorType_IsValid(t *testing.T) {
	valid := []domain.CollectorType{
		domain.CollectorBackend, domain.CollectorPostgres,
		domain.CollectorRedis, domain.CollectorServer,
	}

	for _, ct := range valid {
		if !ct.IsValid() {
			t.Errorf("expected %q to be valid", ct)
		}
	}
}

func TestCollectorType_IsValid_Invalid(t *testing.T) {
	invalid := domain.CollectorType("kafka")
	if invalid.IsValid() {
		t.Error("expected 'kafka' to be invalid")
	}
}

// --- Thresholds Tests ---

func TestThresholds_Evaluate_Healthy(t *testing.T) {
	warn := 80.0
	crit := 95.0
	th := domain.Thresholds{WarningMax: &warn, CriticalMax: &crit}

	status := th.Evaluate(50.0)
	if status != domain.StatusHealthy {
		t.Errorf("expected healthy, got %s", status)
	}
}

func TestThresholds_Evaluate_Degraded_Max(t *testing.T) {
	warn := 80.0
	crit := 95.0
	th := domain.Thresholds{WarningMax: &warn, CriticalMax: &crit}

	status := th.Evaluate(85.0)
	if status != domain.StatusDegraded {
		t.Errorf("expected degraded, got %s", status)
	}
}

func TestThresholds_Evaluate_Critical_Max(t *testing.T) {
	warn := 80.0
	crit := 95.0
	th := domain.Thresholds{WarningMax: &warn, CriticalMax: &crit}

	status := th.Evaluate(96.0)
	if status != domain.StatusCritical {
		t.Errorf("expected critical, got %s", status)
	}
}

func TestThresholds_Evaluate_Degraded_Min(t *testing.T) {
	warnMin := 20.0
	th := domain.Thresholds{WarningMin: &warnMin}

	status := th.Evaluate(15.0)
	if status != domain.StatusDegraded {
		t.Errorf("expected degraded, got %s", status)
	}
}

func TestThresholds_Evaluate_Critical_Min(t *testing.T) {
	critMin := 10.0
	th := domain.Thresholds{CriticalMin: &critMin}

	status := th.Evaluate(5.0)
	if status != domain.StatusCritical {
		t.Errorf("expected critical, got %s", status)
	}
}

func TestThresholds_Evaluate_NoThresholds(t *testing.T) {
	th := domain.Thresholds{}
	status := th.Evaluate(999.0)
	if status != domain.StatusHealthy {
		t.Errorf("expected healthy with no thresholds, got %s", status)
	}
}

// --- NewHealthMetric Tests ---

func TestNewHealthMetric_Valid(t *testing.T) {
	tenantID := "0194ffa0-0000-7000-8000-000000000001"
	componentID := "0194ffa0-0000-7000-8000-000000000002"

	m, err := domain.NewHealthMetric(tenantID, componentID, domain.MetricCPU, domain.CollectorServer, "CPU Usage", "%", domain.Thresholds{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.TenantID != tenantID {
		t.Errorf("tenant_id mismatch: got %s, want %s", m.TenantID, tenantID)
	}
	if m.ComponentID != componentID {
		t.Errorf("component_id mismatch: got %s, want %s", m.ComponentID, componentID)
	}
	if m.MetricType != domain.MetricCPU {
		t.Errorf("metric_type mismatch: got %s, want %s", m.MetricType, domain.MetricCPU)
	}
	if !m.Enabled {
		t.Error("expected metric to be enabled by default")
	}
	if m.ID == "" {
		t.Error("expected ID to be generated")
	}
}

func TestNewHealthMetric_InvalidTenantID(t *testing.T) {
	_, err := domain.NewHealthMetric("invalid", "0194ffa0-0000-7000-8000-000000000002", domain.MetricCPU, domain.CollectorServer, "CPU", "%", domain.Thresholds{})
	if err == nil {
		t.Fatal("expected error for invalid tenant ID")
	}
}

func TestNewHealthMetric_InvalidComponentID(t *testing.T) {
	_, err := domain.NewHealthMetric("0194ffa0-0000-7000-8000-000000000001", "bad-id", domain.MetricCPU, domain.CollectorServer, "CPU", "%", domain.Thresholds{})
	if err == nil {
		t.Fatal("expected error for invalid component ID")
	}
}

func TestNewHealthMetric_InvalidMetricType(t *testing.T) {
	_, err := domain.NewHealthMetric("0194ffa0-0000-7000-8000-000000000001", "0194ffa0-0000-7000-8000-000000000002", domain.MetricType("bad"), domain.CollectorServer, "CPU", "%", domain.Thresholds{})
	if err == nil {
		t.Fatal("expected error for invalid metric type")
	}
}

func TestNewHealthMetric_InvalidCollectorType(t *testing.T) {
	_, err := domain.NewHealthMetric("0194ffa0-0000-7000-8000-000000000001", "0194ffa0-0000-7000-8000-000000000002", domain.MetricCPU, domain.CollectorType("bad"), "CPU", "%", domain.Thresholds{})
	if err == nil {
		t.Fatal("expected error for invalid collector type")
	}
}

func TestNewHealthMetric_EmptyName(t *testing.T) {
	_, err := domain.NewHealthMetric("0194ffa0-0000-7000-8000-000000000001", "0194ffa0-0000-7000-8000-000000000002", domain.MetricCPU, domain.CollectorServer, "", "%", domain.Thresholds{})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

// --- NewHealthScore Tests ---

func TestNewHealthScore_Healthy(t *testing.T) {
	hs := domain.NewHealthScore("tenant-1", "comp-1", 95, nil)
	if hs.Score != 95 {
		t.Errorf("expected score 95, got %d", hs.Score)
	}
	if hs.Status != domain.StatusHealthy {
		t.Errorf("expected healthy, got %s", hs.Status)
	}
	if len(hs.Reasons) != 0 {
		t.Errorf("expected empty reasons, got %v", hs.Reasons)
	}
}

func TestNewHealthScore_Degraded(t *testing.T) {
	hs := domain.NewHealthScore("tenant-1", "comp-1", 80, []string{"high latency"})
	if hs.Status != domain.StatusDegraded {
		t.Errorf("expected degraded, got %s", hs.Status)
	}
}

func TestNewHealthScore_Critical(t *testing.T) {
	hs := domain.NewHealthScore("tenant-1", "comp-1", 50, []string{"service down"})
	if hs.Status != domain.StatusCritical {
		t.Errorf("expected critical, got %s", hs.Status)
	}
}

func TestNewHealthScore_ClampNegative(t *testing.T) {
	hs := domain.NewHealthScore("tenant-1", "comp-1", -10, nil)
	if hs.Score != 0 {
		t.Errorf("expected clamped to 0, got %d", hs.Score)
	}
}

func TestNewHealthScore_ClampAbove100(t *testing.T) {
	hs := domain.NewHealthScore("tenant-1", "comp-1", 150, nil)
	if hs.Score != 100 {
		t.Errorf("expected clamped to 100, got %d", hs.Score)
	}
}

// --- Anomaly Tests ---

func TestNewAnomaly(t *testing.T) {
	a := domain.NewAnomaly("t1", "c1", "m1", domain.MetricCPU, domain.SeverityHigh, "CPU spike", "Exceeded 3σ", 50.0, 99.0)
	if a.Resolved {
		t.Error("expected anomaly to be unresolved")
	}
	if a.ResolvedAt != nil {
		t.Error("expected resolved_at to be nil")
	}
	if a.Severity != domain.SeverityHigh {
		t.Errorf("expected high severity, got %s", a.Severity)
	}
}

func TestAnomaly_Resolve(t *testing.T) {
	a := domain.NewAnomaly("t1", "c1", "m1", domain.MetricCPU, domain.SeverityMedium, "Spike", "Desc", 50.0, 80.0)
	a.Resolve()

	if !a.Resolved {
		t.Error("expected anomaly to be resolved")
	}
	if a.ResolvedAt == nil {
		t.Error("expected resolved_at to be set")
	}
}

// --- ComponentHealth Tests ---

func TestNewComponentHealth(t *testing.T) {
	ch := domain.NewComponentHealth("t1", "c1", "Backend API", domain.CollectorBackend)
	if ch.Status != domain.StatusUnknown {
		t.Errorf("expected unknown status, got %s", ch.Status)
	}
	if ch.Score != 100 {
		t.Errorf("expected default score 100, got %d", ch.Score)
	}
}

func TestComponentHealth_Update(t *testing.T) {
	ch := domain.NewComponentHealth("t1", "c1", "Backend API", domain.CollectorBackend)
	ch.Update(75, domain.StatusDegraded)

	if ch.Score != 75 {
		t.Errorf("expected score 75, got %d", ch.Score)
	}
	if ch.Status != domain.StatusDegraded {
		t.Errorf("expected degraded, got %s", ch.Status)
	}
}
