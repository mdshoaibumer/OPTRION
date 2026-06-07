package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/tenant/domain"
	"github.com/optrion/optrion/internal/tenant/port"
)

// AuditRepository implements port.AuditRepository using PostgreSQL.
type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository creates a new PostgreSQL-backed audit repository.
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// Create persists a new audit event.
func (r *AuditRepository) Create(ctx context.Context, event *domain.AuditEvent) error {
	q := querier(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO audit_events (id, tenant_id, actor_id, action, entity_type, entity_id, payload, occurred_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		event.ID, event.TenantID, event.ActorID, event.Action,
		event.EntityType, event.EntityID, event.Payload, event.OccurredAt,
	)
	if err != nil {
		return fmt.Errorf("inserting audit event: %w", err)
	}
	return nil
}

// ListByEntity retrieves audit events for a specific entity.
func (r *AuditRepository) ListByEntity(ctx context.Context, tenantID, entityType, entityID string, limit int) ([]*domain.AuditEvent, error) {
	q := querier(ctx, r.pool)

	if limit <= 0 {
		limit = 50
	}

	rows, err := q.Query(ctx,
		`SELECT id, tenant_id, actor_id, action, entity_type, entity_id, payload, occurred_at
		 FROM audit_events
		 WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		 ORDER BY occurred_at DESC
		 LIMIT $4`,
		tenantID, entityType, entityID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("listing audit events: %w", err)
	}
	defer rows.Close()

	var events []*domain.AuditEvent
	for rows.Next() {
		var e domain.AuditEvent
		if err := rows.Scan(&e.ID, &e.TenantID, &e.ActorID, &e.Action,
			&e.EntityType, &e.EntityID, &e.Payload, &e.OccurredAt); err != nil {
			return nil, fmt.Errorf("scanning audit event: %w", err)
		}
		events = append(events, &e)
	}
	return events, rows.Err()
}

// ListByTenant retrieves paginated audit events for a tenant with optional filtering.
func (r *AuditRepository) ListByTenant(ctx context.Context, tenantID string, filter port.AuditFilter) ([]*domain.AuditEvent, int, error) {
	q := querier(ctx, r.pool)

	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Build dynamic WHERE clause
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Action != nil {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, *filter.Action)
		argIdx++
	}
	if filter.EntityType != nil {
		where += fmt.Sprintf(" AND entity_type = $%d", argIdx)
		args = append(args, *filter.EntityType)
		argIdx++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_events %s", where)
	var total int
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting audit events: %w", err)
	}

	// Fetch page
	query := fmt.Sprintf(
		`SELECT id, tenant_id, actor_id, action, entity_type, entity_id, payload, occurred_at
		 FROM audit_events %s
		 ORDER BY occurred_at DESC
		 LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing audit events: %w", err)
	}
	defer rows.Close()

	var events []*domain.AuditEvent
	for rows.Next() {
		var e domain.AuditEvent
		if err := rows.Scan(&e.ID, &e.TenantID, &e.ActorID, &e.Action,
			&e.EntityType, &e.EntityID, &e.Payload, &e.OccurredAt); err != nil {
			return nil, 0, fmt.Errorf("scanning audit event: %w", err)
		}
		events = append(events, &e)
	}
	return events, total, rows.Err()
}
