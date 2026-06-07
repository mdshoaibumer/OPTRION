"use client";

import { motion } from "framer-motion";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { StatsCards } from "@/components/dashboard/stats-cards";
import { HealthScoreRing } from "@/components/health/health-score-ring";
import { IncidentList } from "@/components/incidents/incident-list";
import { AlertFeed } from "@/components/alerts/alert-feed";
import { StatsSkeleton, ListSkeleton, QueryError } from "@/components/ui/loading";
import { useDashboardStats, useIncidents, useAlerts } from "@/lib/hooks";
import type { DashboardStats, Incident, Alert } from "@/lib/types";
import { Sparkline } from "@/components/ui/sparkline";

const pageTransition = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.4, ease: [0.25, 0.46, 0.45, 0.94] },
};

const staggerContainer = {
  animate: { transition: { staggerChildren: 0.08 } },
};

const staggerItem = {
  initial: { opacity: 0, y: 16 },
  animate: { opacity: 1, y: 0, transition: { duration: 0.5, ease: [0.25, 0.46, 0.45, 0.94] } },
};

export default function DashboardPage() {
  const { data: stats, isLoading: statsLoading, error: statsError, refetch: refetchStats } = useDashboardStats();
  const { data: incidents, isLoading: incidentsLoading, error: incidentsError } = useIncidents({ limit: 5 });
  const { data: alerts, isLoading: alertsLoading, error: alertsError } = useAlerts();

  const displayStats: DashboardStats = stats || {
    total_components: 0, healthy_components: 0, degraded_components: 0,
    unhealthy_components: 0, open_incidents: 0, active_alerts: 0, average_health_score: 0,
  };

  return (
    <div className="flex min-h-screen noise-overlay">
      <Sidebar />
      <div className="flex-1 ml-16 lg:ml-64">
        <Header />
        <main className="p-6 lg:p-8 space-y-8 dot-grid-bg min-h-[calc(100vh-4rem)]">
          {/* Page title */}
          <motion.div {...pageTransition}>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold tracking-tight">Engineering Intelligence</h1>
              <motion.div
                className="h-2 w-2 rounded-full bg-success"
                animate={{ opacity: [1, 0.4, 1] }}
                transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
              />
            </div>
            <p className="text-sm text-muted mt-1">
              Real-time health monitoring and incident detection
            </p>
          </motion.div>

          {/* Stats row — Bento Grid */}
          {statsLoading ? (
            <StatsSkeleton />
          ) : statsError ? (
            <QueryError error={statsError as Error} onRetry={() => refetchStats()} />
          ) : (
            <StatsCards stats={displayStats} />
          )}

          {/* Main content grid */}
          <motion.div
            variants={staggerContainer}
            initial="initial"
            animate="animate"
            className="grid grid-cols-1 lg:grid-cols-3 gap-6"
          >
            {/* Health Overview */}
            <motion.div
              variants={staggerItem}
              className="lg:col-span-1 rounded-2xl border border-(--glass-border) bg-card p-6 gradient-border group"
            >
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-sm font-semibold">System Health</h2>
                <div className="flex items-center gap-1.5">
                  <Sparkline width={48} height={16} color="var(--success)" />
                </div>
              </div>
              <div className="flex flex-col items-center py-4">
                <HealthScoreRing
                  score={displayStats.average_health_score}
                  status={
                    displayStats.average_health_score >= 80
                      ? "healthy"
                      : displayStats.average_health_score >= 50
                      ? "degraded"
                      : "unhealthy"
                  }
                  size="lg"
                  label="Overall Score"
                />
              </div>
              {/* Component breakdown bar */}
              <div className="mt-6 pt-4 border-t border-(--glass-border)">
                <div className="flex items-center justify-between text-xs text-muted mb-2">
                  <span>Component Status</span>
                  <span className="font-mono">{displayStats.total_components} total</span>
                </div>
                <div className="flex h-2 rounded-full overflow-hidden bg-background">
                  {displayStats.total_components > 0 && (
                    <>
                      <motion.div
                        className="h-full bg-success"
                        initial={{ width: 0 }}
                        animate={{ width: `${(displayStats.healthy_components / displayStats.total_components) * 100}%` }}
                        transition={{ duration: 1, delay: 0.5, ease: "easeOut" }}
                      />
                      <motion.div
                        className="h-full bg-warning"
                        initial={{ width: 0 }}
                        animate={{ width: `${(displayStats.degraded_components / displayStats.total_components) * 100}%` }}
                        transition={{ duration: 1, delay: 0.7, ease: "easeOut" }}
                      />
                      <motion.div
                        className="h-full bg-danger"
                        initial={{ width: 0 }}
                        animate={{ width: `${(displayStats.unhealthy_components / displayStats.total_components) * 100}%` }}
                        transition={{ duration: 1, delay: 0.9, ease: "easeOut" }}
                      />
                    </>
                  )}
                </div>
                <div className="flex items-center gap-4 mt-2 text-[10px] text-muted">
                  <span className="flex items-center gap-1">
                    <span className="h-2 w-2 rounded-full bg-success" />Healthy
                  </span>
                  <span className="flex items-center gap-1">
                    <span className="h-2 w-2 rounded-full bg-warning" />Degraded
                  </span>
                  <span className="flex items-center gap-1">
                    <span className="h-2 w-2 rounded-full bg-danger" />Unhealthy
                  </span>
                </div>
              </div>
            </motion.div>

            {/* Incidents */}
            <motion.div
              variants={staggerItem}
              className="lg:col-span-1 rounded-2xl border border-(--glass-border) bg-card p-6 gradient-border"
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-semibold">Active Incidents</h2>
                {incidents && (
                  <motion.span
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    className="text-xs px-2.5 py-1 rounded-full bg-red-500/10 text-red-400 font-mono border border-red-500/20"
                  >
                    {incidents.filter((i) => i.status !== "resolved" && i.status !== "closed").length}
                  </motion.span>
                )}
              </div>
              {incidentsLoading ? (
                <ListSkeleton count={3} />
              ) : incidentsError ? (
                <QueryError error={incidentsError as Error} />
              ) : (
                <IncidentList incidents={incidents || []} />
              )}
            </motion.div>

            {/* Alert Feed */}
            <motion.div
              variants={staggerItem}
              className="lg:col-span-1 rounded-2xl border border-(--glass-border) bg-card p-6 gradient-border"
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-semibold">Alert Feed</h2>
                {alerts && (
                  <motion.span
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    className="text-xs px-2.5 py-1 rounded-full bg-amber-500/10 text-amber-400 font-mono border border-amber-500/20"
                  >
                    {alerts.length}
                  </motion.span>
                )}
              </div>
              {alertsLoading ? (
                <ListSkeleton count={3} />
              ) : alertsError ? (
                <QueryError error={alertsError as Error} />
              ) : (
                <AlertFeed alerts={alerts || []} />
              )}
            </motion.div>
          </motion.div>
        </main>
      </div>
    </div>
  );
}
