package scoring

import (
	"fmt"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
)

// Rule defines a scoring penalty for a specific metric condition.
type Rule struct {
	MetricType   domain.MetricType
	Condition    Condition
	Penalty      int
	ReasonFormat string
}

// Condition defines when a rule should apply.
type Condition struct {
	Operator string // "gt", "lt", "eq", "gte", "lte"
	Value    float64
}

// Matches checks if a metric value satisfies this condition.
func (c Condition) Matches(value float64) bool {
	switch c.Operator {
	case "gt":
		return value > c.Value
	case "gte":
		return value >= c.Value
	case "lt":
		return value < c.Value
	case "lte":
		return value <= c.Value
	case "eq":
		return value == c.Value
	default:
		return false
	}
}

// ScoringConfig holds the scoring rules for a component.
type ScoringConfig struct {
	BaseScore int
	Rules     []Rule
}

// DefaultBackendConfig returns the default scoring rules for a backend component.
func DefaultBackendConfig() ScoringConfig {
	return ScoringConfig{
		BaseScore: 100,
		Rules: []Rule{
			{MetricType: domain.MetricAvailability, Condition: Condition{"eq", 0}, Penalty: 40, ReasonFormat: "Backend is down"},
			{MetricType: domain.MetricResponseTime, Condition: Condition{"gt", 2000}, Penalty: 20, ReasonFormat: "Response time %.0fms exceeds 2000ms threshold"},
			{MetricType: domain.MetricResponseTime, Condition: Condition{"gt", 1000}, Penalty: 10, ReasonFormat: "Response time %.0fms exceeds 1000ms threshold"},
			{MetricType: domain.MetricErrorRate, Condition: Condition{"gt", 0}, Penalty: 15, ReasonFormat: "Error rate detected"},
		},
	}
}

// DefaultPostgresConfig returns the default scoring rules for PostgreSQL.
func DefaultPostgresConfig() ScoringConfig {
	return ScoringConfig{
		BaseScore: 100,
		Rules: []Rule{
			{MetricType: domain.MetricConnectionStatus, Condition: Condition{"eq", 0}, Penalty: 50, ReasonFormat: "Database connection lost"},
			{MetricType: domain.MetricQueryLatency, Condition: Condition{"gt", 500}, Penalty: 15, ReasonFormat: "Query latency %.0fms exceeds 500ms threshold"},
			{MetricType: domain.MetricQueryLatency, Condition: Condition{"gt", 100}, Penalty: 10, ReasonFormat: "Query latency %.0fms exceeds 100ms threshold"},
			{MetricType: domain.MetricSlowQueries, Condition: Condition{"gt", 0}, Penalty: 10, ReasonFormat: "%.0f slow queries detected"},
			{MetricType: domain.MetricDeadlocks, Condition: Condition{"gt", 0}, Penalty: 20, ReasonFormat: "%.0f deadlocks detected"},
			{MetricType: domain.MetricPoolHealth, Condition: Condition{"gt", 80}, Penalty: 15, ReasonFormat: "Connection pool utilization at %.0f%%"},
			{MetricType: domain.MetricIndexUsage, Condition: Condition{"lt", 90}, Penalty: 5, ReasonFormat: "Index usage ratio at %.0f%%"},
		},
	}
}

// DefaultRedisConfig returns the default scoring rules for Redis.
func DefaultRedisConfig() ScoringConfig {
	return ScoringConfig{
		BaseScore: 100,
		Rules: []Rule{
			{MetricType: domain.MetricAvailability, Condition: Condition{"eq", 0}, Penalty: 30, ReasonFormat: "Redis is down"},
			{MetricType: domain.MetricMemoryUsage, Condition: Condition{"gt", 90}, Penalty: 20, ReasonFormat: "Redis memory usage at %.0f%%"},
			{MetricType: domain.MetricMemoryUsage, Condition: Condition{"gt", 75}, Penalty: 10, ReasonFormat: "Redis memory usage at %.0f%%"},
			{MetricType: domain.MetricHitRatio, Condition: Condition{"lt", 80}, Penalty: 10, ReasonFormat: "Cache hit ratio at %.0f%%"},
			{MetricType: domain.MetricEvictions, Condition: Condition{"gt", 100}, Penalty: 15, ReasonFormat: "%.0f evictions detected"},
		},
	}
}

