package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/shared/id"
)

// OutboxEvent represents an event stored in the outbox table.
type OutboxEvent struct {
	ID             string
	TenantID       string
	EventType      string
	AggregateType  string
	AggregateID    string
	Payload        json.RawMessage
	Metadata       json.RawMessage
	Status         string
	Attempts       int
	MaxAttempts    int
	LastError      string
	IdempotencyKey string
	CreatedAt      time.Time
	ProcessedAt    *time.Time
}

// OutboxWriter writes events to the outbox table within an existing transaction.
type OutboxWriter struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewOutboxWriter creates a new outbox writer.
func NewOutboxWriter(pool *pgxpool.Pool, logger *slog.Logger) *OutboxWriter {
	return &OutboxWriter{pool: pool, logger: logger}
}

// WriteEvent writes a single event to the outbox within the provided transaction.
// The event is committed atomically with the business transaction.
func (w *OutboxWriter) WriteEvent(ctx context.Context, tx pgx.Tx, event Event, aggregateType, aggregateID string) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event payload: %w", err)
	}

	idempotencyKey := fmt.Sprintf("%s:%s:%s:%d", event.EventType(), aggregateID, event.TenantID(), time.Now().UnixNano())

	_, err = tx.Exec(ctx,
		`INSERT INTO event_outbox (id, tenant_id, event_type, aggregate_type, aggregate_id, payload, status, idempotency_key, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8)`,
		id.New(), event.TenantID(), event.EventType(), aggregateType, aggregateID, payload, idempotencyKey, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("inserting outbox event: %w", err)
	}

	return nil
}

// WriteEventDirect writes an event to the outbox without an external transaction.
// Uses its own transaction internally.
func (w *OutboxWriter) WriteEventDirect(ctx context.Context, event Event, aggregateType, aggregateID string) error {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning outbox transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := w.WriteEvent(ctx, tx, event, aggregateType, aggregateID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// OutboxWorker polls the outbox table and dispatches events to handlers.
type OutboxWorker struct {
	pool         *pgxpool.Pool
	bus          *Bus
	logger       *slog.Logger
	workerID     string
	batchSize    int
	pollInterval time.Duration
	lockDuration time.Duration

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
}

// OutboxWorkerConfig configures the outbox worker.
type OutboxWorkerConfig struct {
	BatchSize    int
	PollInterval time.Duration
	LockDuration time.Duration
}

// NewOutboxWorker creates a new outbox worker that polls and dispatches events.
func NewOutboxWorker(pool *pgxpool.Pool, bus *Bus, logger *slog.Logger, cfg OutboxWorkerConfig) *OutboxWorker {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 1 * time.Second
	}
	if cfg.LockDuration <= 0 {
		cfg.LockDuration = 30 * time.Second
	}

	return &OutboxWorker{
		pool:         pool,
		bus:          bus,
		logger:       logger,
		workerID:     id.New()[:8],
		batchSize:    cfg.BatchSize,
		pollInterval: cfg.PollInterval,
		lockDuration: cfg.LockDuration,
	}
}

// Start begins the outbox polling loop.
func (w *OutboxWorker) Start(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	ctx, w.cancel = context.WithCancel(ctx)
	w.mu.Unlock()

	w.logger.Info("outbox worker started", "worker_id", w.workerID, "batch_size", w.batchSize, "poll_interval", w.pollInterval)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("outbox worker stopped", "worker_id", w.workerID)
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.Error("outbox batch processing error", "worker_id", w.workerID, "error", err)
			}
		}
	}
}

// Stop halts the outbox worker.
func (w *OutboxWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
	}
	w.running = false
}

func (w *OutboxWorker) processBatch(ctx context.Context) error {
	// Claim a batch of pending events
	rows, err := w.pool.Query(ctx,
		`UPDATE event_outbox
		 SET status = 'processing', locked_until = $1, locked_by = $2
		 WHERE id IN (
		     SELECT id FROM event_outbox
		     WHERE status = 'pending'
		     AND (locked_until IS NULL OR locked_until < NOW())
		     ORDER BY created_at ASC
		     LIMIT $3
		     FOR UPDATE SKIP LOCKED
		 )
		 RETURNING id, tenant_id, event_type, aggregate_type, aggregate_id, payload, attempts`,
		time.Now().Add(w.lockDuration), w.workerID, w.batchSize,
	)
	if err != nil {
		return fmt.Errorf("claiming outbox events: %w", err)
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var evt OutboxEvent
		if err := rows.Scan(&evt.ID, &evt.TenantID, &evt.EventType, &evt.AggregateType, &evt.AggregateID, &evt.Payload, &evt.Attempts); err != nil {
			return fmt.Errorf("scanning outbox event: %w", err)
		}
		events = append(events, evt)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating outbox events: %w", err)
	}

	// Process each event
	for _, evt := range events {
		if err := w.dispatchEvent(ctx, evt); err != nil {
			w.markFailed(ctx, evt, err)
		} else {
			w.markProcessed(ctx, evt)
		}
	}

	return nil
}

func (w *OutboxWorker) dispatchEvent(ctx context.Context, evt OutboxEvent) error {
	// Create a generic event wrapper for the bus
	wrappedEvent := &outboxDispatchEvent{
		eventType: evt.EventType,
		tenantID:  evt.TenantID,
		payload:   evt.Payload,
	}

	// Dispatch synchronously to all handlers
	w.bus.Publish(ctx, wrappedEvent)
	return nil
}

func (w *OutboxWorker) markProcessed(ctx context.Context, evt OutboxEvent) {
	_, err := w.pool.Exec(ctx,
		`UPDATE event_outbox SET status = 'processed', processed_at = $1, locked_until = NULL, locked_by = NULL
		 WHERE id = $2`,
		time.Now().UTC(), evt.ID,
	)
	if err != nil {
		w.logger.Error("failed to mark event as processed", "event_id", evt.ID, "error", err)
	}
}

func (w *OutboxWorker) markFailed(ctx context.Context, evt OutboxEvent, processErr error) {
	newAttempts := evt.Attempts + 1
	newStatus := "failed"
	if newAttempts >= 5 {
		newStatus = "dead_letter"
	}

	_, err := w.pool.Exec(ctx,
		`UPDATE event_outbox SET status = $1, attempts = $2, last_error = $3, locked_until = NULL, locked_by = NULL
		 WHERE id = $4`,
		newStatus, newAttempts, processErr.Error(), evt.ID,
	)
	if err != nil {
		w.logger.Error("failed to mark event as failed", "event_id", evt.ID, "error", err)
	}
}

// outboxDispatchEvent wraps an outbox event for dispatch through the in-process bus.
type outboxDispatchEvent struct {
	eventType string
	tenantID  string
	payload   json.RawMessage
}

func (e *outboxDispatchEvent) EventType() string { return e.eventType }
func (e *outboxDispatchEvent) TenantID() string  { return e.tenantID }
func (e *outboxDispatchEvent) Payload() []byte   { return e.payload }
