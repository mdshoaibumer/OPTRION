-- 000015_create_incident_timeline.up.sql
-- Unified timeline view combining events, comments, and metric changes.

CREATE TABLE IF NOT EXISTS incident_timeline (
    id          UUID PRIMARY KEY,
    tenant_id   UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    incident_id UUID         NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    entry_type  VARCHAR(20)  NOT NULL,
    title       VARCHAR(500) NOT NULL,
    details     TEXT         NOT NULL DEFAULT '',
    actor_id    VARCHAR(255) NOT NULL DEFAULT 'system',
    occurred_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_incident_timeline_incident ON incident_timeline (incident_id, occurred_at ASC);
CREATE INDEX idx_incident_timeline_tenant ON incident_timeline (tenant_id, occurred_at DESC);

ALTER TABLE incident_timeline ADD CONSTRAINT chk_timeline_entry_type
    CHECK (entry_type IN ('event', 'comment', 'metric'));
