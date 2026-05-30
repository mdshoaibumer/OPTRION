-- 000013_create_incident_rules.up.sql
-- Configurable rules that define when incidents should be created.

CREATE TABLE IF NOT EXISTS incident_rules (
    id              UUID PRIMARY KEY,
    tenant_id       UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    component_id    UUID         REFERENCES components(id) ON DELETE SET NULL,
    collector_type  VARCHAR(50)  NOT NULL,
    condition       JSONB        NOT NULL,
    severity        VARCHAR(20)  NOT NULL DEFAULT 'minor',
    cooldown_sec    INTEGER      NOT NULL DEFAULT 300,
    enabled         BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_incident_rules_name_tenant UNIQUE (tenant_id, name)
);

CREATE INDEX idx_incident_rules_tenant ON incident_rules (tenant_id, enabled);
CREATE INDEX idx_incident_rules_component ON incident_rules (component_id) WHERE component_id IS NOT NULL;
CREATE INDEX idx_incident_rules_collector ON incident_rules (collector_type);

ALTER TABLE incident_rules ADD CONSTRAINT chk_incident_rules_severity
    CHECK (severity IN ('info', 'warning', 'minor', 'major', 'critical'));

CREATE TRIGGER trg_incident_rules_updated_at
    BEFORE UPDATE ON incident_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
