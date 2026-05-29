-- 000001_create_tenants.up.sql
-- Creates the tenants table — root of the multi-tenant hierarchy.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS tenants (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    plan        VARCHAR(50)  NOT NULL DEFAULT 'free',
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    settings    JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_tenants_slug ON tenants (slug);
CREATE INDEX idx_tenants_status ON tenants (status);
CREATE INDEX idx_tenants_created_at ON tenants (created_at DESC);

-- Constraints
ALTER TABLE tenants ADD CONSTRAINT chk_tenants_status
    CHECK (status IN ('active', 'suspended', 'deactivated'));

ALTER TABLE tenants ADD CONSTRAINT chk_tenants_plan
    CHECK (plan IN ('free', 'starter', 'professional', 'enterprise'));

ALTER TABLE tenants ADD CONSTRAINT chk_tenants_slug_format
    CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$' AND LENGTH(slug) >= 3);

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
