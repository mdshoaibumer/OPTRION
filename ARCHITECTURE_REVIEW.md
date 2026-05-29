# OPTRION — Architecture Review

**Reviewer:** Principal Engineer Review  
**Date:** 2026-05-29  
**Status:** Pre-Development Architecture Review  
**Verdict:** Conditionally Approved with Critical Corrections Required

---

## Executive Summary

OPTRION positions itself as an "Engineering Intelligence Platform" — not a monitoring tool. This distinction is **strategically correct** but **architecturally dangerous** if not enforced from Day 1 in the domain model.

The biggest risk is not technical — it is **scope collapse**. The gap between "health monitoring with alerting" (Phase 1) and "predictive intelligence with auto-remediation" (Future) is enormous. If the foundational domain model does not account for the future, you will rewrite the core within 12 months.

**Key findings:**

1. The multi-tenant model needs a third axis: Tenant → Product → Environment → **Service**
2. "Health Scoring" is a derived domain, not a core domain — this changes where logic lives
3. Event architecture must be internal-first before external (Telegram, webhooks)
4. The first customer (GymFlow Track) will distort your domain if you let it
5. Redis is premature for Phase 1 — PostgreSQL handles the MVP load

---

## Architecture Review

### What Is Right

| Decision | Assessment |
|----------|-----------|
| Go backend | Correct. Low allocation, excellent concurrency, strong for long-running health checks |
| PostgreSQL | Correct. JSONB for flexible telemetry, strong partitioning for time-series-like data |
| Clean + Hexagonal Architecture | Correct. Essential for a platform that must support multiple integration types |
| Multi-tenant from Day 1 | Correct. Retrofitting tenancy is a rewrite |
| Next.js for embedded dashboard | Acceptable. Embeddable via iframe/web component is viable |
| Telegram for alerts | Pragmatic for Phase 1. Correct |

### What Is Wrong

| Decision | Problem |
|----------|---------|
| Redis in Phase 1 | Premature. PostgreSQL with materialized views handles MVP scoring. Redis adds operational overhead with zero benefit at <10 tenants |
| "Health Monitoring" as a single concept | Too broad. "Health Check Execution" and "Health State Management" are two separate subdomains |
| No explicit data retention strategy | Time-series health data grows unbounded. Without TTL/partitioning strategy, PostgreSQL will degrade within 3 months of production use |
| No explicit ingestion model | Are health checks pull-based (OPTRION polls targets), push-based (targets report to OPTRION), or both? This is a foundational architectural decision |
| Gemini as "future AI" | Coupling to a single LLM provider is a design error. The AI integration point should be provider-agnostic |

### What Is Risky

| Risk | Impact | Mitigation |
|------|--------|-----------|
| First customer shapes the platform | GymFlow Track's needs become hardcoded assumptions | Strict bounded context separation; customer-specific logic lives in an adapter, never in the core domain |
| "Intelligence" without data volume | AI/ML requires months of historical data. Premature AI integration wastes effort | Ensure data model supports historical queries from Day 1, but defer AI to Phase 3+ |
| Alert fatigue | Without deduplication, correlation, and suppression — alerts become noise | Build alert lifecycle (open → acknowledged → resolved → suppressed) into the domain model now |
| Embedded dashboard security | Embedding a dashboard in customer apps creates XSS, CSRF, and token leakage vectors | Design the embed model as a first-class security boundary with scoped read-only tokens |

### What Will Break Later

1. **Health Score without context** — A score of 85/100 means nothing without baselines. You need historical context to make scores meaningful. Design for baseline comparison from Day 1.
2. **Flat alert model** — Alerts must support escalation, grouping, and silencing. A flat "fire and forget" Telegram message won't survive the second customer.
3. **No service catalog** — Without knowing what services exist, health checks are disconnected data points. A lightweight service registry is required in the core domain.
4. **Single-region assumption** — If any future customer has services in multiple regions, your health check execution model must account for check execution locality.

---

## Domain Discovery

### Core Domains (Your Competitive Advantage)

