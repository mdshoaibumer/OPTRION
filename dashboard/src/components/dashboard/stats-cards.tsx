"use client";

import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import { AnimatedCounter } from "@/components/ui/animated-counter";
import { Sparkline } from "@/components/ui/sparkline";
import type { DashboardStats } from "@/lib/types";
import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  Heart,
  Server,
  TrendingUp,
  TrendingDown,
  XCircle,
} from "lucide-react";

interface StatsCardsProps {
  stats: DashboardStats;
}

const cardVariants = {
  hidden: { opacity: 0, y: 20, scale: 0.95 },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    scale: 1,
    transition: {
      delay: i * 0.06,
      duration: 0.5,
      type: "spring",
      stiffness: 260,
      damping: 20,
    },
  }),
};

export function StatsCards({ stats }: StatsCardsProps) {
  const healthPercent = stats.total_components > 0
    ? Math.round((stats.healthy_components / stats.total_components) * 100)
    : 0;

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
      {/* Hero Card: Health Score — spans 2 columns */}
      <motion.div
        custom={0}
        variants={cardVariants}
        initial="hidden"
        animate="visible"
        whileHover={{ scale: 1.02, y: -2 }}
        transition={{ type: "spring", stiffness: 400, damping: 25 }}
        className="col-span-2 gradient-border group cursor-default"
      >
        <div className="relative p-6 rounded-2xl overflow-hidden">
          {/* Ambient glow */}
          <div
            className="absolute -top-12 -right-12 w-32 h-32 rounded-full blur-3xl opacity-20 transition-opacity duration-500 group-hover:opacity-40"
            style={{
              background: stats.average_health_score >= 80
                ? "var(--success)"
                : stats.average_health_score >= 50
                ? "var(--warning)"
                : "var(--danger)",
            }}
          />

          <div className="relative flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <div
                  className="rounded-lg p-2"
                  style={{
                    backgroundColor: stats.average_health_score >= 80
                      ? "var(--success-glow)"
                      : stats.average_health_score >= 50
                      ? "var(--warning-glow)"
                      : "var(--danger-glow)",
                  }}
                >
                  <Heart
                    className="h-4 w-4"
                    style={{
                      color: stats.average_health_score >= 80
                        ? "var(--success)"
                        : stats.average_health_score >= 50
                        ? "var(--warning)"
                        : "var(--danger)",
                    }}
                  />
                </div>
                <span className="text-xs text-muted font-medium">Avg Health Score</span>
              </div>
              <div className="flex items-baseline gap-2 mt-3">
                <AnimatedCounter
                  value={stats.average_health_score}
                  className="text-4xl font-bold font-mono"
                />
                <span className="text-lg text-muted font-mono">/100</span>
              </div>
              <div className="flex items-center gap-1.5 mt-2">
                <TrendingUp className="h-3 w-3 text-success" />
                <span className="text-xs text-success font-medium">+2.4%</span>
                <span className="text-xs text-muted">vs last hour</span>
              </div>
            </div>
            <div className="hidden sm:block">
              <Sparkline
                width={100}
                height={48}
                color={
                  stats.average_health_score >= 80
                    ? "var(--success)"
                    : stats.average_health_score >= 50
                    ? "var(--warning)"
                    : "var(--danger)"
                }
              />
            </div>
          </div>
        </div>
      </motion.div>

      {/* Hero Card: Open Incidents — spans 2 columns */}
      <motion.div
        custom={1}
        variants={cardVariants}
        initial="hidden"
        animate="visible"
        whileHover={{ scale: 1.02, y: -2 }}
        transition={{ type: "spring", stiffness: 400, damping: 25 }}
        className="col-span-2 gradient-border group cursor-default"
      >
        <div className="relative p-6 rounded-2xl overflow-hidden">
          {/* Ambient glow for incidents */}
          <div
            className="absolute -top-12 -right-12 w-32 h-32 rounded-full blur-3xl opacity-20 transition-opacity duration-500 group-hover:opacity-40"
            style={{
              background: stats.open_incidents > 0 ? "var(--danger)" : "var(--success)",
            }}
          />

          <div className="relative flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <div
                  className="rounded-lg p-2"
                  style={{
                    backgroundColor: stats.open_incidents > 0
                      ? "var(--danger-glow)"
                      : "var(--success-glow)",
                  }}
                >
                  <Activity
                    className="h-4 w-4"
                    style={{
                      color: stats.open_incidents > 0 ? "var(--danger)" : "var(--success)",
                    }}
                  />
                </div>
                <span className="text-xs text-muted font-medium">Open Incidents</span>
              </div>
              <div className="flex items-baseline gap-2 mt-3">
                <AnimatedCounter
                  value={stats.open_incidents}
                  className={cn(
                    "text-4xl font-bold font-mono",
                    stats.open_incidents > 0 ? "text-danger" : "text-success"
                  )}
                />
                {stats.open_incidents > 0 && (
                  <span className="text-xs px-2 py-0.5 rounded-full bg-red-500/10 text-red-400 font-mono animate-pulse">
                    ACTIVE
                  </span>
                )}
              </div>
              <div className="flex items-center gap-1.5 mt-2">
                {stats.open_incidents > 0 ? (
                  <>
                    <TrendingDown className="h-3 w-3 text-danger" />
                    <span className="text-xs text-danger font-medium">Needs attention</span>
                  </>
                ) : (
                  <>
                    <CheckCircle2 className="h-3 w-3 text-success" />
                    <span className="text-xs text-success font-medium">All clear</span>
                  </>
                )}
              </div>
            </div>
            <div className="hidden sm:block">
              <Sparkline
                width={100}
                height={48}
                color={stats.open_incidents > 0 ? "var(--danger)" : "var(--success)"}
              />
            </div>
          </div>
        </div>
      </motion.div>

      {/* Compact Card: Total Components */}
      <CompactStatCard
        index={2}
        label="Total"
        value={stats.total_components}
        icon={Server}
        color="var(--accent)"
        bg="var(--accent-glow)"
      />

      {/* Compact Card: Healthy */}
      <CompactStatCard
        index={3}
        label="Healthy"
        value={stats.healthy_components}
        icon={CheckCircle2}
        color="var(--success)"
        bg="var(--success-glow)"
        suffix={stats.total_components > 0 ? `${healthPercent}%` : undefined}
      />

      {/* Compact Card: Degraded */}
      <CompactStatCard
        index={4}
        label="Degraded"
        value={stats.degraded_components}
        icon={AlertTriangle}
        color="var(--warning)"
        bg="var(--warning-glow)"
      />

      {/* Compact Card: Unhealthy */}
      <CompactStatCard
        index={5}
        label="Unhealthy"
        value={stats.unhealthy_components}
        icon={XCircle}
        color="var(--danger)"
        bg="var(--danger-glow)"
      />
    </div>
  );
}

