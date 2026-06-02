package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantConn wraps a pgx connection with RLS tenant context set.
// It ensures SET LOCAL app.tenant_id is executed at the start of each transaction.
type TenantConn struct {
	pool *pgxpool.Pool
}

// NewTenantConn creates a new tenant-aware connection wrapper.
func NewTenantConn(pool *pgxpool.Pool) *TenantConn {
	return &TenantConn{pool: pool}
}

// WithTenant executes a function within a transaction that has the RLS tenant context set.
// This ensures all queries within the function are scoped to the specified tenant.
func (tc *TenantConn) WithTenant(ctx context.Context, tenantID string, fn func(tx pgx.Tx) error) error {
	tx, err := tc.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tenant transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// SET LOCAL only affects the current transaction
	if _, err := tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID); err != nil {
		return fmt.Errorf("setting tenant context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// QueryWithTenant executes a query within a short-lived transaction that has tenant context set.
// For read-only queries that need RLS enforcement.
func (tc *TenantConn) QueryWithTenant(ctx context.Context, tenantID string, sql string, args ...interface{}) (pgx.Rows, error) {
	tx, err := tc.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tenant query: %w", err)
	}

	if _, err := tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID); err != nil {
		tx.Rollback(ctx) //nolint:errcheck
		return nil, fmt.Errorf("setting tenant context: %w", err)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		tx.Rollback(ctx) //nolint:errcheck
		return nil, err
	}

	// Note: caller must close rows, then we commit
	// For proper resource management, use WithTenant + fn pattern instead
	return rows, nil
}

// SetTenantContext sets the RLS session variable on an existing transaction.
// Use this when you already have a transaction and need to set tenant scope.
func SetTenantContext(ctx context.Context, tx pgx.Tx, tenantID string) error {
	_, err := tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
	if err != nil {
		return fmt.Errorf("setting tenant context: %w", err)
	}
	return nil
}