| Domain | Responsibility | Why Core |
|--------|---------------|----------|
| **Health Intelligence** | Compute, score, and interpret system health | This IS the product. The scoring algorithm, anomaly detection, and state interpretation are your IP |
| **Incident Detection** | Detect, classify, and lifecycle-manage incidents | Moving from "alert" to "incident" is the intelligence layer. This differentiates you from uptime checkers |

### Supporting Domains (Required, Not Differentiating)

| Domain | Responsibility |
|--------|---------------|
| **Health Check Execution** | Execute checks against targets (HTTP, TCP, DB, custom). Pull and/or push models |
| **Alerting & Notification** | Deliver notifications through channels (Telegram, email, webhook). Manages delivery, retries, preferences |
| **Tenant Management** | Onboarding, plans, quotas, API keys |
| **Service Registry** | Catalog of monitored services, their environments, dependencies |
| **Dashboard & Visualization** | Read-only projections of health state for embedded and standalone UIs |

### Generic Domains (Buy or Use Libraries)

| Domain | Recommendation |
|--------|---------------|
| **Authentication & Authorization** | Use existing libraries. Don't build an auth system |
| **Rate Limiting** | Middleware concern. Use proven libraries |
| **Scheduling** | Use proven cron/scheduler libraries for health check intervals |

---

## Bounded Contexts

