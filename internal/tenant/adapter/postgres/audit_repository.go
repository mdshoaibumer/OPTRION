package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/tenant/domain"
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
