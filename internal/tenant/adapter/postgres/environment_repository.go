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

// EnvironmentRepository implements port.EnvironmentRepository using PostgreSQL.
type EnvironmentRepository struct {
	pool *pgxpool.Pool
}

// NewEnvironmentRepository creates a new PostgreSQL-backed environment repository.
func NewEnvironmentRepository(pool *pgxpool.Pool) *EnvironmentRepository {
	return &EnvironmentRepository{pool: pool}
}

// Create persists a new environment.
func (r *EnvironmentRepository) Create(ctx context.Context, env *domain.Environment) error {
	q := querier(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO environments (id, tenant_id, product_id, name, slug, tier, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		env.ID, env.TenantID, env.ProductID, env.Name, env.Slug.String(),
		string(env.Tier), string(env.Status), env.CreatedAt, env.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEnvironmentSlugTaken{ProductID: env.ProductID, Slug: env.Slug.String()}
		}
		return fmt.Errorf("inserting environment: %w", err)
	}
	return nil
}

// GetByID retrieves an environment by ID.
func (r *EnvironmentRepository) GetByID(ctx context.Context, id string) (*domain.Environment, error) {
	q := querier(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT id, tenant_id, product_id, name, slug, tier, status, created_at, updated_at
		 FROM environments WHERE id = $1`, id,
	)
	return scanEnvironment(row)
}

// ListByProduct retrieves environments belonging to a product.
func (r *EnvironmentRepository) ListByProduct(ctx context.Context, productID string, filter port.EnvironmentFilter) ([]*domain.Environment, error) {
	q := querier(ctx, r.pool)

	query := `SELECT id, tenant_id, product_id, name, slug, tier, status, created_at, updated_at
	          FROM environments WHERE product_id = $1`
	args := []any{productID}
	argIdx := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.Tier != nil {
		query += fmt.Sprintf(" AND tier = $%d", argIdx)
		args = append(args, string(*filter.Tier))
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
		return nil, fmt.Errorf("listing environments: %w", err)
	}
	defer rows.Close()

	var envs []*domain.Environment
	for rows.Next() {
		e, err := scanEnvironmentRow(rows)
		if err != nil {
			return nil, err
		}
		envs = append(envs, e)
	}
	return envs, rows.Err()
}

// Update persists changes to an existing environment.
func (r *EnvironmentRepository) Update(ctx context.Context, env *domain.Environment) error {
	q := querier(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE environments SET name = $2, tier = $3, status = $4, updated_at = $5
		 WHERE id = $1`,
		env.ID, env.Name, string(env.Tier), string(env.Status), env.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating environment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEnvironmentNotFound{ID: env.ID}
	}
	return nil
}

// ExistsBySlug checks if an environment with the given slug exists within a product.
func (r *EnvironmentRepository) ExistsBySlug(ctx context.Context, productID, slug string) (bool, error) {
	q := querier(ctx, r.pool)
	var exists bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM environments WHERE product_id = $1 AND slug = $2)`, productID, slug,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking environment slug existence: %w", err)
	}
	return exists, nil
}

func scanEnvironment(row pgx.Row) (*domain.Environment, error) {
	var e domain.Environment
	var slug, tier, status string
	err := row.Scan(&e.ID, &e.TenantID, &e.ProductID, &e.Name, &slug, &tier, &status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEnvironmentNotFound{}
		}
		return nil, fmt.Errorf("scanning environment: %w", err)
	}
	e.Slug = domain.Slug(slug)
	e.Tier = domain.Tier(tier)
	e.Status = domain.Status(status)
	return &e, nil
}

func scanEnvironmentRow(rows pgx.Rows) (*domain.Environment, error) {
	var e domain.Environment
	var slug, tier, status string
	err := rows.Scan(&e.ID, &e.TenantID, &e.ProductID, &e.Name, &slug, &tier, &status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scanning environment row: %w", err)
	}
	e.Slug = domain.Slug(slug)
	e.Tier = domain.Tier(tier)
	e.Status = domain.Status(status)
	return &e, nil
}