```
┌─────────────────────────────────────────────────────────────────────┐
│                         OPTRION PLATFORM                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────┐  │
│  │  Tenant Context  │    │  Catalog Context │    │  Execution   │  │
│  │                  │    │                  │    │  Context     │  │
│  │  - Tenant        │    │  - Product       │    │              │  │
│  │  - Plan          │    │  - Environment   │    │  - Check     │  │
│  │  - APIKey        │    │  - Service       │    │  - Schedule  │  │
│  │  - User          │    │  - Endpoint      │    │  - Result    │  │
│  │                  │    │  - Dependency     │    │  - Probe     │  │
│  └──────────────────┘    └──────────────────┘    └──────────────┘  │
│                                                                     │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────┐  │
│  │  Intelligence    │    │  Incident        │    │  Notification│  │
│  │  Context         │    │  Context         │    │  Context     │  │
│  │                  │    │                  │    │              │  │
│  │  - HealthScore   │    │  - Incident      │    │  - Channel   │  │
│  │  - Baseline      │    │  - Alert         │    │  - Rule      │  │
│  │  - Trend         │    │  - Escalation    │    │  - Delivery  │  │
│  │  - Anomaly       │    │  - Timeline      │    │  - Template  │  │
│  │                  │    │                  │    │              │  │
│  └──────────────────┘    └──────────────────┘    └──────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Context Relationships

| Upstream | Downstream | Relationship |
|----------|-----------|-------------|
| Tenant Context | All Contexts | Every entity is tenant-scoped |
| Catalog Context | Execution Context | Execution needs to know what to check |
| Execution Context | Intelligence Context | Intelligence scores raw check results |
| Intelligence Context | Incident Context | Anomalies trigger incident detection |
| Incident Context | Notification Context | Incidents trigger notifications |

### Anti-Corruption Layers Required

- Between **Catalog** and **Execution**: Execution should not depend on Catalog's internal model. Use a shared kernel or published language.
- Between **Incident** and **Notification**: Notification must not know incident internals. It receives a "notify" command with channel, template, and payload.

---

## Domain Events

### Execution Context Events

| Event | Trigger | Consumers |
|-------|---------|-----------|
| `HealthCheckExecuted` | A health check completes (success or failure) | Intelligence Context |
| `HealthCheckFailed` | A check returns non-healthy status | Intelligence Context, Incident Context |
| `HealthCheckRecovered` | A previously failing check returns healthy | Intelligence Context, Incident Context |
| `CheckTimeoutExceeded` | Check did not complete within SLA | Incident Context |

### Intelligence Context Events

| Event | Trigger | Consumers |
|-------|---------|-----------|
| `HealthScoreComputed` | Periodic score recalculation | Dashboard Context |
| `AnomalyDetected` | Score deviates from baseline beyond threshold | Incident Context |
| `BaselineUpdated` | Rolling baseline recalculated | Internal (Intelligence) |
| `DegradationDetected` | Gradual health decline detected over time window | Incident Context |

### Incident Context Events

| Event | Trigger | Consumers |
|-------|---------|-----------|
| `IncidentOpened` | Anomaly or failure threshold breached | Notification Context |
| `IncidentAcknowledged` | Operator acknowledges | Notification Context (stop escalation) |
| `IncidentResolved` | Health recovered and stable | Notification Context, Intelligence Context |
| `IncidentEscalated` | Acknowledgement SLA breached | Notification Context |
| `AlertFired` | Immediate notification required | Notification Context |

### Notification Context Events

| Event | Trigger | Consumers |
|-------|---------|-----------|
| `NotificationSent` | Message delivered to channel | Audit log |
| `NotificationFailed` | Delivery failed after retries | Incident Context (meta-incident) |

### Tenant Context Events

| Event | Trigger | Consumers |
|-------|---------|-----------|
| `TenantOnboarded` | New tenant registered | All Contexts (provision) |
| `TenantSuspended` | Plan expired or violation | Execution Context (pause checks) |
| `PlanChanged` | Tenant upgrades/downgrades | Execution Context (adjust quotas) |

---

## Aggregates

### Tenant Context

| Aggregate | Root Entity | Child Entities / Value Objects |
|-----------|-------------|-------------------------------|
| **Tenant** | Tenant | Plan, Quota, BillingInfo |
| **APICredential** | APIKey | Scope, Expiry |

### Catalog Context

| Aggregate | Root Entity | Child Entities / Value Objects |
|-----------|-------------|-------------------------------|
| **Product** | Product | — |
| **Environment** | Environment | EnvironmentType (prod/staging/dev) |
| **Service** | Service | Endpoint[], Dependency[], HealthCheckConfig |

### Execution Context

| Aggregate | Root Entity | Child Entities / Value Objects |
|-----------|-------------|-------------------------------|
| **HealthCheck** | HealthCheck | Schedule, CheckType, Target, Timeout, RetryPolicy |
| **CheckResult** | CheckResult | Latency, StatusCode, ResponseBody (truncated), Timestamp |

### Intelligence Context

| Aggregate | Root Entity | Child Entities / Value Objects |
|-----------|-------------|-------------------------------|
| **ServiceHealth** | ServiceHealth | HealthScore, ComponentScores[], Baseline, Trend |

### Incident Context

| Aggregate | Root Entity | Child Entities / Value Objects |
|-----------|-------------|-------------------------------|
| **Incident** | Incident | Alert[], TimelineEntry[], Severity, Status, AssignedTo |

### Notification Context

| Aggregate | Root Entity | Child Entities / Value Objects |
|-----------|-------------|-------------------------------|
| **NotificationRule** | NotificationRule | Channel, Condition, Template, CooldownPeriod |
| **DeliveryAttempt** | DeliveryAttempt | Status, Timestamp, ErrorMessage |

---

## Entity Relationships

```
Tenant (1) ──────────── (*) Product
Product (1) ─────────── (*) Environment
Environment (1) ──────── (*) Service
Service (1) ─────────── (*) HealthCheck
HealthCheck (1) ──────── (*) CheckResult
Service (1) ─────────── (1) ServiceHealth
ServiceHealth (1) ────── (*) HealthScore (time-series)
Service (*) ─────────── (*) Incident
Incident (1) ────────── (*) Alert
Incident (1) ────────── (*) TimelineEntry
Tenant (1) ──────────── (*) NotificationRule
NotificationRule (1) ── (*) DeliveryAttempt
```

---

## Multi-Tenant Strategy

### Tenancy Model: **Shared Database, Shared Schema, Row-Level Isolation**

**Rationale:** At MVP scale (<50 tenants), schema-per-tenant adds operational complexity with zero benefit. Row-level isolation with `tenant_id` foreign keys on every table is sufficient.

### Hierarchy

```
Tenant
└── Product (e.g., "GymFlow Track")
    └── Environment (e.g., production, staging, development)
        └── Service (e.g., "Booking API", "Payment Service")
            └── Endpoint (e.g., "/health", "/api/v1/status")
