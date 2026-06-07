"use client";

import { motion } from "framer-motion";
import { useState } from "react";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { AlertFeed } from "@/components/alerts/alert-feed";
import { ListSkeleton, QueryError } from "@/components/ui/loading";
import { useAlerts } from "@/lib/hooks";
import type { Alert } from "@/lib/types";

export default function AlertsPage() {
  const { data: alerts, isLoading, error, refetch } = useAlerts();
  const [filter, setFilter] = useState<string>("all");

  const filteredAlerts = (alerts || []).filter(
    (a) => filter === "all" || a.severity === filter
  );

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
            <h1 className="text-2xl font-bold">Alert Feed</h1>
            <p className="text-sm text-muted mt-1">
              Real-time alert stream with severity filtering and delivery status
            </p>
          </motion.div>

          {/* Filters */}
          <div className="flex items-center gap-3">
            {["all", "critical", "major", "warning", "info"].map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-colors capitalize ${
                  filter === f
                    ? "border-accent text-accent bg-accent/10"
                    : "border-card-border hover:border-accent hover:text-accent"
                }`}
              >
                {f}
              </button>
            ))}
          </div>

          <div className="max-w-3xl">
            {isLoading ? (
              <ListSkeleton count={5} />
            ) : error ? (
              <QueryError error={error as Error} onRetry={() => refetch()} />
            ) : (
              <AlertFeed alerts={filteredAlerts} />
            )}
          </div>
        </main>
      </div>
    </div>
  );
}