interface CompactStatCardProps {
  index: number;
  label: string;
  value: number;
  icon: React.ComponentType<{ className?: string; style?: React.CSSProperties }>;
  color: string;
  bg: string;
  suffix?: string;
}

function CompactStatCard({ index, label, value, icon: Icon, color, bg, suffix }: CompactStatCardProps) {
  return (
    <motion.div
      custom={index}
      variants={cardVariants}
      initial="hidden"
      animate="visible"
      whileHover={{ scale: 1.03, y: -2 }}
      transition={{ type: "spring", stiffness: 400, damping: 25 }}
      className="col-span-1 gradient-border group cursor-default"
    >
      <div className="p-4 rounded-2xl h-full flex flex-col justify-between">
        <div className="flex items-center justify-between">
          <div className="rounded-lg p-1.5" style={{ backgroundColor: bg }}>
            <Icon className="h-4 w-4" style={{ color }} />
          </div>
          {suffix && (
            <span className="text-[10px] font-mono text-muted">{suffix}</span>
          )}
        </div>
        <div className="mt-3">
          <AnimatedCounter
            value={value}
            className="text-2xl font-bold font-mono block"
          />
          <p className="text-[11px] text-muted mt-1 font-medium">{label}</p>
        </div>
      </div>
    </motion.div>
  );
}
