-- 000014_create_incident_comments.up.sql
-- Human-authored notes attached to incidents for operational context.

CREATE TABLE IF NOT EXISTS incident_comments (
    id          UUID PRIMARY KEY,
    tenant_id   UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    incident_id UUID         NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    author_id   VARCHAR(255) NOT NULL DEFAULT 'system',
    content     TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_incident_comments_incident ON incident_comments (incident_id, created_at ASC);
CREATE INDEX idx_incident_comments_tenant ON incident_comments (tenant_id, created_at DESC);
