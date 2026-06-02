// Optrion API Client
// Communicates with the Go backend via REST API

const API_BASE = process.env.NEXT_PUBLIC_OPTRION_API_URL || "http://localhost:8080";

class OptrionAPIError extends Error {
  constructor(
    public status: number,
    public body: { error: string; details?: string[] }
  ) {
    super(body.error);
    this.name = "OptrionAPIError";
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const apiKey = typeof window !== "undefined"
    ? localStorage.getItem("optrion_api_key") || ""
    : "";

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(apiKey ? { Authorization: `Bearer ${apiKey}` } : {}),
      ...options.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Unknown error" }));
    throw new OptrionAPIError(res.status, body);
  }

  return res.json();
}

// Health & Info
export const api = {
  // System
  healthz: () => request<{ status: string }>("/healthz"),
  readyz: () => request<{ status: string; checks: Record<string, string> }>("/readyz"),
  info: () => request<{ name: string; version: string; environment: string }>("/api/v1/info"),

  // Tenants
  listTenants: () => request<{ data: import("./types").Tenant[]; count: number }>("/api/v1/tenants"),
  getTenant: (id: string) => request<import("./types").Tenant>(`/api/v1/tenants/${encodeURIComponent(id)}`),

  // Products
  listProducts: () => request<{ data: import("./types").Product[]; count: number }>("/api/v1/products"),

  // Environments
  listEnvironments: (productId: string) =>
    request<{ data: import("./types").Environment[]; count: number }>(
      `/api/v1/environments?product_id=${encodeURIComponent(productId)}`
    ),

  // Components
  listComponents: (environmentId: string) =>
    request<{ data: import("./types").Component[]; count: number }>(
      `/api/v1/components?environment_id=${encodeURIComponent(environmentId)}`
    ),

  // Health
  getHealthScores: (componentId: string) =>
    request<import("./types").HealthScore>(`/api/v1/health/scores/${encodeURIComponent(componentId)}`),
  getHealthMetrics: (componentId: string) =>
    request<{ data: import("./types").MetricSnapshot[] }>(
      `/api/v1/health/metrics?component_id=${encodeURIComponent(componentId)}`
    ),

  // Incidents
  listIncidents: (params?: { status?: string; limit?: number }) => {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.set("status", params.status);
    if (params?.limit) searchParams.set("limit", params.limit.toString());
    const qs = searchParams.toString();
    return request<{ data: import("./types").Incident[]; count: number }>(
      `/api/v1/incidents${qs ? `?${qs}` : ""}`
    );
  },
  getIncident: (id: string) => request<import("./types").Incident>(`/api/v1/incidents/${encodeURIComponent(id)}`),
  acknowledgeIncident: (id: string) =>
    request<import("./types").Incident>(`/api/v1/incidents/${encodeURIComponent(id)}/acknowledge`, { method: "POST" }),
  resolveIncident: (id: string) =>
    request<import("./types").Incident>(`/api/v1/incidents/${encodeURIComponent(id)}/resolve`, { method: "POST" }),

  // Alerts
  listAlerts: () => request<{ data: import("./types").Alert[] }>("/api/v1/alerts"),

  // AI Analysis
  getIncidentAnalysis: (incidentId: string) =>
    request<import("./types").RootCauseReport>(
      `/api/v1/incidents/${encodeURIComponent(incidentId)}/analysis`
    ),
  triggerAnalysis: (incidentId: string) =>
    request<import("./types").AIAnalysis>(
      `/api/v1/incidents/${encodeURIComponent(incidentId)}/analyze`,
      { method: "POST" }
    ),

  // Recommendations
  getIncidentRecommendations: (incidentId: string) =>
    request<{ data: import("./types").Recommendation[] }>(
      `/api/v1/incidents/${encodeURIComponent(incidentId)}/recommendations`
    ),
  triggerRecommendations: (incidentId: string) =>
    request<unknown>(
      `/api/v1/incidents/${encodeURIComponent(incidentId)}/recommend`,
      { method: "POST" }
    ),
};

export { OptrionAPIError };
