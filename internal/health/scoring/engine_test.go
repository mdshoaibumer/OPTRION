package scoring_test

import (
	"testing"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
	"github.com/optrion/optrion/internal/health/scoring"
)

func TestEngine_Compute_HealthyBackend(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 1},
			{MetricType: domain.MetricResponseTime, Value: 200},
			{MetricType: domain.MetricErrorRate, Value: 0},
			{MetricType: domain.MetricThroughput, Value: 100},
		},
	}

	sr := engine.Compute(result)
	if sr.Score != 100 {
		t.Errorf("expected score 100, got %d", sr.Score)
	}
	if sr.Status != domain.StatusHealthy {
		t.Errorf("expected healthy, got %s", sr.Status)
	}
	if len(sr.Reasons) != 0 {
		t.Errorf("expected no reasons, got %v", sr.Reasons)
	}
}

func TestEngine_Compute_BackendDown(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 0},
		},
	}

	sr := engine.Compute(result)
	if sr.Score != 60 {
		t.Errorf("expected score 60 (100-40), got %d", sr.Score)
	}
	if sr.Status != domain.StatusCritical {
		t.Errorf("expected critical, got %s", sr.Status)
	}
	if len(sr.Reasons) != 1 || sr.Reasons[0] != "Backend is down" {
		t.Errorf("expected reason 'Backend is down', got %v", sr.Reasons)
	}
}

func TestEngine_Compute_BackendSlowResponse(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 1},
			{MetricType: domain.MetricResponseTime, Value: 2500},
		},
	}

	sr := engine.Compute(result)
	// Response time > 2000ms → penalty 20 (first matching rule for response_time)
	if sr.Score != 80 {
		t.Errorf("expected score 80 (100-20), got %d", sr.Score)
	}
	if sr.Status != domain.StatusDegraded {
		t.Errorf("expected degraded, got %s", sr.Status)
	}
}

func TestEngine_Compute_BackendModerateResponse(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 1},
			{MetricType: domain.MetricResponseTime, Value: 1500},
		},
	}

	sr := engine.Compute(result)
	// Response time > 1000ms (but not > 2000ms) → penalty 10
	// Wait: rules are ordered, first matching applies. >2000 is checked first, 1500 < 2000, so check >1000 next
	if sr.Score != 90 {
		t.Errorf("expected score 90 (100-10), got %d", sr.Score)
	}
}

func TestEngine_Compute_BackendMultiplePenalties(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 0},
			{MetricType: domain.MetricResponseTime, Value: 3000},
			{MetricType: domain.MetricErrorRate, Value: 5},
		},
	}

	sr := engine.Compute(result)
	// Availability=0 → -40, ResponseTime>2000 → -20, ErrorRate>0 → -15 = 100-75 = 25
	if sr.Score != 25 {
		t.Errorf("expected score 25 (100-40-20-15), got %d", sr.Score)
	}
	if sr.Status != domain.StatusCritical {
		t.Errorf("expected critical, got %s", sr.Status)
	}
	if len(sr.Reasons) != 3 {
		t.Errorf("expected 3 reasons, got %d", len(sr.Reasons))
	}
}

func TestEngine_Compute_PostgresConnectionLost(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-db",
		CollectorType: domain.CollectorPostgres,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricConnectionStatus, Value: 0},
		},
	}

	sr := engine.Compute(result)
	if sr.Score != 50 {
		t.Errorf("expected score 50 (100-50), got %d", sr.Score)
	}
	if sr.Status != domain.StatusCritical {
		t.Errorf("expected critical, got %s", sr.Status)
	}
}

func TestEngine_Compute_PostgresSlowQueries(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-db",
		CollectorType: domain.CollectorPostgres,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricConnectionStatus, Value: 1},
			{MetricType: domain.MetricQueryLatency, Value: 200},
			{MetricType: domain.MetricSlowQueries, Value: 3},
			{MetricType: domain.MetricDeadlocks, Value: 0},
		},
	}

	sr := engine.Compute(result)
	// QueryLatency 200 > 100ms → penalty 10 (first match, >500 doesn't match)
	// SlowQueries 3 > 0 → penalty 10
	// Deadlocks 0 ≤ 0 → no match
	if sr.Score != 80 {
		t.Errorf("expected score 80 (100-10-10), got %d", sr.Score)
	}
}

