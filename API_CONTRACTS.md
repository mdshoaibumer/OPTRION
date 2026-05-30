# Recommendation Intelligence API Contracts

## Recommendation APIs

### GET /api/v1/incidents/{id}/recommendations
- Get recommendations for an incident
- Response: RecommendationReportDTO

### POST /api/v1/incidents/{id}/recommend
- Request recommendations for an incident
- Response: RecommendationReportDTO

### GET /api/v1/recommendations
- List all recommendations for the tenant
- Response: List of RecommendationDTO

---

## DTOs

### RecommendationDTO
- id: string (UUID)
- tenant_id: string (UUID)
- incident_id: string (UUID)
- report_id: string (UUID)
- category: string
- priority: string
- title: string
- description: string
- confidence: number
- risk_level: string
- evidence: RecommendationEvidenceDTO[]
- created_at: string (RFC3339)

### RecommendationReportDTO
- id: string (UUID)
- tenant_id: string (UUID)
- incident_id: string (UUID)
- recommendations: RecommendationDTO[]
- created_at: string (RFC3339)

### RecommendationEvidenceDTO
- id: string (UUID)
- tenant_id: string (UUID)
- incident_id: string (UUID)
- recommendation_id: string (UUID)
- description: string
- created_at: string (RFC3339)

---

## Dashboard DTOs

### RecommendationFeedDTO
- recommendations: RecommendationDTO[]

### RecommendationHistoryDTO
- incident_id: string (UUID)
- reports: RecommendationReportDTO[]
# AI Root Cause Intelligence API Contracts

## Analysis APIs

### GET /api/v1/incidents/{id}/analysis
- Get root cause analysis for an incident
- Response: RootCauseReportDTO

### POST /api/v1/incidents/{id}/analyze
- Request AI analysis for an incident
- Response: AIAnalysisDTO

### GET /api/v1/analysis
- List all AI analyses for the tenant
- Response: List of AIAnalysisDTO

---

## DTOs

### AIAnalysisDTO
- id: string (UUID)
- tenant_id: string (UUID)
- incident_id: string (UUID)
- context_id: string (UUID)
- report_id: string (UUID)
- provider: string
- requested_at: string (RFC3339)
- completed_at: string (RFC3339)
- status: string
- created_at: string (RFC3339)

### RootCauseReportDTO
- id: string (UUID)
- tenant_id: string (UUID)
- incident_id: string (UUID)
- likely_cause: string
- affected_components: string[]
- confidence: number
- investigation_hints: string[]
- created_at: string (RFC3339)

### AIContextDTO
- id: string (UUID)
- tenant_id: string (UUID)
- incident_id: string (UUID)
- snapshot: object
- created_at: string (RFC3339)

---

## Dashboard DTOs

### RootCauseFeedDTO
- analyses: RootCauseReportDTO[]

### InvestigationGuidanceDTO
- likely_cause: string
- confidence: number
- affected_components: string[]
- investigation_hints: string[]

### AnalysisHistoryDTO
- incident_id: string (UUID)
- analyses: AIAnalysisDTO[]
# Alert Intelligence Platform API Contracts

## Alert APIs

### GET /api/v1/alerts
- List all alerts for the tenant.
- Query params: severity, status, incident_id, rule_id, channel_id, time range
- Response: List of AlertDTO

### GET /api/v1/alerts/{id}
- Get alert by ID
- Response: AlertDTO

## Alert Rule APIs

### GET /api/v1/alert-rules
- List all alert rules for the tenant
- Response: List of AlertRuleDTO

### POST /api/v1/alert-rules
- Create a new alert rule
- Request: AlertRuleCreateDTO
- Response: AlertRuleDTO

### PATCH /api/v1/alert-rules/{id}
- Update an alert rule
- Request: AlertRuleUpdateDTO
- Response: AlertRuleDTO

## Escalation Policy APIs

### GET /api/v1/escalation-policies
- List all escalation policies for the tenant
- Response: List of EscalationPolicyDTO

### POST /api/v1/escalation-policies
- Create a new escalation policy
- Request: EscalationPolicyCreateDTO
- Response: EscalationPolicyDTO

---

## DTOs

### AlertDTO
- id: string (UUID)
- tenant_id: string (UUID)
- rule_id: string (UUID)
- incident_id: string (UUID)
- severity: string
- status: string
- message: string
- deliveries: AlertDeliveryDTO[]
- created_at: string (RFC3339)
- updated_at: string (RFC3339)

### AlertRuleDTO
- id: string (UUID)
- tenant_id: string (UUID)
- name: string
- description: string
- severity: string
- enabled: boolean
- conditions: object[]
- channels: string[] (UUID)
- escalation_policy_id: string (UUID)
- created_at: string (RFC3339)
- updated_at: string (RFC3339)

### EscalationPolicyDTO
- id: string (UUID)
- tenant_id: string (UUID)
- name: string
- description: string
- steps: object[]
- created_at: string (RFC3339)
- updated_at: string (RFC3339)

### AlertDeliveryDTO
- id: string (UUID)
- channel_id: string (UUID)
- status: string
- attempts: number
- last_error: string
- history: object[]
- created_at: string (RFC3339)
- updated_at: string (RFC3339)

---

## Dashboard DTOs

### AlertFeedDTO
- alerts: AlertDTO[]

### DeliveryStatusDTO
- delivery_id: string (UUID)
- status: string
- attempts: number
- last_error: string
- history: object[]

### EscalationStatusDTO
- policy_id: string (UUID)
- current_step: number
- next_escalation_at: string (RFC3339)

### NotificationHistoryDTO
- alert_id: string (UUID)
- deliveries: AlertDeliveryDTO[]
# OPTRION — API & Contract Architecture

**Author:** Principal API Architect  
**Date:** 2026-05-29  
**Version:** 1.0  
**Status:** Contract Design Complete

---

## API Strategy

### Communication Model Decision

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **REST only** | Simple, universal, SDK-friendly, cacheable | No real-time push, polling for updates | ✅ Phase 1 |
| **REST + WebSocket** | REST for commands/queries, WS for real-time dashboard push | Additional connection management | ✅ Phase 2 |
| **REST + gRPC** | High-performance internal communication | Dual protocol complexity, browser unfriendly | ❌ Not until service extraction |
| **REST + Server-Sent Events** | One-way real-time for dashboards | Limited browser connection pool | Evaluate in Phase 2 |

### Recommendation: REST-First with Event Webhooks

**Phase 1:** REST APIs for all operations. Webhooks for push notifications to customer systems.  
**Phase 2:** WebSocket for real-time dashboard updates.  
**Phase 3+:** gRPC only if internal microservice extraction demands it.

**Rationale:**
1. Every SDK in every language can speak REST
2. Webhooks are the standard for SaaS-to-SaaS communication
3. gRPC adds complexity with zero Phase 1 benefit (single binary, no inter-service calls)
4. WebSocket is only needed when embedded dashboards require <1s refresh

### API Design Principles

| Principle | Rule |
|-----------|------|
| **Resource-oriented** | URLs represent nouns, not verbs. `/components` not `/getComponents` |
| **Consistent verbs** | GET (read), POST (create), PUT (full replace), PATCH (partial update), DELETE (remove) |
| **Idempotent writes** | PUT and DELETE are idempotent. POST with `Idempotency-Key` header for at-least-once safety |
| **Envelope responses** | All responses wrapped in standard envelope with `data`, `meta`, `errors` |
| **Cursor pagination** | No offset-based pagination. Cursor-based for stable performance at scale |
| **UTC timestamps** | All times in RFC 3339 format, always UTC: `2026-05-29T10:15:30Z` |
| **Consistent naming** | snake_case for JSON fields. Plural nouns for collections |
| **HATEOAS-lite** | Include `self` link on every resource. No full hypermedia (overkill for SDK consumers) |

---

## Authentication Strategy

### Option Analysis

| Method | Use Case | Security Level | SDK Friendliness |
|--------|----------|---------------|-----------------|
| **API Key** | Machine-to-machine. SDKs, CI/CD, automation | High (if properly scoped and rotated) | Excellent — single header |
| **JWT (short-lived)** | Human users via dashboard UI | High (with refresh rotation) | Medium — requires token refresh logic |
| **Service Token** | Internal service-to-service (future) | High | N/A (not SDK-facing) |

### Recommendation: API Key for SDK/API + JWT for Dashboard

| Actor | Auth Method | Header | Lifetime |
|-------|-------------|--------|----------|
| SDK / external integration | API Key | `X-API-Key: op_live_abc123...` | Long-lived (manual rotation) |
| Dashboard user (future) | JWT | `Authorization: Bearer eyJ...` | 15 minutes (with refresh token) |
| Webhook verification | HMAC signature | `X-Optrion-Signature: sha256=...` | Per-request |

