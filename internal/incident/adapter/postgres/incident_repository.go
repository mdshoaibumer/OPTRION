package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/incident/domain"
	"github.com/optrion/optrion/internal/incident/port"
)

// IncidentRepository implements port.IncidentRepository using PostgreSQL.
type IncidentRepository struct {
	pool *pgxpool.Pool
}

// NewIncidentRepository creates a new PostgreSQL incident repository.
func NewIncidentRepository(pool *pgxpool.Pool) *IncidentRepository {
	return &IncidentRepository{pool: pool}
}

func (r *IncidentRepository) Create(ctx context.Context, inc *domain.Incident) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO incidents (id, tenant_id, component_id, title, description, status, severity, rule_id, correlation_id, occurred_at, acknowledged_at, resolved_at, closed_at, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, inc.ID, inc.TenantID, inc.ComponentID, inc.Title, inc.Description,
		inc.Status, inc.Severity, nullableString(inc.RuleID), inc.CorrelationID,
		inc.OccurredAt, inc.AcknowledgedAt, inc.ResolvedAt, inc.ClosedAt,
		inc.Version, inc.CreatedAt, inc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting incident: %w", err)
	}
	return nil
}

func (r *IncidentRepository) GetByID(ctx context.Context, id string) (*domain.Incident, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, component_id, title, description, status, severity,
		       rule_id, correlation_id, occurred_at, acknowledged_at, resolved_at, closed_at,
		       version, created_at, updated_at
		FROM incidents WHERE id = $1
	`, id)

	return scanIncident(row)
}

func (r *IncidentRepository) Update(ctx context.Context, inc *domain.Incident) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE incidents SET
			status = $1, severity = $2, correlation_id = $3,
			acknowledged_at = $4, resolved_at = $5, closed_at = $6,
			version = $7, updated_at = $8
		WHERE id = $9 AND version = $10
	`, inc.Status, inc.Severity, inc.CorrelationID,
		inc.AcknowledgedAt, inc.ResolvedAt, inc.ClosedAt,
		inc.Version, inc.UpdatedAt, inc.ID, inc.Version-1)
	if err != nil {
		return fmt.Errorf("updating incident: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("optimistic concurrency conflict on incident %s", inc.ID)
	}
	return nil
}

