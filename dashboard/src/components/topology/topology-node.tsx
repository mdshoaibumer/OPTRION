"use client";

import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import type { ComponentKind, HealthStatus } from "@/lib/types";
import { Database, Globe, HardDrive, Layers, Server, Wifi, type LucideIcon } from "lucide-react";

interface TopologyNodeProps {
  id: string;
  name: string;
  kind: ComponentKind;
  status: HealthStatus;
  score: number;
  selected?: boolean;
  onClick?: (id: string) => void;
}

const kindIcons: Record<ComponentKind, LucideIcon> = {
  database: Database,
  cache: Layers,
  api: Globe,
  queue: Wifi,
  service: Server,
  server: HardDrive,
};

const statusStyles: Record<HealthStatus, string> = {
  healthy: "border-[var(--success)] bg-[var(--success-glow)]",
  degraded: "border-[var(--warning)] bg-[var(--warning-glow)]",
  unhealthy: "border-[var(--danger)] bg-[var(--danger-glow)]",
  unknown: "border-[var(--muted)] bg-[var(--card)]",
};

const statusPulse: Record<HealthStatus, string> = {
  healthy: "",
  degraded: "pulse-warning",
  unhealthy: "pulse-critical",
  unknown: "",
};

export function TopologyNode({
  id,
  name,
  kind,
  status,
  score,
  selected = false,
  onClick,
}: TopologyNodeProps) {
  const Icon = kindIcons[kind] || Server;

  return (
    <motion.div
      layoutId={`topology-node-${id}`}
      initial={{ scale: 0, opacity: 0 }}
      animate={{ scale: 1, opacity: 1 }}
      transition={{ type: "spring", stiffness: 300, damping: 25 }}
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.95 }}
      onClick={() => onClick?.(id)}
      className={cn(
        "relative cursor-pointer rounded-xl border-2 p-4 transition-all duration-200",
        "flex flex-col items-center gap-2 w-32",
        statusStyles[status],
        statusPulse[status],
        selected && "ring-2 ring-accent ring-offset-2 ring-offset-background"
      )}
    >
      <Icon className="h-8 w-8" color={`var(--${status === "healthy" ? "success" : status === "degraded" ? "warning" : status === "unhealthy" ? "danger" : "muted"})`} />
      <span className="text-xs font-medium text-center truncate w-full">{name}</span>
      <div className="flex items-center gap-1">
        <div
          className="h-2 w-2 rounded-full"
          style={{
            backgroundColor: `var(--${status === "healthy" ? "success" : status === "degraded" ? "warning" : status === "unhealthy" ? "danger" : "muted"})`,
          }}
        />
        <span className="text-[10px] font-mono text-muted">{score}</span>
      </div>
    </motion.div>
  );
}
