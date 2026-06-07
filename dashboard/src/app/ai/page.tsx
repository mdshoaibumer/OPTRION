"use client";

import { motion } from "framer-motion";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { ListSkeleton, QueryError, EmptyState } from "@/components/ui/loading";
import { useIncidents, useIncidentAnalysis, useIncidentRecommendations } from "@/lib/hooks";
import { Brain, Lightbulb, Target, Zap } from "lucide-react";

export default function AIPage() {
  const { data: incidents, isLoading, error, refetch } = useIncidents();

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex-1 ml-16 lg:ml-64">
        <Header />
        <main className="p-6 space-y-6">
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
          >
            <h1 className="text-2xl font-bold">AI Insights</h1>
            <p className="text-sm text-muted mt-1">
              Root cause analysis and intelligent recommendations powered by AI
            </p>
          </motion.div>

          {isLoading ? (
            <ListSkeleton count={2} />
          ) : error ? (
            <QueryError error={error as Error} onRetry={() => refetch()} />
          ) : !incidents || incidents.length === 0 ? (
            <EmptyState message="No incidents to analyze. AI insights will appear here when incidents are detected." icon={<Brain className="h-8 w-8" />} />
          ) : (
            <div className="space-y-6">
              {incidents.map((incident, idx) => (
                <AIInsightCard key={incident.id} incident={incident} idx={idx} />
              ))}
            </div>
          )}
        </main>
      </div>
    </div>
  );
}

function AIInsightCard({ incident, idx }: { incident: { id: string; title: string }; idx: number }) {
  const { data: analysis } = useIncidentAnalysis(incident.id);
  const { data: recommendations } = useIncidentRecommendations(incident.id);

  if (!analysis) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: idx * 0.1 }}
        className="rounded-xl border border-card-border bg-card p-6"
      >
        <div className="flex items-center gap-3">
          <div className="rounded-lg p-2 bg-(--accent-glow)">
            <Brain className="h-5 w-5 text-accent" />
          </div>
          <div>
            <h3 className="text-sm font-semibold">{incident.title}</h3>
            <span className="text-[10px] text-muted font-mono">{incident.id}</span>
          </div>
        </div>
        <p className="text-xs text-muted mt-4">No AI analysis available yet. Run analysis from the Incident War Room.</p>
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: idx * 0.1 }}
      className="rounded-xl border border-card-border bg-card overflow-hidden"
    >
      {/* Header */}
      <div className="px-6 py-4 border-b border-card-border flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="rounded-lg p-2 bg-(--accent-glow)">
            <Brain className="h-5 w-5 text-accent" />
          </div>
          <div>
            <h3 className="text-sm font-semibold">{incident.title}</h3>
            <span className="text-[10px] text-muted font-mono">{incident.id}</span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted">Confidence</span>
          <span
            className="text-sm font-bold font-mono"
            style={{
              color: analysis.confidence >= 0.8 ? "var(--success)" : analysis.confidence >= 0.6 ? "var(--warning)" : "var(--danger)",
            }}
          >
            {Math.round(analysis.confidence * 100)}%
          </span>
        </div>
      </div>

      <div className="p-6 grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Root Cause */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Target className="h-4 w-4 text-danger" />
            <h4 className="text-xs uppercase tracking-wider font-semibold text-muted">Root Cause</h4>
          </div>
          <p className="text-sm leading-relaxed">{analysis.likely_cause}</p>

          <div className="mt-4">
            <h4 className="text-xs uppercase tracking-wider font-semibold text-muted mb-2">Affected Components</h4>
            <div className="flex flex-wrap gap-2">
              {(analysis.affected_components || []).map((comp: string) => (
                <span key={comp} className="text-xs px-2 py-1 rounded-md bg-card-border text-foreground">{comp}</span>
              ))}
            </div>
          </div>

          {analysis.investigation_hints && (
            <div className="mt-4">
              <div className="flex items-center gap-2 mb-2">
                <Lightbulb className="h-4 w-4 text-warning" />
                <h4 className="text-xs uppercase tracking-wider font-semibold text-muted">Investigation Hints</h4>
              </div>
              <ul className="space-y-1.5">
                {analysis.investigation_hints.map((hint: string, i: number) => (
                  <li key={i} className="text-xs text-muted flex items-start gap-2">
                    <span className="text-accent mt-0.5">→</span>
                    {hint}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {/* Recommendations */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Zap className="h-4 w-4 text-accent" />
            <h4 className="text-xs uppercase tracking-wider font-semibold text-muted">Recommendations</h4>
          </div>
          <div className="space-y-3">
            {(recommendations || []).map((rec, i) => (
              <div key={i} className="rounded-lg border border-card-border p-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">{rec.title}</span>
                </div>
                <div className="flex items-center gap-3 mt-2">
                  <span
                    className="text-[10px] uppercase font-semibold px-2 py-0.5 rounded-full"
                    style={{
                      backgroundColor: rec.priority === "high" ? "var(--danger-glow)" : rec.priority === "medium" ? "var(--warning-glow)" : "var(--success-glow)",
                      color: rec.priority === "high" ? "var(--danger)" : rec.priority === "medium" ? "var(--warning)" : "var(--success)",
                    }}
                  >
                    {rec.priority}
                  </span>
                  <span className="text-[10px] text-muted">Risk: {rec.risk_level}</span>
                </div>
              </div>
            ))}
            {(!recommendations || recommendations.length === 0) && (
              <p className="text-xs text-muted">No recommendations generated yet.</p>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  );
}
