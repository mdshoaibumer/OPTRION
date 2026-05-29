-- 000002_create_products.up.sql
-- Products belong to a tenant. A product is a logical grouping of services.

CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY,
    tenant_id   UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_products_tenant_slug UNIQUE (tenant_id, slug)
);

-- Indexes
CREATE INDEX idx_products_tenant_id ON products (tenant_id);
CREATE INDEX idx_products_status ON products (status);

-- Constraints
ALTER TABLE products ADD CONSTRAINT chk_products_status
    CHECK (status IN ('active', 'archived'));

ALTER TABLE products ADD CONSTRAINT chk_products_slug_format
    CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$' AND LENGTH(slug) >= 3);

-- Updated_at trigger
CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
