-- 000010_create_anomalies.up.sql
-- Records detected anomalies (deviations from expected behavior).

CREATE TABLE IF NOT EXISTS anomalies (
    id              UUID PRIMARY KEY,
    tenant_id       UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    component_id    UUID         NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    metric_id       UUID         NOT NULL REFERENCES health_metrics(id) ON DELETE CASCADE,
    metric_type     VARCHAR(50)  NOT NULL,
    severity        VARCHAR(20)  NOT NULL DEFAULT 'medium',
    title           VARCHAR(500) NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    expected_value  DOUBLE PRECISION NOT NULL,
    actual_value    DOUBLE PRECISION NOT NULL,
    resolved        BOOLEAN      NOT NULL DEFAULT FALSE,
    detected_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ
);

CREATE INDEX idx_anomalies_tenant_time ON anomalies (tenant_id, detected_at DESC);
CREATE INDEX idx_anomalies_component ON anomalies (component_id, detected_at DESC);
CREATE INDEX idx_anomalies_unresolved ON anomalies (tenant_id, resolved) WHERE resolved = FALSE;
CREATE INDEX idx_anomalies_severity ON anomalies (severity, detected_at DESC);

ALTER TABLE anomalies ADD CONSTRAINT chk_anomalies_severity
    CHECK (severity IN ('low', 'medium', 'high', 'critical'));
