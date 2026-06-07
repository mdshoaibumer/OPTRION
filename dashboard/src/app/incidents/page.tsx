"use client";

import { motion } from "framer-motion";
import { useState } from "react";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { IncidentList } from "@/components/incidents/incident-list";
import { ListSkeleton, QueryError } from "@/components/ui/loading";
import { useIncidents, useAcknowledgeIncident, useResolveIncident, useTriggerAnalysis } from "@/lib/hooks";
import type { Incident } from "@/lib/types";
import { Brain, CheckCircle2, Clock, MessageSquare, Users } from "lucide-react";

export default function IncidentsPage() {
  const [selected, setSelected] = useState<Incident | null>(null);
  const { data: incidents, isLoading, error, refetch } = useIncidents();
  const acknowledgeMutation = useAcknowledgeIncident();
  const resolveMutation = useResolveIncident();
  const analysisMutation = useTriggerAnalysis();

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1 ml-16 lg:ml-64">
        <Header />
        <main className="p-6">
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6"
          >
            <h1 className="text-2xl font-bold">Incident War Room</h1>
            <p className="text-sm text-muted mt-1">
              Manage active incidents with AI-powered root cause analysis
            </p>
          </motion.div>

          <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
            {/* Incident list */}
            <div className="lg:col-span-2">
              {isLoading ? (
                <ListSkeleton count={4} />
              ) : error ? (
                <QueryError error={error as Error} onRetry={() => refetch()} />
              ) : (
                <IncidentList
                  incidents={incidents || []}
                  onSelect={setSelected}
                  selectedId={selected?.id}
                />
              )}
            </div>

            {/* Incident detail panel */}
            <motion.div
              className="lg:col-span-3 rounded-xl border border-card-border bg-card p-6"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
            >
              {selected ? (
                <div className="space-y-6">
                  {/* Header */}
                  <div>
                    <div className="flex items-center gap-2 mb-2">
                      <span className="text-xs font-mono text-muted">
                        {selected.id}
                      </span>
                      <span className="text-xs uppercase px-2 py-0.5 rounded-full bg-red-500/10 text-red-400 border border-red-500/20">
                        {selected.severity}
                      </span>
                    </div>
                    <h2 className="text-lg font-bold">{selected.title}</h2>
                    <p className="text-sm text-muted mt-2">
                      {selected.description}
                    </p>
                  </div>

                  {/* Timeline */}
                  <div>
                    <h3 className="text-xs uppercase tracking-wider text-muted font-semibold mb-3">
                      Timeline
                    </h3>
                    <div className="space-y-3">
                      <TimelineEntry
                        icon={<Clock className="h-3.5 w-3.5" />}
                        label="Detected"
                        time={selected.occurred_at}
                        color="var(--danger)"
                      />
                      {selected.acknowledged_at && (
                        <TimelineEntry
                          icon={<Users className="h-3.5 w-3.5" />}
                          label="Acknowledged"
                          time={selected.acknowledged_at}
                          color="var(--warning)"
                        />
                      )}
                      {selected.resolved_at && (
                        <TimelineEntry
                          icon={<CheckCircle2 className="h-3.5 w-3.5" />}
                          label="Resolved"
                          time={selected.resolved_at}
                          color="var(--success)"
                        />
                      )}
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex gap-3">
                    <button
                      onClick={() => analysisMutation.mutate(selected.id)}
                      disabled={analysisMutation.isPending}
                      className="flex items-center gap-2 px-4 py-2 rounded-lg bg-accent text-white text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
                    >
                      <Brain className="h-4 w-4" />
                      {analysisMutation.isPending ? "Analyzing..." : "Run AI Analysis"}
                    </button>
                    <button
                      onClick={() => acknowledgeMutation.mutate(selected.id)}
                      disabled={acknowledgeMutation.isPending || selected.status === "acknowledged"}
                      className="flex items-center gap-2 px-4 py-2 rounded-lg border border-card-border text-sm font-medium hover:bg-card-border transition-colors disabled:opacity-50"
                    >
                      <CheckCircle2 className="h-4 w-4" />
                      Acknowledge
                    </button>
                    <button
                      onClick={() => resolveMutation.mutate(selected.id)}
                      disabled={resolveMutation.isPending || selected.status === "resolved"}
                      className="flex items-center gap-2 px-4 py-2 rounded-lg border border-card-border text-sm font-medium hover:bg-card-border transition-colors disabled:opacity-50"
                    >
                      <MessageSquare className="h-4 w-4" />
                      Resolve
                    </button>
                  </div>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-64 text-muted">
                  <p className="text-sm">Select an incident to view details</p>
                </div>
              )}
            </motion.div>
          </div>
        </main>
      </div>
    </div>
  );
}

function TimelineEntry({
  icon,
  label,
  time,
  color,
}: {
  icon: React.ReactNode;
  label: string;
  time: string;
  color: string;
}) {
  return (
    <div className="flex items-center gap-3">
      <div
        className="rounded-full p-1.5"
        style={{ backgroundColor: `${color}20`, color }}
      >
        {icon}
      </div>
      <div className="flex-1">
        <span className="text-xs font-medium">{label}</span>
      </div>
      <span className="text-xs text-muted font-mono">
        {new Date(time).toLocaleString()}
      </span>
    </div>
  );
}
