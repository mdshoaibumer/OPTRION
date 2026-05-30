-- Recommendation Intelligence Layer Migrations (UP)
-- 1. recommendations
CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    incident_id UUID NOT NULL,
    report_id UUID NOT NULL,
    category VARCHAR(64) NOT NULL,
    priority VARCHAR(32) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    confidence FLOAT8 NOT NULL,
    risk_level VARCHAR(32) NOT NULL,
    evidence_ids UUID[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 2. recommendation_reports
CREATE TABLE IF NOT EXISTS recommendation_reports (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    incident_id UUID NOT NULL,
    recommendations UUID[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 3. recommendation_evidence
CREATE TABLE IF NOT EXISTS recommendation_evidence (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    incident_id UUID NOT NULL,
    recommendation_id UUID NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for scalability
CREATE INDEX IF NOT EXISTS idx_recommendations_tenant_id ON recommendations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_recommendation_reports_tenant_id ON recommendation_reports(tenant_id);
CREATE INDEX IF NOT EXISTS idx_recommendation_evidence_tenant_id ON recommendation_evidence(tenant_id);

-- All tables are immutable and auditable.
