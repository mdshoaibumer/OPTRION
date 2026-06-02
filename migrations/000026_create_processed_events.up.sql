-- 000026_create_processed_events.up.sql
-- Tracks processed events for idempotency guarantees.
-- Ensures at-least-once delivery doesn't result in duplicate processing.

CREATE TABLE IF NOT EXISTS processed_events (
    idempotency_key VARCHAR(255) PRIMARY KEY,
    event_type      VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'processing'
                    CHECK (status IN ('processing', 'processed', 'failed')),
    last_error      TEXT,
    claimed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ
);

-- Index for cleanup of old processed events
CREATE INDEX idx_processed_events_processed_at ON processed_events (processed_at)
    WHERE status = 'processed';

-- Index for finding failed events for retry
CREATE INDEX idx_processed_events_failed ON processed_events (claimed_at)
    WHERE status = 'failed';

-- Function to purge old processed events (keep for 24h for dedup window)
CREATE OR REPLACE FUNCTION purge_processed_events(retention_hours INT DEFAULT 24)
RETURNS BIGINT AS $$
DECLARE
    deleted_count BIGINT;
BEGIN
    DELETE FROM processed_events
    WHERE status = 'processed'
    AND processed_at < NOW() - (interval '1 hour' * retention_hours);
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