### API Key Design

**Key Format:**
```
op_{environment}_{random_32_chars}

Examples:
  op_live_k8f2m9x7...    (production)
  op_test_p3n6v1w4...    (testing/staging)
```

**Key Properties:**

| Property | Design |
|----------|--------|
| Prefix | `op_live_` or `op_test_` — immediately identifies environment and platform |
| Length | 48 characters total (prefix + 32 random chars) |
| Storage | Only SHA-256 hash stored in database. Full key shown ONCE at creation |
| Scopes | `ingest`, `read`, `manage`, `admin` — assigned at creation |
| Rotation | New key can be created before revoking old key (overlap period) |
| Rate limiting | Per-key rate limits based on tenant plan |
| Identification | First 8 characters stored as `key_prefix` for debugging ("which key is making this request?") |

### Authentication Flow

```
SDK Request
    │
    ▼
[API Gateway / Middleware]
    │  1. Extract X-API-Key header
    │  2. Hash the key (SHA-256)
    │  3. Lookup hash in api_keys table
    │  4. Verify key is active and not expired
    │  5. Load tenant_id and scopes
    │  6. Set tenant context for RLS
    │
    ▼
[Authorized Request with TenantID in context]
```

### Scope Matrix

| Scope | Capabilities |
|-------|-------------|
| `ingest` | POST to ingestion endpoints only. Cannot read data |
| `read` | GET on all endpoints. Cannot modify anything |
| `manage` | Full CRUD on components, rules, channels. Cannot manage tenant settings |
| `admin` | All operations including tenant configuration and key management |
| `ingest,read` | Combined — typical SDK key for full integration |

### Rate Limiting

| Plan | Requests/minute | Ingestion points/minute | Burst |
|------|----------------|------------------------|-------|
| Free | 60 | 100 | 2x for 10s |
| Starter | 300 | 1,000 | 2x for 30s |
| Professional | 1,000 | 10,000 | 3x for 60s |
| Enterprise | 10,000 | 100,000 | Custom |

Rate limit headers on every response:
```
X-RateLimit-Limit: 300
X-RateLimit-Remaining: 247
X-RateLimit-Reset: 1716984600
X-RateLimit-RetryAfter: 0
```

---

## API Versioning

### Strategy: URL Path Versioning

```
/api/v1/...
/api/v2/...
```

**Why URL path (not header-based):**
1. Immediately visible in logs, docs, and debugging
2. SDKs hardcode the version at compile time — explicit is better
3. No content negotiation complexity
4. Cacheable (different URL = different cache entry)
5. Can run multiple versions simultaneously in the same binary

### Version Lifecycle

| Phase | Status | Duration | Behavior |
|-------|--------|----------|----------|
| **Active** | Current recommended version | Indefinite | Full support, new features added |
| **Deprecated** | Superseded by newer version | 12 months minimum | Functional but no new features. `Sunset` header on responses |
| **Retired** | No longer available | After deprecation period | Returns 410 Gone with migration guide URL |

### Deprecation Header (on deprecated endpoints):

```
Sunset: Sat, 29 May 2027 00:00:00 GMT
Deprecation: true
Link: <https://docs.optrion.io/migration/v1-to-v2>; rel="successor-version"
```

### Breaking vs Non-Breaking Changes

| Change Type | Breaking? | Handling |
|-------------|-----------|----------|
| Adding a new optional field to response | No | Add freely to current version |
| Adding a new endpoint | No | Add to current version |
| Adding a new optional query parameter | No | Add freely |
| Removing a response field | YES | New version required |
| Changing field type | YES | New version required |
| Renaming a field | YES | New version required |
| Changing error codes | YES | New version required |
| Making an optional field required | YES | New version required |
| Adding a required field to request | YES | New version required |

### Evolution Strategy

**v1** — Launched with MVP. Health monitoring, incidents, alerts.  
**v2** (future) — When AI features require new response shapes or when ingestion model fundamentally changes.

**Rule:** Avoid v2 as long as possible. Additive changes to v1 are preferred.

---

## Ingestion Contracts

### Design Philosophy

Ingestion is the entry point for all customer systems. It must be:
- **Simple** — One endpoint, one payload shape
- **Batched** — Accept multiple data points per request (reduce HTTP overhead)
- **Typed** — Clear metric types with validation
- **Flexible** — Support infrastructure metrics AND business metrics
- **Idempotent** — Safe to retry on network failure

---

### `POST /api/v1/ingest`

**Purpose:** Universal ingestion endpoint. Accepts health snapshots, metrics, and status updates in a single batch.

**Headers:**
```
X-API-Key: op_live_...
Content-Type: application/json
Idempotency-Key: {client-generated UUID}  (optional, enables safe retries)
```

**Request Body:**

```json
{
  "environment": "production",
  "component": "backend-api",
  "timestamp": "2026-05-29T10:15:30Z",
  "metrics": [
    {
      "name": "http_response_time_ms",
      "type": "gauge",
      "value": 142.5,
      "unit": "ms",
      "tags": {
        "endpoint": "/api/v1/bookings",
        "method": "POST"
      }
    },
    {
      "name": "http_error_rate",
      "type": "gauge",
      "value": 0.02,
      "unit": "percent"
    },
    {
      "name": "active_connections",
      "type": "gauge",
      "value": 47,
      "unit": "count"
    }
  ],
  "health": {
    "status": "healthy",
    "message": "All systems operational",
    "checks": [
      {
        "name": "database_connection",
        "status": "healthy",
        "latency_ms": 3
      },
      {
        "name": "redis_connection",
        "status": "healthy",
        "latency_ms": 1
      },
      {
        "name": "disk_usage",
        "status": "degraded",
        "message": "Disk usage at 82%",
        "value": 82,
        "threshold": 90
      }
    ]
  }
}
```

**Response (202 Accepted):**

```json
{
  "data": {
    "accepted": true,
    "points_received": 3,
    "health_checks_received": 3,
    "ingested_at": "2026-05-29T10:15:30.142Z"
  },
  "meta": {
    "request_id": "req_abc123",
    "processing": "async"
  }
}
```

**Why 202 (not 200/201):** Ingestion is acknowledged but processed asynchronously. The platform may batch, aggregate, and score in the background.

---

### `POST /api/v1/ingest/batch`

**Purpose:** High-volume batch ingestion. Accept multiple components in a single request.

**Request Body:**

```json
{
  "environment": "production",
  "timestamp": "2026-05-29T10:15:30Z",
  "items": [
    {
      "component": "backend-api",
      "metrics": [
        { "name": "response_time_p95", "type": "gauge", "value": 245, "unit": "ms" }
      ],
      "health": { "status": "healthy" }
    },
    {
      "component": "postgresql",
      "metrics": [
        { "name": "active_connections", "type": "gauge", "value": 23, "unit": "count" },
        { "name": "query_time_p95", "type": "gauge", "value": 12, "unit": "ms" }
      ],
      "health": { "status": "healthy" }
    },
    {
      "component": "redis",
      "metrics": [
        { "name": "memory_usage", "type": "gauge", "value": 67, "unit": "percent" }
      ],
      "health": { "status": "healthy" }
    }
  ]
}
```

**Response (202 Accepted):**

```json
{
  "data": {
    "accepted": true,
    "items_received": 3,
    "total_metrics": 4,
    "total_health_checks": 0,
    "ingested_at": "2026-05-29T10:15:30.142Z"
  },
  "meta": {
    "request_id": "req_def456"
  }
}
```

**Limits:**
- Maximum 100 items per batch request
- Maximum 50 metrics per item
- Maximum 20 health checks per item
- Maximum request body size: 1MB

---

### `POST /api/v1/ingest/event`

**Purpose:** Push business events (deployments, config changes, scaling events) that provide context for incidents.

**Request Body:**

```json
{
  "environment": "production",
  "component": "backend-api",
  "event_type": "deployment",
  "title": "Deployed v2.3.1",
  "description": "Feature: new booking flow. Fix: payment timeout handling.",
  "severity": "info",
  "timestamp": "2026-05-29T10:00:00Z",
  "metadata": {
    "version": "2.3.1",
    "commit": "abc123f",
    "deployer": "ci/cd"
  }
}
```

**Response (202 Accepted):**

```json
{
  "data": {
    "event_id": "evt_789xyz",
    "accepted": true
  },
  "meta": {
    "request_id": "req_ghi789"
  }
}
```

---

### Metric Types

