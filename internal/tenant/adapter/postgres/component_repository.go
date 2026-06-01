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

// ComponentRepository implements port.ComponentRepository using PostgreSQL.
type ComponentRepository struct {
	pool *pgxpool.Pool
}

// NewComponentRepository creates a new PostgreSQL-backed component repository.
func NewComponentRepository(pool *pgxpool.Pool) *ComponentRepository {
	return &ComponentRepository{pool: pool}
}

// Create persists a new component.
func (r *ComponentRepository) Create(ctx context.Context, comp *domain.Component) error {
	q := querier(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO components (id, tenant_id, product_id, environment_id, name, slug, kind, endpoint_url, status, metadata, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		comp.ID, comp.TenantID, comp.ProductID, comp.EnvironmentID,
		comp.Name, comp.Slug.String(), string(comp.Kind), comp.EndpointURL,
		string(comp.Status), comp.Metadata, comp.CreatedAt, comp.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrComponentSlugTaken{EnvironmentID: comp.EnvironmentID, Slug: comp.Slug.String()}
		}
		return fmt.Errorf("inserting component: %w", err)
	}
	return nil
}

// GetByID retrieves a component by ID.
func (r *ComponentRepository) GetByID(ctx context.Context, id string) (*domain.Component, error) {
	q := querier(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT id, tenant_id, product_id, environment_id, name, slug, kind, endpoint_url, status, metadata, created_at, updated_at
		 FROM components WHERE id = $1`, id,
	)
	return scanComponent(row)
}

// ListByEnvironment retrieves components belonging to an environment.
func (r *ComponentRepository) ListByEnvironment(ctx context.Context, environmentID string, filter port.ComponentFilter) ([]*domain.Component, error) {
	q := querier(ctx, r.pool)

	query := `SELECT id, tenant_id, product_id, environment_id, name, slug, kind, endpoint_url, status, metadata, created_at, updated_at
	          FROM components WHERE environment_id = $1`
	args := []any{environmentID}
	argIdx := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.Kind != nil {
		query += fmt.Sprintf(" AND kind = $%d", argIdx)
		args = append(args, string(*filter.Kind))
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
		return nil, fmt.Errorf("listing components: %w", err)
	}
	defer rows.Close()

	var comps []*domain.Component
	for rows.Next() {
		c, err := scanComponentRow(rows)
		if err != nil {
			return nil, err
		}
		comps = append(comps, c)
	}
	return comps, rows.Err()
}

// ListByTenant returns all components belonging to a tenant.
func (r *ComponentRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Component, error) {
	q := querier(ctx, r.pool)

	rows, err := q.Query(ctx,
		`SELECT id, tenant_id, product_id, environment_id, name, slug, kind, endpoint_url, status, metadata, created_at, updated_at
		 FROM components WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing components by tenant: %w", err)
	}
	defer rows.Close()

	var comps []*domain.Component
	for rows.Next() {
		c, err := scanComponentRow(rows)
		if err != nil {
			return nil, err
		}
		comps = append(comps, c)
	}
	return comps, rows.Err()
}

// Update persists changes to an existing component.
func (r *ComponentRepository) Update(ctx context.Context, comp *domain.Component) error {
	q := querier(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE components SET name = $2, kind = $3, endpoint_url = $4, status = $5, metadata = $6, updated_at = $7
		 WHERE id = $1`,
		comp.ID, comp.Name, string(comp.Kind), comp.EndpointURL,
		string(comp.Status), comp.Metadata, comp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating component: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrComponentNotFound{ID: comp.ID}
	}
	return nil
}

// ExistsBySlug checks if a component with the given slug exists within an environment.
func (r *ComponentRepository) ExistsBySlug(ctx context.Context, environmentID, slug string) (bool, error) {
	q := querier(ctx, r.pool)
	var exists bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM components WHERE environment_id = $1 AND slug = $2)`, environmentID, slug,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking component slug existence: %w", err)
	}
	return exists, nil
}

func scanComponent(row pgx.Row) (*domain.Component, error) {
	var c domain.Component
	var slug, kind, status string
	err := row.Scan(&c.ID, &c.TenantID, &c.ProductID, &c.EnvironmentID,
		&c.Name, &slug, &kind, &c.EndpointURL, &status, &c.Metadata, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrComponentNotFound{}
		}
		return nil, fmt.Errorf("scanning component: %w", err)
	}
	c.Slug = domain.Slug(slug)
	c.Kind = domain.ComponentKind(kind)
	c.Status = domain.Status(status)
	return &c, nil
}

func scanComponentRow(rows pgx.Rows) (*domain.Component, error) {
	var c domain.Component
	var slug, kind, status string
	err := rows.Scan(&c.ID, &c.TenantID, &c.ProductID, &c.EnvironmentID,
		&c.Name, &slug, &kind, &c.EndpointURL, &status, &c.Metadata, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scanning component row: %w", err)
	}
	c.Slug = domain.Slug(slug)
	c.Kind = domain.ComponentKind(kind)
	c.Status = domain.Status(status)
	return &c, nil
}
