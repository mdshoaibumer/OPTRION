"use client";

import { motion, AnimatePresence } from "framer-motion";
import { TopologyNode } from "./topology-node";
import type { Component } from "@/lib/types";
import { useState } from "react";

interface TopologyMapProps {
  components: Component[];
  onNodeClick?: (id: string) => void;
}

export function TopologyMap({ components, onNodeClick }: TopologyMapProps) {
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  const handleNodeClick = (id: string) => {
    setSelectedNode(id === selectedNode ? null : id);
    onNodeClick?.(id);
  };

  // Group components by kind for layout
  const grouped = components.reduce(
    (acc, comp) => {
      const group = acc[comp.kind] || [];
      group.push(comp);
      acc[comp.kind] = group;
      return acc;
    },
    {} as Record<string, Component[]>
  );

  const kindOrder = ["api", "service", "database", "cache", "queue", "server"];
  const sortedGroups = kindOrder.filter((k) => grouped[k]);

  return (
    <div className="relative w-full min-h-100 bg-card rounded-2xl border border-card-border p-8 overflow-hidden">
      {/* Background grid */}
      <div
        className="absolute inset-0 opacity-5"
        style={{
          backgroundImage:
            "radial-gradient(circle, var(--muted) 1px, transparent 1px)",
          backgroundSize: "24px 24px",
        }}
      />

      {/* Connection lines (SVG overlay) */}
      <svg className="absolute inset-0 w-full h-full pointer-events-none">
        {components.length > 1 &&
          components.slice(0, -1).map((comp, i) => {
            const nextComp = components[i + 1];
            if (!nextComp) return null;
            const x1 = 100 + (i % 4) * 200 + 64;
            const y1 = 60 + Math.floor(i / 4) * 160 + 50;
            const x2 = 100 + ((i + 1) % 4) * 200 + 64;
            const y2 = 60 + Math.floor((i + 1) / 4) * 160 + 50;
            return (
              <motion.line
                key={`edge-${comp.id}-${nextComp.id}`}
                x1={x1}
                y1={y1}
                x2={x2}
                y2={y2}
                stroke="var(--card-border)"
                strokeWidth="2"
                strokeDasharray="4 4"
                initial={{ pathLength: 0 }}
                animate={{ pathLength: 1 }}
                transition={{ duration: 0.8, delay: i * 0.1 }}
              />
            );
          })}
      </svg>

      {/* Nodes */}
      <div className="relative z-10">
        <AnimatePresence mode="popLayout">
          {sortedGroups.map((kind, groupIdx) => (
            <div key={kind} className="mb-8">
              <h3 className="text-xs uppercase tracking-wider text-muted mb-4 font-semibold">
                {kind}
              </h3>
              <div className="flex flex-wrap gap-4">
                {grouped[kind].map((comp) => (
                  <TopologyNode
                    key={comp.id}
                    id={comp.id}
                    name={comp.name}
                    kind={comp.kind}
                    status={comp.status || "unknown"}
                    score={comp.health_score || 0}
                    selected={selectedNode === comp.id}
                    onClick={handleNodeClick}
                  />
                ))}
              </div>
            </div>
          ))}
        </AnimatePresence>
      </div>

      {/* Empty state */}
      {components.length === 0 && (
        <div className="flex flex-col items-center justify-center h-64 text-muted">
          <p className="text-sm">No components registered</p>
          <p className="text-xs mt-1">Register components via the CLI or API</p>
        </div>
      )}
    </div>
  );
}