func (r *IncidentRepository) ListByTenant(ctx context.Context, tenantID string, filter port.IncidentFilter) ([]*domain.Incident, error) {
	query, args := buildIncidentQuery(tenantID, filter, false)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*domain.Incident
	for rows.Next() {
		inc, err := scanIncidentFromRows(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, rows.Err()
}

func (r *IncidentRepository) CountByTenant(ctx context.Context, tenantID string, filter port.IncidentFilter) (int, error) {
	query, args := buildIncidentQuery(tenantID, filter, true)
	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting incidents: %w", err)
	}
	return count, nil
}

func (r *IncidentRepository) FindActiveByRuleAndComponent(ctx context.Context, ruleID, componentID string) (*domain.Incident, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, component_id, title, description, status, severity,
		       rule_id, correlation_id, occurred_at, acknowledged_at, resolved_at, closed_at,
		       version, created_at, updated_at
		FROM incidents
		WHERE rule_id = $1 AND component_id = $2 AND status IN ('open', 'acknowledged', 'investigating')
		ORDER BY occurred_at DESC
		LIMIT 1
	`, ruleID, componentID)

	inc, err := scanIncident(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return inc, err
}

func (r *IncidentRepository) FindByCorrelation(ctx context.Context, correlationID string) ([]*domain.Incident, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, component_id, title, description, status, severity,
		       rule_id, correlation_id, occurred_at, acknowledged_at, resolved_at, closed_at,
		       version, created_at, updated_at
		FROM incidents WHERE correlation_id = $1
		ORDER BY occurred_at ASC
	`, correlationID)
	if err != nil {
		return nil, fmt.Errorf("finding correlated incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*domain.Incident
	for rows.Next() {
		inc, err := scanIncidentFromRows(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, rows.Err()
}

// --- Event Store ---

// EventRepository implements port.IncidentEventRepository using PostgreSQL.
type EventRepository struct {
	pool *pgxpool.Pool
}

// NewEventRepository creates a new PostgreSQL event repository.
func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func (r *EventRepository) Append(ctx context.Context, events []domain.IncidentEvent) error {
	batch := &pgx.Batch{}
	for _, evt := range events {
		metadata, _ := json.Marshal(evt.Metadata)
		batch.Queue(`
			INSERT INTO incident_events (id, tenant_id, incident_id, event_type, metadata, occurred_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, evt.ID, evt.TenantID, evt.IncidentID, evt.EventType, metadata, evt.OccurredAt)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range events {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("appending event: %w", err)
		}
	}
	return nil
}

func (r *EventRepository) ListByIncident(ctx context.Context, incidentID string) ([]domain.IncidentEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, incident_id, event_type, metadata, occurred_at
		FROM incident_events WHERE incident_id = $1
		ORDER BY occurred_at ASC
	`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("listing events: %w", err)
	}
	defer rows.Close()

	var events []domain.IncidentEvent
	for rows.Next() {
		var evt domain.IncidentEvent
		var metadata []byte
		if err := rows.Scan(&evt.ID, &evt.TenantID, &evt.IncidentID, &evt.EventType, &metadata, &evt.OccurredAt); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		if err := json.Unmarshal(metadata, &evt.Metadata); err != nil {
			evt.Metadata = map[string]string{"raw": string(metadata)}
		}
		events = append(events, evt)
	}
	return events, rows.Err()
}

func (r *EventRepository) ListByTenant(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]domain.IncidentEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, incident_id, event_type, metadata, occurred_at
		FROM incident_events
		WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at <= $3
		ORDER BY occurred_at DESC LIMIT $4
	`, tenantID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("listing tenant events: %w", err)
	}
	defer rows.Close()

	var events []domain.IncidentEvent
	for rows.Next() {
		var evt domain.IncidentEvent
		var metadata []byte
		if err := rows.Scan(&evt.ID, &evt.TenantID, &evt.IncidentID, &evt.EventType, &metadata, &evt.OccurredAt); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		json.Unmarshal(metadata, &evt.Metadata) //nolint:errcheck
		events = append(events, evt)
	}
	return events, rows.Err()
}

// --- Helpers ---

func buildIncidentQuery(tenantID string, filter port.IncidentFilter, countOnly bool) (string, []interface{}) {
	var selectClause string
	if countOnly {
		selectClause = "SELECT COUNT(*)"
	} else {
		selectClause = `SELECT id, tenant_id, component_id, title, description, status, severity,
		       rule_id, correlation_id, occurred_at, acknowledged_at, resolved_at, closed_at,
		       version, created_at, updated_at`
	}

	query := selectClause + " FROM incidents WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Severity != nil {
		query += fmt.Sprintf(" AND severity = $%d", argIdx)
		args = append(args, *filter.Severity)
		argIdx++
	}
	if filter.ComponentID != nil {
		query += fmt.Sprintf(" AND component_id = $%d", argIdx)
		args = append(args, *filter.ComponentID)
		argIdx++
	}
	if filter.From != nil {
		query += fmt.Sprintf(" AND occurred_at >= $%d", argIdx)
		args = append(args, *filter.From)
		argIdx++
	}
	if filter.To != nil {
		query += fmt.Sprintf(" AND occurred_at <= $%d", argIdx)
		args = append(args, *filter.To)
		argIdx++
	}

	if !countOnly {
		query += " ORDER BY occurred_at DESC"
		if filter.Limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d", argIdx)
			args = append(args, filter.Limit)
			argIdx++
		}
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIdx)
			args = append(args, filter.Offset)
		}
	}

	return query, args
}

func scanIncident(row pgx.Row) (*domain.Incident, error) {
	var inc domain.Incident
	var ruleID *string
	err := row.Scan(
		&inc.ID, &inc.TenantID, &inc.ComponentID, &inc.Title, &inc.Description,
		&inc.Status, &inc.Severity, &ruleID, &inc.CorrelationID,
		&inc.OccurredAt, &inc.AcknowledgedAt, &inc.ResolvedAt, &inc.ClosedAt,
		&inc.Version, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ruleID != nil {
		inc.RuleID = *ruleID
	}
	return &inc, nil
}

func scanIncidentFromRows(rows pgx.Rows) (*domain.Incident, error) {
	var inc domain.Incident
	var ruleID *string
	err := rows.Scan(
		&inc.ID, &inc.TenantID, &inc.ComponentID, &inc.Title, &inc.Description,
		&inc.Status, &inc.Severity, &ruleID, &inc.CorrelationID,
		&inc.OccurredAt, &inc.AcknowledgedAt, &inc.ResolvedAt, &inc.ClosedAt,
		&inc.Version, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning incident row: %w", err)
	}
	if ruleID != nil {
		inc.RuleID = *ruleID
	}
	return &inc, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