| Type | Meaning | Example |
|------|---------|---------|
| `gauge` | Point-in-time value (can go up or down) | Response time, memory usage, connection count |
| `counter` | Monotonically increasing value (resets on restart) | Total requests, total errors |
| `histogram` | Pre-computed percentiles | `{ "p50": 100, "p95": 250, "p99": 500 }` |

### Auto-Registration Behavior

When an ingestion request references a `component` or `environment` that doesn't exist:

| Setting | Behavior |
|---------|----------|
| `auto_register: true` (default) | Component is automatically created with inferred type |
| `auto_register: false` | Returns 422 with error "component not registered" |

Configured per-tenant in tenant settings.

---

## Health APIs

### `GET /api/v1/health/summary`

**Purpose:** Top-level health overview for the tenant. Single call gives the full picture.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `product` | string | No | Filter by product name/ID |
| `environment` | string | No | Filter by environment name |

**Response (200):**

```json
{
  "data": {
    "overall_score": 87,
    "overall_status": "degraded",
    "products": [
      {
        "id": "prod_abc",
        "name": "GymFlow Track",
        "score": 87,
        "status": "degraded",
        "environments": [
          {
            "id": "env_prod",
            "name": "production",
            "type": "production",
            "score": 82,
            "status": "degraded",
            "component_count": 4,
            "healthy_count": 3,
            "degraded_count": 1,
            "unhealthy_count": 0
          },
          {
            "id": "env_stg",
            "name": "staging",
            "type": "staging",
            "score": 100,
            "status": "healthy",
            "component_count": 3,
            "healthy_count": 3,
            "degraded_count": 0,
            "unhealthy_count": 0
          }
        ]
      }
    ],
    "active_incidents": 1,
    "computed_at": "2026-05-29T10:15:30Z"
  },
  "meta": {
    "request_id": "req_sum001"
  }
}
```

---

### `GET /api/v1/health/components`

