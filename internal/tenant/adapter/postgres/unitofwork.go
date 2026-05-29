package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// txKey is the context key for carrying transactions.
type txKey struct{}

// UnitOfWork implements port.UnitOfWork using pgx transactions.
type UnitOfWork struct {
	pool *pgxpool.Pool
}

// NewUnitOfWork creates a new UnitOfWork.
func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{pool: pool}
}

// Begin starts a new transaction and stores it in the context.
func (u *UnitOfWork) Begin(ctx context.Context) (context.Context, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return ctx, fmt.Errorf("beginning transaction: %w", err)
	}
	return context.WithValue(ctx, txKey{}, tx), nil
}

// Commit commits the transaction stored in the context.
func (u *UnitOfWork) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction in context")
	}
	return tx.Commit(ctx)
}

// Rollback rolls back the transaction stored in the context.
func (u *UnitOfWork) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction in context")
	}
	return tx.Rollback(ctx)
}

// querier returns the transaction if present, otherwise the pool.
func querier(ctx context.Context, pool *pgxpool.Pool) queriers {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}

// queriers is a shared interface between pgxpool.Pool and pgx.Tx.
type queriers interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// isUniqueViolation checks if a PostgreSQL error is a unique constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