// DefaultServerConfig returns the default scoring rules for server resources.
func DefaultServerConfig() ScoringConfig {
	return ScoringConfig{
		BaseScore: 100,
		Rules: []Rule{
			{MetricType: domain.MetricCPU, Condition: Condition{"gt", 90}, Penalty: 20, ReasonFormat: "CPU usage at %.0f%%"},
			{MetricType: domain.MetricCPU, Condition: Condition{"gt", 75}, Penalty: 10, ReasonFormat: "CPU usage at %.0f%%"},
			{MetricType: domain.MetricRAM, Condition: Condition{"gt", 90}, Penalty: 20, ReasonFormat: "RAM usage at %.0f%%"},
			{MetricType: domain.MetricRAM, Condition: Condition{"gt", 75}, Penalty: 10, ReasonFormat: "RAM usage at %.0f%%"},
		},
	}
}

// Engine computes health scores based on collected metrics.
type Engine struct {
	configs map[domain.CollectorType]ScoringConfig
}

// NewEngine creates a scoring engine with default configurations.
func NewEngine() *Engine {
	return &Engine{
		configs: map[domain.CollectorType]ScoringConfig{
			domain.CollectorBackend:  DefaultBackendConfig(),
			domain.CollectorPostgres: DefaultPostgresConfig(),
			domain.CollectorRedis:    DefaultRedisConfig(),
			domain.CollectorServer:   DefaultServerConfig(),
		},
	}
}

// SetConfig overrides the scoring config for a collector type.
func (e *Engine) SetConfig(collectorType domain.CollectorType, config ScoringConfig) {
	e.configs[collectorType] = config
}

// ScoreResult holds the output of a scoring computation.
type ScoreResult struct {
	Score   int
	Status  domain.MetricStatus
	Reasons []string
}

// Compute calculates a health score from collector results.
func (e *Engine) Compute(result *port.CollectorResult) ScoreResult {
	cfg, ok := e.configs[result.CollectorType]
	if !ok {
		return ScoreResult{Score: 100, Status: domain.StatusHealthy, Reasons: []string{}}
	}

	score := cfg.BaseScore
	reasons := make([]string, 0)

	// Index metrics by type for rule evaluation
	metricValues := make(map[domain.MetricType]float64)
	for _, m := range result.Metrics {
		metricValues[m.MetricType] = m.Value
	}

	// Evaluate rules (only apply the first matching rule per metric type to avoid double-counting)
	appliedMetrics := make(map[domain.MetricType]bool)
	for _, rule := range cfg.Rules {
		if appliedMetrics[rule.MetricType] {
			continue
		}

		value, exists := metricValues[rule.MetricType]
		if !exists {
			continue
		}

		if rule.Condition.Matches(value) {
			score -= rule.Penalty
			reason := rule.ReasonFormat
			if containsFormatVerb(reason) {
				reason = fmt.Sprintf(rule.ReasonFormat, value)
			}
			reasons = append(reasons, reason)
			appliedMetrics[rule.MetricType] = true
		}
	}

	if score < 0 {
		score = 0
	}

	status := domain.StatusHealthy
	if score < 70 {
		status = domain.StatusCritical
	} else if score < 90 {
		status = domain.StatusDegraded
	}

	return ScoreResult{
		Score:   score,
		Status:  status,
		Reasons: reasons,
	}
}

func containsFormatVerb(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '%' && s[i+1] != '%' {
			return true
		}
	}
	return false
}
