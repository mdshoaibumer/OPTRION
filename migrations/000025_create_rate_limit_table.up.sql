-- 000025_create_rate_limit_table.up.sql
-- Persistent rate limiting that survives restarts and works across instances.
-- Uses sliding window counter pattern stored in PostgreSQL.

CREATE TABLE IF NOT EXISTS rate_limit_windows (
    key         VARCHAR(255) NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    request_count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (key, window_start)
);

-- Index for cleanup of old windows
CREATE INDEX idx_rate_limit_windows_expiry ON rate_limit_windows (window_start);

-- Function to check and increment rate limit atomically
CREATE OR REPLACE FUNCTION check_rate_limit(
    p_key VARCHAR(255),
    p_window_seconds INT,
    p_max_requests INT
) RETURNS BOOLEAN AS $$
DECLARE
    current_window TIMESTAMPTZ;
    current_count INT;
BEGIN
    -- Calculate current window start (truncate to window boundary)
    current_window := DATE_TRUNC('second', NOW()) 
        - (EXTRACT(EPOCH FROM DATE_TRUNC('second', NOW()))::INT % p_window_seconds) * interval '1 second';
    
    -- Upsert the counter and get the new count
    INSERT INTO rate_limit_windows (key, window_start, request_count)
    VALUES (p_key, current_window, 1)
    ON CONFLICT (key, window_start)
    DO UPDATE SET request_count = rate_limit_windows.request_count + 1
    RETURNING request_count INTO current_count;
    
    -- Return whether the request is allowed
    RETURN current_count <= p_max_requests;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up expired rate limit windows
CREATE OR REPLACE FUNCTION cleanup_rate_limit_windows(retention_minutes INT DEFAULT 10)
RETURNS BIGINT AS $$
DECLARE
    deleted_count BIGINT;
BEGIN
    DELETE FROM rate_limit_windows
    WHERE window_start < NOW() - (interval '1 minute' * retention_minutes);
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
