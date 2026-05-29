package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/optrion/optrion/internal/platform/config"
)

// PostgreSQL wraps a pgxpool.Pool with lifecycle management.
type PostgreSQL struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
	cfg    config.DatabaseConfig
}

// NewPostgreSQL creates and validates a PostgreSQL connection pool.
func NewPostgreSQL(ctx context.Context, cfg config.DatabaseConfig, logger *slog.Logger) (*PostgreSQL, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing database DSN: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	// Verify connectivity
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	logger.Info("database connected",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Name,
		"max_conns", cfg.MaxOpenConns,
		"min_conns", cfg.MaxIdleConns,
	)

	return &PostgreSQL{
		pool:   pool,
		logger: logger,
		cfg:    cfg,
	}, nil
}

// Pool returns the underlying connection pool for use by repositories.
func (pg *PostgreSQL) Pool() *pgxpool.Pool {
	return pg.pool
}

// Health checks database connectivity. Returns an error if the database is unreachable.
func (pg *PostgreSQL) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pg.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Close gracefully shuts down the connection pool.
func (pg *PostgreSQL) Close() {
	pg.logger.Info("closing database connection pool")
	pg.pool.Close()
}

// Stats returns current pool statistics for observability.
func (pg *PostgreSQL) Stats() *PoolStats {
	stat := pg.pool.Stat()
	return &PoolStats{
		TotalConns:       int(stat.TotalConns()),
		IdleConns:        int(stat.IdleConns()),
		AcquiredConns:    int(stat.AcquiredConns()),
		MaxConns:         int(stat.MaxConns()),
		AcquireCount:     stat.AcquireCount(),
		EmptyAcquires:    stat.EmptyAcquireCount(),
		CanceledAcquires: stat.CanceledAcquireCount(),
	}
}

// PoolStats holds connection pool statistics.
type PoolStats struct {
	TotalConns       int   `json:"total_conns"`
	IdleConns        int   `json:"idle_conns"`
	AcquiredConns    int   `json:"acquired_conns"`
	MaxConns         int   `json:"max_conns"`
	AcquireCount     int64 `json:"acquire_count"`
	EmptyAcquires    int64 `json:"empty_acquires"`
	CanceledAcquires int64 `json:"canceled_acquires"`
}
