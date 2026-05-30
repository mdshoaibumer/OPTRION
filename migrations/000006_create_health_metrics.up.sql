-- 000006_create_health_metrics.up.sql
-- Defines which metrics are tracked per component.

CREATE TABLE IF NOT EXISTS health_metrics (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    component_id    UUID NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    metric_type     VARCHAR(50)  NOT NULL,
    collector_type  VARCHAR(50)  NOT NULL,
    name            VARCHAR(255) NOT NULL,
    unit            VARCHAR(50)  NOT NULL DEFAULT '',
    thresholds      JSONB        NOT NULL DEFAULT '{}',
    enabled         BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_health_metrics_component_type UNIQUE (component_id, metric_type)
);

CREATE INDEX idx_health_metrics_tenant ON health_metrics (tenant_id);
CREATE INDEX idx_health_metrics_component ON health_metrics (component_id);
CREATE INDEX idx_health_metrics_type ON health_metrics (metric_type);
CREATE INDEX idx_health_metrics_enabled ON health_metrics (enabled) WHERE enabled = TRUE;

CREATE TRIGGER trg_health_metrics_updated_at
    BEFORE UPDATE ON health_metrics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
