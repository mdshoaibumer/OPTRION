package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/health/domain"
)

// ComponentHealthRepository implements port.ComponentHealthRepository.
type ComponentHealthRepository struct {
	pool *pgxpool.Pool
}

// NewComponentHealthRepository creates a new component health repository.
func NewComponentHealthRepository(pool *pgxpool.Pool) *ComponentHealthRepository {
	return &ComponentHealthRepository{pool: pool}
}

func (r *ComponentHealthRepository) Upsert(ctx context.Context, status *domain.ComponentHealth) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO component_status (id, tenant_id, component_id, component_name, collector_type, status, score, last_check_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (component_id) DO UPDATE SET
			status = EXCLUDED.status,
			score = EXCLUDED.score,
			last_check_at = EXCLUDED.last_check_at,
			updated_at = NOW()`,
		status.ID, status.TenantID, status.ComponentID, status.ComponentName,
		status.CollectorType, status.Status, status.Score, status.LastCheckAt, status.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upserting component status: %w", err)
	}
	return nil
}

func (r *ComponentHealthRepository) GetByComponent(ctx context.Context, componentID string) (*domain.ComponentHealth, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, component_id, component_name, collector_type, status, score, last_check_at, updated_at
		 FROM component_status WHERE component_id = $1`, componentID)

	var ch domain.ComponentHealth
	err := row.Scan(
		&ch.ID, &ch.TenantID, &ch.ComponentID, &ch.ComponentName,
		&ch.CollectorType, &ch.Status, &ch.Score, &ch.LastCheckAt, &ch.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning component status: %w", err)
	}
	return &ch, nil
}

func (r *ComponentHealthRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.ComponentHealth, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, component_id, component_name, collector_type, status, score, last_check_at, updated_at
		 FROM component_status WHERE tenant_id = $1 ORDER BY component_name`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("querying component statuses: %w", err)
	}
	defer rows.Close()

	var statuses []*domain.ComponentHealth
	for rows.Next() {
		var ch domain.ComponentHealth
		err := rows.Scan(
			&ch.ID, &ch.TenantID, &ch.ComponentID, &ch.ComponentName,
			&ch.CollectorType, &ch.Status, &ch.Score, &ch.LastCheckAt, &ch.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning component status row: %w", err)
		}
		statuses = append(statuses, &ch)
	}
	return statuses, rows.Err()
}
