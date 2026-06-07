"use client";

import { useMemo } from "react";
import { motion } from "framer-motion";

interface SparklineProps {
  data?: number[];
  width?: number;
  height?: number;
  color?: string;
  filled?: boolean;
}

export function Sparkline({
  data,
  width = 80,
  height = 32,
  color = "var(--accent)",
  filled = true,
}: SparklineProps) {
  const points = useMemo(() => {
    const values = data && data.length > 2 ? data : generateFakeData();
    const max = Math.max(...values);
    const min = Math.min(...values);
    const range = max - min || 1;
    const padding = 2;

    return values.map((v, i) => ({
      x: padding + (i / (values.length - 1)) * (width - padding * 2),
      y: padding + (1 - (v - min) / range) * (height - padding * 2),
    }));
  }, [data, width, height]);

  const linePath = points.map((p, i) => `${i === 0 ? "M" : "L"} ${p.x} ${p.y}`).join(" ");
  const fillPath = `${linePath} L ${points[points.length - 1].x} ${height} L ${points[0].x} ${height} Z`;

  return (
    <svg width={width} height={height} className="overflow-visible">
      {filled && (
        <motion.path
          d={fillPath}
          fill={color}
          opacity={0.1}
          initial={{ opacity: 0 }}
          animate={{ opacity: 0.1 }}
          transition={{ delay: 0.5, duration: 0.8 }}
        />
      )}
      <motion.path
        d={linePath}
        fill="none"
        stroke={color}
        strokeWidth={1.5}
        strokeLinecap="round"
        strokeLinejoin="round"
        initial={{ pathLength: 0, opacity: 0 }}
        animate={{ pathLength: 1, opacity: 1 }}
        transition={{ duration: 1.2, ease: [0.25, 0.46, 0.45, 0.94], delay: 0.3 }}
      />
      {/* End dot */}
      <motion.circle
        cx={points[points.length - 1].x}
        cy={points[points.length - 1].y}
        r={2.5}
        fill={color}
        initial={{ scale: 0, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ delay: 1.5, duration: 0.3, type: "spring" }}
      />
    </svg>
  );
}

function generateFakeData(): number[] {
  const len = 12;
  const data: number[] = [];
  let val = 50 + Math.random() * 30;
  for (let i = 0; i < len; i++) {
    val += (Math.random() - 0.4) * 10;
    data.push(Math.max(10, Math.min(100, val)));
  }
  return data;
}
