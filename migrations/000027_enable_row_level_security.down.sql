-- Disable Row-Level Security on all tables
ALTER TABLE tenants DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_tenants ON tenants;

ALTER TABLE products DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_products ON products;

ALTER TABLE environments DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_environments ON environments;

ALTER TABLE components DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_components ON components;

ALTER TABLE audit_events DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_audit_events ON audit_events;

ALTER TABLE health_metrics DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_health_metrics ON health_metrics;

ALTER TABLE metric_snapshots DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_metric_snapshots ON metric_snapshots;

ALTER TABLE health_scores DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_health_scores ON health_scores;

ALTER TABLE component_status DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_component_status ON component_status;

ALTER TABLE anomalies DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_anomalies ON anomalies;

ALTER TABLE incidents DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_incidents ON incidents;

ALTER TABLE incident_events DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_incident_events ON incident_events;

ALTER TABLE incident_rules DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_incident_rules ON incident_rules;

ALTER TABLE incident_comments DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_incident_comments ON incident_comments;

ALTER TABLE incident_timeline DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_incident_timeline ON incident_timeline;

ALTER TABLE alert_rules DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_alert_rules ON alert_rules;

ALTER TABLE alerts DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_alerts ON alerts;

ALTER TABLE alert_channels DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_alert_channels ON alert_channels;

ALTER TABLE alert_deliveries DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_alert_deliveries ON alert_deliveries;

ALTER TABLE ai_analyses DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_ai_analyses ON ai_analyses;

ALTER TABLE ai_context_snapshots DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_ai_context_snapshots ON ai_context_snapshots;

ALTER TABLE ai_reports DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_ai_reports ON ai_reports;

ALTER TABLE recommendations DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_recommendations ON recommendations;

ALTER TABLE recommendation_reports DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_recommendation_reports ON recommendation_reports;

ALTER TABLE recommendation_evidence DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_recommendation_evidence ON recommendation_evidence;

ALTER TABLE api_keys DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_api_keys ON api_keys;

-- Remove NO FORCE from tables
ALTER TABLE tenants NO FORCE ROW LEVEL SECURITY;
ALTER TABLE products NO FORCE ROW LEVEL SECURITY;
ALTER TABLE environments NO FORCE ROW LEVEL SECURITY;
ALTER TABLE components NO FORCE ROW LEVEL SECURITY;
ALTER TABLE audit_events NO FORCE ROW LEVEL SECURITY;
ALTER TABLE health_metrics NO FORCE ROW LEVEL SECURITY;
ALTER TABLE metric_snapshots NO FORCE ROW LEVEL SECURITY;
ALTER TABLE health_scores NO FORCE ROW LEVEL SECURITY;
ALTER TABLE component_status NO FORCE ROW LEVEL SECURITY;
ALTER TABLE anomalies NO FORCE ROW LEVEL SECURITY;
ALTER TABLE incidents NO FORCE ROW LEVEL SECURITY;
ALTER TABLE incident_events NO FORCE ROW LEVEL SECURITY;
ALTER TABLE incident_rules NO FORCE ROW LEVEL SECURITY;
ALTER TABLE incident_comments NO FORCE ROW LEVEL SECURITY;
ALTER TABLE incident_timeline NO FORCE ROW LEVEL SECURITY;
ALTER TABLE alert_rules NO FORCE ROW LEVEL SECURITY;
ALTER TABLE alerts NO FORCE ROW LEVEL SECURITY;
ALTER TABLE alert_channels NO FORCE ROW LEVEL SECURITY;
ALTER TABLE alert_deliveries NO FORCE ROW LEVEL SECURITY;
ALTER TABLE ai_analyses NO FORCE ROW LEVEL SECURITY;
ALTER TABLE ai_context_snapshots NO FORCE ROW LEVEL SECURITY;
ALTER TABLE ai_reports NO FORCE ROW LEVEL SECURITY;
ALTER TABLE recommendations NO FORCE ROW LEVEL SECURITY;
ALTER TABLE recommendation_reports NO FORCE ROW LEVEL SECURITY;
ALTER TABLE recommendation_evidence NO FORCE ROW LEVEL SECURITY;
ALTER TABLE api_keys NO FORCE ROW LEVEL SECURITY;
