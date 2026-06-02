"use client";

import { motion } from "framer-motion";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { StatsCards } from "@/components/dashboard/stats-cards";
import { HealthScoreRing } from "@/components/health/health-score-ring";
import { IncidentList } from "@/components/incidents/incident-list";
import { AlertFeed } from "@/components/alerts/alert-feed";
import type { DashboardStats, Incident, Alert } from "@/lib/types";

// Mock data for initial render (replaced by API calls with TanStack Query)
const mockStats: DashboardStats = {
  total_components: 6,
  healthy_components: 4,
  degraded_components: 1,
  unhealthy_components: 1,
  open_incidents: 2,
  active_alerts: 3,
  average_health_score: 78,
};

const mockIncidents: Incident[] = [
  {
    id: "inc-001",
    tenant_id: "t-001",
    component_id: "comp-001",
    rule_id: "rule-001",
    title: "PostgreSQL connection pool exhausted",
    description: "Connection pool utilization exceeded 95%",
    severity: "critical",
    status: "investigating",
    occurred_at: "2026-06-02T20:00:00.000Z",
    version: 3,
  },
  {
    id: "inc-002",
    tenant_id: "t-001",
    component_id: "comp-002",
    rule_id: "rule-002",
    title: "Redis cache hit ratio below threshold",
    description: "Cache hit ratio dropped to 62%",
    severity: "warning",
    status: "open",
    occurred_at: "2026-06-02T19:20:00.000Z",
    version: 1,
  },
  {
    id: "inc-003",
    tenant_id: "t-001",
    component_id: "comp-003",
    rule_id: "rule-003",
    title: "API response latency spike (P99 > 2s)",
    description: "Backend API P99 latency increased to 2.4s",
    severity: "major",
    status: "acknowledged",
    occurred_at: "2026-06-02T18:00:00.000Z",
    acknowledged_at: "2026-06-02T18:20:00.000Z",
    version: 2,
  },
];

const mockAlerts: Alert[] = [
  {
    id: "alert-001",
    tenant_id: "t-001",
    incident_id: "inc-001",
    rule_id: "rule-001",
    severity: "critical",
    status: "delivered",
    title: "CRITICAL: PostgreSQL pool exhausted",
    message: "Connection pool on prod-db-primary has exceeded 95% capacity. Immediate action required.",
    created_at: "2026-06-02T20:00:00.000Z",
    delivered_at: "2026-06-02T20:00:10.000Z",
  },
  {
    id: "alert-002",
    tenant_id: "t-001",
    incident_id: "inc-003",
    rule_id: "rule-003",
    severity: "major",
    status: "delivered",
    title: "MAJOR: API latency exceeds SLO",
    message: "P99 latency for /api/v1/members endpoint is 2.4s (SLO: 1s)",
    created_at: "2026-06-02T18:00:00.000Z",
    delivered_at: "2026-06-02T18:00:05.000Z",
  },
  {
    id: "alert-003",
    tenant_id: "t-001",
    incident_id: "inc-002",
    rule_id: "rule-002",
    severity: "warning",
    status: "pending",
    title: "WARNING: Redis cache degraded",
    message: "Cache hit ratio below 80% threshold",
    created_at: "2026-06-02T19:20:00.000Z",
  },
];

export default function DashboardPage() {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1 ml-16 lg:ml-64">
        <Header />
        <main className="p-6 space-y-6">
          {/* Page title */}
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
          >
            <h1 className="text-2xl font-bold">Engineering Intelligence</h1>
            <p className="text-sm text-muted mt-1">
              Real-time health monitoring and incident detection
            </p>
          </motion.div>

          {/* Stats row */}
          <StatsCards stats={mockStats} />

          {/* Main content grid */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Health Overview */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="lg:col-span-1 rounded-xl border border-card-border bg-card p-6"
            >
              <h2 className="text-sm font-semibold mb-6">System Health</h2>
              <div className="flex flex-col items-center">
                <HealthScoreRing
                  score={mockStats.average_health_score}
                  status={
                    mockStats.average_health_score >= 80
                      ? "healthy"
                      : mockStats.average_health_score >= 50
                      ? "degraded"
                      : "unhealthy"
                  }
                  size="lg"
                  label="Overall Score"
                />
                <div className="grid grid-cols-3 gap-6 mt-8 w-full">
                  <HealthScoreRing score={95} status="healthy" size="sm" label="API" />
                  <HealthScoreRing score={62} status="degraded" size="sm" label="DB" />
                  <HealthScoreRing score={88} status="healthy" size="sm" label="Cache" />
                </div>
              </div>
            </motion.div>

            {/* Incidents */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 }}
              className="lg:col-span-1 rounded-xl border border-card-border bg-card p-6"
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-semibold">Active Incidents</h2>
                <span className="text-xs px-2 py-0.5 rounded-full bg-red-500/10 text-red-400 font-mono">
                  {mockIncidents.filter((i) => i.status !== "resolved" && i.status !== "closed").length}
                </span>
              </div>
              <IncidentList incidents={mockIncidents} />
            </motion.div>

            {/* Alert Feed */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4 }}
              className="lg:col-span-1 rounded-xl border border-card-border bg-card p-6"
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-semibold">Alert Feed</h2>
                <span className="text-xs px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 font-mono">
                  {mockAlerts.length}
                </span>
              </div>
              <AlertFeed alerts={mockAlerts} />
            </motion.div>
          </div>
        </main>
      </div>
    </div>
  );
}
