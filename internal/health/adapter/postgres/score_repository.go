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

// ScoreRepository implements port.HealthScoreRepository.
type ScoreRepository struct {
	pool *pgxpool.Pool
}

// NewScoreRepository creates a new health score repository.
func NewScoreRepository(pool *pgxpool.Pool) *ScoreRepository {
	return &ScoreRepository{pool: pool}
}

func (r *ScoreRepository) Create(ctx context.Context, score *domain.HealthScore) error {
	reasonsJSON, err := json.Marshal(score.Reasons)
	if err != nil {
		return fmt.Errorf("marshaling reasons: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO health_scores (id, tenant_id, component_id, score, status, reasons, computed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		score.ID, score.TenantID, score.ComponentID, score.Score,
		score.Status, reasonsJSON, score.ComputedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting health score: %w", err)
	}
	return nil
}

func (r *ScoreRepository) GetLatestByComponent(ctx context.Context, componentID string) (*domain.HealthScore, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, component_id, score, status, reasons, computed_at
		 FROM health_scores
		 WHERE component_id = $1
		 ORDER BY computed_at DESC
		 LIMIT 1`, componentID)

	return r.scanScore(row)
}

func (r *ScoreRepository) ListByTenant(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.HealthScore, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, component_id, score, status, reasons, computed_at
		 FROM health_scores
		 WHERE tenant_id = $1 AND computed_at >= $2 AND computed_at <= $3
		 ORDER BY computed_at DESC
		 LIMIT $4`,
		tenantID, from, to, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying scores by tenant: %w", err)
	}
	defer rows.Close()
	return r.scanScores(rows)
}

func (r *ScoreRepository) ListByComponent(ctx context.Context, componentID string, from, to time.Time, limit int) ([]*domain.HealthScore, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, component_id, score, status, reasons, computed_at
		 FROM health_scores
		 WHERE component_id = $1 AND computed_at >= $2 AND computed_at <= $3
		 ORDER BY computed_at DESC
		 LIMIT $4`,
		componentID, from, to, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying scores by component: %w", err)
	}
	defer rows.Close()
	return r.scanScores(rows)
}

func (r *ScoreRepository) scanScore(row pgx.Row) (*domain.HealthScore, error) {
	var s domain.HealthScore
	var reasonsJSON []byte
	err := row.Scan(&s.ID, &s.TenantID, &s.ComponentID, &s.Score, &s.Status, &reasonsJSON, &s.ComputedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning health score: %w", err)
	}
	if err := json.Unmarshal(reasonsJSON, &s.Reasons); err != nil {
		return nil, fmt.Errorf("unmarshaling reasons: %w", err)
	}
	return &s, nil
}

func (r *ScoreRepository) scanScores(rows pgx.Rows) ([]*domain.HealthScore, error) {
	var scores []*domain.HealthScore
	for rows.Next() {
		var s domain.HealthScore
		var reasonsJSON []byte
		err := rows.Scan(&s.ID, &s.TenantID, &s.ComponentID, &s.Score, &s.Status, &reasonsJSON, &s.ComputedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning score row: %w", err)
		}
		if err := json.Unmarshal(reasonsJSON, &s.Reasons); err != nil {
			return nil, fmt.Errorf("unmarshaling reasons: %w", err)
		}
		scores = append(scores, &s)
	}
	return scores, rows.Err()
}
