-- AI Root Cause Intelligence Layer Migrations (UP)
-- 1. ai_analyses
CREATE TABLE IF NOT EXISTS ai_analyses (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    incident_id UUID NOT NULL,
    context_id UUID NOT NULL,
    report_id UUID NOT NULL,
    provider VARCHAR(64) NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(32) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 2. ai_context_snapshots
CREATE TABLE IF NOT EXISTS ai_context_snapshots (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    incident_id UUID NOT NULL,
    snapshot BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 3. ai_reports
CREATE TABLE IF NOT EXISTS ai_reports (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    incident_id UUID NOT NULL,
    likely_cause TEXT NOT NULL,
    affected_components TEXT[] NOT NULL,
    confidence FLOAT8 NOT NULL,
    investigation_hints TEXT[] NOT NULL,
    raw_output BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for scalability
CREATE INDEX IF NOT EXISTS idx_ai_analyses_tenant_id ON ai_analyses(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_context_snapshots_tenant_id ON ai_context_snapshots(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_reports_tenant_id ON ai_reports(tenant_id);

-- All tables are immutable and auditable.