**Purpose:** List all components with their current health status.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `environment` | string | No | Filter by environment |
| `status` | string | No | Filter: `healthy`, `degraded`, `unhealthy`, `unknown` |
| `type` | string | No | Filter: `database`, `cache`, `api`, `frontend`, `queue` |
| `sort` | string | No | `score_asc`, `score_desc`, `name`, `last_check` |
| `cursor` | string | No | Pagination cursor |
| `limit` | integer | No | Default 20, max 100 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "comp_api01",
      "name": "backend-api",
      "type": "api",
      "environment": {
        "id": "env_prod",
        "name": "production"
      },
      "health": {
        "status": "healthy",
        "score": 95,
        "last_check_at": "2026-05-29T10:15:00Z",
        "consecutive_successes": 482,
        "consecutive_failures": 0
      },
      "metrics": {
        "availability": 99.97,
        "latency_p95_ms": 245,
        "error_rate": 0.03
      },
      "self": "/api/v1/health/components/comp_api01"
    },
    {
      "id": "comp_pg01",
      "name": "postgresql",
      "type": "database",
      "environment": {
        "id": "env_prod",
        "name": "production"
      },
      "health": {
        "status": "degraded",
        "score": 72,
        "last_check_at": "2026-05-29T10:15:00Z",
        "consecutive_successes": 0,
        "consecutive_failures": 3
      },
      "metrics": {
        "availability": 99.5,
        "latency_p95_ms": 890,
        "error_rate": 0.5
      },
      "self": "/api/v1/health/components/comp_pg01"
    }
  ],
  "meta": {
    "request_id": "req_comp001",
    "pagination": {
      "cursor": "eyJpZCI6ImNvbXBfcGcwMSJ9",
      "has_more": false,
      "total_count": 2
    }
  }
}
```

---

### `GET /api/v1/health/components/{id}`

**Purpose:** Detailed health view for a single component.

**Response (200):**

```json
{
  "data": {
    "id": "comp_pg01",
    "name": "postgresql",
    "type": "database",
    "environment": {
      "id": "env_prod",
      "name": "production"
    },
    "product": {
      "id": "prod_abc",
      "name": "GymFlow Track"
    },
    "health": {
      "status": "degraded",
      "score": 72,
      "factors": [
        { "name": "availability", "weight": 0.4, "score": 90 },
        { "name": "latency", "weight": 0.3, "score": 55 },
        { "name": "consistency", "weight": 0.2, "score": 60 },
        { "name": "freshness", "weight": 0.1, "score": 100 }
      ],
      "last_check_at": "2026-05-29T10:15:00Z",
      "last_success_at": "2026-05-29T10:14:00Z",
      "last_failure_at": "2026-05-29T10:15:00Z"
    },
    "checks": [
      {
        "id": "chk_conn01",
        "name": "connection_pool",
        "type": "tcp",
        "status": "healthy",
        "last_latency_ms": 3,
        "schedule_seconds": 30
      },
      {
        "id": "chk_query01",
        "name": "query_performance",
        "type": "custom",
        "status": "degraded",
        "last_latency_ms": 890,
        "schedule_seconds": 60
      }
    ],
    "dependencies": [
      {
        "id": "comp_disk01",
        "name": "data-volume",
        "type": "hard",
        "status": "healthy"
      }
    ],
    "active_incidents": [
      {
        "id": "INC-GF-042",
        "title": "High query latency on postgresql",
        "severity": "medium",
        "opened_at": "2026-05-29T10:10:00Z"
      }
    ],
    "self": "/api/v1/health/components/comp_pg01"
  },
  "meta": {
    "request_id": "req_comp002"
  }
}
```

---

### `GET /api/v1/health/components/{id}/history`

**Purpose:** Health score time-series for graphing.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `window` | string | No | `1h`, `6h`, `24h`, `7d`, `30d`. Default `24h` |
| `resolution` | string | No | `1m`, `5m`, `15m`, `1h`, `1d`. Auto-determined if omitted |

**Response (200):**

```json
{
  "data": {
    "component_id": "comp_pg01",
    "window": "24h",
    "resolution": "15m",
    "points": [
      { "timestamp": "2026-05-28T10:15:00Z", "score": 95, "status": "healthy" },
      { "timestamp": "2026-05-28T10:30:00Z", "score": 95, "status": "healthy" },
      { "timestamp": "2026-05-28T10:45:00Z", "score": 88, "status": "degraded" },
      { "timestamp": "2026-05-28T11:00:00Z", "score": 72, "status": "degraded" }
    ]
  },
  "meta": {
    "request_id": "req_hist001",
    "point_count": 96
  }
}
```

---

### `GET /api/v1/health/scores`

**Purpose:** Health scores at any aggregation level (component, environment, product).

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `subject_type` | string | Yes | `component`, `environment`, `product` |
| `subject_id` | string | Yes | ID of the subject |
| `window` | string | No | Default `24h` |

---

## Incident APIs

### `GET /api/v1/incidents`

**Purpose:** List incidents with filtering.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | No | `opened`, `acknowledged`, `investigating`, `resolved`, `closed`. Comma-separated for multiple |
| `severity` | string | No | `critical`, `high`, `medium`, `low` |
| `component_id` | string | No | Filter by component |
| `environment` | string | No | Filter by environment |
| `opened_after` | datetime | No | Only incidents opened after this time |
| `opened_before` | datetime | No | Only incidents opened before this time |
| `sort` | string | No | `opened_at_desc` (default), `severity_desc`, `duration_desc` |
| `cursor` | string | No | Pagination cursor |
| `limit` | integer | No | Default 20, max 100 |

**Response (200):**

```json
{
  "data": [
    {
      "id": "INC-GF-042",
      "title": "High query latency on postgresql",
      "severity": "medium",
      "status": "opened",
      "component": {
        "id": "comp_pg01",
        "name": "postgresql",
        "type": "database"
      },
      "environment": {
        "id": "env_prod",
        "name": "production"
      },
      "trigger": {
        "type": "threshold",
        "metric": "latency_p95",
        "threshold": 500,
        "actual_value": 890
      },
      "opened_at": "2026-05-29T10:10:00Z",
      "acknowledged_at": null,
      "resolved_at": null,
      "duration_seconds": null,
      "event_count": 3,
      "self": "/api/v1/incidents/INC-GF-042"
    }
  ],
  "meta": {
    "request_id": "req_inc001",
    "pagination": {
      "cursor": "eyJpZCI6IklOQy1HRi0wNDIifQ",
      "has_more": true,
      "total_count": 15
    }
  }
}
```

---

### `GET /api/v1/incidents/{id}`

**Purpose:** Full incident detail with timeline.

**Response (200):**

```json
{
  "data": {
    "id": "INC-GF-042",
    "title": "High query latency on postgresql",
    "description": "P95 query latency exceeded 500ms threshold. Current value: 890ms.",
    "severity": "medium",
    "status": "opened",
    "component": {
      "id": "comp_pg01",
      "name": "postgresql",
      "type": "database",
      "environment": "production"
    },
    "trigger": {
      "type": "threshold",
      "metric": "latency_p95",
      "threshold": 500,
      "actual_value": 890,
      "consecutive_breaches": 5
    },
    "timeline": [
      {
        "event_type": "opened",
        "actor": "system",
        "message": "Incident opened: P95 latency exceeded 500ms for 5 consecutive checks",
        "occurred_at": "2026-05-29T10:10:00Z"
      },
      {
        "event_type": "alert_sent",
        "actor": "system",
        "message": "Alert sent via Telegram to #ops-alerts",
        "occurred_at": "2026-05-29T10:10:01Z"
      },
      {
        "event_type": "severity_changed",
        "actor": "system",
        "message": "Severity escalated from low to medium (threshold sustained > 5min)",
        "occurred_at": "2026-05-29T10:15:00Z"
      }
    ],
    "related_metrics": {
      "latency_p95_at_trigger": 890,
      "latency_p95_current": 920,
      "error_rate_at_trigger": 0.5,
      "score_at_trigger": 72
    },
    "opened_at": "2026-05-29T10:10:00Z",
    "acknowledged_at": null,
    "resolved_at": null,
    "closed_at": null,
    "duration_seconds": null,
    "self": "/api/v1/incidents/INC-GF-042"
  },
  "meta": {
    "request_id": "req_inc002"
  }
}
```

---

### `POST /api/v1/incidents/{id}/acknowledge`

**Purpose:** Acknowledge an incident (stops escalation timer).

**Request Body:**

```json
{
  "acknowledged_by": "john@gymflow.com",
  "note": "Investigating. Likely related to recent migration."
}
```

**Response (200):**

```json
{
  "data": {
    "id": "INC-GF-042",
    "status": "acknowledged",
    "acknowledged_at": "2026-05-29T10:20:00Z",
    "acknowledged_by": "john@gymflow.com"
  },
  "meta": {
    "request_id": "req_inc003"
  }
}
```

---

### `POST /api/v1/incidents/{id}/resolve`

**Purpose:** Manually resolve an incident.

**Request Body:**

```json
{
  "resolved_by": "john@gymflow.com",
  "resolution_note": "Slow queries caused by missing index on bookings table. Index added.",
  "resolution_type": "manual"
}
```

**Response (200):**

```json
{
  "data": {
    "id": "INC-GF-042",
    "status": "resolved",
    "resolved_at": "2026-05-29T10:45:00Z",
    "resolved_by": "john@gymflow.com",
    "duration_seconds": 2100
  },
  "meta": {
    "request_id": "req_inc004"
  }
}
```

---

### `POST /api/v1/incidents/{id}/close`

**Purpose:** Close a resolved incident (no further updates possible).

**Request Body:**

```json
{
  "closed_by": "john@gymflow.com",
  "postmortem_url": "https://notion.so/gymflow/postmortem-042"
}
```

---

### `POST /api/v1/incidents/{id}/note`

**Purpose:** Add a timeline note to an incident.

**Request Body:**

```json
{
  "author": "john@gymflow.com",
  "message": "Running EXPLAIN ANALYZE on slow queries. Seeing sequential scan on bookings.created_at."
}
```

---

## Alert APIs

### `GET /api/v1/alerts/rules`

**Purpose:** List all configured alert rules.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | No | `active`, `muted`, `disabled` |
| `scope_type` | string | No | `component`, `environment`, `product` |
| `scope_id` | string | No | Filter by specific scope |

**Response (200):**

```json
{
  "data": [
    {
      "id": "rule_001",
      "name": "High latency alert",
      "status": "active",
      "scope": {
        "type": "component",
        "id": "comp_api01",
        "name": "backend-api"
      },
      "condition": {
        "metric": "latency_p95",
        "operator": "greater_than",
        "threshold": 500,
        "unit": "ms",
        "window": "5m",
        "consecutive_breaches": 3
      },
      "severity": "high",
      "channels": [
        { "id": "ch_001", "name": "Ops Telegram", "type": "telegram" }
      ],
      "cooldown_seconds": 300,
      "last_triggered_at": "2026-05-28T14:30:00Z",
      "trigger_count_24h": 2,
      "self": "/api/v1/alerts/rules/rule_001"
    }
  ],
  "meta": {
    "request_id": "req_alr001",
    "pagination": {
      "cursor": null,
      "has_more": false,
      "total_count": 5
    }
  }
}
```

---

### `POST /api/v1/alerts/rules`

**Purpose:** Create a new alert rule.

**Request Body:**

```json
{
  "name": "Database connection saturation",
  "scope": {
    "type": "component",
    "id": "comp_pg01"
  },
  "condition": {
    "metric": "active_connections",
    "operator": "greater_than",
    "threshold": 80,
    "unit": "percent",
    "window": "5m",
    "consecutive_breaches": 2
  },
  "severity": "critical",
  "channel_ids": ["ch_001", "ch_002"],
  "cooldown_seconds": 600,
  "enabled": true
}
```

**Response (201):**

```json
{
  "data": {
    "id": "rule_002",
    "name": "Database connection saturation",
    "status": "active",
    "created_at": "2026-05-29T10:30:00Z",
    "self": "/api/v1/alerts/rules/rule_002"
  },
  "meta": {
    "request_id": "req_alr002"
  }
}
```

---

### `PATCH /api/v1/alerts/rules/{id}`

**Purpose:** Partial update of an alert rule.

**Request Body (only changed fields):**

```json
{
  "condition": {
    "threshold": 70
  },
  "cooldown_seconds": 900
}
```

**Response (200):** Returns full updated rule.

---

### `POST /api/v1/alerts/rules/{id}/mute`

**Purpose:** Temporarily mute a rule.

**Request Body:**

```json
{
  "duration_minutes": 60,
  "reason": "Planned maintenance on database"
}
```

---

### `POST /api/v1/alerts/rules/{id}/unmute`

**Purpose:** Unmute a muted rule immediately.

---

### `DELETE /api/v1/alerts/rules/{id}`

**Purpose:** Delete an alert rule.

**Response (204 No Content)**

---

### `GET /api/v1/alerts/history`

**Purpose:** View past alert firings.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `rule_id` | string | No | Filter by specific rule |
| `severity` | string | No | Filter by severity |
| `suppressed` | boolean | No | Include suppressed alerts |
| `window` | string | No | `24h`, `7d`, `30d`. Default `7d` |

**Response (200):**

```json
{
  "data": [
    {
      "id": "alh_001",
      "rule": {
        "id": "rule_001",
        "name": "High latency alert"
      },
      "severity": "high",
      "trigger_value": 890,
      "channels": [
        { "type": "telegram", "status": "delivered" }
      ],
      "suppressed": false,
      "incident_id": "INC-GF-042",
      "triggered_at": "2026-05-29T10:10:01Z"
    }
  ],
  "meta": {
    "request_id": "req_alh001"
  }
}
```

---

### `GET /api/v1/alerts/channels`

**Purpose:** List notification channels.

**Response (200):**

```json
{
  "data": [
    {
      "id": "ch_001",
      "name": "Ops Telegram",
      "type": "telegram",
      "verified": true,
      "status": "active",
      "last_used_at": "2026-05-29T10:10:01Z",
      "self": "/api/v1/alerts/channels/ch_001"
    }
  ]
}
```

---

### `POST /api/v1/alerts/channels`

**Purpose:** Create a notification channel.

**Request Body:**

```json
{
  "name": "Ops Telegram",
  "type": "telegram",
  "config": {
    "chat_id": "-100123456789",
    "bot_token": "will_be_encrypted"
  }
}
```

**Note:** `config` is stored encrypted. Never returned in GET responses. Only a masked version shown.

---

### `POST /api/v1/alerts/channels/{id}/test`

**Purpose:** Send a test notification to verify channel connectivity.

**Response (200):**

```json
{
  "data": {
    "test_result": "success",
    "delivered_at": "2026-05-29T10:35:00Z",
    "message": "Test notification delivered successfully"
  }
}
```

---

### Maintenance Windows

### `POST /api/v1/maintenance`

**Request Body:**

```json
{
  "scope": {
    "type": "environment",
    "id": "env_prod"
  },
  "reason": "Database migration - adding indexes",
  "starts_at": "2026-05-30T02:00:00Z",
  "ends_at": "2026-05-30T03:00:00Z"
}
```

### `GET /api/v1/maintenance`

### `DELETE /api/v1/maintenance/{id}`

---

## Analytics APIs

### `GET /api/v1/analytics/uptime`

**Purpose:** Uptime percentage for a component or environment over time.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `subject_type` | string | Yes | `component`, `environment`, `product` |
| `subject_id` | string | Yes | ID |
| `window` | string | No | `24h`, `7d`, `30d`, `90d`. Default `30d` |

**Response (200):**

```json
{
  "data": {
    "subject": {
      "type": "component",
      "id": "comp_api01",
      "name": "backend-api"
    },
    "window": "30d",
    "uptime_percent": 99.94,
    "total_downtime_seconds": 1555,
    "incident_count": 3,
    "longest_outage_seconds": 900,
    "daily_breakdown": [
      { "date": "2026-05-29", "uptime_percent": 99.5, "incidents": 1 },
      { "date": "2026-05-28", "uptime_percent": 100, "incidents": 0 }
    ]
  },
  "meta": {
    "request_id": "req_ana001"
  }
}
```

---

### `GET /api/v1/analytics/trends`

**Purpose:** Metric trends over time for graph visualization.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `component_id` | string | Yes | Component to analyze |
| `metrics` | string | Yes | Comma-separated: `latency_p95,error_rate,availability` |
| `window` | string | No | Default `7d` |
| `resolution` | string | No | `5m`, `15m`, `1h`, `1d`. Auto if omitted |

**Response (200):**

```json
{
  "data": {
    "component_id": "comp_api01",
    "window": "7d",
    "resolution": "1h",
    "series": {
      "latency_p95": {
        "unit": "ms",
        "points": [
          { "timestamp": "2026-05-22T00:00:00Z", "value": 220 },
          { "timestamp": "2026-05-22T01:00:00Z", "value": 215 }
        ]
      },
      "error_rate": {
        "unit": "percent",
        "points": [
          { "timestamp": "2026-05-22T00:00:00Z", "value": 0.01 },
          { "timestamp": "2026-05-22T01:00:00Z", "value": 0.02 }
        ]
      }
    }
  },
  "meta": {
    "request_id": "req_ana002",
    "point_count": 168
  }
}
```

---

### `GET /api/v1/analytics/reliability`

**Purpose:** Reliability statistics — MTTR, MTBF, incident frequency.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `subject_type` | string | No | Default: tenant-wide |
| `subject_id` | string | No | |
| `window` | string | No | Default `30d` |

**Response (200):**

```json
{
  "data": {
    "window": "30d",
    "mttr_seconds": 1200,
    "mtbf_seconds": 259200,
    "total_incidents": 8,
    "incidents_by_severity": {
      "critical": 1,
      "high": 2,
      "medium": 3,
      "low": 2
    },
    "average_resolution_time_seconds": 1500,
    "auto_resolved_percent": 37.5,
    "top_affected_components": [
      { "id": "comp_pg01", "name": "postgresql", "incident_count": 4 },
      { "id": "comp_api01", "name": "backend-api", "incident_count": 3 }
    ]
  },
  "meta": {
    "request_id": "req_ana003"
  }
}
```

---

### `GET /api/v1/analytics/incidents/stats`

**Purpose:** Incident statistics for reporting.

**Query Parameters:**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `window` | string | No | Default `30d` |
| `group_by` | string | No | `day`, `week`, `severity`, `component` |

**Response (200):**

```json
{
  "data": {
    "window": "30d",
    "total": 8,
    "by_day": [
      { "date": "2026-05-29", "opened": 1, "resolved": 0 },
      { "date": "2026-05-28", "opened": 0, "resolved": 1 }
    ],
    "by_severity": {
      "critical": { "count": 1, "avg_duration_seconds": 3600 },
      "high": { "count": 2, "avg_duration_seconds": 1800 },
      "medium": { "count": 3, "avg_duration_seconds": 1200 },
      "low": { "count": 2, "avg_duration_seconds": 600 }
    },
    "resolution_types": {
      "auto_resolved": 3,
      "manual": 5
    }
  },
  "meta": {
    "request_id": "req_ana004"
  }
}
```

---

## SDK Architecture

### SDK Design Principles

| Principle | Rule |
|-----------|------|
| **Zero-config start** | SDK works with only an API key. Everything else has sensible defaults |
| **Non-blocking** | SDK never blocks the host application. All ingestion is async with buffering |
| **Graceful degradation** | If OPTRION is unreachable, SDK silently buffers and retries. No exceptions thrown |
| **Minimal dependencies** | SDK imports only what it needs. No transitive dependency bloat |
| **Idempotent** | Network retries are safe (Idempotency-Key header) |
| **Tenant-aware** | API key determines tenant. SDK does not need explicit tenant configuration |

### SDK Lifecycle

```
┌──────────────────────────────────────────────────────┐
│                   HOST APPLICATION                     │
│                                                       │
│  1. Initialize                                        │
│     optrion.Init(apiKey, options)                     │
│                                                       │
│  2. Register components (optional if auto-register)   │
│     optrion.RegisterComponent("backend-api", "api")  │
│                                                       │
│  3. Report metrics (continuous)                       │
│     optrion.Gauge("response_time", 142, tags)        │
│     optrion.Counter("requests_total", 1, tags)       │
│                                                       │
│  4. Report health (continuous)                        │
│     optrion.ReportHealth("backend-api", healthy)     │
│                                                       │
│  5. Report events (on change)                         │
│     optrion.Event("deployment", metadata)            │
│                                                       │
│  6. Shutdown (on app exit)                            │
│     optrion.Shutdown()  // Flushes buffered data     │
│                                                       │
└──────────────────────────────────────────────────────┘
         │                              │
         ▼                              ▼
