package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
)

// PostgresCollector monitors a PostgreSQL database.
type PostgresCollector struct {
	tenantID    string
	componentID string
	pool        *pgxpool.Pool
}

// NewPostgresCollector creates a new PostgreSQL health collector.
func NewPostgresCollector(tenantID, componentID string, pool *pgxpool.Pool) *PostgresCollector {
	return &PostgresCollector{
		tenantID:    tenantID,
		componentID: componentID,
		pool:        pool,
	}
}

func (c *PostgresCollector) Type() domain.CollectorType { return domain.CollectorPostgres }
func (c *PostgresCollector) ComponentID() string        { return c.componentID }
func (c *PostgresCollector) TenantID() string           { return c.tenantID }

func (c *PostgresCollector) Collect(ctx context.Context) (*port.CollectorResult, error) {
	result := &port.CollectorResult{
		ComponentID:   c.componentID,
		CollectorType: domain.CollectorPostgres,
		Metrics:       make([]port.MetricReading, 0, 7),
	}

	// 1. Connection status (ping)
	start := time.Now()
	err := c.pool.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		result.Metrics = append(result.Metrics,
			port.MetricReading{MetricType: domain.MetricConnectionStatus, Value: 0, Labels: map[string]string{"error": err.Error()}},
		)
		result.Error = err
		return result, nil
	}

	result.Metrics = append(result.Metrics,
		port.MetricReading{MetricType: domain.MetricConnectionStatus, Value: 1, Labels: map[string]string{}},
	)

	// 2. Query latency (from ping)
	result.Metrics = append(result.Metrics,
		port.MetricReading{MetricType: domain.MetricQueryLatency, Value: latency.Seconds() * 1000, Labels: map[string]string{"unit": "ms"}},
	)

	// 3. Active connections
	var activeConns int
	err = c.pool.QueryRow(ctx, `SELECT count(*) FROM pg_stat_activity WHERE state = 'active'`).Scan(&activeConns)
	if err == nil {
		result.Metrics = append(result.Metrics,
			port.MetricReading{MetricType: domain.MetricActiveConnections, Value: float64(activeConns), Labels: map[string]string{}},
		)
	}

	// 4. Slow queries (queries running > 5 seconds)
	var slowQueries int
	err = c.pool.QueryRow(ctx, `
		SELECT count(*) FROM pg_stat_activity
		WHERE state = 'active'
		AND now() - query_start > interval '5 seconds'
		AND query NOT LIKE '%pg_stat_activity%'
	`).Scan(&slowQueries)
	if err == nil {
		result.Metrics = append(result.Metrics,
			port.MetricReading{MetricType: domain.MetricSlowQueries, Value: float64(slowQueries), Labels: map[string]string{"threshold": "5s"}},
		)
	}

	// 5. Deadlocks
	var deadlocks int64
	err = c.pool.QueryRow(ctx, `SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()`).Scan(&deadlocks)
	if err == nil {
		result.Metrics = append(result.Metrics,
			port.MetricReading{MetricType: domain.MetricDeadlocks, Value: float64(deadlocks), Labels: map[string]string{}},
		)
	}

	// 6. Index usage ratio
	var indexRatio float64
	err = c.pool.QueryRow(ctx, `
		SELECT COALESCE(
			round(sum(idx_blks_hit)::numeric / NULLIF(sum(idx_blks_hit + idx_blks_read), 0) * 100, 2),
			100
		)
		FROM pg_statio_user_indexes
	`).Scan(&indexRatio)
	if err == nil {
		result.Metrics = append(result.Metrics,
			port.MetricReading{MetricType: domain.MetricIndexUsage, Value: indexRatio, Labels: map[string]string{"unit": "percent"}},
		)
	}

	// 7. Connection pool health
	stat := c.pool.Stat()
	poolUtilization := 0.0
	if stat.MaxConns() > 0 {
		poolUtilization = float64(stat.AcquiredConns()) / float64(stat.MaxConns()) * 100
	}
	result.Metrics = append(result.Metrics,
		port.MetricReading{
			MetricType: domain.MetricPoolHealth,
			Value:      poolUtilization,
			Labels: map[string]string{
				"total":    fmt.Sprintf("%d", stat.TotalConns()),
				"idle":     fmt.Sprintf("%d", stat.IdleConns()),
				"acquired": fmt.Sprintf("%d", stat.AcquiredConns()),
				"max":      fmt.Sprintf("%d", stat.MaxConns()),
				"unit":     "percent",
			},
		},
	)

	return result, nil
}
