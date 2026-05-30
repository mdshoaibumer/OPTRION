-- 000009_create_component_status.up.sql
-- Current health status of each monitored component (latest state).

CREATE TABLE IF NOT EXISTS component_status (
    id              UUID PRIMARY KEY,
    tenant_id       UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    component_id    UUID         NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    component_name  VARCHAR(255) NOT NULL,
    collector_type  VARCHAR(50)  NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'unknown',
    score           INTEGER      NOT NULL DEFAULT 100 CHECK (score >= 0 AND score <= 100),
    last_check_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_component_status_component UNIQUE (component_id)
);

CREATE INDEX idx_component_status_tenant ON component_status (tenant_id);
CREATE INDEX idx_component_status_status ON component_status (status) WHERE status != 'healthy';

ALTER TABLE component_status ADD CONSTRAINT chk_component_status_status
    CHECK (status IN ('healthy', 'degraded', 'critical', 'unknown'));

CREATE TRIGGER trg_component_status_updated_at
    BEFORE UPDATE ON component_status
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
