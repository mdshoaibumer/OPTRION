-- 000012_create_incident_events.up.sql
-- Event sourcing store: immutable append-only log of all incident state changes.

CREATE TABLE IF NOT EXISTS incident_events (
    id          UUID PRIMARY KEY,
    tenant_id   UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    incident_id UUID        NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    event_type  VARCHAR(50) NOT NULL,
    metadata    JSONB       NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_incident_events_incident ON incident_events (incident_id, occurred_at ASC);
CREATE INDEX idx_incident_events_tenant ON incident_events (tenant_id, occurred_at DESC);
CREATE INDEX idx_incident_events_type ON incident_events (event_type, occurred_at DESC);
