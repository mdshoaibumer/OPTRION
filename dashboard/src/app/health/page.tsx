"use client";

import { motion } from "framer-motion";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { HealthScoreRing } from "@/components/health/health-score-ring";
import { Activity, TrendingDown, TrendingUp } from "lucide-react";

interface ComponentHealth {
  id: string;
  name: string;
  kind: string;
  score: number;
  status: "healthy" | "degraded" | "unhealthy" | "unknown";
  trend: "up" | "down" | "stable";
  metrics: { name: string; value: string; status: "ok" | "warn" | "crit" }[];
}

const mockHealth: ComponentHealth[] = [
  {
    id: "c1", name: "Backend API", kind: "api", score: 95, status: "healthy", trend: "stable",
    metrics: [
      { name: "Latency P99", value: "245ms", status: "ok" },
      { name: "Error Rate", value: "0.02%", status: "ok" },
      { name: "Throughput", value: "1.2k rps", status: "ok" },
    ],
  },
  {
    id: "c2", name: "PostgreSQL Primary", kind: "database", score: 62, status: "degraded", trend: "down",
    metrics: [
      { name: "Connections", value: "24/25", status: "crit" },
      { name: "Query Latency", value: "450ms", status: "warn" },
      { name: "Deadlocks", value: "0", status: "ok" },
    ],
  },
  {
    id: "c3", name: "Redis Cache", kind: "cache", score: 88, status: "healthy", trend: "up",
    metrics: [
      { name: "Hit Ratio", value: "94%", status: "ok" },
      { name: "Memory", value: "62%", status: "ok" },
      { name: "Evictions", value: "12/min", status: "ok" },
    ],
  },
  {
    id: "c4", name: "Redis Sessions", kind: "cache", score: 28, status: "unhealthy", trend: "down",
    metrics: [
      { name: "Hit Ratio", value: "34%", status: "crit" },
      { name: "Memory", value: "96%", status: "crit" },
      { name: "Evictions", value: "2.4k/min", status: "crit" },
    ],
  },
  {
    id: "c5", name: "NGINX Proxy", kind: "server", score: 99, status: "healthy", trend: "stable",
    metrics: [
      { name: "CPU", value: "12%", status: "ok" },
      { name: "RAM", value: "34%", status: "ok" },
      { name: "Active Conns", value: "847", status: "ok" },
    ],
  },
];

export default function HealthPage() {
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
            <h1 className="text-2xl font-bold">Health Scores</h1>
            <p className="text-sm text-muted mt-1">
              Component-level health monitoring with metric breakdown
            </p>
          </motion.div>

          {/* Component health cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {mockHealth.map((comp, idx) => (
              <motion.div
                key={comp.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.05 }}
                className="rounded-xl border border-card-border bg-card p-5"
              >
                {/* Component header */}
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <h3 className="text-sm font-semibold">{comp.name}</h3>
                    <span className="text-[10px] uppercase tracking-wider text-muted">
                      {comp.kind}
                    </span>
                  </div>
                  <HealthScoreRing
                    score={comp.score}
                    status={comp.status}
                    size="sm"
                    animate={false}
                  />
                </div>

                {/* Trend indicator */}
                <div className="flex items-center gap-1 mb-4">
                  {comp.trend === "up" && (
                    <TrendingUp className="h-3.5 w-3.5 text-success" />
                  )}
                  {comp.trend === "down" && (
                    <TrendingDown className="h-3.5 w-3.5 text-danger" />
                  )}
                  {comp.trend === "stable" && (
                    <Activity className="h-3.5 w-3.5 text-muted" />
                  )}
                  <span className="text-[10px] text-muted capitalize">
                    {comp.trend} trend
                  </span>
                </div>

                {/* Metrics */}
                <div className="space-y-2">
                  {comp.metrics.map((metric) => (
                    <div
                      key={metric.name}
                      className="flex items-center justify-between text-xs"
                    >
                      <span className="text-muted">{metric.name}</span>
                      <span
                        className="font-mono font-medium"
                        style={{
                          color:
                            metric.status === "ok"
                              ? "var(--foreground)"
                              : metric.status === "warn"
                              ? "var(--warning)"
                              : "var(--danger)",
                        }}
                      >
                        {metric.value}
                      </span>
                    </div>
                  ))}
                </div>
              </motion.div>
            ))}
          </div>
        </main>
      </div>
    </div>
  );
}
