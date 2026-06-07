"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, OptrionAPIError } from "./api";
import type {
  DashboardStats,
  Incident,
  Alert,
  Component,
  HealthScore,
} from "./types";

// ─── Auth helpers ──────────────────────────────────────────────────────

export function useIsConfigured(): boolean {
  if (typeof window === "undefined") return false;
  return !!localStorage.getItem("optrion_api_key");
}

// ─── Dashboard / System ────────────────────────────────────────────────

export function useSystemInfo() {
  return useQuery({
    queryKey: ["system", "info"],
    queryFn: () => api.info(),
  });
}

export function useReadiness() {
  return useQuery({
    queryKey: ["system", "readyz"],
    queryFn: () => api.readyz(),
    refetchInterval: 30_000,
  });
}

export function useDashboardStats() {
  return useQuery<DashboardStats>({
    queryKey: ["dashboard", "stats"],
    queryFn: async () => {
      // Compose stats from multiple endpoints
      const [components, incidents, alerts] = await Promise.all([
        api.listComponents("").catch(() => ({ data: [], count: 0 })),
        api.listIncidents({ status: "open" }).catch(() => ({ data: [], count: 0 })),
        api.listAlerts().catch(() => ({ data: [] })),
      ]);

      const comps = components.data || [];
      const healthy = comps.filter((c) => c.status === "healthy").length;
      const degraded = comps.filter((c) => c.status === "degraded").length;
      const unhealthy = comps.filter((c) => c.status === "unhealthy").length;
      const avgScore = comps.length > 0
        ? Math.round(comps.reduce((sum, c) => sum + (c.health_score || 0), 0) / comps.length)
        : 0;

      return {
        total_components: comps.length,
        healthy_components: healthy,
        degraded_components: degraded,
        unhealthy_components: unhealthy,
        open_incidents: incidents.count || incidents.data.length,
        active_alerts: (alerts.data || []).filter((a) => a.status === "pending").length,
        average_health_score: avgScore,
      };
    },
  });
}

// ─── Components ────────────────────────────────────────────────────────

export function useComponents(environmentId?: string) {
  return useQuery<Component[]>({
    queryKey: ["components", environmentId],
    queryFn: async () => {
      const res = await api.listComponents(environmentId || "");
      return res.data;
    },
  });
}

// ─── Health ────────────────────────────────────────────────────────────

export function useHealthScores(componentId: string) {
  return useQuery<HealthScore>({
    queryKey: ["health", "scores", componentId],
    queryFn: () => api.getHealthScores(componentId),
    enabled: !!componentId,
  });
}

// ─── Incidents ─────────────────────────────────────────────────────────

export function useIncidents(params?: { status?: string; limit?: number }) {
  return useQuery<Incident[]>({
    queryKey: ["incidents", params],
    queryFn: async () => {
      const res = await api.listIncidents(params);
      return res.data;
    },
  });
}

export function useIncident(id: string) {
  return useQuery<Incident>({
    queryKey: ["incidents", id],
    queryFn: () => api.getIncident(id),
    enabled: !!id,
  });
}

export function useAcknowledgeIncident() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.acknowledgeIncident(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["incidents"] });
    },
  });
}

export function useResolveIncident() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.resolveIncident(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["incidents"] });
    },
  });
}

// ─── Alerts ────────────────────────────────────────────────────────────

export function useAlerts() {
  return useQuery<Alert[]>({
    queryKey: ["alerts"],
    queryFn: async () => {
      const res = await api.listAlerts();
      return res.data;
    },
  });
}

// ─── AI Analysis ───────────────────────────────────────────────────────

export function useIncidentAnalysis(incidentId: string) {
  return useQuery({
    queryKey: ["ai", "analysis", incidentId],
    queryFn: () => api.getIncidentAnalysis(incidentId),
    enabled: !!incidentId,
  });
}

export function useTriggerAnalysis() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (incidentId: string) => api.triggerAnalysis(incidentId),
    onSuccess: (_data, incidentId) => {
      queryClient.invalidateQueries({ queryKey: ["ai", "analysis", incidentId] });
    },
  });
}

export function useIncidentRecommendations(incidentId: string) {
  return useQuery({
    queryKey: ["ai", "recommendations", incidentId],
    queryFn: async () => {
      const res = await api.getIncidentRecommendations(incidentId);
      return res.data;
    },
    enabled: !!incidentId,
  });
}
