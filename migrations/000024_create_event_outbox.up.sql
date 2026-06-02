-- 000024_create_event_outbox.up.sql
-- Event outbox for guaranteed at-least-once event delivery.
-- Business transactions write events into this table within the same DB transaction.
-- A background worker polls and dispatches events to handlers.
-- Processed events are purged after retention period.

CREATE TABLE IF NOT EXISTS event_outbox (
    id              UUID NOT NULL,
    tenant_id       UUID NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    aggregate_type  VARCHAR(100) NOT NULL,
    aggregate_id    UUID NOT NULL,
    payload         JSONB NOT NULL,
    metadata        JSONB NOT NULL DEFAULT '{}',
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'processing', 'processed', 'failed', 'dead_letter')),
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 5,
    last_error      TEXT,
    idempotency_key VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    locked_until    TIMESTAMPTZ,
    locked_by       VARCHAR(100),
    
    PRIMARY KEY (id, created_at),
    CONSTRAINT uq_event_outbox_idempotency UNIQUE (idempotency_key, created_at)
) PARTITION BY RANGE (created_at);

-- Create partitions: current month and next 2 months
CREATE TABLE event_outbox_y2026m06 PARTITION OF event_outbox
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE event_outbox_y2026m07 PARTITION OF event_outbox
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE event_outbox_y2026m08 PARTITION OF event_outbox
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE event_outbox_default PARTITION OF event_outbox DEFAULT;

-- Index for worker polling: pending events ordered by creation time
CREATE INDEX idx_event_outbox_pending ON event_outbox (created_at ASC)
    WHERE status = 'pending';

-- Index for lock expiry lookup (worker checks locked_until < NOW() at runtime)
CREATE INDEX idx_event_outbox_locked ON event_outbox (locked_until)
    WHERE status = 'pending' AND locked_until IS NOT NULL;

-- Index for retry: failed events that haven't exceeded max attempts
CREATE INDEX idx_event_outbox_retry ON event_outbox (created_at ASC)
    WHERE status = 'failed' AND attempts < max_attempts;

-- Index for cleanup: processed events older than retention
CREATE INDEX idx_event_outbox_processed ON event_outbox (processed_at)
    WHERE status = 'processed';

-- Index for event type filtering
CREATE INDEX idx_event_outbox_type ON event_outbox (event_type, created_at DESC);

-- Index for aggregate lookup (find all events for an aggregate)
CREATE INDEX idx_event_outbox_aggregate ON event_outbox (aggregate_type, aggregate_id, created_at DESC);

-- Function to auto-create outbox partitions
CREATE OR REPLACE FUNCTION create_outbox_partition()
RETURNS void AS $$
DECLARE
    next_month DATE;
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    FOR i IN 0..2 LOOP
        next_month := DATE_TRUNC('month', NOW()) + (interval '1 month' * i);
        partition_name := 'event_outbox_y' || TO_CHAR(next_month, 'YYYY') || 'm' || TO_CHAR(next_month, 'MM');
        start_date := next_month;
        end_date := next_month + interval '1 month';
        
        IF NOT EXISTS (
            SELECT 1 FROM pg_class WHERE relname = partition_name
        ) THEN
            EXECUTE format(
                'CREATE TABLE %I PARTITION OF event_outbox FOR VALUES FROM (%L) TO (%L)',
                partition_name, start_date, end_date
            );
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to purge processed events older than N days
CREATE OR REPLACE FUNCTION purge_processed_outbox_events(retention_days INT DEFAULT 7)
RETURNS BIGINT AS $$
DECLARE
    deleted_count BIGINT;
BEGIN
    DELETE FROM event_outbox
    WHERE status = 'processed'
    AND processed_at < NOW() - (interval '1 day' * retention_days);
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
