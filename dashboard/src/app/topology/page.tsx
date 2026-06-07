"use client";

import { motion } from "framer-motion";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { TopologyMap } from "@/components/topology/topology-map";
import { PageLoader, QueryError } from "@/components/ui/loading";
import { useComponents } from "@/lib/hooks";

export default function TopologyPage() {
  const { data: components, isLoading, error, refetch } = useComponents();

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
            <h1 className="text-2xl font-bold">Infrastructure Topology</h1>
            <p className="text-sm text-muted mt-1">
              Component dependency graph with real-time health status
            </p>
          </motion.div>

          {isLoading ? (
            <PageLoader />
          ) : error ? (
            <QueryError error={error as Error} onRetry={() => refetch()} />
          ) : (
            <TopologyMap components={components || []} />
          )}
        </main>
      </div>
    </div>
  );
}