```

### Tenant Isolation Guarantees

| Layer | Mechanism |
|-------|-----------|
| Database | Every query includes `WHERE tenant_id = ?`. Enforced via repository pattern, never optional |
| API | Tenant derived from API key or JWT. Never from request body or path parameter alone |
| Execution | Health checks tagged with tenant_id. Check results cannot cross tenant boundaries |
| Dashboard | Embed tokens scoped to specific tenant + product + environment |
| Rate Limiting | Per-tenant rate limits. One tenant cannot starve others |
| Data Retention | Per-tenant retention policies (tied to plan) |

### Critical Rule

> **Every database table (except the `tenants` table itself) MUST have a `tenant_id` column. No exceptions. No "shared" tables. Enforce this with a linter or migration check.**

---

## Environment Strategy

### Environment Model

```
Tenant: "GymFlow"
├── Product: "GymFlow Track"
│   ├── Environment: production
│   │   ├── Service: booking-api
│   │   ├── Service: payment-service
│   │   └── Service: notification-service
│   ├── Environment: staging
│   │   ├── Service: booking-api
│   │   └── Service: payment-service
│   └── Environment: development
│       └── Service: booking-api
└── Product: "GymFlow Analytics" (future)
    └── Environment: production
        └── Service: analytics-api
```

### Environment-Specific Behavior

| Aspect | Production | Staging | Development |
|--------|-----------|---------|-------------|
| Check Frequency | High (every 30s) | Medium (every 2min) | Low (every 5min) |
| Alerting | Full escalation | Notification only | Silent (log only) |
| Retention | 90 days | 30 days | 7 days |
| SLA Tracking | Yes | No | No |
| Incident Creation | Automatic | Manual review | Disabled |

### Key Design Decision

Environments are **not** separate deployments of OPTRION. They are a **logical grouping** within a single OPTRION instance. This keeps operations simple while providing logical isolation.

---

## Event Architecture

### Phase 1: In-Process Event Bus

**Do NOT use Kafka, RabbitMQ, or NATS in Phase 1.**

Use an in-process event bus (Go channels + pub/sub pattern) with:
- Synchronous event dispatch within a single request for consistency
- Asynchronous event dispatch for side effects (notifications)
- PostgreSQL as the event store (outbox pattern for reliability)

### Outbox Pattern (Critical for Reliability)

```
1. Business transaction writes entity + event to same DB transaction
2. Background worker polls outbox table
3. Worker dispatches event to handlers
4. Worker marks event as processed
```

This guarantees at-least-once delivery without external message brokers.

### Event Flow

```
[Health Check Executor]
        │
        ▼ (HealthCheckExecuted event)
[Intelligence Service]
        │
        ▼ (AnomalyDetected event)
[Incident Manager]
        │
        ▼ (IncidentOpened event)
[Notification Service]
        │
        ▼ (Telegram / Webhook)
[External Channel]
```

### Future Migration Path (Phase 3+)

When event volume exceeds what PostgreSQL outbox can handle (~10K events/sec), migrate to:
- NATS JetStream (preferred for Go ecosystem) or
- Apache Kafka (if customer requires enterprise compliance)

The in-process bus interface stays the same; only the implementation changes. **This is why hexagonal architecture matters.**

---

## Service Boundaries

### Phase 1: Modular Monolith

**Do NOT build microservices in Phase 1.**

Build a single Go binary with clear module boundaries:

```
optrion/
├── cmd/
│   └── optrion/          # Single binary entry point
├── internal/
│   ├── tenant/           # Tenant bounded context
│   ├── catalog/          # Service catalog bounded context
│   ├── execution/        # Health check execution bounded context
│   ├── intelligence/     # Scoring and analysis bounded context
│   ├── incident/         # Incident management bounded context
│   ├── notification/     # Alert delivery bounded context
│   └── shared/           # Shared kernel (tenant ID, event bus interface)
├── pkg/
│   └── sdk/              # Public SDK for push-based integrations
└── api/
    └── rest/             # HTTP handlers (thin layer, delegates to domains)
