package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/health/domain"
)

// HealthCheckConfigRepository implements port.HealthCheckConfigRepository using PostgreSQL.
type HealthCheckConfigRepository struct {
	pool *pgxpool.Pool
}

// NewHealthCheckConfigRepository creates a new health check config repository.
func NewHealthCheckConfigRepository(pool *pgxpool.Pool) *HealthCheckConfigRepository {
	return &HealthCheckConfigRepository{pool: pool}
}

// Upsert creates or updates a health check configuration.
func (r *HealthCheckConfigRepository) Upsert(ctx context.Context, config *domain.HealthCheckConfig) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO health_check_configs (id, tenant_id, component_id, check_interval_ms, timeout_ms, retries, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (component_id) DO UPDATE SET
		   check_interval_ms = EXCLUDED.check_interval_ms,
		   timeout_ms = EXCLUDED.timeout_ms,
		   retries = EXCLUDED.retries,
		   enabled = EXCLUDED.enabled,
		   updated_at = EXCLUDED.updated_at`,
		config.ID, config.TenantID, config.ComponentID,
		config.CheckInterval.Milliseconds(), config.Timeout.Milliseconds(),
		config.Retries, config.Enabled,
		config.CreatedAt, config.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upserting health check config: %w", err)
	}
	return nil
}

// GetByComponent retrieves the health check config for a component.
func (r *HealthCheckConfigRepository) GetByComponent(ctx context.Context, componentID string) (*domain.HealthCheckConfig, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, component_id, check_interval_ms, timeout_ms, retries, enabled, created_at, updated_at
		 FROM health_check_configs WHERE component_id = $1`,
		componentID,
	)

	var config domain.HealthCheckConfig
	var intervalMS, timeoutMS int64
	err := row.Scan(&config.ID, &config.TenantID, &config.ComponentID,
		&intervalMS, &timeoutMS,
		&config.Retries, &config.Enabled, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying health check config: %w", err)
	}
	config.CheckInterval = time.Duration(intervalMS) * time.Millisecond
	config.Timeout = time.Duration(timeoutMS) * time.Millisecond
	return &config, nil
}

// ListByTenant retrieves all health check configs for a tenant.
func (r *HealthCheckConfigRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.HealthCheckConfig, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, component_id, check_interval_ms, timeout_ms, retries, enabled, created_at, updated_at
		 FROM health_check_configs WHERE tenant_id = $1
		 ORDER BY created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing health check configs: %w", err)
	}
	defer rows.Close()

	var configs []*domain.HealthCheckConfig
	for rows.Next() {
		var c domain.HealthCheckConfig
		var intervalMS, timeoutMS int64
		if err := rows.Scan(&c.ID, &c.TenantID, &c.ComponentID,
			&intervalMS, &timeoutMS,
			&c.Retries, &c.Enabled, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning health check config: %w", err)
		}
		c.CheckInterval = time.Duration(intervalMS) * time.Millisecond
		c.Timeout = time.Duration(timeoutMS) * time.Millisecond
		configs = append(configs, &c)
	}
	return configs, rows.Err()
}
