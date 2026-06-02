-- 000021_partition_metric_snapshots.down.sql
-- Revert partitioned metric_snapshots back to a regular table.

-- Drop partition management functions
DROP FUNCTION IF EXISTS drop_expired_metric_partitions(INT);
DROP FUNCTION IF EXISTS create_metric_snapshot_partition();

-- Recreate as a regular table
CREATE TABLE metric_snapshots_flat (
    id           UUID PRIMARY KEY,
    tenant_id    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    metric_id    UUID         NOT NULL REFERENCES health_metrics(id) ON DELETE CASCADE,
    value        DOUBLE PRECISION NOT NULL,
    status       VARCHAR(20)  NOT NULL DEFAULT 'unknown',
    labels       JSONB        NOT NULL DEFAULT '{}',
    collected_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_metric_snapshots_status
        CHECK (status IN ('healthy', 'degraded', 'critical', 'unknown'))
);

-- Migrate data
INSERT INTO metric_snapshots_flat (id, tenant_id, metric_id, value, status, labels, collected_at)
SELECT id, tenant_id, metric_id, value, status, labels, collected_at FROM metric_snapshots;

-- Drop partitioned table (cascades to all partitions)
DROP TABLE metric_snapshots CASCADE;

-- Rename flat table to original name
ALTER TABLE metric_snapshots_flat RENAME TO metric_snapshots;

-- Recreate indexes
CREATE INDEX idx_metric_snapshots_metric_time ON metric_snapshots (metric_id, collected_at DESC);
CREATE INDEX idx_metric_snapshots_tenant_time ON metric_snapshots (tenant_id, collected_at DESC);
CREATE INDEX idx_metric_snapshots_status ON metric_snapshots (status) WHERE status != 'healthy';
CREATE INDEX idx_metric_snapshots_collected_at ON metric_snapshots (collected_at DESC);
