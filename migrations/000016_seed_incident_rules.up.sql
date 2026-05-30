-- Seed incident rules for the default tenant.
-- This migration inserts standard incident detection rules.
-- They are inserted only if no rules exist yet (idempotent).

-- Rule 1: Database Connection Pool Exhaustion
INSERT INTO incident_rules (id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at)
SELECT
    '019e0000-0001-7000-8000-000000000001',
    t.id,
    'Database Connection Pool Exhaustion',
    'Triggers when database active connections exceed 90% of the pool limit',
    NULL,
    'database',
    '{"metric_type": "connection_pool_usage_percent", "operator": "gt", "threshold": 90}'::jsonb,
    'critical',
    300,
    true,
    NOW(),
    NOW()
FROM tenants t
WHERE t.slug = 'gymflow-track'
AND NOT EXISTS (
    SELECT 1 FROM incident_rules WHERE name = 'Database Connection Pool Exhaustion' AND tenant_id = t.id
);

-- Rule 2: Redis Memory Pressure
INSERT INTO incident_rules (id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at)
SELECT
    '019e0000-0001-7000-8000-000000000002',
    t.id,
    'Redis Memory Pressure',
    'Triggers when Redis memory usage exceeds 85% of maxmemory',
    NULL,
    'redis',
    '{"metric_type": "memory_usage_percent", "operator": "gt", "threshold": 85}'::jsonb,
    'major',
    600,
    true,
    NOW(),
    NOW()
FROM tenants t
WHERE t.slug = 'gymflow-track'
AND NOT EXISTS (
    SELECT 1 FROM incident_rules WHERE name = 'Redis Memory Pressure' AND tenant_id = t.id
);

-- Rule 3: High CPU Usage
INSERT INTO incident_rules (id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at)
SELECT
    '019e0000-0001-7000-8000-000000000003',
    t.id,
    'High CPU Usage',
    'Triggers when server CPU usage exceeds 90% sustained',
    NULL,
    'server',
    '{"metric_type": "cpu_percent", "operator": "gt", "threshold": 90}'::jsonb,
    'major',
    300,
    true,
    NOW(),
    NOW()
FROM tenants t
WHERE t.slug = 'gymflow-track'
AND NOT EXISTS (
    SELECT 1 FROM incident_rules WHERE name = 'High CPU Usage' AND tenant_id = t.id
);

-- Rule 4: Database Replication Lag
INSERT INTO incident_rules (id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at)
SELECT
    '019e0000-0001-7000-8000-000000000004',
    t.id,
    'Database Replication Lag',
    'Triggers when replication lag exceeds 30 seconds',
    NULL,
    'database',
    '{"metric_type": "replication_lag_seconds", "operator": "gt", "threshold": 30}'::jsonb,
    'major',
    600,
    true,
    NOW(),
    NOW()
FROM tenants t
WHERE t.slug = 'gymflow-track'
AND NOT EXISTS (
    SELECT 1 FROM incident_rules WHERE name = 'Database Replication Lag' AND tenant_id = t.id
);

-- Rule 5: Disk Space Critical
INSERT INTO incident_rules (id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at)
SELECT
    '019e0000-0001-7000-8000-000000000005',
    t.id,
    'Disk Space Critical',
    'Triggers when disk usage exceeds 90%',
    NULL,
    'server',
    '{"metric_type": "disk_usage_percent", "operator": "gt", "threshold": 90}'::jsonb,
    'critical',
    900,
    true,
    NOW(),
    NOW()
FROM tenants t
WHERE t.slug = 'gymflow-track'
AND NOT EXISTS (
    SELECT 1 FROM incident_rules WHERE name = 'Disk Space Critical' AND tenant_id = t.id
);

-- Rule 6: Memory Usage Warning
INSERT INTO incident_rules (id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at)
SELECT
    '019e0000-0001-7000-8000-000000000006',
    t.id,
    'Memory Usage Warning',
    'Triggers when memory usage exceeds 85%',
    NULL,
    'server',
    '{"metric_type": "memory_percent", "operator": "gt", "threshold": 85}'::jsonb,
    'warning',
    300,
    true,
    NOW(),
    NOW()
FROM tenants t
WHERE t.slug = 'gymflow-track'
AND NOT EXISTS (
    SELECT 1 FROM incident_rules WHERE name = 'Memory Usage Warning' AND tenant_id = t.id
);
