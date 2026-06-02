-- 000022_enable_row_level_security.down.sql
-- Disable RLS on all tenant-scoped tables.

-- Drop all tenant isolation policies
DROP POLICY IF EXISTS tenant_isolation_products ON products;
DROP POLICY IF EXISTS tenant_isolation_environments ON environments;
DROP POLICY IF EXISTS tenant_isolation_components ON components;
DROP POLICY IF EXISTS tenant_isolation_audit_events ON audit_events;
DROP POLICY IF EXISTS tenant_isolation_health_metrics ON health_metrics;
DROP POLICY IF EXISTS tenant_isolation_metric_snapshots ON metric_snapshots;
DROP POLICY IF EXISTS tenant_isolation_health_scores ON health_scores;
DROP POLICY IF EXISTS tenant_isolation_component_status ON component_status;
DROP POLICY IF EXISTS tenant_isolation_anomalies ON anomalies;
DROP POLICY IF EXISTS tenant_isolation_incidents ON incidents;
DROP POLICY IF EXISTS tenant_isolation_incident_events ON incident_events;
DROP POLICY IF EXISTS tenant_isolation_incident_rules ON incident_rules;
DROP POLICY IF EXISTS tenant_isolation_incident_comments ON incident_comments;
DROP POLICY IF EXISTS tenant_isolation_incident_timeline ON incident_timeline;
DROP POLICY IF EXISTS tenant_isolation_alert_rules ON alert_rules;
DROP POLICY IF EXISTS tenant_isolation_alerts ON alerts;
DROP POLICY IF EXISTS tenant_isolation_alert_channels ON alert_channels;
DROP POLICY IF EXISTS tenant_isolation_alert_deliveries ON alert_deliveries;
DROP POLICY IF EXISTS tenant_isolation_escalation_policies ON escalation_policies;
DROP POLICY IF EXISTS tenant_isolation_notification_templates ON notification_templates;
DROP POLICY IF EXISTS tenant_isolation_ai_analyses ON ai_analyses;
DROP POLICY IF EXISTS tenant_isolation_ai_context_snapshots ON ai_context_snapshots;
DROP POLICY IF EXISTS tenant_isolation_ai_reports ON ai_reports;
DROP POLICY IF EXISTS tenant_isolation_recommendations ON recommendations;
DROP POLICY IF EXISTS tenant_isolation_recommendation_reports ON recommendation_reports;
DROP POLICY IF EXISTS tenant_isolation_recommendation_evidence ON recommendation_evidence;
DROP POLICY IF EXISTS tenant_isolation_api_keys ON api_keys;

-- Disable RLS on all tables
ALTER TABLE products DISABLE ROW LEVEL SECURITY;
ALTER TABLE environments DISABLE ROW LEVEL SECURITY;
ALTER TABLE components DISABLE ROW LEVEL SECURITY;
ALTER TABLE audit_events DISABLE ROW LEVEL SECURITY;
ALTER TABLE health_metrics DISABLE ROW LEVEL SECURITY;
ALTER TABLE metric_snapshots DISABLE ROW LEVEL SECURITY;
ALTER TABLE health_scores DISABLE ROW LEVEL SECURITY;
ALTER TABLE component_status DISABLE ROW LEVEL SECURITY;
ALTER TABLE anomalies DISABLE ROW LEVEL SECURITY;
ALTER TABLE incidents DISABLE ROW LEVEL SECURITY;
ALTER TABLE incident_events DISABLE ROW LEVEL SECURITY;
ALTER TABLE incident_rules DISABLE ROW LEVEL SECURITY;
ALTER TABLE incident_comments DISABLE ROW LEVEL SECURITY;
ALTER TABLE incident_timeline DISABLE ROW LEVEL SECURITY;
ALTER TABLE alert_rules DISABLE ROW LEVEL SECURITY;
ALTER TABLE alerts DISABLE ROW LEVEL SECURITY;
ALTER TABLE alert_channels DISABLE ROW LEVEL SECURITY;
ALTER TABLE alert_deliveries DISABLE ROW LEVEL SECURITY;
ALTER TABLE escalation_policies DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates DISABLE ROW LEVEL SECURITY;
ALTER TABLE ai_analyses DISABLE ROW LEVEL SECURITY;
ALTER TABLE ai_context_snapshots DISABLE ROW LEVEL SECURITY;
ALTER TABLE ai_reports DISABLE ROW LEVEL SECURITY;
ALTER TABLE recommendations DISABLE ROW LEVEL SECURITY;
ALTER TABLE recommendation_reports DISABLE ROW LEVEL SECURITY;
ALTER TABLE recommendation_evidence DISABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys DISABLE ROW LEVEL SECURITY;

-- Drop the application role
DROP ROLE IF EXISTS optrion_app;
