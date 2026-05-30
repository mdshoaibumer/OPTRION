-- 000007_create_metric_snapshots.up.sql
-- Time-series storage for collected metric values.
-- Designed for high-volume inserts and time-range queries.

CREATE TABLE IF NOT EXISTS metric_snapshots (
    id           UUID PRIMARY KEY,
    tenant_id    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    metric_id    UUID         NOT NULL REFERENCES health_metrics(id) ON DELETE CASCADE,
    value        DOUBLE PRECISION NOT NULL,
    status       VARCHAR(20)  NOT NULL DEFAULT 'unknown',
    labels       JSONB        NOT NULL DEFAULT '{}',
    collected_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Time-series optimized indexes
CREATE INDEX idx_metric_snapshots_metric_time ON metric_snapshots (metric_id, collected_at DESC);
CREATE INDEX idx_metric_snapshots_tenant_time ON metric_snapshots (tenant_id, collected_at DESC);
CREATE INDEX idx_metric_snapshots_status ON metric_snapshots (status) WHERE status != 'healthy';
CREATE INDEX idx_metric_snapshots_collected_at ON metric_snapshots (collected_at DESC);

-- Constraint for status values
ALTER TABLE metric_snapshots ADD CONSTRAINT chk_metric_snapshots_status
    CHECK (status IN ('healthy', 'degraded', 'critical', 'unknown'));