```

### Module Communication Rules

1. Modules communicate **only** through domain events or defined interfaces
2. No module imports another module's internal types directly
3. Shared types live in `shared/` kernel (TenantID, EventBus, Clock)
4. Database access is per-module (each module owns its tables)

### Future Extraction Path

When any module's team size or scale justifies independence:
1. Extract module into separate service
2. Replace in-process event bus with network event bus
3. Replace in-process function calls with gRPC
4. Each service gets its own database (if needed)

---

## Future AI Integration Points

### Where AI Adds Value (Future Phases)

| Integration Point | AI Capability | Data Required | Phase |
|-------------------|--------------|---------------|-------|
| **Anomaly Detection** | Unsupervised learning on health metrics | 3+ months of check results per service | Phase 3 |
| **Root Cause Analysis** | Correlation of incidents across services using dependency graph | Service dependency map + incident history | Phase 3 |
| **Recommendation Engine** | "You should add a health check for X" based on similar services | Cross-tenant anonymized patterns | Phase 4 |
| **Predictive Intelligence** | "This service will degrade in 2 hours" | 6+ months of time-series data with seasonal patterns | Phase 4 |
| **Auto Remediation** | Execute runbooks automatically | Runbook definitions + confidence scoring | Phase 5 |
| **Natural Language Queries** | "Why was booking-api slow yesterday?" | All historical data + LLM | Phase 4 |

### AI Architecture Principle

```
[Domain Event Stream] → [AI Adapter Port] → [AI Provider Implementation]
                                                    ├── Gemini
                                                    ├── OpenAI
                                                    └── Local Model
