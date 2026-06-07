"use client";

import { motion } from "framer-motion";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { HealthScoreRing } from "@/components/health/health-score-ring";
import { CardSkeleton, QueryError } from "@/components/ui/loading";
import { useComponents } from "@/lib/hooks";
import { Activity, TrendingDown, TrendingUp } from "lucide-react";
import type { HealthStatus } from "@/lib/types";

export default function HealthPage() {
  const { data: components, isLoading, error, refetch } = useComponents();

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

          {isLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <CardSkeleton key={i} />
              ))}
            </div>
          ) : error ? (
            <QueryError error={error as Error} onRetry={() => refetch()} />
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
              {(components || []).map((comp, idx) => {
                const status: HealthStatus = comp.health_score >= 80
                  ? "healthy"
                  : comp.health_score >= 50
                  ? "degraded"
                  : "unhealthy";

                return (
                  <motion.div
                    key={comp.id}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: idx * 0.05 }}
                    className="rounded-xl border border-card-border bg-card p-5"
                  >
                    <div className="flex items-center justify-between mb-4">
                      <div>
                        <h3 className="text-sm font-semibold">{comp.name}</h3>
                        <span className="text-[10px] uppercase tracking-wider text-muted">
                          {comp.kind}
                        </span>
                      </div>
                      <HealthScoreRing
                        score={comp.health_score}
                        status={status}
                        size="sm"
                        animate={false}
                      />
                    </div>

                    <div className="flex items-center gap-1 mb-4">
                      {status === "healthy" && (
                        <TrendingUp className="h-3.5 w-3.5 text-success" />
                      )}
                      {status === "unhealthy" && (
                        <TrendingDown className="h-3.5 w-3.5 text-danger" />
                      )}
                      {status === "degraded" && (
                        <Activity className="h-3.5 w-3.5 text-muted" />
                      )}
                      <span className="text-[10px] text-muted capitalize">
                        {status}
                      </span>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center justify-between text-xs">
                        <span className="text-muted">Score</span>
                        <span
                          className="font-mono font-medium"
                          style={{
                            color: comp.health_score >= 80
                              ? "var(--foreground)"
                              : comp.health_score >= 50
                              ? "var(--warning)"
                              : "var(--danger)",
                          }}
                        >
                          {comp.health_score}/100
                        </span>
                      </div>
                      <div className="flex items-center justify-between text-xs">
                        <span className="text-muted">Status</span>
                        <span className="font-mono font-medium capitalize">{comp.status}</span>
                      </div>
                      <div className="flex items-center justify-between text-xs">
                        <span className="text-muted">Endpoint</span>
                        <span className="font-mono font-medium text-muted truncate max-w-[140px]">
                          {comp.endpoint_url || "—"}
                        </span>
                      </div>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
