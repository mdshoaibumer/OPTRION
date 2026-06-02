"use client";

import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import type { DashboardStats } from "@/lib/types";
import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  Heart,
  Server,
  XCircle,
} from "lucide-react";

interface StatsCardsProps {
  stats: DashboardStats;
}

export function StatsCards({ stats }: StatsCardsProps) {
  const cards = [
    {
      label: "Total Components",
      value: stats.total_components,
      icon: Server,
      color: "var(--accent)",
      bg: "var(--accent-glow)",
    },
    {
      label: "Healthy",
      value: stats.healthy_components,
      icon: CheckCircle2,
      color: "var(--success)",
      bg: "var(--success-glow)",
    },
    {
      label: "Degraded",
      value: stats.degraded_components,
      icon: AlertTriangle,
      color: "var(--warning)",
      bg: "var(--warning-glow)",
    },
    {
      label: "Unhealthy",
      value: stats.unhealthy_components,
      icon: XCircle,
      color: "var(--danger)",
      bg: "var(--danger-glow)",
    },
    {
      label: "Open Incidents",
      value: stats.open_incidents,
      icon: Activity,
      color: stats.open_incidents > 0 ? "var(--danger)" : "var(--success)",
      bg: stats.open_incidents > 0 ? "var(--danger-glow)" : "var(--success-glow)",
    },
    {
      label: "Avg Health Score",
      value: stats.average_health_score,
      icon: Heart,
      color:
        stats.average_health_score >= 80
          ? "var(--success)"
          : stats.average_health_score >= 50
          ? "var(--warning)"
          : "var(--danger)",
      bg:
        stats.average_health_score >= 80
          ? "var(--success-glow)"
          : stats.average_health_score >= 50
          ? "var(--warning-glow)"
          : "var(--danger-glow)",
    },
  ];

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
      {cards.map((card, idx) => (
        <motion.div
          key={card.label}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: idx * 0.05, duration: 0.4 }}
          className="rounded-xl border border-[var(--card-border)] bg-[var(--card)] p-4"
        >
          <div className="flex items-center gap-2 mb-3">
            <div
              className="rounded-lg p-1.5"
              style={{ backgroundColor: card.bg }}
            >
              <card.icon className="h-4 w-4" style={{ color: card.color }} />
            </div>
          </div>
          <motion.p
            className="text-2xl font-bold font-mono"
            style={{ color: card.color }}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: idx * 0.05 + 0.3 }}
          >
            {card.value}
          </motion.p>
          <p className="text-[11px] text-[var(--muted)] mt-1">{card.label}</p>
        </motion.div>
      ))}
    </div>
  );
}