┌──────────────────┐          ┌─────────────────┐
│  Buffer (in-mem) │──flush──▶│  OPTRION API    │
│  Max: 1000 items │  every   │  POST /ingest   │
│  or 10 seconds   │  10s     │                 │
└──────────────────┘          └─────────────────┘
```

### Go SDK Interface

```go
// Package optrion provides the OPTRION SDK for Go applications.

// Init initializes the OPTRION SDK with the given API key and options.
// Must be called once at application startup.
func Init(apiKey string, opts ...Option) error

// Shutdown gracefully flushes all buffered data and closes connections.
// Must be called before application exits.
func Shutdown(ctx context.Context) error

// --- Configuration Options ---

type Option func(*config)

func WithEnvironment(env string) Option          // "production", "staging", etc.
func WithComponent(name string) Option           // Default component name
func WithFlushInterval(d time.Duration) Option   // Default: 10s
func WithBufferSize(n int) Option                // Default: 1000
func WithEndpoint(url string) Option             // Custom API endpoint
func WithLogger(l Logger) Option                 // Custom logger
func WithRetryPolicy(p RetryPolicy) Option       // Custom retry behavior
func WithAutoRegister(enabled bool) Option       // Auto-register unknown components

// --- Metric Reporting ---

// Gauge records a point-in-time metric value.
func Gauge(name string, value float64, opts ...MetricOption) error

// Counter increments a counter metric.
func Counter(name string, value float64, opts ...MetricOption) error

// Histogram records a pre-computed histogram.
func Histogram(name string, percentiles map[string]float64, opts ...MetricOption) error

type MetricOption func(*metricConfig)

func ForComponent(name string) MetricOption      // Override default component
func WithUnit(unit string) MetricOption          // "ms", "percent", "count"
func WithTags(tags map[string]string) MetricOption

// --- Health Reporting ---

type HealthStatus string
const (
    Healthy   HealthStatus = "healthy"
    Degraded  HealthStatus = "degraded"
    Unhealthy HealthStatus = "unhealthy"
)

// ReportHealth reports the current health status of a component.
func ReportHealth(component string, status HealthStatus, opts ...HealthOption) error

type HealthOption func(*healthConfig)

func WithMessage(msg string) HealthOption
func WithChecks(checks []HealthCheck) HealthOption

type HealthCheck struct {
    Name      string
    Status    HealthStatus
    LatencyMs int
    Message   string
}

// --- Event Reporting ---

// Event reports a significant event (deployment, config change, scaling).
func Event(eventType string, opts ...EventOption) error

type EventOption func(*eventConfig)

func WithTitle(title string) EventOption
func WithDescription(desc string) EventOption
func WithSeverity(sev string) EventOption
func WithMetadata(meta map[string]string) EventOption
```

### JavaScript/TypeScript SDK Interface

```typescript
// Initialize
import { Optrion } from '@optrion/sdk';

const optrion = new Optrion({
  apiKey: 'op_live_...',
  environment: 'production',
  component: 'frontend',
  flushInterval: 10000,  // ms
  bufferSize: 500,
});

