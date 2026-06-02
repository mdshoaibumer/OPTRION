package eventbus

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// IdempotencyGuard prevents duplicate event processing using a persistent store.
// Events are tracked by their idempotency key; if an event has already been
// processed, the handler is skipped.
type IdempotencyGuard struct {
	pool *pgxpool.Pool
}

// NewIdempotencyGuard creates a new idempotency guard backed by PostgreSQL.
func NewIdempotencyGuard(pool *pgxpool.Pool) *IdempotencyGuard {
	return &IdempotencyGuard{pool: pool}
}

// Wrap wraps an event handler with idempotency checking.
// The keyFunc extracts a unique idempotency key from the event.
// If the key has been seen before, the handler is not invoked.
func (g *IdempotencyGuard) Wrap(handler Handler, keyFunc func(Event) string) Handler {
	return func(ctx context.Context, event Event) error {
		key := keyFunc(event)
		if key == "" {
			// No idempotency key — execute handler without deduplication
			return handler(ctx, event)
		}

		// Try to claim this event for processing
		claimed, err := g.claim(ctx, key, event.EventType())
		if err != nil {
			return fmt.Errorf("idempotency check failed: %w", err)
		}
		if !claimed {
			// Already processed — skip silently
			return nil
		}

		// Execute handler
		if err := handler(ctx, event); err != nil {
			// Mark as failed so it can be retried
			_ = g.markFailed(ctx, key, err)
			return err
		}

		// Mark as successfully processed
		return g.markProcessed(ctx, key)
	}
}

func (g *IdempotencyGuard) claim(ctx context.Context, key, eventType string) (bool, error) {
	// Use INSERT ... ON CONFLICT DO NOTHING to atomically claim the event
	result, err := g.pool.Exec(ctx,
		`INSERT INTO processed_events (idempotency_key, event_type, status, claimed_at)
		 VALUES ($1, $2, 'processing', $3)
		 ON CONFLICT (idempotency_key) DO NOTHING`,
		key, eventType, time.Now().UTC(),
	)
	if err != nil {
		return false, fmt.Errorf("claiming event: %w", err)
	}
	// If RowsAffected is 0, the key already existed (already processed/processing)
	return result.RowsAffected() > 0, nil
}

func (g *IdempotencyGuard) markProcessed(ctx context.Context, key string) error {
	_, err := g.pool.Exec(ctx,
		`UPDATE processed_events SET status = 'processed', processed_at = $1 WHERE idempotency_key = $2`,
		time.Now().UTC(), key,
	)
	return err
}

func (g *IdempotencyGuard) markFailed(ctx context.Context, key string, processErr error) error {
	_, err := g.pool.Exec(ctx,
		`UPDATE processed_events SET status = 'failed', last_error = $1 WHERE idempotency_key = $2`,
		processErr.Error(), key,
	)
	return err
}
