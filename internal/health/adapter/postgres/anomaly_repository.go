package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
)

// AnomalyRepository implements port.AnomalyRepository.
type AnomalyRepository struct {
	pool *pgxpool.Pool
}

// NewAnomalyRepository creates a new anomaly repository.
func NewAnomalyRepository(pool *pgxpool.Pool) *AnomalyRepository {
	return &AnomalyRepository{pool: pool}
}

func (r *AnomalyRepository) Create(ctx context.Context, a *domain.Anomaly) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO anomalies (id, tenant_id, component_id, metric_id, metric_type, severity, title, description, expected_value, actual_value, resolved, detected_at, resolved_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		a.ID, a.TenantID, a.ComponentID, a.MetricID, a.MetricType,
		a.Severity, a.Title, a.Description, a.ExpectedValue, a.ActualValue,
		a.Resolved, a.DetectedAt, a.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting anomaly: %w", err)
	}
	return nil
}

func (r *AnomalyRepository) GetByID(ctx context.Context, id string) (*domain.Anomaly, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, component_id, metric_id, metric_type, severity, title, description, expected_value, actual_value, resolved, detected_at, resolved_at
		 FROM anomalies WHERE id = $1`, id)

	return r.scanAnomaly(row)
}

func (r *AnomalyRepository) ListByTenant(ctx context.Context, tenantID string, filter port.AnomalyFilter) ([]*domain.Anomaly, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
	args = append(args, tenantID)
	argIdx++

	if filter.ComponentID != nil {
		conditions = append(conditions, fmt.Sprintf("component_id = $%d", argIdx))
		args = append(args, *filter.ComponentID)
		argIdx++
	}
	if filter.Severity != nil {
		conditions = append(conditions, fmt.Sprintf("severity = $%d", argIdx))
		args = append(args, *filter.Severity)
		argIdx++
	}
	if filter.Resolved != nil {
		conditions = append(conditions, fmt.Sprintf("resolved = $%d", argIdx))
		args = append(args, *filter.Resolved)
		argIdx++
	}
	if filter.From != nil {
		conditions = append(conditions, fmt.Sprintf("detected_at >= $%d", argIdx))
		args = append(args, *filter.From)
		argIdx++
	}
	if filter.To != nil {
		conditions = append(conditions, fmt.Sprintf("detected_at <= $%d", argIdx))
		args = append(args, *filter.To)
		argIdx++
	}

	query := fmt.Sprintf(
		`SELECT id, tenant_id, component_id, metric_id, metric_type, severity, title, description, expected_value, actual_value, resolved, detected_at, resolved_at
		 FROM anomalies WHERE %s ORDER BY detected_at DESC LIMIT $%d OFFSET $%d`,
		strings.Join(conditions, " AND "), argIdx, argIdx+1,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying anomalies: %w", err)
	}
	defer rows.Close()
	return r.scanAnomalies(rows)
}

func (r *AnomalyRepository) ListUnresolved(ctx context.Context, tenantID string) ([]*domain.Anomaly, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, component_id, metric_id, metric_type, severity, title, description, expected_value, actual_value, resolved, detected_at, resolved_at
		 FROM anomalies WHERE tenant_id = $1 AND resolved = FALSE ORDER BY detected_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("querying unresolved anomalies: %w", err)
	}
	defer rows.Close()
	return r.scanAnomalies(rows)
}

func (r *AnomalyRepository) Resolve(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE anomalies SET resolved = TRUE, resolved_at = $1 WHERE id = $2`,
		time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("resolving anomaly: %w", err)
	}
	return nil
}

func (r *AnomalyRepository) scanAnomaly(row pgx.Row) (*domain.Anomaly, error) {
	var a domain.Anomaly
	err := row.Scan(
		&a.ID, &a.TenantID, &a.ComponentID, &a.MetricID, &a.MetricType,
		&a.Severity, &a.Title, &a.Description, &a.ExpectedValue, &a.ActualValue,
		&a.Resolved, &a.DetectedAt, &a.ResolvedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("anomaly not found")
		}
		return nil, fmt.Errorf("scanning anomaly: %w", err)
	}
	return &a, nil
}

func (r *AnomalyRepository) scanAnomalies(rows pgx.Rows) ([]*domain.Anomaly, error) {
	var anomalies []*domain.Anomaly
	for rows.Next() {
		var a domain.Anomaly
		err := rows.Scan(
			&a.ID, &a.TenantID, &a.ComponentID, &a.MetricID, &a.MetricType,
			&a.Severity, &a.Title, &a.Description, &a.ExpectedValue, &a.ActualValue,
			&a.Resolved, &a.DetectedAt, &a.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning anomaly row: %w", err)
		}
		anomalies = append(anomalies, &a)
	}
	return anomalies, rows.Err()
}