// Report metrics
optrion.gauge('page_load_time', 1200, { unit: 'ms', tags: { page: '/dashboard' } });
optrion.counter('user_actions', 1, { tags: { action: 'booking_created' } });

// Report health
optrion.reportHealth('frontend', 'healthy', {
  checks: [
    { name: 'api_reachable', status: 'healthy', latencyMs: 50 },
    { name: 'websocket_connected', status: 'healthy' },
  ],
});

// Report events
optrion.event('deployment', {
  title: 'Frontend v3.2.0 deployed',
  metadata: { version: '3.2.0' },
});

// Shutdown (flush buffer)
await optrion.shutdown();
```

### SDK Configuration Defaults

| Setting | Default | Rationale |
|---------|---------|-----------|
| Flush interval | 10 seconds | Balance between latency and request count |
| Buffer size | 1000 items | Prevents unbounded memory in outage scenarios |
| Retry attempts | 3 | With exponential backoff (1s, 2s, 4s) |
| Request timeout | 5 seconds | Don't block host app |
| Batch size | 100 items per request | Stay within ingestion limits |
| Auto-register | true | Minimal configuration for new users |

### SDK Failure Modes

| Scenario | SDK Behavior |
|----------|-------------|
| OPTRION unreachable | Buffer locally. Retry with backoff. Drop oldest if buffer full |
| Invalid API key | Log error once. Stop sending (don't spam failed auth) |
| Rate limited (429) | Respect `Retry-After` header. Resume when allowed |
| Network timeout | Retry up to 3 times with backoff |
| SDK not initialized | Return error immediately. Never panic |
| Application shutting down | `Shutdown()` flushes remaining buffer with 5s timeout |

---

## Event Contracts

### Webhook Delivery (Outgoing Events to Customer Systems)

Customers can register webhook URLs to receive events from OPTRION.

### Event Envelope (Webhook Payload)

```json
{
  "id": "evt_abc123",
  "type": "incident.opened",
  "version": "1.0",
  "tenant_id": "ten_xyz",
  "occurred_at": "2026-05-29T10:10:00Z",
  "delivered_at": "2026-05-29T10:10:01Z",
  "data": { }
}
```

### Event: `metric.collected`

```json
{
  "id": "evt_m001",
  "type": "metric.collected",
  "version": "1.0",
  "tenant_id": "ten_xyz",
  "occurred_at": "2026-05-29T10:15:30Z",
  "data": {
    "component_id": "comp_api01",
    "component_name": "backend-api",
    "environment": "production",
    "metric_name": "latency_p95",
    "metric_type": "gauge",
    "value": 245,
    "unit": "ms"
  }
}
```

### Event: `health_score.calculated`

```json
{
  "id": "evt_hs001",
  "type": "health_score.calculated",
  "version": "1.0",
  "tenant_id": "ten_xyz",
  "occurred_at": "2026-05-29T10:15:30Z",
  "data": {
    "subject_type": "component",
    "subject_id": "comp_api01",
    "subject_name": "backend-api",
    "score": 95,
    "previous_score": 98,
    "status": "healthy",
    "factors": [
      { "name": "availability", "score": 99 },
      { "name": "latency", "score": 88 },
      { "name": "consistency", "score": 95 },
      { "name": "freshness", "score": 100 }
    ]
  }
}
```

### Event: `incident.opened`

```json
{
  "id": "evt_io001",
  "type": "incident.opened",
  "version": "1.0",
  "tenant_id": "ten_xyz",
  "occurred_at": "2026-05-29T10:10:00Z",
  "data": {
    "incident_id": "INC-GF-042",
    "title": "High query latency on postgresql",
    "severity": "medium",
    "component_id": "comp_pg01",
    "component_name": "postgresql",
    "environment": "production",
    "trigger": {
      "type": "threshold",
      "metric": "latency_p95",
      "threshold": 500,
      "actual_value": 890
    }
  }
}
```

### Event: `incident.acknowledged`

```json
{
  "id": "evt_ia001",
  "type": "incident.acknowledged",
  "version": "1.0",
  "tenant_id": "ten_xyz",
  "occurred_at": "2026-05-29T10:20:00Z",
  "data": {
    "incident_id": "INC-GF-042",
    "acknowledged_by": "john@gymflow.com",
    "time_to_acknowledge_seconds": 600
  }
}
```

### Event: `incident.resolved`

```json
{
  "id": "evt_ir001",
  "type": "incident.resolved",
  "version": "1.0",
  "tenant_id": "ten_xyz",
  "occurred_at": "2026-05-29T10:45:00Z",
  "data": {
    "incident_id": "INC-GF-042",
    "resolved_by": "john@gymflow.com",
    "resolution_type": "manual",
    "resolution_note": "Missing index added to bookings table",
    "duration_seconds": 2100,
    "severity": "medium"
  }
}
```

### Event: `alert.sent`

```json
{
  "id": "evt_as001",
  "type": "alert.sent",
  "version": "1.0",
  "tenant_id": "ten_xyz",
  "occurred_at": "2026-05-29T10:10:01Z",
  "data": {
    "alert_rule_id": "rule_001",
    "alert_rule_name": "High latency alert",
    "severity": "high",
    "channel_type": "telegram",
    "delivery_status": "delivered",
    "incident_id": "INC-GF-042"
  }
}
```

### Event: `health.degraded`

```json
{
  "id": "evt_hd001",
  "type": "health.degraded",
  "version": "1.0",
  "tenant_id": "ten_xyz",
  "occurred_at": "2026-05-29T10:09:00Z",
  "data": {
    "component_id": "comp_pg01",
    "component_name": "postgresql",
    "environment": "production",
    "previous_status": "healthy",
    "new_status": "degraded",
    "previous_score": 95,
    "new_score": 72,
    "reason": "P95 latency exceeded threshold"
  }
}
```

### Event: `health.restored`

```json
{
  "id": "evt_hr001",
  "type": "health.restored",
  "version": "1.0",
  "tenant_id": "ten_xyz",
  "occurred_at": "2026-05-29T11:00:00Z",
  "data": {
    "component_id": "comp_pg01",
    "component_name": "postgresql",
    "environment": "production",
    "previous_status": "degraded",
    "new_status": "healthy",
    "new_score": 96,
    "degradation_duration_seconds": 3060
  }
}
```

### Webhook Security

| Mechanism | Purpose |
|-----------|---------|
| HMAC-SHA256 signature | Verify payload authenticity. Header: `X-Optrion-Signature: sha256=abc123...` |
| Timestamp validation | Reject replayed events older than 5 minutes. Header: `X-Optrion-Timestamp: 1716984930` |
| Retry with backoff | 3 retries (5s, 30s, 120s) on non-2xx responses |
| Event deduplication | Consumer should deduplicate by `id` field (idempotency) |

---

## Business Metrics Framework

### Design Philosophy

OPTRION must support arbitrary business KPIs alongside infrastructure metrics. Business metrics provide context that pure infrastructure monitoring cannot.

### `POST /api/v1/metrics/business`

**Purpose:** Ingest business-specific KPIs that contribute to overall "system health" from a business perspective.

**Request Body:**

```json
{
  "environment": "production",
  "timestamp": "2026-05-29T10:15:00Z",
  "kpis": [
    {
      "name": "checkin_success_rate",
      "value": 98.5,
      "unit": "percent",
      "context": {
        "total_attempts": 200,
        "successful": 197,
        "failed": 3
      },
      "thresholds": {
        "healthy": { "operator": "gte", "value": 95 },
        "degraded": { "operator": "gte", "value": 85 },
        "unhealthy": { "operator": "lt", "value": 85 }
      }
    },
    {
      "name": "invoice_generation_rate",
      "value": 100,
      "unit": "percent",
      "context": {
        "total_expected": 50,
        "generated": 50,
        "failed": 0
      },
      "thresholds": {
        "healthy": { "operator": "gte", "value": 99 },
        "degraded": { "operator": "gte", "value": 90 },
        "unhealthy": { "operator": "lt", "value": 90 }
      }
    },
    {
      "name": "booking_completion_time",
      "value": 3.2,
      "unit": "seconds",
      "thresholds": {
        "healthy": { "operator": "lte", "value": 5 },
        "degraded": { "operator": "lte", "value": 10 },
        "unhealthy": { "operator": "gt", "value": 10 }
      }
    }
  ]
}
```

**Response (202 Accepted):**

```json
{
  "data": {
    "accepted": true,
    "kpis_received": 3,
    "health_impacts": [
      { "name": "checkin_success_rate", "derived_status": "healthy" },
      { "name": "invoice_generation_rate", "derived_status": "healthy" },
      { "name": "booking_completion_time", "derived_status": "healthy" }
    ]
  }
}
```

### KPI Definition Contract

### `POST /api/v1/metrics/business/definitions`

**Purpose:** Pre-define a business KPI with its thresholds and contribution to health score.

```json
{
  "name": "checkin_success_rate",
  "display_name": "Check-in Success Rate",
  "description": "Percentage of successful member check-ins at gym entrance",
  "unit": "percent",
  "direction": "higher_is_better",
  "component_id": "comp_api01",
  "weight": 0.3,
  "thresholds": {
    "healthy": { "operator": "gte", "value": 95 },
    "degraded": { "operator": "gte", "value": 85 },
    "unhealthy": { "operator": "lt", "value": 85 }
  },
  "alert_on": "degraded",
  "collection_interval": "5m"
}
```

### Business KPI Integration with Health Score

```
Component Health Score = weighted_average(
  infrastructure_score × 0.6,    # From check_results (latency, availability, errors)
  business_kpi_score × 0.4       # From business metrics (success rates, SLOs)
)
```

Business KPIs contribute directly to the health score of the component they're attached to. A system can be "infrastructure healthy" but "business unhealthy" if KPIs are degraded.

### Example: GymFlow Track Business KPIs

| KPI | Meaning | Healthy | Degraded | Unhealthy |
|-----|---------|---------|----------|-----------|
| `checkin_success_rate` | Members successfully checking in | ≥ 95% | ≥ 85% | < 85% |
| `attendance_accuracy` | Attendance records correctly computed | ≥ 99% | ≥ 95% | < 95% |
| `invoice_success_rate` | Invoices generated without error | ≥ 99% | ≥ 90% | < 90% |
| `booking_completion_time` | Time to complete a booking | ≤ 5s | ≤ 10s | > 10s |
| `payment_success_rate` | Payments processed without failure | ≥ 98% | ≥ 90% | < 90% |

### Example: HRMS Business KPIs

| KPI | Meaning | Healthy | Degraded | Unhealthy |
|-----|---------|---------|----------|-----------|
| `payroll_processing_accuracy` | Payroll computed correctly | ≥ 99.9% | ≥ 99% | < 99% |
| `leave_approval_latency` | Time from request to approval | ≤ 2h | ≤ 8h | > 8h |
| `attendance_sync_success` | Biometric sync completion | ≥ 99% | ≥ 95% | < 95% |

---

## Future AI Contracts

### `POST /api/v1/ai/analyze/incident/{id}`

**Purpose:** Request AI root cause analysis for a specific incident.

**Request Body:**

```json
{
  "analysis_type": "root_cause",
  "context_window": "1h",
  "include_dependencies": true
}
```

**Response (202 Accepted — Analysis is async):**

```json
{
  "data": {
    "analysis_id": "anl_001",
    "status": "processing",
    "estimated_completion_seconds": 15,
    "poll_url": "/api/v1/ai/analyses/anl_001"
  }
}
```

### `GET /api/v1/ai/analyses/{id}`

**Response (200 — When complete):**

```json
{
  "data": {
    "id": "anl_001",
    "incident_id": "INC-GF-042",
    "analysis_type": "root_cause",
    "status": "completed",
    "confidence": 0.87,
    "findings": {
      "probable_root_cause": "Missing index on bookings.created_at column causing sequential scans under load",
      "evidence": [
        "P95 latency spike correlates with daily peak traffic at 10:00",
        "PostgreSQL active connections doubled during the period",
        "No deployment or config change in the 24h prior",
        "Similar incident INC-GF-035 resolved with index addition"
      ],
      "affected_components": [
        { "id": "comp_pg01", "name": "postgresql", "role": "root_cause" },
        { "id": "comp_api01", "name": "backend-api", "role": "affected" }
      ],
      "dependency_chain": "backend-api → postgresql (query latency propagation)"
    },
    "recommendations": [
      {
        "priority": "high",
        "action": "Add index on bookings(created_at, status)",
        "reasoning": "Sequential scan on 2M+ rows table. Index reduces query time from ~800ms to ~3ms",
        "risk": "low",
        "estimated_impact": "Resolves 95% of latency spikes"
      },
      {
        "priority": "medium",
        "action": "Add connection pool tuning (reduce max_connections, increase pool_size)",
        "reasoning": "Connection saturation is a secondary symptom",
        "risk": "low",
        "estimated_impact": "Prevents connection exhaustion during peak"
      }
    ],
    "similar_incidents": [
      { "id": "INC-GF-035", "title": "Slow bookings API", "resolution": "Added index on users.email", "similarity": 0.82 }
    ],
    "computed_at": "2026-05-29T10:11:15Z",
    "model_version": "optrion-rca-v1",
    "provider": "gemini"
  }
}
```

---

### `GET /api/v1/ai/recommendations`

**Purpose:** List proactive AI recommendations for the tenant.

**Response (200):**

```json
{
  "data": [
    {
      "id": "rec_001",
      "type": "add_health_check",
      "priority": "medium",
      "title": "Add disk usage monitoring to postgresql",
      "description": "Your postgresql component has no disk usage monitoring. 67% of database incidents involve disk pressure. Adding this check enables earlier detection.",
      "component_id": "comp_pg01",
      "confidence": 0.79,
      "evidence": "Based on incident patterns from similar database components",
      "status": "pending",
      "created_at": "2026-05-29T08:00:00Z",
      "actions": {
        "accept_url": "/api/v1/ai/recommendations/rec_001/accept",
        "dismiss_url": "/api/v1/ai/recommendations/rec_001/dismiss"
      }
    },
    {
      "id": "rec_002",
      "type": "adjust_threshold",
      "priority": "low",
      "title": "Lower latency threshold on backend-api",
      "description": "Your latency threshold is 500ms but P95 baseline is 220ms. A threshold at 350ms would detect degradation 5 minutes earlier.",
      "component_id": "comp_api01",
      "confidence": 0.85,
      "status": "pending",
      "created_at": "2026-05-29T08:00:00Z"
    }
  ]
}
```

---

### `GET /api/v1/ai/predictions`

**Purpose:** View AI-generated predictions about future system behavior.

**Response (200):**

```json
{
  "data": [
    {
      "id": "pred_001",
      "type": "degradation",
      "component_id": "comp_pg01",
      "component_name": "postgresql",
      "title": "Predicted degradation in 48 hours",
      "description": "Based on current growth rate of active connections (+12/day), postgresql will reach connection saturation (100/100) by 2026-05-31T10:00:00Z",
      "confidence": 0.72,
      "predicted_for": "2026-05-31T10:00:00Z",
      "recommended_action": "Increase max_connections or implement connection pooling",
      "status": "pending",
      "created_at": "2026-05-29T08:00:00Z"
    }
  ]
}
```

---

### `POST /api/v1/ai/recommendations/{id}/accept`

### `POST /api/v1/ai/recommendations/{id}/dismiss`

**Purpose:** Feedback loop — tell the AI whether recommendations were useful.

```json
{
  "reason": "Already handled by our DBA"
}
```

---

## Error Model

### Standard Error Response

Every error response follows this structure:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Request validation failed",
    "details": [
      {
        "field": "condition.threshold",
        "issue": "must be a positive number",
        "value": -5
      },
      {
        "field": "channel_ids",
        "issue": "must contain at least one channel",
        "value": []
      }
    ],
    "request_id": "req_err001",
    "documentation_url": "https://docs.optrion.io/errors/VALIDATION_FAILED"
  }
}
```

