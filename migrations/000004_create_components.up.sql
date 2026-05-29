-- 000004_create_components.up.sql
-- Components belong to an environment. They are the monitored entities.

CREATE TABLE IF NOT EXISTS components (
    id              UUID PRIMARY KEY,
    tenant_id       UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id      UUID         NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    environment_id  UUID         NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) NOT NULL,
    kind            VARCHAR(50)  NOT NULL,
    endpoint_url    TEXT         NOT NULL DEFAULT '',
    status          VARCHAR(20)  NOT NULL DEFAULT 'active',
    metadata        JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_components_environment_slug UNIQUE (environment_id, slug)
);

-- Indexes
CREATE INDEX idx_components_tenant_id ON components (tenant_id);
CREATE INDEX idx_components_product_id ON components (product_id);
CREATE INDEX idx_components_environment_id ON components (environment_id);
CREATE INDEX idx_components_status ON components (status);
CREATE INDEX idx_components_kind ON components (kind);

-- Constraints
ALTER TABLE components ADD CONSTRAINT chk_components_status
    CHECK (status IN ('active', 'degraded', 'inactive', 'archived'));

ALTER TABLE components ADD CONSTRAINT chk_components_kind
    CHECK (kind IN ('database', 'cache', 'api', 'web', 'queue', 'storage', 'service', 'external'));

ALTER TABLE components ADD CONSTRAINT chk_components_slug_format
    CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$' AND LENGTH(slug) >= 3);

-- Updated_at trigger
CREATE TRIGGER trg_components_updated_at
    BEFORE UPDATE ON components
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