func TestEngine_Compute_RedisDown(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-redis",
		CollectorType: domain.CollectorRedis,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 0},
			{MetricType: domain.MetricMemoryUsage, Value: 95},
		},
	}

	sr := engine.Compute(result)
	// Availability=0 → -30, MemoryUsage>90 → -20 = 100-50 = 50
	if sr.Score != 50 {
		t.Errorf("expected score 50 (100-30-20), got %d", sr.Score)
	}
}

func TestEngine_Compute_ServerHighCPU(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-server",
		CollectorType: domain.CollectorServer,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricCPU, Value: 95},
			{MetricType: domain.MetricRAM, Value: 60},
		},
	}

	sr := engine.Compute(result)
	// CPU>90 → -20, RAM ok
	if sr.Score != 80 {
		t.Errorf("expected score 80 (100-20), got %d", sr.Score)
	}
}

func TestEngine_Compute_UnknownCollectorType(t *testing.T) {
	engine := scoring.NewEngine()
	result := &port.CollectorResult{
		ComponentID:   "comp-x",
		CollectorType: domain.CollectorType("unknown"),
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricCPU, Value: 99},
		},
	}

	sr := engine.Compute(result)
	if sr.Score != 100 {
		t.Errorf("expected score 100 for unknown collector, got %d", sr.Score)
	}
}

func TestEngine_Compute_ScoreNeverBelowZero(t *testing.T) {
	engine := scoring.NewEngine()
	// Custom config with extreme penalties
	engine.SetConfig(domain.CollectorBackend, scoring.ScoringConfig{
		BaseScore: 100,
		Rules: []scoring.Rule{
			{MetricType: domain.MetricAvailability, Condition: scoring.Condition{Operator: "eq", Value: 0}, Penalty: 60, ReasonFormat: "Down"},
			{MetricType: domain.MetricErrorRate, Condition: scoring.Condition{Operator: "gt", Value: 0}, Penalty: 60, ReasonFormat: "Errors"},
		},
	})

	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 0},
			{MetricType: domain.MetricErrorRate, Value: 50},
		},
	}

	sr := engine.Compute(result)
	if sr.Score != 0 {
		t.Errorf("expected score clamped to 0, got %d", sr.Score)
	}
}

func TestEngine_Compute_OnlyFirstMatchingRulePerMetric(t *testing.T) {
	engine := scoring.NewEngine()
	// Backend config: response_time has two rules (>2000 → -20, >1000 → -10)
	// If value is 3000, only the first rule (>2000) should apply, not both
	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 1},
			{MetricType: domain.MetricResponseTime, Value: 3000},
		},
	}

	sr := engine.Compute(result)
	// Only -20 for response time (not -20 AND -10)
	if sr.Score != 80 {
		t.Errorf("expected score 80 (only first matching rule), got %d", sr.Score)
	}
}

func TestEngine_SetConfig(t *testing.T) {
	engine := scoring.NewEngine()
	custom := scoring.ScoringConfig{
		BaseScore: 50,
		Rules: []scoring.Rule{
			{MetricType: domain.MetricCPU, Condition: scoring.Condition{Operator: "gt", Value: 50}, Penalty: 25, ReasonFormat: "High CPU"},
		},
	}
	engine.SetConfig(domain.CollectorServer, custom)

	result := &port.CollectorResult{
		ComponentID:   "comp-server",
		CollectorType: domain.CollectorServer,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricCPU, Value: 60},
		},
	}

	sr := engine.Compute(result)
	if sr.Score != 25 {
		t.Errorf("expected score 25 (50-25), got %d", sr.Score)
	}
}

// --- Condition Tests ---

func TestCondition_Matches(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		value    float64
		input    float64
		expected bool
	}{
		{"gt true", "gt", 10, 15, true},
		{"gt false", "gt", 10, 10, false},
		{"gte true equal", "gte", 10, 10, true},
		{"gte true above", "gte", 10, 11, true},
		{"gte false", "gte", 10, 9, false},
		{"lt true", "lt", 10, 5, true},
		{"lt false", "lt", 10, 10, false},
		{"lte true equal", "lte", 10, 10, true},
		{"lte true below", "lte", 10, 9, true},
		{"lte false", "lte", 10, 11, false},
		{"eq true", "eq", 0, 0, true},
		{"eq false", "eq", 0, 1, false},
		{"unknown operator", "between", 10, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := scoring.Condition{Operator: tt.op, Value: tt.value}
			got := c.Matches(tt.input)
			if got != tt.expected {
				t.Errorf("Condition{%s, %v}.Matches(%v) = %v, want %v", tt.op, tt.value, tt.input, got, tt.expected)
			}
		})
	}
}
