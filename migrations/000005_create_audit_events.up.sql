-- 000005_create_audit_events.up.sql
-- Audit events track all state changes in the platform.

CREATE TABLE IF NOT EXISTS audit_events (
    id           UUID PRIMARY KEY,
    tenant_id    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    actor_id     VARCHAR(255) NOT NULL DEFAULT 'system',
    action       VARCHAR(100) NOT NULL,
    entity_type  VARCHAR(100) NOT NULL,
    entity_id    UUID         NOT NULL,
    payload      JSONB        NOT NULL DEFAULT '{}',
    occurred_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Indexes (optimized for time-range queries per tenant)
CREATE INDEX idx_audit_events_tenant_id ON audit_events (tenant_id);
CREATE INDEX idx_audit_events_occurred_at ON audit_events (occurred_at DESC);
CREATE INDEX idx_audit_events_entity ON audit_events (entity_type, entity_id);
CREATE INDEX idx_audit_events_action ON audit_events (action);

-- Composite index for common query pattern: "show me audit for entity X in tenant Y"
CREATE INDEX idx_audit_events_tenant_entity ON audit_events (tenant_id, entity_type, entity_id, occurred_at DESC);
