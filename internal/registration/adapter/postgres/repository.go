package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/registration/domain"
	"github.com/optrion/optrion/internal/shared/id"
)

// RegistrationRepository implements port.RegistrationRepository using PostgreSQL.
type RegistrationRepository struct {
	pool *pgxpool.Pool
}

// NewRegistrationRepository creates a new registration repository.
func NewRegistrationRepository(pool *pgxpool.Pool) *RegistrationRepository {
	return &RegistrationRepository{pool: pool}
}

// CreateAudit stores a registration audit record.
func (r *RegistrationRepository) CreateAudit(ctx context.Context, audit *domain.RegistrationAudit) error {
	if audit.ID == "" {
		audit.ID = id.New()
	}

	payload, err := json.Marshal(audit.Request)
	if err != nil {
		return fmt.Errorf("marshaling audit payload: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO audit_events (id, tenant_id, action, entity_type, entity_id, payload, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		audit.ID, audit.TenantID, "registration."+audit.Status, "registration", audit.ID, payload, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("inserting registration audit: %w", err)
	}
	return nil
}

// GetAuditByID retrieves a registration audit by ID.
func (r *RegistrationRepository) GetAuditByID(ctx context.Context, auditID string) (*domain.RegistrationAudit, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, action, payload, created_at
		 FROM audit_events WHERE id = $1 AND entity_type = 'registration'`, auditID,
	)

	var audit domain.RegistrationAudit
	var payload []byte
	var action string
	var createdAt time.Time
	err := row.Scan(&audit.ID, &audit.TenantID, &action, &payload, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying registration audit: %w", err)
	}

	_ = json.Unmarshal(payload, &audit.Request)
	return &audit, nil
}

// ListAuditsByTenant retrieves all registration audits for a tenant.
func (r *RegistrationRepository) ListAuditsByTenant(ctx context.Context, tenantID string) ([]*domain.RegistrationAudit, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, action, payload, created_at
		 FROM audit_events WHERE tenant_id = $1 AND entity_type = 'registration'
		 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying registration audits: %w", err)
	}
	defer rows.Close()

	var audits []*domain.RegistrationAudit
	for rows.Next() {
		var audit domain.RegistrationAudit
		var payload []byte
		var action string
		var createdAt time.Time
		if err := rows.Scan(&audit.ID, &audit.TenantID, &action, &payload, &createdAt); err != nil {
			return nil, fmt.Errorf("scanning registration audit: %w", err)
		}
		_ = json.Unmarshal(payload, &audit.Request)
		audits = append(audits, &audit)
	}
	return audits, rows.Err()
}
