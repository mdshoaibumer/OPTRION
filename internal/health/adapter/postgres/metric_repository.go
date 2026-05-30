package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/health/domain"
)

// HealthMetricRepository implements port.HealthMetricRepository.
type HealthMetricRepository struct {
	pool *pgxpool.Pool
}

// NewHealthMetricRepository creates a new PostgreSQL-backed health metric repository.
func NewHealthMetricRepository(pool *pgxpool.Pool) *HealthMetricRepository {
	return &HealthMetricRepository{pool: pool}
}

func (r *HealthMetricRepository) Create(ctx context.Context, metric *domain.HealthMetric) error {
	thresholdsJSON, err := json.Marshal(metric.Thresholds)
	if err != nil {
		return fmt.Errorf("marshaling thresholds: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO health_metrics (id, tenant_id, component_id, metric_type, collector_type, name, unit, thresholds, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		metric.ID, metric.TenantID, metric.ComponentID, metric.MetricType,
		metric.CollectorType, metric.Name, metric.Unit, thresholdsJSON,
		metric.Enabled, metric.CreatedAt, metric.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting health metric: %w", err)
	}
	return nil
}

func (r *HealthMetricRepository) GetByID(ctx context.Context, id string) (*domain.HealthMetric, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, component_id, metric_type, collector_type, name, unit, thresholds, enabled, created_at, updated_at
		 FROM health_metrics WHERE id = $1`, id)

	return r.scanMetric(row)
}

func (r *HealthMetricRepository) ListByComponent(ctx context.Context, componentID string) ([]*domain.HealthMetric, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, component_id, metric_type, collector_type, name, unit, thresholds, enabled, created_at, updated_at
		 FROM health_metrics WHERE component_id = $1 ORDER BY metric_type`, componentID)
	if err != nil {
		return nil, fmt.Errorf("querying metrics by component: %w", err)
	}
	defer rows.Close()
	return r.scanMetrics(rows)
}

func (r *HealthMetricRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.HealthMetric, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, component_id, metric_type, collector_type, name, unit, thresholds, enabled, created_at, updated_at
		 FROM health_metrics WHERE tenant_id = $1 ORDER BY component_id, metric_type`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("querying metrics by tenant: %w", err)
	}
	defer rows.Close()
	return r.scanMetrics(rows)
}

func (r *HealthMetricRepository) ListEnabled(ctx context.Context, tenantID string) ([]*domain.HealthMetric, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, component_id, metric_type, collector_type, name, unit, thresholds, enabled, created_at, updated_at
		 FROM health_metrics WHERE tenant_id = $1 AND enabled = TRUE ORDER BY component_id, metric_type`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("querying enabled metrics: %w", err)
	}
	defer rows.Close()
	return r.scanMetrics(rows)
}

func (r *HealthMetricRepository) Update(ctx context.Context, metric *domain.HealthMetric) error {
	thresholdsJSON, err := json.Marshal(metric.Thresholds)
	if err != nil {
		return fmt.Errorf("marshaling thresholds: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`UPDATE health_metrics SET name = $1, unit = $2, thresholds = $3, enabled = $4, updated_at = $5 WHERE id = $6`,
		metric.Name, metric.Unit, thresholdsJSON, metric.Enabled, time.Now().UTC(), metric.ID,
	)
	if err != nil {
		return fmt.Errorf("updating health metric: %w", err)
	}
	return nil
}

func (r *HealthMetricRepository) Upsert(ctx context.Context, metric *domain.HealthMetric) error {
	thresholdsJSON, err := json.Marshal(metric.Thresholds)
	if err != nil {
		return fmt.Errorf("marshaling thresholds: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO health_metrics (id, tenant_id, component_id, metric_type, collector_type, name, unit, thresholds, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (component_id, metric_type) DO UPDATE SET
			name = EXCLUDED.name,
			unit = EXCLUDED.unit,
			thresholds = EXCLUDED.thresholds,
			enabled = EXCLUDED.enabled,
			updated_at = NOW()`,
		metric.ID, metric.TenantID, metric.ComponentID, metric.MetricType,
		metric.CollectorType, metric.Name, metric.Unit, thresholdsJSON,
		metric.Enabled, metric.CreatedAt, metric.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upserting health metric: %w", err)
	}
	return nil
}

func (r *HealthMetricRepository) scanMetric(row pgx.Row) (*domain.HealthMetric, error) {
	var m domain.HealthMetric
	var thresholdsJSON []byte

	err := row.Scan(
		&m.ID, &m.TenantID, &m.ComponentID, &m.MetricType, &m.CollectorType,
		&m.Name, &m.Unit, &thresholdsJSON, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("health metric not found")
		}
		return nil, fmt.Errorf("scanning health metric: %w", err)
	}

	if err := json.Unmarshal(thresholdsJSON, &m.Thresholds); err != nil {
		return nil, fmt.Errorf("unmarshaling thresholds: %w", err)
	}
	return &m, nil
}

func (r *HealthMetricRepository) scanMetrics(rows pgx.Rows) ([]*domain.HealthMetric, error) {
	var metrics []*domain.HealthMetric
	for rows.Next() {
		var m domain.HealthMetric
		var thresholdsJSON []byte
		err := rows.Scan(
			&m.ID, &m.TenantID, &m.ComponentID, &m.MetricType, &m.CollectorType,
			&m.Name, &m.Unit, &thresholdsJSON, &m.Enabled, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning health metric row: %w", err)
		}
		if err := json.Unmarshal(thresholdsJSON, &m.Thresholds); err != nil {
			return nil, fmt.Errorf("unmarshaling thresholds: %w", err)
		}
		metrics = append(metrics, &m)
	}
	return metrics, rows.Err()
}
