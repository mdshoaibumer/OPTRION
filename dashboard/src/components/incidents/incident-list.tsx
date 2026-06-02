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
  type LucideIcon,
} from "lucide-react";

interface IncidentListProps {
  incidents: Incident[];
  onSelect?: (incident: Incident) => void;
  selectedId?: string;
}

const severityConfig: Record<Severity, { color: string; icon: LucideIcon }> = {
  critical: { color: "var(--danger)", icon: XCircle },
  major: { color: "var(--danger)", icon: AlertTriangle },
  minor: { color: "var(--warning)", icon: AlertTriangle },
  warning: { color: "var(--warning)", icon: Eye },
  info: { color: "var(--info)", icon: Eye },
};

const statusBadge: Record<string, { label: string; className: string }> = {
  open: { label: "Open", className: "bg-red-500/10 text-red-400 border-red-500/20" },
  acknowledged: { label: "Ack", className: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20" },
  investigating: { label: "Inv", className: "bg-blue-500/10 text-blue-400 border-blue-500/20" },
  resolved: { label: "Resolved", className: "bg-green-500/10 text-green-400 border-green-500/20" },
  closed: { label: "Closed", className: "bg-gray-500/10 text-gray-400 border-gray-500/20" },
};

export function IncidentList({ incidents, onSelect, selectedId }: IncidentListProps) {
  return (
    <div className="space-y-2">
      <AnimatePresence mode="popLayout">
        {incidents.map((incident, idx) => {
          const severity = severityConfig[incident.severity] || severityConfig.info;
          const status = statusBadge[incident.status] || statusBadge.open;
          const SeverityIcon = severity.icon;
          const isSelected = selectedId === incident.id;

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
                delay: idx * 0.03,
              }}
              onClick={() => onSelect?.(incident)}
              className={cn(
                "group cursor-pointer rounded-xl border p-4 transition-all duration-200",
                "hover:border-accent/40 hover:bg-card",
                isSelected
                  ? "border-accent bg-(--accent-glow)"
                  : "border-card-border bg-card/50"
              )}
            >
              <div className="flex items-start gap-3">
                {/* Severity indicator */}
                <div
                  className="mt-0.5 rounded-full p-1"
                  style={{ backgroundColor: `${severity.color}20` }}
                >
                  <SeverityIcon
                    className="h-4 w-4"
                    color={severity.color}
                  />
                </div>

                {/* Content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h4 className="text-sm font-medium truncate">
                      {incident.title}
                    </h4>
                    <span
                      className={cn(
                        "inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium border",
                        status.className
                      )}
                    >
                      {status.label}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 mt-1.5 text-xs text-muted">
                    <span className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      <RelativeTime date={incident.occurred_at} />
                    </span>
                    <span className="uppercase font-mono text-[10px]">
                      {incident.severity}
                    </span>
                  </div>
                </div>

                {/* Actions */}
                <div className="opacity-0 group-hover:opacity-100 transition-opacity">
                  <button className="p-1.5 rounded-lg hover:bg-card-border">
                    <Search className="h-3.5 w-3.5 text-muted" />
                  </button>
                </div>
              </div>
            </motion.div>
          );
        })}
      </AnimatePresence>

      {incidents.length === 0 && (
        <div className="flex flex-col items-center justify-center py-12 text-muted">
          <CheckCircle2 className="h-8 w-8 mb-2 text-success" />
          <p className="text-sm">No active incidents</p>
          <p className="text-xs mt-1">All systems operating normally</p>
        </div>
      )}
    </div>
  );
}
