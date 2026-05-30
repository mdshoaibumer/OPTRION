-- 000011_create_incidents.up.sql
-- Core incident table: the aggregate root for incident management.

CREATE TABLE IF NOT EXISTS incidents (
    id              UUID PRIMARY KEY,
    tenant_id       UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    component_id    UUID         NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    title           VARCHAR(500) NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    status          VARCHAR(30)  NOT NULL DEFAULT 'open',
    severity        VARCHAR(20)  NOT NULL DEFAULT 'minor',
    rule_id         UUID,
    correlation_id  UUID         NOT NULL,
    occurred_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    acknowledged_at TIMESTAMPTZ,
    resolved_at     TIMESTAMPTZ,
    closed_at       TIMESTAMPTZ,
    version         INTEGER      NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_incidents_tenant_status ON incidents (tenant_id, status);
CREATE INDEX idx_incidents_tenant_time ON incidents (tenant_id, occurred_at DESC);
CREATE INDEX idx_incidents_component ON incidents (component_id, status);
CREATE INDEX idx_incidents_correlation ON incidents (correlation_id);
CREATE INDEX idx_incidents_rule ON incidents (rule_id) WHERE rule_id IS NOT NULL;
CREATE INDEX idx_incidents_severity ON incidents (severity, occurred_at DESC);
CREATE INDEX idx_incidents_active ON incidents (tenant_id, status)
    WHERE status IN ('open', 'acknowledged', 'investigating');

ALTER TABLE incidents ADD CONSTRAINT chk_incidents_status
    CHECK (status IN ('open', 'acknowledged', 'investigating', 'resolved', 'closed'));

ALTER TABLE incidents ADD CONSTRAINT chk_incidents_severity
    CHECK (severity IN ('info', 'warning', 'minor', 'major', 'critical'));

CREATE TRIGGER trg_incidents_updated_at
    BEFORE UPDATE ON incidents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
