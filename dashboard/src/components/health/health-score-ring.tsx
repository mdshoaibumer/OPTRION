"use client";

import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import { AnimatedCounter } from "@/components/ui/animated-counter";
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

const statusGlow: Record<HealthStatus, string> = {
  healthy: "var(--success-glow)",
  degraded: "var(--warning-glow)",
  unhealthy: "var(--danger-glow)",
  unknown: "transparent",
};

const statusLabel: Record<HealthStatus, string> = {
  healthy: "Operational",
  degraded: "Degraded",
  unhealthy: "Critical",
  unknown: "Unknown",
};

const sizes = {
  sm: { ring: 56, stroke: 4, innerStroke: 3, text: "text-sm", label: "text-[10px]", gap: 6 },
  md: { ring: 96, stroke: 6, innerStroke: 4, text: "text-2xl", label: "text-xs", gap: 8 },
  lg: { ring: 140, stroke: 8, innerStroke: 5, text: "text-4xl", label: "text-sm", gap: 10 },
};

export function HealthScoreRing({
  score,
  status,
  size = "md",
  label,
  animate = true,
}: HealthScoreRingProps) {
  const { ring, stroke, innerStroke, text, label: labelSize, gap } = sizes[size];
  const radius = (ring - stroke) / 2;
  const circumference = 2 * Math.PI * radius;
  const progress = (score / 100) * circumference;
  const color = statusColors[status];
  const glow = statusGlow[status];

  // Inner ring (secondary metric - simulating sub-health)
  const innerRadius = radius - stroke - gap;
  const innerCircumference = 2 * Math.PI * innerRadius;
  const innerProgress = (Math.min(score + 15, 100) / 100) * innerCircumference;

  return (
    <div className="relative inline-flex flex-col items-center group">
      {/* Radial glow behind the ring */}
      <motion.div
        className="absolute rounded-full blur-2xl"
        style={{
          width: ring * 0.7,
          height: ring * 0.7,
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          background: glow,
        }}
        initial={{ opacity: 0, scale: 0.5 }}
        animate={{ opacity: 0.6, scale: 1 }}
        transition={{ delay: 0.8, duration: 1 }}
      />

      <svg width={ring} height={ring} className="-rotate-90 relative z-10">
        {/* Outer background ring */}
        <circle
          cx={ring / 2}
          cy={ring / 2}
          r={radius}
          fill="none"
          stroke="var(--card-border)"
          strokeWidth={stroke}
          opacity={0.5}
        />
        {/* Inner background ring */}
        <circle
          cx={ring / 2}
          cy={ring / 2}
          r={innerRadius}
          fill="none"
          stroke="var(--card-border)"
          strokeWidth={innerStroke}
          opacity={0.3}
        />
        {/* Outer progress ring */}
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
          transition={animate ? { duration: 1.4, type: "spring", stiffness: 60, damping: 15 } : { duration: 0 }}
          style={{ filter: `drop-shadow(0 0 6px ${color})` }}
        />
        {/* Inner progress ring (secondary) */}
        <motion.circle
          cx={ring / 2}
          cy={ring / 2}
          r={innerRadius}
          fill="none"
          stroke={color}
          strokeWidth={innerStroke}
          strokeLinecap="round"
          strokeDasharray={innerCircumference}
          initial={{ strokeDashoffset: innerCircumference }}
          animate={{ strokeDashoffset: innerCircumference - innerProgress }}
          transition={animate ? { duration: 1.6, type: "spring", stiffness: 50, damping: 15, delay: 0.2 } : { duration: 0 }}
          opacity={0.4}
        />
      </svg>

      {/* Score text */}
      <div className="absolute inset-0 flex flex-col items-center justify-center z-10">
        <AnimatedCounter
          value={score}
          className={cn("font-bold font-mono", text)}
          duration={1.5}
        />
        <motion.span
          className={cn("text-muted font-medium mt-0.5", labelSize)}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1 }}
        >
          {statusLabel[status]}
        </motion.span>
      </div>

      {label && (
        <motion.span
          className={cn("mt-3 text-muted font-medium", labelSize)}
          initial={{ opacity: 0, y: 5 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.2 }}
        >
          {label}
        </motion.span>
      )}
    </div>
  );
}
