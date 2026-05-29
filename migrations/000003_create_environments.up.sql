-- 000003_create_environments.up.sql
-- Environments belong to a product. Represent deployment stages.

CREATE TABLE IF NOT EXISTS environments (
    id          UUID PRIMARY KEY,
    tenant_id   UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id  UUID         NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) NOT NULL,
    tier        VARCHAR(20)  NOT NULL DEFAULT 'development',
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_environments_product_slug UNIQUE (product_id, slug)
);

-- Indexes
CREATE INDEX idx_environments_tenant_id ON environments (tenant_id);
CREATE INDEX idx_environments_product_id ON environments (product_id);
CREATE INDEX idx_environments_status ON environments (status);

-- Constraints
ALTER TABLE environments ADD CONSTRAINT chk_environments_status
    CHECK (status IN ('active', 'archived'));

ALTER TABLE environments ADD CONSTRAINT chk_environments_tier
    CHECK (tier IN ('development', 'staging', 'production'));

ALTER TABLE environments ADD CONSTRAINT chk_environments_slug_format
    CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$' AND LENGTH(slug) >= 3);

-- Updated_at trigger
CREATE TRIGGER trg_environments_updated_at
    BEFORE UPDATE ON environments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
