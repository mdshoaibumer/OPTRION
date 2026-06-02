-- 000021_partition_metric_snapshots.up.sql
-- Convert metric_snapshots to a partitioned table for time-series scalability.
-- Partitioned by RANGE on collected_at (monthly partitions).
-- This prevents unbounded growth from degrading PostgreSQL performance.

-- Step 1: Rename existing table
ALTER TABLE metric_snapshots RENAME TO metric_snapshots_old;

-- Step 2: Drop the foreign key constraint on the old table's indexes
DROP INDEX IF EXISTS idx_metric_snapshots_metric_time;
DROP INDEX IF EXISTS idx_metric_snapshots_tenant_time;
DROP INDEX IF EXISTS idx_metric_snapshots_status;
DROP INDEX IF EXISTS idx_metric_snapshots_collected_at;

-- Step 3: Create partitioned table
CREATE TABLE metric_snapshots (
    id           UUID         NOT NULL,
    tenant_id    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    metric_id    UUID         NOT NULL REFERENCES health_metrics(id) ON DELETE CASCADE,
    value        DOUBLE PRECISION NOT NULL,
    status       VARCHAR(20)  NOT NULL DEFAULT 'unknown',
    labels       JSONB        NOT NULL DEFAULT '{}',
    collected_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, collected_at),
    CONSTRAINT chk_metric_snapshots_status_part
        CHECK (status IN ('healthy', 'degraded', 'critical', 'unknown'))
) PARTITION BY RANGE (collected_at);

-- Step 4: Create default partition for any data that doesn't match date ranges
CREATE TABLE metric_snapshots_default PARTITION OF metric_snapshots DEFAULT;

-- Step 5: Create partitions for current and next 3 months
CREATE TABLE metric_snapshots_y2026m06 PARTITION OF metric_snapshots
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE metric_snapshots_y2026m07 PARTITION OF metric_snapshots
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE metric_snapshots_y2026m08 PARTITION OF metric_snapshots
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE metric_snapshots_y2026m09 PARTITION OF metric_snapshots
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

-- Step 6: Migrate existing data
INSERT INTO metric_snapshots (id, tenant_id, metric_id, value, status, labels, collected_at)
SELECT id, tenant_id, metric_id, value, status, labels, collected_at FROM metric_snapshots_old;

-- Step 7: Drop old table
DROP TABLE metric_snapshots_old;

-- Step 8: Recreate indexes on the partitioned table
CREATE INDEX idx_metric_snapshots_metric_time ON metric_snapshots (metric_id, collected_at DESC);
CREATE INDEX idx_metric_snapshots_tenant_time ON metric_snapshots (tenant_id, collected_at DESC);
CREATE INDEX idx_metric_snapshots_status ON metric_snapshots (status) WHERE status != 'healthy';
CREATE INDEX idx_metric_snapshots_collected_at ON metric_snapshots (collected_at DESC);

-- Step 9: Create a function to auto-create monthly partitions
CREATE OR REPLACE FUNCTION create_metric_snapshot_partition()
RETURNS void AS $$
DECLARE
    next_month DATE;
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    -- Create partitions for the next 2 months ahead
    FOR i IN 0..2 LOOP
        next_month := DATE_TRUNC('month', NOW()) + (interval '1 month' * i);
        partition_name := 'metric_snapshots_y' || TO_CHAR(next_month, 'YYYY') || 'm' || TO_CHAR(next_month, 'MM');
        start_date := next_month;
        end_date := next_month + interval '1 month';
        
        IF NOT EXISTS (
            SELECT 1 FROM pg_class WHERE relname = partition_name
        ) THEN
            EXECUTE format(
                'CREATE TABLE %I PARTITION OF metric_snapshots FOR VALUES FROM (%L) TO (%L)',
                partition_name, start_date, end_date
            );
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Step 10: Create a function to drop partitions older than retention period
CREATE OR REPLACE FUNCTION drop_expired_metric_partitions(retention_months INT DEFAULT 3)
RETURNS void AS $$
DECLARE
    partition_record RECORD;
    cutoff_date DATE;
BEGIN
    cutoff_date := DATE_TRUNC('month', NOW()) - (interval '1 month' * retention_months);
    
    FOR partition_record IN
        SELECT inhrelid::regclass::text AS partition_name
        FROM pg_inherits
        WHERE inhparent = 'metric_snapshots'::regclass
        AND inhrelid::regclass::text != 'metric_snapshots_default'
    LOOP
        -- Extract date from partition name and drop if older than cutoff
        IF partition_record.partition_name < ('metric_snapshots_y' || TO_CHAR(cutoff_date, 'YYYY') || 'm' || TO_CHAR(cutoff_date, 'MM')) THEN
            EXECUTE format('DROP TABLE IF EXISTS %I', partition_record.partition_name);
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
