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

const severityColors: Record<Severity, string> = {
  critical: "border-l-red-500 bg-red-500/5",
  major: "border-l-orange-500 bg-orange-500/5",
  minor: "border-l-yellow-500 bg-yellow-500/5",
  warning: "border-l-amber-500 bg-amber-500/5",
  info: "border-l-blue-500 bg-blue-500/5",
};

export function AlertFeed({ alerts }: AlertFeedProps) {
  return (
    <div className="space-y-2 max-h-150 overflow-y-auto">
      <AnimatePresence mode="popLayout">
        {alerts.map((alert, idx) => {
          const Icon = severityIcons[alert.severity] || Info;
          const colorClass = severityColors[alert.severity] || severityColors.info;

          return (
            <motion.div
              key={alert.id}
              initial={{ opacity: 0, y: -10, height: 0 }}
              animate={{ opacity: 1, y: 0, height: "auto" }}
              exit={{ opacity: 0, y: 10, height: 0 }}
              transition={{
                type: "spring",
                stiffness: 400,
                damping: 25,
                delay: idx * 0.05,
              }}
              className={cn(
                "rounded-lg border border-card-border border-l-4 p-3",
                colorClass
              )}
            >
              <div className="flex items-start gap-3">
                <Icon className="h-4 w-4 mt-0.5 shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{alert.title}</p>
                  <p className="text-xs text-muted mt-0.5 line-clamp-2">
                    {alert.message}
                  </p>
                  <div className="flex items-center gap-2 mt-2 text-[10px] text-muted">
                    <span className="uppercase font-mono font-semibold">
                      {alert.severity}
                    </span>
                    <span>·</span>
                    <RelativeTime date={alert.created_at} />
                    {alert.delivered_at && (
                      <>
                        <span>·</span>
                        <span className="text-success">Delivered</span>
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
        <div className="flex flex-col items-center justify-center py-12 text-muted">
          <Bell className="h-8 w-8 mb-2" />
          <p className="text-sm">No alerts</p>
          <p className="text-xs mt-1">Alert rules will trigger notifications here</p>
        </div>
      )}
    </div>
  );
}
