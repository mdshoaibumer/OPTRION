"use client";

import { motion } from "framer-motion";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { TopologyMap } from "@/components/topology/topology-map";
import type { Component } from "@/lib/types";

const mockComponents: Component[] = [
  { id: "c1", tenant_id: "t1", environment_id: "e1", name: "Backend API", slug: "backend-api", kind: "api", endpoint_url: "http://localhost:3000", status: "healthy", health_score: 95, created_at: "" },
  { id: "c2", tenant_id: "t1", environment_id: "e1", name: "Auth Service", slug: "auth-service", kind: "service", endpoint_url: "http://localhost:3001", status: "healthy", health_score: 92, created_at: "" },
  { id: "c3", tenant_id: "t1", environment_id: "e1", name: "PostgreSQL Primary", slug: "pg-primary", kind: "database", endpoint_url: "postgres://localhost:5432", status: "degraded", health_score: 62, created_at: "" },
  { id: "c4", tenant_id: "t1", environment_id: "e1", name: "PostgreSQL Replica", slug: "pg-replica", kind: "database", endpoint_url: "postgres://localhost:5433", status: "healthy", health_score: 88, created_at: "" },
  { id: "c5", tenant_id: "t1", environment_id: "e1", name: "Redis Cache", slug: "redis-cache", kind: "cache", endpoint_url: "redis://localhost:6379", status: "healthy", health_score: 91, created_at: "" },
  { id: "c6", tenant_id: "t1", environment_id: "e1", name: "Redis Sessions", slug: "redis-sessions", kind: "cache", endpoint_url: "redis://localhost:6380", status: "unhealthy", health_score: 28, created_at: "" },
  { id: "c7", tenant_id: "t1", environment_id: "e1", name: "NGINX Proxy", slug: "nginx", kind: "server", endpoint_url: "http://localhost:80", status: "healthy", health_score: 99, created_at: "" },
];

export default function TopologyPage() {
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
            <p className="text-sm text-[var(--muted)] mt-1">
              Component dependency graph with real-time health status
            </p>
          </motion.div>

          <TopologyMap components={mockComponents} />
        </main>
      </div>
    </div>
  );
}