### Error Code Catalog

| HTTP Status | Error Code | Meaning |
|-------------|-----------|---------|
| 400 | `BAD_REQUEST` | Malformed JSON, invalid content-type |
| 400 | `VALIDATION_FAILED` | Request body fails validation rules |
| 401 | `AUTHENTICATION_REQUIRED` | No API key provided |
| 401 | `INVALID_API_KEY` | API key is invalid or expired |
| 403 | `INSUFFICIENT_SCOPE` | API key lacks required scope for this operation |
| 403 | `TENANT_SUSPENDED` | Tenant account is suspended |
| 404 | `RESOURCE_NOT_FOUND` | The requested resource does not exist (in this tenant) |
| 409 | `CONFLICT` | Resource already exists (duplicate creation) |
| 409 | `INVALID_STATE_TRANSITION` | Operation not allowed in current state (e.g., resolve an already-closed incident) |
| 422 | `UNPROCESSABLE_ENTITY` | Semantically invalid (e.g., start_time after end_time) |
| 422 | `COMPONENT_NOT_REGISTERED` | Ingestion references unknown component (when auto-register is off) |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests. Respect `Retry-After` header |
| 429 | `QUOTA_EXCEEDED` | Tenant plan limit reached (e.g., max components) |
| 500 | `INTERNAL_ERROR` | Unexpected server error. Includes request_id for support |
| 502 | `UPSTREAM_ERROR` | Dependency failure (e.g., notification channel unreachable) |
| 503 | `SERVICE_UNAVAILABLE` | Server is temporarily unable to handle requests |

