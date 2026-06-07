"use client";

import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";
import { RelativeTime } from "@/components/ui/relative-time";
import type { Incident, Severity } from "@/lib/types";
import {
  AlertTriangle,
  CheckCircle2,
  Clock,
  Eye,
  Search,
  XCircle,
  Zap,
  type LucideIcon,
} from "lucide-react";

interface IncidentListProps {
  incidents: Incident[];
  onSelect?: (incident: Incident) => void;
  selectedId?: string;
}

const severityConfig: Record<Severity, { color: string; glow: string; icon: LucideIcon }> = {
  critical: { color: "var(--danger)", glow: "var(--danger-glow)", icon: XCircle },
  major: { color: "var(--danger)", glow: "var(--danger-glow)", icon: AlertTriangle },
  minor: { color: "var(--warning)", glow: "var(--warning-glow)", icon: AlertTriangle },
  warning: { color: "var(--warning)", glow: "var(--warning-glow)", icon: Eye },
  info: { color: "var(--info)", glow: "var(--info-glow)", icon: Eye },
};

const statusBadge: Record<string, { label: string; className: string; dotColor: string }> = {
  open: { label: "Open", className: "bg-red-500/10 text-red-400 border-red-500/20", dotColor: "bg-red-400" },
  acknowledged: { label: "Ack", className: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20", dotColor: "bg-yellow-400" },
  investigating: { label: "Investigating", className: "bg-blue-500/10 text-blue-400 border-blue-500/20", dotColor: "bg-blue-400" },
  resolved: { label: "Resolved", className: "bg-green-500/10 text-green-400 border-green-500/20", dotColor: "bg-green-400" },
  closed: { label: "Closed", className: "bg-gray-500/10 text-gray-400 border-gray-500/20", dotColor: "bg-gray-400" },
};

export function IncidentList({ incidents, onSelect, selectedId }: IncidentListProps) {
  return (
    <div className="relative space-y-1">
      {/* Timeline line */}
      {incidents.length > 0 && (
        <div className="absolute left-4.75 top-4 bottom-4 w-px bg-linear-to-b from-card-border via-card-border to-transparent" />
      )}

      <AnimatePresence mode="popLayout">
        {incidents.map((incident, idx) => {
          const severity = severityConfig[incident.severity] || severityConfig.info;
          const status = statusBadge[incident.status] || statusBadge.open;
          const SeverityIcon = severity.icon;
          const isSelected = selectedId === incident.id;
          const isActive = incident.status !== "resolved" && incident.status !== "closed";

          return (
            <motion.div
              key={incident.id}
              layout
              initial={{ opacity: 0, x: -20, scale: 0.95 }}
              animate={{ opacity: 1, x: 0, scale: 1 }}
              exit={{ opacity: 0, x: 20, scale: 0.95 }}
              transition={{
                type: "spring",
                stiffness: 500,
                damping: 30,
                delay: idx * 0.04,
              }}
              onClick={() => onSelect?.(incident)}
              className={cn(
                "group relative cursor-pointer rounded-xl border p-4 pl-10 transition-all duration-200",
                "hover:border-card-border-hover",
                isSelected
                  ? "border-accent/40 bg-(--accent-glow)"
                  : "border-transparent bg-transparent hover:bg-(--glass)"
              )}
            >
              {/* Timeline dot */}
              <div
                className={cn(
                  "absolute left-3 top-5 h-3 w-3 rounded-full border-2 z-10",
                  isActive ? "border-current" : "border-card-border"
                )}
                style={{
                  borderColor: isActive ? severity.color : undefined,
                  backgroundColor: isActive ? severity.glow : "var(--card-solid)",
                  boxShadow: isActive ? `0 0 8px ${severity.glow}` : undefined,
                }}
              >
                {isActive && (
                  <motion.div
                    className="absolute inset-0 rounded-full"
                    style={{ backgroundColor: severity.color }}
                    animate={{ scale: [1, 1.8, 1], opacity: [0.6, 0, 0.6] }}
                    transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
                  />
                )}
              </div>

              <div className="flex items-start gap-3">
                {/* Severity icon */}
                <div
                  className="rounded-lg p-1.5 shrink-0"
                  style={{ backgroundColor: severity.glow }}
                >
                  <SeverityIcon className="h-3.5 w-3.5" style={{ color: severity.color }} />
                </div>

                {/* Content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h4 className="text-sm font-medium truncate group-hover:text-foreground transition-colors">
                      {incident.title}
                    </h4>
                  </div>
                  <div className="flex items-center gap-2 mt-1.5">
                    <span
                      className={cn(
                        "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium border",
                        status.className
                      )}
                    >
                      <span className={cn("h-1.5 w-1.5 rounded-full", status.dotColor)} />
                      {status.label}
                    </span>
                    <span className="text-[10px] uppercase font-mono text-muted font-semibold">
                      {incident.severity}
                    </span>
                    <span className="text-muted">·</span>
                    <span className="flex items-center gap-1 text-xs text-muted">
                      <Clock className="h-3 w-3" />
                      <RelativeTime date={incident.occurred_at} />
                    </span>
                  </div>
                </div>

                {/* Hover action */}
                <div className="opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                  <button className="p-1.5 rounded-lg hover:bg-card-border transition-colors">
                    <Search className="h-3.5 w-3.5 text-muted" />
                  </button>
                </div>
              </div>
            </motion.div>
          );
        })}
      </AnimatePresence>

      {incidents.length === 0 && (
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="flex flex-col items-center justify-center py-12 text-muted"
        >
          <div className="rounded-full p-3 bg-success/10 mb-3">
            <CheckCircle2 className="h-6 w-6 text-success" />
          </div>
          <p className="text-sm font-medium">No active incidents</p>
          <p className="text-xs mt-1 text-muted">All systems operating normally</p>
        </motion.div>
      )}
    </div>
  );
}
