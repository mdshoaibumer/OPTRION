"use client";

import { motion } from "framer-motion";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { Brain, Lightbulb, Target, Zap } from "lucide-react";

interface AIInsight {
  incidentId: string;
  incidentTitle: string;
  rootCause: string;
  confidence: number;
  affectedComponents: string[];
  hints: string[];
  recommendations: { title: string; priority: string; risk: string }[];
}

const mockInsights: AIInsight[] = [
  {
    incidentId: "inc-001",
    incidentTitle: "PostgreSQL connection pool exhausted",
    rootCause: "Long-running transaction from the membership sync job is holding connections open beyond pool timeout. Query on members table with missing index on (tenant_id, status) causes sequential scan.",
    confidence: 0.87,
    affectedComponents: ["PostgreSQL Primary", "Backend API"],
    hints: [
      "Check pg_stat_activity for long-running queries",
      "Review membership_sync_job transaction boundaries",
      "Verify index exists on members(tenant_id, status)",
    ],
    recommendations: [
      { title: "Add composite index on members(tenant_id, status)", priority: "high", risk: "low" },
      { title: "Set statement_timeout to 30s for sync jobs", priority: "high", risk: "low" },
      { title: "Increase pool size from 25 to 50", priority: "medium", risk: "medium" },
    ],
  },
  {
    incidentId: "inc-002",
    incidentTitle: "Redis cache hit ratio below threshold",
    rootCause: "Memory pressure causing aggressive eviction of cached member profiles. Recent deployment increased object sizes by ~40% without adjusting maxmemory configuration.",
    confidence: 0.72,
    affectedComponents: ["Redis Sessions"],
    hints: [
      "Compare Redis INFO memory before/after recent deploy",
      "Check serialized object sizes for MemberProfile",
      "Review eviction policy (allkeys-lru vs volatile-lru)",
    ],
    recommendations: [
      { title: "Increase maxmemory from 256MB to 512MB", priority: "high", risk: "low" },
      { title: "Compress cached objects with msgpack", priority: "medium", risk: "low" },
      { title: "Add TTL to member profile cache entries", priority: "low", risk: "low" },
    ],
  },
];

export default function AIPage() {
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
            <h1 className="text-2xl font-bold">AI Insights</h1>
            <p className="text-sm text-muted mt-1">
              Root cause analysis and intelligent recommendations powered by AI
            </p>
          </motion.div>

          {/* AI Insights */}
          <div className="space-y-6">
            {mockInsights.map((insight, idx) => (
              <motion.div
                key={insight.incidentId}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.1 }}
                className="rounded-xl border border-card-border bg-card overflow-hidden"
              >
                {/* Header */}
                <div className="px-6 py-4 border-b border-card-border flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="rounded-lg p-2 bg-(--accent-glow)">
                      <Brain className="h-5 w-5 text-accent" />
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold">{insight.incidentTitle}</h3>
                      <span className="text-[10px] text-muted font-mono">
                        {insight.incidentId}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted">Confidence</span>
                    <span
                      className="text-sm font-bold font-mono"
                      style={{
                        color: insight.confidence >= 0.8 ? "var(--success)" : insight.confidence >= 0.6 ? "var(--warning)" : "var(--danger)",
                      }}
                    >
                      {Math.round(insight.confidence * 100)}%
                    </span>
                  </div>
                </div>

                <div className="p-6 grid grid-cols-1 lg:grid-cols-2 gap-6">
                  {/* Root Cause */}
                  <div>
                    <div className="flex items-center gap-2 mb-3">
                      <Target className="h-4 w-4 text-danger" />
                      <h4 className="text-xs uppercase tracking-wider font-semibold text-muted">
                        Root Cause
                      </h4>
                    </div>
                    <p className="text-sm leading-relaxed">{insight.rootCause}</p>

                    <div className="mt-4">
                      <h4 className="text-xs uppercase tracking-wider font-semibold text-muted mb-2">
                        Affected Components
                      </h4>
                      <div className="flex flex-wrap gap-2">
                        {insight.affectedComponents.map((comp) => (
                          <span
                            key={comp}
                            className="text-xs px-2 py-1 rounded-md bg-card-border text-foreground"
                          >
                            {comp}
                          </span>
                        ))}
                      </div>
                    </div>

                    <div className="mt-4">
                      <div className="flex items-center gap-2 mb-2">
                        <Lightbulb className="h-4 w-4 text-warning" />
                        <h4 className="text-xs uppercase tracking-wider font-semibold text-muted">
                          Investigation Hints
                        </h4>
                      </div>
                      <ul className="space-y-1.5">
                        {insight.hints.map((hint, i) => (
                          <li
                            key={i}
                            className="text-xs text-muted flex items-start gap-2"
                          >
                            <span className="text-accent mt-0.5">→</span>
                            {hint}
                          </li>
                        ))}
                      </ul>
                    </div>
                  </div>

                  {/* Recommendations */}
                  <div>
                    <div className="flex items-center gap-2 mb-3">
                      <Zap className="h-4 w-4 text-accent" />
                      <h4 className="text-xs uppercase tracking-wider font-semibold text-muted">
                        Recommendations
                      </h4>
                    </div>
                    <div className="space-y-3">
                      {insight.recommendations.map((rec, i) => (
                        <div
                          key={i}
                          className="rounded-lg border border-card-border p-3"
                        >
                          <div className="flex items-center justify-between">
                            <span className="text-sm font-medium">{rec.title}</span>
                          </div>
                          <div className="flex items-center gap-3 mt-2">
                            <span
                              className="text-[10px] uppercase font-semibold px-2 py-0.5 rounded-full"
                              style={{
                                backgroundColor: rec.priority === "high" ? "var(--danger-glow)" : rec.priority === "medium" ? "var(--warning-glow)" : "var(--success-glow)",
                                color: rec.priority === "high" ? "var(--danger)" : rec.priority === "medium" ? "var(--warning)" : "var(--success)",
                              }}
                            >
                              {rec.priority}
                            </span>
                            <span className="text-[10px] text-muted">
                              Risk: {rec.risk}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </main>
      </div>
    </div>
  );
}
