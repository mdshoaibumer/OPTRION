package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/tenant/domain"
	"github.com/optrion/optrion/internal/tenant/port"
)

// TenantRepository implements port.TenantRepository using PostgreSQL.
type TenantRepository struct {
	pool *pgxpool.Pool
}

// NewTenantRepository creates a new PostgreSQL-backed tenant repository.
func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

// Create persists a new tenant.
func (r *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	q := querier(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO tenants (id, name, slug, plan, status, settings, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenant.ID, tenant.Name, tenant.Slug.String(), string(tenant.Plan),
		string(tenant.Status), tenant.Settings, tenant.CreatedAt, tenant.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrTenantSlugTaken{Slug: tenant.Slug.String()}
		}
		return fmt.Errorf("inserting tenant: %w", err)
	}
	return nil
}

// GetByID retrieves a tenant by ID.
func (r *TenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	q := querier(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT id, name, slug, plan, status, settings, created_at, updated_at
		 FROM tenants WHERE id = $1`, id,
	)
	return scanTenant(row)
}

// GetBySlug retrieves a tenant by slug.
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	q := querier(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT id, name, slug, plan, status, settings, created_at, updated_at
		 FROM tenants WHERE slug = $1`, slug,
	)
	return scanTenant(row)
}

// List retrieves tenants matching the given filter.
func (r *TenantRepository) List(ctx context.Context, filter port.TenantFilter) ([]*domain.Tenant, error) {
	q := querier(ctx, r.pool)

	query := `SELECT id, name, slug, plan, status, settings, created_at, updated_at FROM tenants WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		t, err := scanTenantRow(rows)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

// Update persists changes to an existing tenant.
func (r *TenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	q := querier(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE tenants SET name = $2, plan = $3, status = $4, settings = $5, updated_at = $6
		 WHERE id = $1`,
		tenant.ID, tenant.Name, string(tenant.Plan), string(tenant.Status),
		tenant.Settings, tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating tenant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTenantNotFound{ID: tenant.ID}
	}
	return nil
}

// ExistsBySlug checks if a tenant with the given slug exists.
func (r *TenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	q := querier(ctx, r.pool)
	var exists bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM tenants WHERE slug = $1)`, slug,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking tenant slug existence: %w", err)
	}
	return exists, nil
}

func scanTenant(row pgx.Row) (*domain.Tenant, error) {
	var t domain.Tenant
	var slug, plan, status string
	err := row.Scan(&t.ID, &t.Name, &slug, &plan, &status, &t.Settings, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTenantNotFound{}
		}
		return nil, fmt.Errorf("scanning tenant: %w", err)
	}
	t.Slug = domain.Slug(slug)
	t.Plan = domain.Plan(plan)
	t.Status = domain.Status(status)
	return &t, nil
}

func scanTenantRow(rows pgx.Rows) (*domain.Tenant, error) {
	var t domain.Tenant
	var slug, plan, status string
	err := rows.Scan(&t.ID, &t.Name, &slug, &plan, &status, &t.Settings, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scanning tenant row: %w", err)
	}
	t.Slug = domain.Slug(slug)
	t.Plan = domain.Plan(plan)
	t.Status = domain.Status(status)
	return &t, nil
}