```

**Never couple domain logic to a specific AI provider.** The AI integration is a **port** in hexagonal architecture, with provider-specific **adapters**.

### Data Pipeline Requirement

AI models need clean, structured, labeled data. From Day 1:
- Store raw check results (not just pass/fail)
- Store latency percentiles (p50, p95, p99)
- Store response size trends
- Tag data with environment, service, time

This data becomes the training set for future ML models.

---

## Security Risks

### Critical (Must Address Before Launch)

| Risk | Description | Mitigation |
|------|-------------|-----------|
| **Tenant Data Leakage** | Missing `tenant_id` filter in any query exposes another tenant's data | Repository pattern enforces tenant scoping. Integration tests verify isolation. Never trust application code alone — add PostgreSQL Row-Level Security as defense-in-depth |
| **API Key Exposure** | API keys stored in plaintext or leaked in logs | Hash API keys (SHA-256). Store only hash. Display prefix only. Never log full keys |
| **SSRF via Health Checks** | Attacker registers health check pointing to internal infrastructure (169.254.169.254, localhost, internal services) | Validate and deny-list private IP ranges, link-local addresses, and localhost. DNS rebinding protection |
| **Embedded Dashboard XSS** | Dashboard embedded in customer sites can be exploited for XSS | CSP headers, sandboxed iframe, scoped read-only tokens with short TTL |
| **Stored Credential Exposure** | Health checks may require auth tokens for target services | Encrypt at rest with per-tenant keys. Use envelope encryption. Never return credentials in API responses |

### High (Must Address Before Multi-Tenant)

| Risk | Description | Mitigation |
|------|-------------|-----------|
| **Noisy Neighbor** | One tenant's aggressive check schedule starves others | Per-tenant execution quotas enforced at scheduler level |
| **Denial of Wallet** | Attacker creates thousands of health checks to inflate resource usage | Hard limits per plan. Rate limit check creation API |
| **Event Injection** | If push-based ingestion is supported, malicious payloads in check results | Validate and sanitize all ingested data. Size limits on response bodies stored |

---

## Scalability Risks

### Data Growth (Most Critical)

| Data Type | Growth Rate (per service) | 1 Year at 100 Services | Mitigation |
|-----------|--------------------------|------------------------|-----------|
| Check Results | 2,880/day (30s interval) | ~105M rows | Time-partitioned tables (daily/weekly). Auto-drop old partitions per retention policy |
| Health Scores | 96/day (15min recalc) | ~3.5M rows | Materialized views for dashboards. Archive to cold storage after 90 days |
| Events | ~100/day | ~3.6M rows | Outbox table with TTL. Processed events purged after 7 days |
| Incidents | ~5/day | ~182K rows | Minimal growth concern |

### Compute Scaling

| Component | Bottleneck | Scaling Strategy |
|-----------|-----------|-----------------|
| Health Check Execution | Network I/O, goroutine limits | Worker pool pattern. Horizontal scale by adding workers |
| Score Computation | CPU for aggregation | Materialized views. Background workers. Eventually CQRS |
| API Serving | Connection count | Standard horizontal scaling. Stateless handlers |
| Event Processing | Outbox polling frequency | Batch processing. Partitioned outbox per tenant |

### Key Scalability Decision

**PostgreSQL partitioning strategy MUST be defined before the first migration runs.** Partitioning by `(tenant_id, created_at)` for check_results. This cannot be retrofitted without downtime.

---

## Risks Summary

### Ranked by Impact × Likelihood

| # | Risk | Impact | Likelihood | Priority |
|---|------|--------|-----------|----------|
| 1 | Tenant data leakage | Critical | Medium | P0 |
| 2 | Unbounded data growth crashes PostgreSQL | Critical | High | P0 |
| 3 | SSRF via health check targets | High | High | P0 |
| 4 | Scope creep from first customer | High | High | P1 |
| 5 | Alert fatigue without deduplication | Medium | High | P1 |
| 6 | No baseline = meaningless health scores | Medium | High | P1 |
| 7 | Premature Redis dependency | Low | Medium | P2 |
| 8 | Over-engineering event system | Medium | Medium | P2 |

---

## Recommendations

### Do Immediately (Before First Line of Code)

1. **Define the ingestion model** — Pull-based (OPTRION polls), push-based (SDK reports), or both. This shapes the entire execution domain.

2. **Design the check_results partitioning strategy** — Monthly partitions by tenant_id range. Define retention policy per plan tier.

3. **Write the tenant isolation contract** — Every repository method takes `TenantID` as first parameter. No exceptions. Write a linter to enforce this.

4. **Define "healthy"** — What does healthy mean? HTTP 200? Response time < 500ms? Custom expression? This is a product decision that shapes the domain model.

5. **Choose pull vs. push for Phase 1** — Recommendation: Pull-only for Phase 1 (OPTRION executes HTTP checks). Push via SDK in Phase 2.

### Do During Development

6. **Implement the outbox pattern from Day 1** — Even if you only have 3 events, the pattern prevents data loss and enables future extraction.

7. **Build alert deduplication into the incident aggregate** — An incident groups related alerts. Never fire duplicate notifications for the same root issue.

8. **Design health scores as time-series** — Store every score computation with timestamp. This becomes your baseline for anomaly detection.

9. **Add SSRF protection to the check executor** — Deny-list RFC 1918 ranges, link-local, loopback. Validate DNS resolution results, not just the hostname.

### Do NOT Do

10. **Do NOT add Redis until you have proven PostgreSQL cannot handle the load** — Materialized views and LISTEN/NOTIFY cover MVP caching and pub/sub needs.

11. **Do NOT build a plugin system** — Extensibility through well-defined ports/adapters is sufficient. A plugin system is a product, not a feature.

12. **Do NOT build multi-region check execution** — Single-region execution with a clean interface is enough. Multi-region is a Phase 3+ concern.

---

## MVP Scope (Phase 1)

### In Scope

| Feature | Bounded Context | Complexity |
|---------|----------------|-----------|
| Tenant registration + API key | Tenant | Low |
| Product + Environment CRUD | Catalog | Low |
| Service registration with endpoints | Catalog | Medium |
| HTTP health check configuration | Execution | Medium |
| Scheduled health check execution (pull-based) | Execution | Medium |
| Health score computation (simple weighted average) | Intelligence | Medium |
| Incident detection (threshold-based) | Incident | Medium |
| Telegram alerting | Notification | Low |
| REST API for all above | API | Medium |
| Embedded health dashboard (read-only) | Dashboard | Medium |

### MVP Health Score Formula (Keep It Simple)

```
ServiceHealthScore = weighted_average(
    availability_score,    # % of checks passing in window
    latency_score,         # p95 latency vs configured threshold
    consistency_score      # variance in response times
)
```

Do NOT build ML-based scoring in Phase 1. A deterministic, explainable formula is better for debugging and customer trust.

### MVP Incident Detection (Keep It Simple)

```
IF consecutive_failures >= threshold THEN open_incident
IF recovered AND stable_for >= cooldown THEN resolve_incident
```

Do NOT build correlation-based detection in Phase 1.

---

## What Should NOT Be Built Yet

| Feature | Why Not | When |
|---------|---------|------|
| Push-based SDK ingestion | Adds protocol design + auth complexity | Phase 2 |
| Root cause analysis | Requires dependency graph + correlation engine + data volume | Phase 3 |
| Predictive intelligence | Requires 6+ months of historical data + ML pipeline | Phase 4 |
| Auto remediation | Requires extremely high confidence + runbook system + liability model | Phase 5 |
| Multi-region check execution | Adds distributed coordination complexity | Phase 3 |
| Custom check types (TCP, gRPC, DB) | HTTP covers 90% of use cases. Add others when demanded | Phase 2 |
| User management / RBAC | Single API key per tenant is sufficient for MVP | Phase 2 |
| Billing / usage metering | Manual billing for first 5 customers | Phase 2 |
| Webhooks (outgoing) | Telegram is sufficient for Phase 1 | Phase 2 |
| Public status pages | Different product surface. Not MVP | Phase 3 |
| Engineering memory / knowledge base | Requires AI + structured data pipeline | Phase 4 |
| SLA tracking / SLO burn rates | Valuable but not MVP | Phase 2 |

---

## Future Roadmap

### Phase 1: Foundation (MVP)
- Pull-based HTTP health checks
- Deterministic health scoring
- Threshold-based incident detection
- Telegram alerts
- Embedded dashboard
- Single tenant proven, multi-tenant ready

### Phase 2: Platform
- Push-based SDK ingestion
- Additional check types (TCP, DNS, gRPC)
- Webhook notifications
- User management / RBAC
- SLO/SLA tracking
- Usage metering + billing

### Phase 3: Intelligence
- Service dependency mapping
- Cross-service incident correlation
- Root cause analysis (rule-based first, then ML)
- Multi-region check execution
- Historical trend analysis
- Baseline learning (statistical, not ML)

### Phase 4: AI-Powered
- LLM-powered root cause explanation
- Predictive degradation alerts
- Natural language queries
- Recommendation engine
- Cross-tenant pattern learning (anonymized)

### Phase 5: Autonomous
- Auto-remediation with runbooks
- Self-healing suggestions
- Capacity planning predictions
- Architecture recommendations

---

## Final Verdict

OPTRION has a **sound strategic vision** and a **pragmatic Phase 1 scope**. The key risks are:

1. **Letting the first customer distort the domain model** — enforce bounded context boundaries ruthlessly
2. **Ignoring data growth** — partition strategy must be decided before first migration
3. **Premature complexity** — Redis, Kafka, microservices, AI are all Phase 2+ concerns
4. **Security blind spots** — SSRF and tenant isolation are non-negotiable for a multi-tenant platform that executes network requests on behalf of customers

Build the modular monolith. Get the domain model right. Ship to one customer. Then evolve.

---

*End of Architecture Review*
