"use client";

import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";
import { RelativeTime } from "@/components/ui/relative-time";
import type { Alert, Severity } from "@/lib/types";
import { Bell, Shield, AlertTriangle, XCircle, Info } from "lucide-react";

interface AlertFeedProps {
  alerts: Alert[];
}

const severityIcons: Record<Severity, React.ComponentType<{ className?: string }>> = {
  critical: XCircle,
  major: AlertTriangle,
  minor: AlertTriangle,
  warning: Shield,
  info: Info,
};

const severityStyles: Record<Severity, { border: string; bg: string; icon: string; glow: string }> = {
  critical: { border: "border-l-red-500", bg: "bg-red-500/5", icon: "text-red-400", glow: "shadow-red-500/10" },
  major: { border: "border-l-orange-500", bg: "bg-orange-500/5", icon: "text-orange-400", glow: "shadow-orange-500/10" },
  minor: { border: "border-l-yellow-500", bg: "bg-yellow-500/5", icon: "text-yellow-400", glow: "shadow-yellow-500/10" },
  warning: { border: "border-l-amber-500", bg: "bg-amber-500/5", icon: "text-amber-400", glow: "shadow-amber-500/10" },
  info: { border: "border-l-blue-500", bg: "bg-blue-500/5", icon: "text-blue-400", glow: "shadow-blue-500/10" },
};

export function AlertFeed({ alerts }: AlertFeedProps) {
  return (
    <div className="space-y-2 max-h-150 overflow-y-auto pr-1">
      <AnimatePresence mode="popLayout">
        {alerts.map((alert, idx) => {
          const Icon = severityIcons[alert.severity] || Info;
          const styles = severityStyles[alert.severity] || severityStyles.info;

          return (
            <motion.div
              key={alert.id}
              initial={{ opacity: 0, x: 30, scale: 0.95 }}
              animate={{ opacity: 1, x: 0, scale: 1 }}
              exit={{ opacity: 0, x: -20, scale: 0.9, height: 0 }}
              transition={{
                type: "spring",
                stiffness: 400,
                damping: 25,
                delay: idx * 0.05,
              }}
              whileHover={{ x: 2 }}
              className={cn(
                "rounded-xl border border-(--glass-border) border-l-4 p-3.5 transition-all duration-200",
                "hover:border-(--glass-border-hover) hover:shadow-lg",
                styles.border,
                styles.bg,
                styles.glow
              )}
            >
              <div className="flex items-start gap-3">
                <div className={cn("mt-0.5 shrink-0 rounded-lg p-1.5", styles.bg)}>
                  <Icon className={cn("h-3.5 w-3.5", styles.icon)} />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate leading-tight">{alert.title}</p>
                  <p className="text-xs text-muted mt-1 line-clamp-2 leading-relaxed">
                    {alert.message}
                  </p>
                  <div className="flex items-center gap-2 mt-2.5">
                    <span className={cn(
                      "uppercase font-mono text-[10px] font-semibold px-1.5 py-0.5 rounded",
                      styles.bg, styles.icon
                    )}>
                      {alert.severity}
                    </span>
                    <span className="text-[10px] text-muted">·</span>
                    <span className="text-[10px] text-muted">
                      <RelativeTime date={alert.created_at} />
                    </span>
                    {alert.delivered_at && (
                      <>
                        <span className="text-[10px] text-muted">·</span>
                        <span className="text-[10px] text-success font-medium flex items-center gap-1">
                          <span className="h-1.5 w-1.5 rounded-full bg-success" />
                          Delivered
                        </span>
                      </>
                    )}
                  </div>
                </div>
              </div>
            </motion.div>
          );
        })}
      </AnimatePresence>

      {alerts.length === 0 && (
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="flex flex-col items-center justify-center py-12 text-muted"
        >
          <div className="rounded-full p-3 bg-(--glass) mb-3">
            <Bell className="h-6 w-6 text-muted" />
          </div>
          <p className="text-sm font-medium">No alerts</p>
          <p className="text-xs mt-1">Alert rules will trigger notifications here</p>
        </motion.div>
      )}
    </div>
  );
}
