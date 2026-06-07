CREATE TABLE IF NOT EXISTS health_check_configs (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    component_id    UUID NOT NULL UNIQUE REFERENCES components(id) ON DELETE CASCADE,
    check_interval_ms BIGINT NOT NULL DEFAULT 60000,
    timeout_ms      BIGINT NOT NULL DEFAULT 10000,
    retries         INT NOT NULL DEFAULT 3,
    enabled         BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_check_configs_tenant ON health_check_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_health_check_configs_component ON health_check_configs(component_id);
