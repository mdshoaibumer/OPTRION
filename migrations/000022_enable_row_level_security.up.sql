-- 000022_enable_row_level_security.up.sql
-- Defense-in-depth: PostgreSQL Row-Level Security enforces tenant isolation
-- at the database level, preventing tenant data leakage even if application
-- code has a bug in the WHERE clause.

-- Enable RLS on all tenant-scoped tables
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE environments ENABLE ROW LEVEL SECURITY;
ALTER TABLE components ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE health_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE metric_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE health_scores ENABLE ROW LEVEL SECURITY;
ALTER TABLE component_status ENABLE ROW LEVEL SECURITY;
ALTER TABLE anomalies ENABLE ROW LEVEL SECURITY;
ALTER TABLE incidents ENABLE ROW LEVEL SECURITY;
ALTER TABLE incident_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE incident_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE incident_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE incident_timeline ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_channels ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_deliveries ENABLE ROW LEVEL SECURITY;
ALTER TABLE escalation_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_analyses ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_context_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_reports ENABLE ROW LEVEL SECURITY;
ALTER TABLE recommendations ENABLE ROW LEVEL SECURITY;
ALTER TABLE recommendation_reports ENABLE ROW LEVEL SECURITY;
ALTER TABLE recommendation_evidence ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;

-- Create application role for OPTRION application connections
-- (skip if already exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'optrion_app') THEN
        CREATE ROLE optrion_app;
    END IF;
END
$$;

-- RLS policies: application can only see rows matching current tenant
-- The application sets the tenant context via: SET LOCAL app.current_tenant_id = 'tenant-uuid';

-- Products
CREATE POLICY tenant_isolation_products ON products
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Environments
CREATE POLICY tenant_isolation_environments ON environments
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Components
CREATE POLICY tenant_isolation_components ON components
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Audit Events
CREATE POLICY tenant_isolation_audit_events ON audit_events
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Health Metrics
CREATE POLICY tenant_isolation_health_metrics ON health_metrics
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Metric Snapshots
CREATE POLICY tenant_isolation_metric_snapshots ON metric_snapshots
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Health Scores
CREATE POLICY tenant_isolation_health_scores ON health_scores
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Component Status
CREATE POLICY tenant_isolation_component_status ON component_status
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Anomalies
CREATE POLICY tenant_isolation_anomalies ON anomalies
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Incidents
CREATE POLICY tenant_isolation_incidents ON incidents
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Incident Events
CREATE POLICY tenant_isolation_incident_events ON incident_events
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Incident Rules
CREATE POLICY tenant_isolation_incident_rules ON incident_rules
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Incident Comments
CREATE POLICY tenant_isolation_incident_comments ON incident_comments
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Incident Timeline
CREATE POLICY tenant_isolation_incident_timeline ON incident_timeline
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Alert Rules
CREATE POLICY tenant_isolation_alert_rules ON alert_rules
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Alerts
CREATE POLICY tenant_isolation_alerts ON alerts
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Alert Channels
CREATE POLICY tenant_isolation_alert_channels ON alert_channels
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Alert Deliveries
CREATE POLICY tenant_isolation_alert_deliveries ON alert_deliveries
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Escalation Policies
CREATE POLICY tenant_isolation_escalation_policies ON escalation_policies
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Notification Templates
CREATE POLICY tenant_isolation_notification_templates ON notification_templates
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- AI Analyses
CREATE POLICY tenant_isolation_ai_analyses ON ai_analyses
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- AI Context Snapshots
CREATE POLICY tenant_isolation_ai_context_snapshots ON ai_context_snapshots
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- AI Reports
CREATE POLICY tenant_isolation_ai_reports ON ai_reports
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Recommendations
CREATE POLICY tenant_isolation_recommendations ON recommendations
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Recommendation Reports
CREATE POLICY tenant_isolation_recommendation_reports ON recommendation_reports
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Recommendation Evidence
CREATE POLICY tenant_isolation_recommendation_evidence ON recommendation_evidence
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- API Keys
CREATE POLICY tenant_isolation_api_keys ON api_keys
    FOR ALL TO optrion_app
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- IMPORTANT: The superuser/owner role bypasses RLS.
-- Only the optrion_app role is subject to these policies.
-- The migration runner uses the superuser role and is NOT affected by RLS.

COMMENT ON POLICY tenant_isolation_products ON products IS 
    'Defense-in-depth: ensures tenant isolation at database level even if application layer has a bug';
