"use client";

import { motion } from "framer-motion";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { AlertFeed } from "@/components/alerts/alert-feed";
import type { Alert } from "@/lib/types";

const mockAlerts: Alert[] = [
  {
    id: "alert-001", tenant_id: "t-001", incident_id: "inc-001", rule_id: "rule-001",
    severity: "critical", status: "delivered",
    title: "CRITICAL: PostgreSQL pool exhausted",
    message: "Connection pool on prod-db-primary has exceeded 95% capacity. Immediate action required.",
    created_at: "2026-06-02T20:00:00.000Z",
    delivered_at: "2026-06-02T20:00:10.000Z",
  },
  {
    id: "alert-002", tenant_id: "t-001", incident_id: "inc-003", rule_id: "rule-003",
    severity: "major", status: "delivered",
    title: "MAJOR: API latency exceeds SLO",
    message: "P99 latency for /api/v1/members endpoint is 2.4s (SLO: 1s). Degraded user experience detected.",
    created_at: "2026-06-02T18:00:00.000Z",
    delivered_at: "2026-06-02T18:00:05.000Z",
  },
  {
    id: "alert-003", tenant_id: "t-001", incident_id: "inc-002", rule_id: "rule-002",
    severity: "warning", status: "pending",
    title: "WARNING: Redis cache degraded",
    message: "Cache hit ratio below 80% threshold. Current: 62%. Eviction rate elevated.",
    created_at: "2026-06-02T19:20:00.000Z",
  },
  {
    id: "alert-004", tenant_id: "t-001", incident_id: "inc-004", rule_id: "rule-004",
    severity: "info", status: "delivered",
    title: "INFO: Scheduled maintenance window starting",
    message: "Maintenance window for prod-db-primary starts in 30 minutes. Alerts will be suppressed.",
    created_at: "2026-06-02T20:30:00.000Z",
    delivered_at: "2026-06-02T20:30:05.000Z",
  },
  {
    id: "alert-005", tenant_id: "t-001", incident_id: "inc-001", rule_id: "rule-001",
    severity: "critical", status: "delivered",
    title: "CRITICAL: Incident escalated to Major",
    message: "PostgreSQL pool exhaustion incident escalated after 15 minutes without acknowledgment.",
    created_at: "2026-06-02T20:35:00.000Z",
    delivered_at: "2026-06-02T20:35:02.000Z",
  },
];

export default function AlertsPage() {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1 ml-16 lg:ml-64">
        <Header />
        <main className="p-6 space-y-6">
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
          >
            <h1 className="text-2xl font-bold">Alert Feed</h1>
            <p className="text-sm text-muted mt-1">
              Real-time alert stream with severity filtering and delivery status
            </p>
          </motion.div>

          {/* Filters */}
          <div className="flex items-center gap-3">
            {["all", "critical", "major", "warning", "info"].map((filter) => (
              <button
                key={filter}
                className="px-3 py-1.5 rounded-lg text-xs font-medium border border-card-border hover:border-accent hover:text-accent transition-colors capitalize"
              >
                {filter}
              </button>
            ))}
          </div>

          <div className="max-w-3xl">
            <AlertFeed alerts={mockAlerts} />
          </div>
        </main>
      </div>
    </div>
  );
}
