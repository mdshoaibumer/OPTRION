// Core domain types for Optrion Dashboard

export type HealthStatus = "healthy" | "degraded" | "unhealthy" | "unknown";
export type IncidentStatus = "open" | "acknowledged" | "investigating" | "resolved" | "closed";
export type Severity = "info" | "warning" | "minor" | "major" | "critical";
export type ComponentKind = "database" | "cache" | "api" | "queue" | "service" | "server";

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  plan: string;
  status: string;
  created_at: string;
}

export interface Product {
  id: string;
  tenant_id: string;
  name: string;
  slug: string;
  description: string;
  created_at: string;
}

export interface Environment {
  id: string;
  tenant_id: string;
  product_id: string;
  name: string;
  slug: string;
  tier: string;
  created_at: string;
}

export interface Component {
  id: string;
  tenant_id: string;
  environment_id: string;
  name: string;
  slug: string;
  kind: ComponentKind;
  endpoint_url: string;
  status: HealthStatus;
  health_score: number;
  created_at: string;
}

export interface HealthScore {
  component_id: string;
  score: number;
  status: HealthStatus;
  last_check_at: string;
  metrics: MetricSnapshot[];
}

export interface MetricSnapshot {
  id: string;
  component_id: string;
  metric_name: string;
  value: number;
  unit: string;
  collected_at: string;
}

export interface Incident {
  id: string;
  tenant_id: string;
  component_id: string;
  rule_id: string;
  title: string;
  description: string;
  severity: Severity;
  status: IncidentStatus;
  occurred_at: string;
  acknowledged_at?: string;
  resolved_at?: string;
  closed_at?: string;
  correlation_id?: string;
  version: number;
}

export interface IncidentEvent {
  id: string;
  incident_id: string;
  type: string;
  payload: Record<string, unknown>;
  occurred_at: string;
}

export interface Alert {
  id: string;
  tenant_id: string;
  incident_id: string;
  rule_id: string;
  severity: Severity;
  status: string;
  title: string;
  message: string;
  created_at: string;
  delivered_at?: string;
}

export interface AIAnalysis {
  id: string;
  tenant_id: string;
  incident_id: string;
  provider: string;
  status: string;
  created_at: string;
  completed_at?: string;
}

export interface RootCauseReport {
  id: string;
  tenant_id: string;
  incident_id: string;
  likely_cause: string;
  affected_components: string[];
  confidence: number;
  investigation_hints: string[];
  created_at: string;
}

export interface Recommendation {
  id: string;
  tenant_id: string;
  incident_id: string;
  category: string;
  priority: string;
  title: string;
  description: string;
  confidence: number;
  risk_level: string;
  created_at: string;
}

// Dashboard-specific types
export interface DashboardStats {
  total_components: number;
  healthy_components: number;
  degraded_components: number;
  unhealthy_components: number;
  open_incidents: number;
  active_alerts: number;
  average_health_score: number;
}

export interface TopologyNode {
  id: string;
  name: string;
  kind: ComponentKind;
  status: HealthStatus;
  score: number;
  x?: number;
  y?: number;
}

export interface TopologyEdge {
  source: string;
  target: string;
  status: "active" | "degraded" | "broken";
}