### Error Design Rules

| Rule | Rationale |
|------|-----------|
| Never expose internal implementation details | "PostgreSQL connection timeout" → `INTERNAL_ERROR` |
| Always include `request_id` | Enables support debugging without sharing sensitive data |
| Never include stack traces in production | Security risk |
| Include `documentation_url` for all codes | Self-service debugging |
| Field-level errors for validation failures | SDK can map errors to specific input fields |
| Use `Retry-After` header for 429 and 503 | SDKs can implement automatic retry |
| `404` for wrong-tenant access (not `403`) | Don't reveal resource existence to unauthorized tenants |

### Error Response by HTTP Method

| Method | Success Code | Error Behavior |
|--------|-------------|---------------|
| GET | 200 | 404 if not found |
| POST (create) | 201 | 409 if duplicate |
| POST (action) | 200 | 409 if invalid state |
| POST (ingest) | 202 | 422 if unprocessable |
| PATCH | 200 | 404 if not found, 409 if conflict |
| DELETE | 204 | 404 if not found |

---

## Risks & Recommendations

### Breaking Change Risks

| Risk | Scenario | Mitigation |
|------|----------|-----------|
| Adding required fields to ingestion | Future metric types need new required fields | All new fields must be optional. Default values in backend |
| Changing health score formula | Score interpretation changes for existing dashboards | Version the scoring algorithm. Allow tenants to pin a version |
| Removing deprecated fields | Older SDK versions break | 12-month deprecation window. Monitor usage of deprecated fields before removal |
| Event schema changes | Webhook consumers break | Version field in events. Consumers can filter by version |

### Scaling Risks

| Risk | Breaking Point | Mitigation |
|------|---------------|-----------|
| Ingestion burst from SDK flush | 1000 tenants × SDK flush every 10s = 100 req/s sustained | Accept 202, process async. Batch internally. Never sync-process ingestion |
| Dashboard polling storm | 1000 tenants × 5 dashboard tabs × 10s poll = 500 req/s | Server-side caching (10s TTL). Phase 2: WebSocket push eliminates polling |
| Analytics query complexity | 30-day trend across 50 components = large scan | Pre-computed rollups. Never query raw check_results for analytics |
| Webhook delivery at scale | 1000 tenants × 10 incidents/day × 3 webhooks = 30K deliveries/day | Background delivery queue. Retry with backoff. Circuit breaker per endpoint |

### Security Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| API key leaked in client-side JavaScript | Full API access for attacker | Provide scoped `ingest`-only keys for frontend SDKs. Never use `admin` keys in client code |
| Webhook payload contains sensitive data | Data exposure if webhook URL is compromised | Minimal payload in webhook. Full data fetched via authenticated API call |
| SSRF via webhook URL registration | Attacker registers internal URL as webhook | Validate webhook URLs against deny-list. Verify with challenge-response |
| Rate limit bypass via multiple keys | One tenant creates many keys to multiply rate limits | Rate limiting is per-tenant (not per-key) |
| Replay attacks on webhook | Captured webhook payload replayed | Timestamp + signature verification. Reject events older than 5 minutes |

### SDK Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| SDK buffer memory leak | Host app OOM if OPTRION is down for extended period | Hard buffer cap (1000 items default). Drop oldest when full |
| SDK thread/goroutine leak | Host app resource exhaustion | Clean shutdown API. Timeout on all network calls |
| SDK blocks host app hot path | Latency added to customer's critical paths | All SDK operations are async. Fire-and-forget with buffering |
| SDK version fragmentation | Old SDKs sending deprecated formats | SDK version sent in `User-Agent` header. Monitor version distribution. Emit deprecation warnings |
| SDK initialization failure crashes host | Customer app fails to start because OPTRION is unreachable | `Init()` must never fail fatally. Degrade gracefully — no-op mode if OPTRION unreachable |

### Recommendations

| Priority | Recommendation |
|----------|---------------|
| **P0** | Implement rate limiting from Day 1. Per-tenant, not per-key. Return proper 429 with `Retry-After` |
| **P0** | Never return `403` for wrong-tenant access. Always `404`. Prevents resource enumeration |
| **P0** | Validate all ingestion payloads strictly. Reject oversized JSONB, invalid timestamps, negative metrics |
| **P1** | Add `Idempotency-Key` support on all POST endpoints from launch. SDKs send it automatically |
| **P1** | Version event schemas from Day 1. Every event payload includes `"version": "1.0"` |
| **P1** | SDK `User-Agent` must include version: `optrion-go-sdk/1.0.0`. Track in access logs |
| **P2** | Add `Prefer: return=minimal` header support for ingestion (skip response body to reduce bandwidth) |
| **P2** | Add request compression support (`Content-Encoding: gzip`) for batch ingestion |
| **P2** | Add API usage analytics (requests per endpoint per tenant per day) for future billing |

### Complete API Endpoint Summary

| Method | Path | Purpose | Auth Scope |
|--------|------|---------|-----------|
| POST | `/api/v1/ingest` | Single-component metric + health ingestion | `ingest` |
| POST | `/api/v1/ingest/batch` | Multi-component batch ingestion | `ingest` |
| POST | `/api/v1/ingest/event` | Business event ingestion | `ingest` |
| POST | `/api/v1/metrics/business` | Business KPI ingestion | `ingest` |
| POST | `/api/v1/metrics/business/definitions` | Define KPI schema | `manage` |
| GET | `/api/v1/health/summary` | Tenant health overview | `read` |
| GET | `/api/v1/health/components` | List components with health | `read` |
| GET | `/api/v1/health/components/{id}` | Component detail | `read` |
| GET | `/api/v1/health/components/{id}/history` | Score time-series | `read` |
| GET | `/api/v1/health/scores` | Health scores by subject | `read` |
| GET | `/api/v1/incidents` | List incidents | `read` |
| GET | `/api/v1/incidents/{id}` | Incident detail + timeline | `read` |
| POST | `/api/v1/incidents/{id}/acknowledge` | Acknowledge | `manage` |
| POST | `/api/v1/incidents/{id}/resolve` | Resolve | `manage` |
| POST | `/api/v1/incidents/{id}/close` | Close | `manage` |
| POST | `/api/v1/incidents/{id}/note` | Add timeline note | `manage` |
| GET | `/api/v1/alerts/rules` | List alert rules | `read` |
| POST | `/api/v1/alerts/rules` | Create rule | `manage` |
| PATCH | `/api/v1/alerts/rules/{id}` | Update rule | `manage` |
| DELETE | `/api/v1/alerts/rules/{id}` | Delete rule | `manage` |
| POST | `/api/v1/alerts/rules/{id}/mute` | Mute rule | `manage` |
| POST | `/api/v1/alerts/rules/{id}/unmute` | Unmute rule | `manage` |
| GET | `/api/v1/alerts/history` | Alert history | `read` |
| GET | `/api/v1/alerts/channels` | List channels | `read` |
| POST | `/api/v1/alerts/channels` | Create channel | `manage` |
| POST | `/api/v1/alerts/channels/{id}/test` | Test channel | `manage` |
| POST | `/api/v1/maintenance` | Create maintenance window | `manage` |
| GET | `/api/v1/maintenance` | List windows | `read` |
| DELETE | `/api/v1/maintenance/{id}` | Cancel window | `manage` |
| GET | `/api/v1/analytics/uptime` | Uptime stats | `read` |
| GET | `/api/v1/analytics/trends` | Metric trends | `read` |
| GET | `/api/v1/analytics/reliability` | MTTR/MTBF stats | `read` |
| GET | `/api/v1/analytics/incidents/stats` | Incident statistics | `read` |
| POST | `/api/v1/ai/analyze/incident/{id}` | Request AI analysis | `manage` |
| GET | `/api/v1/ai/analyses/{id}` | Get analysis result | `read` |
| GET | `/api/v1/ai/recommendations` | List recommendations | `read` |
| POST | `/api/v1/ai/recommendations/{id}/accept` | Accept recommendation | `manage` |
| POST | `/api/v1/ai/recommendations/{id}/dismiss` | Dismiss recommendation | `manage` |
| GET | `/api/v1/ai/predictions` | List predictions | `read` |

**Total: 38 endpoints (Phase 1: 28, Future AI: 6, Future analytics: 4)**

---

*End of API & Contract Architecture*
