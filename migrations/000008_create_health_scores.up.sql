-- 000008_create_health_scores.up.sql
-- Computed health scores per component, stored for historical analysis.

CREATE TABLE IF NOT EXISTS health_scores (
    id           UUID PRIMARY KEY,
    tenant_id    UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    component_id UUID         NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    score        INTEGER      NOT NULL CHECK (score >= 0 AND score <= 100),
    status       VARCHAR(20)  NOT NULL DEFAULT 'unknown',
    reasons      JSONB        NOT NULL DEFAULT '[]',
    computed_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_health_scores_component_time ON health_scores (component_id, computed_at DESC);
CREATE INDEX idx_health_scores_tenant_time ON health_scores (tenant_id, computed_at DESC);
CREATE INDEX idx_health_scores_status ON health_scores (status) WHERE status != 'healthy';
CREATE INDEX idx_health_scores_score ON health_scores (score) WHERE score < 90;

ALTER TABLE health_scores ADD CONSTRAINT chk_health_scores_status
    CHECK (status IN ('healthy', 'degraded', 'critical', 'unknown'));
