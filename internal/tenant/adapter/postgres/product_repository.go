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

// ProductRepository implements port.ProductRepository using PostgreSQL.
type ProductRepository struct {
	pool *pgxpool.Pool
}

// NewProductRepository creates a new PostgreSQL-backed product repository.
func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

// Create persists a new product.
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	q := querier(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO products (id, tenant_id, name, slug, description, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		product.ID, product.TenantID, product.Name, product.Slug.String(),
		product.Description, string(product.Status), product.CreatedAt, product.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrProductSlugTaken{TenantID: product.TenantID, Slug: product.Slug.String()}
		}
		return fmt.Errorf("inserting product: %w", err)
	}
	return nil
}

// GetByID retrieves a product by ID.
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	q := querier(ctx, r.pool)
	row := q.QueryRow(ctx,
		`SELECT id, tenant_id, name, slug, description, status, created_at, updated_at
		 FROM products WHERE id = $1`, id,
	)
	return scanProduct(row)
}

// ListByTenant retrieves products belonging to a tenant.
func (r *ProductRepository) ListByTenant(ctx context.Context, tenantID string, filter port.ProductFilter) ([]*domain.Product, error) {
	q := querier(ctx, r.pool)

	query := `SELECT id, tenant_id, name, slug, description, status, created_at, updated_at
	          FROM products WHERE tenant_id = $1`
	args := []any{tenantID}
	argIdx := 2

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
		return nil, fmt.Errorf("listing products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		p, err := scanProductRow(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

// Update persists changes to an existing product.
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	q := querier(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE products SET name = $2, description = $3, status = $4, updated_at = $5
		 WHERE id = $1`,
		product.ID, product.Name, product.Description, string(product.Status), product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProductNotFound{ID: product.ID}
	}
	return nil
}

// ExistsBySlug checks if a product with the given slug exists within a tenant.
func (r *ProductRepository) ExistsBySlug(ctx context.Context, tenantID, slug string) (bool, error) {
	q := querier(ctx, r.pool)
	var exists bool
	err := q.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM products WHERE tenant_id = $1 AND slug = $2)`, tenantID, slug,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking product slug existence: %w", err)
	}
	return exists, nil
}

func scanProduct(row pgx.Row) (*domain.Product, error) {
	var p domain.Product
	var slug, status string
	err := row.Scan(&p.ID, &p.TenantID, &p.Name, &slug, &p.Description, &status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound{}
		}
		return nil, fmt.Errorf("scanning product: %w", err)
	}
	p.Slug = domain.Slug(slug)
	p.Status = domain.Status(status)
	return &p, nil
}

func scanProductRow(rows pgx.Rows) (*domain.Product, error) {
	var p domain.Product
	var slug, status string
	err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &slug, &p.Description, &status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scanning product row: %w", err)
	}
	p.Slug = domain.Slug(slug)
	p.Status = domain.Status(status)
	return &p, nil
}
