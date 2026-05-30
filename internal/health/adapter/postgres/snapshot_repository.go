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

// SnapshotRepository implements port.MetricSnapshotRepository.
type SnapshotRepository struct {
	pool *pgxpool.Pool
}

// NewSnapshotRepository creates a new snapshot repository.
func NewSnapshotRepository(pool *pgxpool.Pool) *SnapshotRepository {
	return &SnapshotRepository{pool: pool}
}

func (r *SnapshotRepository) Create(ctx context.Context, snapshot *domain.MetricSnapshot) error {
	labelsJSON, err := json.Marshal(snapshot.Labels)
	if err != nil {
		return fmt.Errorf("marshaling labels: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO metric_snapshots (id, tenant_id, metric_id, value, status, labels, collected_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		snapshot.ID, snapshot.TenantID, snapshot.MetricID,
		snapshot.Value, snapshot.Status, labelsJSON, snapshot.CollectedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting metric snapshot: %w", err)
	}
	return nil
}

func (r *SnapshotRepository) CreateBatch(ctx context.Context, snapshots []*domain.MetricSnapshot) error {
	batch := &pgx.Batch{}
	for _, s := range snapshots {
		labelsJSON, err := json.Marshal(s.Labels)
		if err != nil {
			return fmt.Errorf("marshaling labels: %w", err)
		}
		batch.Queue(
			`INSERT INTO metric_snapshots (id, tenant_id, metric_id, value, status, labels, collected_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			s.ID, s.TenantID, s.MetricID, s.Value, s.Status, labelsJSON, s.CollectedAt,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range snapshots {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("batch inserting snapshot: %w", err)
		}
	}
	return nil
}

func (r *SnapshotRepository) ListByMetric(ctx context.Context, metricID string, from, to time.Time, limit int) ([]*domain.MetricSnapshot, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, metric_id, value, status, labels, collected_at
		 FROM metric_snapshots
		 WHERE metric_id = $1 AND collected_at >= $2 AND collected_at <= $3
		 ORDER BY collected_at DESC
		 LIMIT $4`,
		metricID, from, to, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying snapshots by metric: %w", err)
	}
	defer rows.Close()
	return r.scanSnapshots(rows)
}

func (r *SnapshotRepository) ListByTenant(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.MetricSnapshot, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, metric_id, value, status, labels, collected_at
		 FROM metric_snapshots
		 WHERE tenant_id = $1 AND collected_at >= $2 AND collected_at <= $3
		 ORDER BY collected_at DESC
		 LIMIT $4`,
		tenantID, from, to, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying snapshots by tenant: %w", err)
	}
	defer rows.Close()
	return r.scanSnapshots(rows)
}

func (r *SnapshotRepository) GetLatestByMetric(ctx context.Context, metricID string) (*domain.MetricSnapshot, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, metric_id, value, status, labels, collected_at
		 FROM metric_snapshots
		 WHERE metric_id = $1
		 ORDER BY collected_at DESC
		 LIMIT 1`, metricID)

	var s domain.MetricSnapshot
	var labelsJSON []byte
	err := row.Scan(&s.ID, &s.TenantID, &s.MetricID, &s.Value, &s.Status, &labelsJSON, &s.CollectedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning latest snapshot: %w", err)
	}
	if err := json.Unmarshal(labelsJSON, &s.Labels); err != nil {
		return nil, fmt.Errorf("unmarshaling labels: %w", err)
	}
	return &s, nil
}

func (r *SnapshotRepository) scanSnapshots(rows pgx.Rows) ([]*domain.MetricSnapshot, error) {
	var snapshots []*domain.MetricSnapshot
	for rows.Next() {
		var s domain.MetricSnapshot
		var labelsJSON []byte
		err := rows.Scan(&s.ID, &s.TenantID, &s.MetricID, &s.Value, &s.Status, &labelsJSON, &s.CollectedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning snapshot row: %w", err)
		}
		if err := json.Unmarshal(labelsJSON, &s.Labels); err != nil {
			return nil, fmt.Errorf("unmarshaling labels: %w", err)
		}
		snapshots = append(snapshots, &s)
	}
	return snapshots, rows.Err()
}
