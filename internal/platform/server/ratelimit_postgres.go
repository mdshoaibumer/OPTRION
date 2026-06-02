package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRateLimiter implements RateLimiter using PostgreSQL for persistence.
// Works across multiple instances and survives restarts.
// Uses a sliding window counter pattern stored in the database.
type PostgresRateLimiter struct {
	pool          *pgxpool.Pool
	maxRequests   int
	windowSeconds int
	logger        *slog.Logger
}

// NewPostgresRateLimiter creates a rate limiter backed by PostgreSQL.
// maxRequests is the maximum requests per window period.
// windowDuration is the sliding window time period.
func NewPostgresRateLimiter(pool *pgxpool.Pool, maxRequests int, windowDuration time.Duration, logger *slog.Logger) *PostgresRateLimiter {
	if maxRequests <= 0 {
		maxRequests = 100
	}
	windowSeconds := int(windowDuration.Seconds())
	if windowSeconds <= 0 {
		windowSeconds = 60
	}

	rl := &PostgresRateLimiter{
		pool:          pool,
		maxRequests:   maxRequests,
		windowSeconds: windowSeconds,
		logger:        logger,
	}

	// Start background cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if the given key is allowed to proceed.
func (rl *PostgresRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	var allowed bool
	err := rl.pool.QueryRow(ctx,
		`SELECT check_rate_limit($1, $2, $3)`,
		key, rl.windowSeconds, rl.maxRequests,
	).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("checking rate limit: %w", err)
	}
	return allowed, nil
}

func (rl *PostgresRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var deleted int64
		err := rl.pool.QueryRow(ctx, `SELECT cleanup_rate_limit_windows(10)`).Scan(&deleted)
		if err != nil {
			rl.logger.Error("rate limit cleanup failed", "error", err)
		} else if deleted > 0 {
			rl.logger.Debug("cleaned up rate limit windows", "deleted", deleted)
		}
		cancel()
	}
}

// HybridRateLimiter combines in-memory and PostgreSQL rate limiting.
// Uses in-memory for hot path (fast rejection) and PostgreSQL for durability.
// This provides the performance of in-memory with the consistency of database-backed.
type HybridRateLimiter struct {
	memory   *InMemoryRateLimiter
	postgres *PostgresRateLimiter
	logger   *slog.Logger
}

// NewHybridRateLimiter creates a rate limiter that uses in-memory for speed
// and falls back to PostgreSQL for accuracy across instances.
func NewHybridRateLimiter(pool *pgxpool.Pool, rps int, logger *slog.Logger) *HybridRateLimiter {
	return &HybridRateLimiter{
		memory:   NewInMemoryRateLimiter(rps),
		postgres: NewPostgresRateLimiter(pool, rps*60, 60*time.Second, logger),
		logger:   logger,
	}
}

// Allow checks both the in-memory and PostgreSQL rate limiters.
// If the in-memory limiter rejects, the request is rejected immediately (fast path).
// If the in-memory limiter allows, verify against PostgreSQL (consistency check).
func (h *HybridRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	// Fast path: in-memory rejection
	allowed, err := h.memory.Allow(ctx, key)
	if err != nil {
		h.logger.Error("in-memory rate limiter error", "error", err)
		// Fall through to PostgreSQL
	} else if !allowed {
		return false, nil
	}

	// Consistency check: PostgreSQL (durable, cross-instance)
	allowed, err = h.postgres.Allow(ctx, key)
	if err != nil {
		// If PostgreSQL is unavailable, fall back to in-memory result (fail-open for availability)
		h.logger.Warn("postgres rate limiter unavailable, using in-memory only", "error", err)
		return true, nil
	}

	return allowed, nil
}
