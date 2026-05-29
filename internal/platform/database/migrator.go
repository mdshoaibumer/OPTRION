package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrator runs SQL migration files against the database.
type Migrator struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewMigrator creates a new migration runner.
func NewMigrator(pool *pgxpool.Pool, logger *slog.Logger) *Migrator {
	return &Migrator{pool: pool, logger: logger}
}

// MigrateFS runs all .up.sql migrations from an embedded filesystem.
func (m *Migrator) MigrateFS(ctx context.Context, migrations embed.FS, dir string) error {
	// Ensure migration tracking table exists
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	// Gather migration files
	entries, err := fs.ReadDir(migrations, dir)
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}
	sort.Strings(upFiles)

	for _, file := range upFiles {
		version := strings.Split(file, "_")[0]

		// Check if already applied
		applied, err := m.isApplied(ctx, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		// Read and execute
		content, err := fs.ReadFile(migrations, filepath.Join(dir, file))
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", file, err)
		}

		m.logger.Info("applying migration", "file", file, "version", version)

		if _, err := m.pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("executing migration %s: %w", file, err)
		}

		// Record as applied
		if err := m.recordMigration(ctx, version, file); err != nil {
			return err
		}

		m.logger.Info("migration applied", "file", file)
	}

	return nil
}

func (m *Migrator) ensureTable(ctx context.Context) error {
	_, err := m.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    VARCHAR(20) PRIMARY KEY,
			filename   VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}
	return nil
}

func (m *Migrator) isApplied(ctx context.Context, version string) (bool, error) {
	var exists bool
	err := m.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking migration version %s: %w", version, err)
	}
	return exists, nil
}

func (m *Migrator) recordMigration(ctx context.Context, version, filename string) error {
	_, err := m.pool.Exec(ctx,
		`INSERT INTO schema_migrations (version, filename) VALUES ($1, $2)`,
		version, filename,
	)
	if err != nil {
		return fmt.Errorf("recording migration %s: %w", version, err)
	}
	return nil
}
