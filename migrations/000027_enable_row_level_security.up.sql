-- Enable Row-Level Security on all tenant-scoped tables
-- This provides defense-in-depth: even if application code is bypassed,
-- the database itself enforces tenant isolation.

-- Helper: Set tenant context variable before queries
-- Usage: SET LOCAL app.tenant_id = '<uuid>';

-- TENANTS (tenants can only see themselves)
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_tenants ON tenants
    USING (id::text = current_setting('app.tenant_id', true));

-- PRODUCTS
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_products ON products
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- ENVIRONMENTS
ALTER TABLE environments ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_environments ON environments
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- COMPONENTS
ALTER TABLE components ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_components ON components
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- AUDIT EVENTS
ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_audit_events ON audit_events
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- HEALTH METRICS
ALTER TABLE health_metrics ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_health_metrics ON health_metrics
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- METRIC SNAPSHOTS
ALTER TABLE metric_snapshots ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_metric_snapshots ON metric_snapshots
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- HEALTH SCORES
ALTER TABLE health_scores ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_health_scores ON health_scores
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- COMPONENT STATUS
ALTER TABLE component_status ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_component_status ON component_status
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- ANOMALIES
ALTER TABLE anomalies ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_anomalies ON anomalies
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- INCIDENTS
ALTER TABLE incidents ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_incidents ON incidents
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- INCIDENT EVENTS
ALTER TABLE incident_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_incident_events ON incident_events
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- INCIDENT RULES
ALTER TABLE incident_rules ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_incident_rules ON incident_rules
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- INCIDENT COMMENTS
ALTER TABLE incident_comments ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_incident_comments ON incident_comments
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- INCIDENT TIMELINE
ALTER TABLE incident_timeline ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_incident_timeline ON incident_timeline
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- ALERT TABLES
ALTER TABLE alert_rules ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_alert_rules ON alert_rules
    USING (tenant_id::text = current_setting('app.tenant_id', true));

ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_alerts ON alerts
    USING (tenant_id::text = current_setting('app.tenant_id', true));

ALTER TABLE alert_channels ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_alert_channels ON alert_channels
    USING (tenant_id::text = current_setting('app.tenant_id', true));

ALTER TABLE alert_deliveries ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_alert_deliveries ON alert_deliveries
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- AI TABLES
ALTER TABLE ai_analyses ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_ai_analyses ON ai_analyses
    USING (tenant_id::text = current_setting('app.tenant_id', true));

ALTER TABLE ai_context_snapshots ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_ai_context_snapshots ON ai_context_snapshots
    USING (tenant_id::text = current_setting('app.tenant_id', true));

ALTER TABLE ai_reports ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_ai_reports ON ai_reports
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- RECOMMENDATION TABLES
ALTER TABLE recommendations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_recommendations ON recommendations
    USING (tenant_id::text = current_setting('app.tenant_id', true));

ALTER TABLE recommendation_reports ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_recommendation_reports ON recommendation_reports
    USING (tenant_id::text = current_setting('app.tenant_id', true));

ALTER TABLE recommendation_evidence ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_recommendation_evidence ON recommendation_evidence
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- API KEYS
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_api_keys ON api_keys
    USING (tenant_id::text = current_setting('app.tenant_id', true));

-- BYPASS POLICY for the application superuser role
-- The application connection uses this role for migrations and system operations
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'optrion_app') THEN
        CREATE ROLE optrion_app;
    END IF;
END
$$;

-- Superuser bypass: The application's main role bypasses RLS
-- Individual queries set app.tenant_id to scope access
ALTER TABLE tenants FORCE ROW LEVEL SECURITY;
ALTER TABLE products FORCE ROW LEVEL SECURITY;
ALTER TABLE environments FORCE ROW LEVEL SECURITY;
ALTER TABLE components FORCE ROW LEVEL SECURITY;
ALTER TABLE audit_events FORCE ROW LEVEL SECURITY;
ALTER TABLE health_metrics FORCE ROW LEVEL SECURITY;
ALTER TABLE metric_snapshots FORCE ROW LEVEL SECURITY;
ALTER TABLE health_scores FORCE ROW LEVEL SECURITY;
ALTER TABLE component_status FORCE ROW LEVEL SECURITY;
ALTER TABLE anomalies FORCE ROW LEVEL SECURITY;
ALTER TABLE incidents FORCE ROW LEVEL SECURITY;
ALTER TABLE incident_events FORCE ROW LEVEL SECURITY;
ALTER TABLE incident_rules FORCE ROW LEVEL SECURITY;
ALTER TABLE incident_comments FORCE ROW LEVEL SECURITY;
ALTER TABLE incident_timeline FORCE ROW LEVEL SECURITY;
ALTER TABLE alert_rules FORCE ROW LEVEL SECURITY;
ALTER TABLE alerts FORCE ROW LEVEL SECURITY;
ALTER TABLE alert_channels FORCE ROW LEVEL SECURITY;
ALTER TABLE alert_deliveries FORCE ROW LEVEL SECURITY;
ALTER TABLE ai_analyses FORCE ROW LEVEL SECURITY;
ALTER TABLE ai_context_snapshots FORCE ROW LEVEL SECURITY;
ALTER TABLE ai_reports FORCE ROW LEVEL SECURITY;
ALTER TABLE recommendations FORCE ROW LEVEL SECURITY;
ALTER TABLE recommendation_reports FORCE ROW LEVEL SECURITY;
ALTER TABLE recommendation_evidence FORCE ROW LEVEL SECURITY;
ALTER TABLE api_keys FORCE ROW LEVEL SECURITY;
