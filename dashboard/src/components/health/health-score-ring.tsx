"use client";

import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import type { HealthStatus } from "@/lib/types";

interface HealthScoreRingProps {
  score: number;
  status: HealthStatus;
  size?: "sm" | "md" | "lg";
  label?: string;
  animate?: boolean;
}

const statusColors: Record<HealthStatus, string> = {
  healthy: "var(--success)",
  degraded: "var(--warning)",
  unhealthy: "var(--danger)",
  unknown: "var(--muted)",
};

const statusPulse: Record<HealthStatus, string> = {
  healthy: "pulse-healthy",
  degraded: "pulse-warning",
  unhealthy: "pulse-critical",
  unknown: "",
};

const sizes = {
  sm: { ring: 48, stroke: 4, text: "text-sm", label: "text-[10px]" },
  md: { ring: 80, stroke: 6, text: "text-xl", label: "text-xs" },
  lg: { ring: 120, stroke: 8, text: "text-3xl", label: "text-sm" },
};

export function HealthScoreRing({
  score,
  status,
  size = "md",
  label,
  animate = true,
}: HealthScoreRingProps) {
  const { ring, stroke, text, label: labelSize } = sizes[size];
  const radius = (ring - stroke) / 2;
  const circumference = 2 * Math.PI * radius;
  const progress = (score / 100) * circumference;
  const color = statusColors[status];

  return (
    <div className={cn("relative inline-flex flex-col items-center", statusPulse[status])}>
      <svg width={ring} height={ring} className="-rotate-90">
        {/* Background ring */}
        <circle
          cx={ring / 2}
          cy={ring / 2}
          r={radius}
          fill="none"
          stroke="var(--card-border)"
          strokeWidth={stroke}
        />
        {/* Progress ring */}
        <motion.circle
          cx={ring / 2}
          cy={ring / 2}
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth={stroke}
          strokeLinecap="round"
          strokeDasharray={circumference}
          initial={{ strokeDashoffset: circumference }}
          animate={{ strokeDashoffset: circumference - progress }}
          transition={animate ? { duration: 1, ease: [0.34, 1.56, 0.64, 1] } : { duration: 0 }}
        />
      </svg>
      {/* Score text */}
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <motion.span
          className={cn("font-bold font-mono", text)}
          initial={animate ? { opacity: 0, scale: 0.5 } : {}}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.5, duration: 0.3 }}
          style={{ color }}
        >
          {score}
        </motion.span>
      </div>
      {label && (
        <span className={cn("mt-1 text-[var(--muted)]", labelSize)}>{label}</span>
      )}
    </div>
  );
}
