"use client";

import { motion } from "framer-motion";
import { useState } from "react";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { IncidentList } from "@/components/incidents/incident-list";
import type { Incident } from "@/lib/types";
import { Brain, CheckCircle2, Clock, MessageSquare, Users } from "lucide-react";

const mockIncidents: Incident[] = [
  {
    id: "inc-001", tenant_id: "t-001", component_id: "comp-001", rule_id: "rule-001",
    title: "PostgreSQL connection pool exhausted",
    description: "Connection pool utilization exceeded 95% on prod-db-primary. Active connections: 24/25. Queries queuing.",
    severity: "critical", status: "investigating",
    occurred_at: "2026-06-02T20:00:00.000Z", version: 3,
  },
  {
    id: "inc-002", tenant_id: "t-001", component_id: "comp-002", rule_id: "rule-002",
    title: "Redis cache hit ratio below threshold",
    description: "Cache hit ratio dropped from 94% to 62% over the last 15 minutes. Eviction rate spiked.",
    severity: "warning", status: "open",
    occurred_at: "2026-06-02T19:20:00.000Z", version: 1,
  },
  {
    id: "inc-003", tenant_id: "t-001", component_id: "comp-003", rule_id: "rule-003",
    title: "API response latency spike (P99 > 2s)",
    description: "Backend API P99 latency increased to 2.4s. Correlated with database pool saturation.",
    severity: "major", status: "acknowledged",
    occurred_at: "2026-06-02T18:00:00.000Z",
    acknowledged_at: "2026-06-02T18:20:00.000Z", version: 2,
  },
];

export default function IncidentsPage() {
  const [selected, setSelected] = useState<Incident | null>(null);

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1 ml-16 lg:ml-64">
        <Header />
        <main className="p-6">
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6"
          >
            <h1 className="text-2xl font-bold">Incident War Room</h1>
            <p className="text-sm text-[var(--muted)] mt-1">
              Manage active incidents with AI-powered root cause analysis
            </p>
          </motion.div>

          <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
            {/* Incident list */}
            <div className="lg:col-span-2">
              <IncidentList
                incidents={mockIncidents}
                onSelect={setSelected}
                selectedId={selected?.id}
              />
            </div>

            {/* Incident detail panel */}
            <motion.div
              className="lg:col-span-3 rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-6"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
            >
              {selected ? (
                <div className="space-y-6">
                  {/* Header */}
                  <div>
                    <div className="flex items-center gap-2 mb-2">
                      <span className="text-xs font-mono text-[var(--muted)]">
                        {selected.id}
                      </span>
                      <span className="text-xs uppercase px-2 py-0.5 rounded-full bg-red-500/10 text-red-400 border border-red-500/20">
                        {selected.severity}
                      </span>
                    </div>
                    <h2 className="text-lg font-bold">{selected.title}</h2>
                    <p className="text-sm text-[var(--muted)] mt-2">
                      {selected.description}
                    </p>
                  </div>

                  {/* Timeline */}
                  <div>
                    <h3 className="text-xs uppercase tracking-wider text-[var(--muted)] font-semibold mb-3">
                      Timeline
                    </h3>
                    <div className="space-y-3">
                      <TimelineEntry
                        icon={<Clock className="h-3.5 w-3.5" />}
                        label="Detected"
                        time={selected.occurred_at}
                        color="var(--danger)"
                      />
                      {selected.acknowledged_at && (
                        <TimelineEntry
                          icon={<Users className="h-3.5 w-3.5" />}
                          label="Acknowledged"
                          time={selected.acknowledged_at}
                          color="var(--warning)"
                        />
                      )}
                      {selected.resolved_at && (
                        <TimelineEntry
                          icon={<CheckCircle2 className="h-3.5 w-3.5" />}
                          label="Resolved"
                          time={selected.resolved_at}
                          color="var(--success)"
                        />
                      )}
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex gap-3">
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg bg-[var(--accent)] text-white text-sm font-medium hover:opacity-90 transition-opacity">
                      <Brain className="h-4 w-4" />
                      Run AI Analysis
                    </button>
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg border border-[var(--card-border)] text-sm font-medium hover:bg-[var(--card-border)] transition-colors">
                      <CheckCircle2 className="h-4 w-4" />
                      Acknowledge
                    </button>
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg border border-[var(--card-border)] text-sm font-medium hover:bg-[var(--card-border)] transition-colors">
                      <MessageSquare className="h-4 w-4" />
                      Comment
                    </button>
                  </div>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-64 text-[var(--muted)]">
                  <p className="text-sm">Select an incident to view details</p>
                </div>
              )}
            </motion.div>
          </div>
        </main>
      </div>
    </div>
  );
}

function TimelineEntry({
  icon,
  label,
  time,
  color,
}: {
  icon: React.ReactNode;
  label: string;
  time: string;
  color: string;
}) {
  return (
    <div className="flex items-center gap-3">
      <div
        className="rounded-full p-1.5"
        style={{ backgroundColor: `${color}20`, color }}
      >
        {icon}
      </div>
      <div className="flex-1">
        <span className="text-xs font-medium">{label}</span>
      </div>
      <span className="text-xs text-[var(--muted)] font-mono">
        {new Date(time).toLocaleString()}
      </span>
    </div>
  );
}
