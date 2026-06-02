# OPTRION — Domain Model

**Author:** Principal Domain Architect  
**Date:** 2026-05-29 (Updated: 2026-06-02)  
**Version:** 2.0  
**Status:** Implementation Complete

---

## Domain Overview

OPTRION's domain model is designed around ten bounded contexts that collaborate through domain events. The model supports multi-tenancy at every level, assumes eventual consistency between contexts, and includes AI-powered root cause analysis and recommendation intelligence.

### Bounded Contexts (All Implemented)

| # | Context | Responsibility |
|---|---------|---------------|
| 1 | **Tenant** | Identity, API keys, plans, audit |
| 2 | **Health** | Monitoring, scoring, anomaly detection |
| 3 | **Incident** | Event-sourced lifecycle, deduplication, correlation |
| 4 | **Alert** | Rules, channels, escalation, Telegram |
| 5 | **AI** | Root cause analysis (Gemini, OpenAI, Anthropic, Ollama) |
| 6 | **Recommendation** | Evidence-based recommendations with safety validation |
| 7 | **Registration** | Bulk plug-and-play onboarding |
| 8 | **AutoDiscovery** | Auto-detect PostgreSQL, Redis, HTTP services |
| 9 | **Validation** | Integration validation |
| 10 | **Config** | YAML configuration loading for CLI |

### Design Principles Applied

1. **Aggregate boundaries enforce consistency** — Only data that MUST be consistent lives in the same aggregate
2. **Events cross boundaries, not entities** — Bounded contexts communicate through published events
3. **Tenant isolation is structural, not behavioral** — TenantID is part of every aggregate root's identity
4. **Time is a first-class concept** — Every metric, score, and event is timestamped with collection time AND processing time
5. **Immutability where possible** — Check results, metric snapshots, and events are append-only

---

## Entities

### 1. Tenant

| Aspect | Detail |
|--------|--------|
| **Purpose** | Root organizational unit. Represents a paying customer of OPTRION |
| **Responsibilities** | Owns all resources. Defines plan limits. Holds billing relationship. Scopes all data access |
| **Lifecycle** | Created → Active → Suspended → Terminated |
| **Relationships** | Has many Products. Has many NotificationChannels. Has many Users (future). Has one Plan |

### 2. Product

| Aspect | Detail |
|--------|--------|
| **Purpose** | A distinct software product or system owned by a tenant. The primary organizational grouping |
| **Responsibilities** | Groups environments. Defines product-level health aggregation boundary. Scopes dashboards |
| **Lifecycle** | Created → Active → Archived |
| **Relationships** | Belongs to Tenant. Has many Environments |

### 3. Environment

| Aspect | Detail |
|--------|--------|
| **Purpose** | A deployment context (production, staging, development) for a product |
| **Responsibilities** | Defines operational context. Controls alerting behavior (production alerts vs dev silence). Sets check frequency and retention policies |
| **Lifecycle** | Created → Active → Decommissioned |
| **Relationships** | Belongs to Product. Has many Components. Has one EnvironmentPolicy |

### 4. Component

| Aspect | Detail |
|--------|--------|
| **Purpose** | A monitorable unit within an environment — a database, cache, API, queue, service |
| **Responsibilities** | The atomic unit of health monitoring. Owns health checks. Produces metrics. Has a health score. Declares dependencies |
| **Lifecycle** | Registered → Active → Degraded → Unhealthy → Deregistered |
| **Relationships** | Belongs to Environment. Has many HealthChecks. Has many MetricSnapshots. Has one ComponentHealth. May depend on other Components |

### 5. HealthCheck

| Aspect | Detail |
|--------|--------|
| **Purpose** | A configured probe that tests a specific aspect of a component's health |
| **Responsibilities** | Defines what to check (target URL, port, query), how to check (HTTP, TCP, custom), when to check (schedule), and what constitutes success (assertion) |
| **Lifecycle** | Configured → Active → Paused → Disabled → Deleted |
| **Relationships** | Belongs to Component. Produces many CheckResults. Has one Schedule. Has one Assertion |

### 6. CheckResult

| Aspect | Detail |
|--------|--------|
| **Purpose** | The immutable outcome of a single health check execution |
| **Responsibilities** | Records success/failure, latency, response metadata, timestamp. Serves as raw input for health scoring and incident detection |
| **Lifecycle** | Created (immutable, append-only) → Expired (TTL-based purge) |
| **Relationships** | Belongs to HealthCheck. May trigger MetricSnapshot creation. May contribute to Incident detection |

### 7. MetricSnapshot

| Aspect | Detail |
|--------|--------|
| **Purpose** | A point-in-time measurement of a component's operational metric (latency, error rate, saturation, availability) |
| **Responsibilities** | Stores typed metric values. Provides time-series data for scoring, trending, and anomaly detection |
| **Lifecycle** | Collected (immutable) → Aggregated → Archived → Purged |
| **Relationships** | Belongs to Component. Has one MetricType. Contributes to HealthScore calculation |

### 8. ComponentHealth

| Aspect | Detail |
|--------|--------|
| **Purpose** | The current computed health state of a single component — the "live" view |
| **Responsibilities** | Maintains current health score, status, last check time, streak (consecutive failures/successes). Updated on every check result |
| **Lifecycle** | Initialized → Continuously updated (never deleted while component exists) |
| **Relationships** | Belongs to Component (1:1). Derived from CheckResults and MetricSnapshots |

### 9. HealthScore

| Aspect | Detail |
|--------|--------|
| **Purpose** | A historical record of computed health scores at various aggregation levels (component, environment, product) |
| **Responsibilities** | Records score value, computation timestamp, contributing factors, aggregation level. Enables trending and baseline calculation |
| **Lifecycle** | Computed (immutable) → Archived → Purged per retention |
| **Relationships** | References Component or Environment or Product (polymorphic via ScoreSubject). Used by Intelligence context for baseline comparison |

### 10. Incident

| Aspect | Detail |
|--------|--------|
| **Purpose** | A detected problem that requires attention. Groups related failures into a single actionable unit |
| **Responsibilities** | Tracks lifecycle of a problem from detection to resolution. Aggregates related events. Manages severity, assignment, and escalation |
| **Lifecycle** | Opened → Acknowledged → Investigating → Resolved → Closed |
| **Relationships** | Belongs to Component (primary). May span multiple Components (correlated). Has many IncidentEvents. Has many IncidentAlerts. May reference a RootCause (future) |

### 11. IncidentEvent

| Aspect | Detail |
|--------|--------|
| **Purpose** | An immutable record of something that happened during an incident's lifecycle |
| **Responsibilities** | Captures state transitions, annotations, escalations, and evidence. Forms the incident timeline |
| **Lifecycle** | Created (immutable, append-only) |
| **Relationships** | Belongs to Incident. Has one EventType. May reference an Actor (system or user) |

### 12. AlertRule

| Aspect | Detail |
|--------|--------|
| **Purpose** | Defines conditions under which notifications should be triggered |
| **Responsibilities** | Evaluates conditions against health state changes. Determines severity. References notification channels. Manages cooldown/deduplication |
| **Lifecycle** | Created → Active → Muted → Disabled → Deleted |
| **Relationships** | Belongs to Tenant. Scoped to Component or Environment or Product. References NotificationChannels. Produces AlertHistory entries |

### 13. AlertHistory

| Aspect | Detail |
|--------|--------|
| **Purpose** | Immutable record of an alert that was evaluated and/or fired |
| **Responsibilities** | Records what triggered, when, what was sent, to whom, and delivery status |
| **Lifecycle** | Created (immutable) → Purged per retention |
| **Relationships** | Belongs to AlertRule. References Incident (if incident-triggered). Has delivery status per channel |

### 14. NotificationChannel

| Aspect | Detail |
|--------|--------|
| **Purpose** | A configured delivery mechanism for alerts (Telegram, email, webhook, Slack) |
| **Responsibilities** | Stores channel-specific configuration (chat ID, webhook URL). Validates connectivity. Tracks delivery success rate |
| **Lifecycle** | Created → Verified → Active → Disabled → Deleted |
| **Relationships** | Belongs to Tenant. Referenced by many AlertRules |

### 15. ComponentDependency

| Aspect | Detail |
|--------|--------|
| **Purpose** | Declares that one component depends on another (e.g., Backend API depends on PostgreSQL) |
| **Responsibilities** | Enables dependency-aware health scoring. Supports future root cause analysis. Prevents false positive incidents when a dependency is the actual root cause |
| **Lifecycle** | Declared → Active → Removed |
| **Relationships** | References two Components (source depends on target). Used by Intelligence and Incident contexts |

### 16. EnvironmentPolicy

| Aspect | Detail |
|--------|--------|
| **Purpose** | Defines operational rules for an environment — check frequency, retention, alerting behavior, SLA expectations |
| **Responsibilities** | Controls how aggressively the platform monitors this environment. Determines whether incidents auto-create or require review |
| **Lifecycle** | Created with Environment → Updated as needs change |
| **Relationships** | Belongs to Environment (1:1) |

### 17. Baseline

| Aspect | Detail |
|--------|--------|
| **Purpose** | A statistical profile of "normal" behavior for a component over a defined time window |
| **Responsibilities** | Stores expected ranges for latency, error rate, availability. Used to detect anomalies (deviation from baseline = potential incident). Updated periodically |
| **Lifecycle** | Computing → Active → Stale → Recomputing |
| **Relationships** | Belongs to Component. Computed from MetricSnapshots. Used by Intelligence context for anomaly detection |

### 18. MaintenanceWindow

| Aspect | Detail |
|--------|--------|
| **Purpose** | A scheduled period during which alerts should be suppressed for a scope (component, environment, product) |
| **Responsibilities** | Prevents false positive alerting during deployments, planned maintenance. Recorded in incident timeline if an incident occurs during window |
| **Lifecycle** | Scheduled → Active → Completed → Cancelled |
| **Relationships** | Belongs to Tenant. Scoped to Component or Environment or Product |

### 19. AuditEntry

| Aspect | Detail |
|--------|--------|
| **Purpose** | Immutable record of significant actions taken within the platform (configuration changes, incident actions, rule modifications) |
| **Responsibilities** | Provides accountability and change tracking. Required for compliance in enterprise customers |
| **Lifecycle** | Created (immutable, append-only) → Archived |
| **Relationships** | References Actor, Action, Subject, Tenant |

---

## Aggregates

### Aggregate 1: Tenant Aggregate

**Root:** Tenant

**Contains:**
- Tenant (root)
- Plan (value object embedded)

**Boundary Justification:**  
Tenant is the top-level identity. Plan changes must be atomic with tenant state (you cannot have an active tenant without a plan). Products are NOT inside this aggregate because they have independent lifecycles and their own consistency needs.

**Invariants:**
- A tenant must always have exactly one active plan
- Tenant slug must be unique globally
- Tenant cannot be terminated if it has active products (soft-delete cascade required)

---

### Aggregate 2: Product Aggregate

**Root:** Product

**Contains:**
- Product (root)

**Boundary Justification:**  
Products are independently manageable. Creating/archiving a product does not require locking environments or components. Environments reference the product but are a separate aggregate for independent lifecycle management.

**Invariants:**
- Product name must be unique within tenant
- Product cannot be archived if it has active environments

---

### Aggregate 3: Environment Aggregate

**Root:** Environment

**Contains:**
- Environment (root)
- EnvironmentPolicy (embedded value object / child entity)

**Boundary Justification:**  
Environment and its policy must change atomically — you cannot have an environment without a policy. Components are NOT inside because they have independent, high-frequency state changes that should not require locking the environment.

**Invariants:**
- Environment name must be unique within product
- Environment must have exactly one policy
- Environment type must be one of: production, staging, development, custom

---

### Aggregate 4: Component Aggregate

**Root:** Component

**Contains:**
- Component (root)
- HealthCheck[] (child entities)
- ComponentHealth (child entity, 1:1)
- ComponentDependency[] (child entities)

**Boundary Justification:**  
This is the most critical aggregate. A component's health checks define what IS a component from a monitoring perspective. They must be consistent — you cannot have a component with zero health checks in "active" state. ComponentHealth is the derived current state and must be atomically consistent with the component.

**Why HealthChecks are inside:** Adding/removing a health check changes the health score formula. This must be atomic.

**Why CheckResults are OUTSIDE:** Results are append-only, high-volume, and don't affect component consistency. They are a separate aggregate.

**Invariants:**
- Active component must have at least one active health check
- Component name must be unique within environment
- Health score must be recomputed when health checks are added/removed
- Circular dependencies must be rejected

---

### Aggregate 5: CheckResult Aggregate

**Root:** CheckResult

**Contains:**
- CheckResult (root, immutable)

**Boundary Justification:**  
Check results are immutable, high-volume, append-only. They have no internal consistency requirements beyond themselves. Keeping them as their own aggregate prevents locking the Component aggregate on every check execution (which happens every 30 seconds).

**Invariants:**
- Must reference a valid HealthCheck
- Timestamp must not be in the future
- Immutable after creation

---

### Aggregate 6: MetricSnapshot Aggregate

**Root:** MetricSnapshot

**Contains:**
- MetricSnapshot (root, immutable)

**Boundary Justification:**  
Same reasoning as CheckResult — high volume, immutable, append-only. Separate aggregate prevents contention.

**Invariants:**
- Must reference a valid Component
- Must have a valid MetricType
- Value must be within type-specific bounds
- Immutable after creation

---

### Aggregate 7: Incident Aggregate

**Root:** Incident

**Contains:**
- Incident (root)
- IncidentEvent[] (child entities, append-only)

**Boundary Justification:**  
An incident and its events form a consistency boundary. You cannot add an event to a non-existent incident. State transitions (opened → acknowledged → resolved) must be atomic with the event that records the transition. This is a classic event-sourced aggregate candidate.

**Invariants:**
- Status transitions must follow allowed state machine
- Severity can only increase during an active incident (never auto-decrease)
- Resolution requires at least one event explaining the resolution
- Cannot add events to a closed incident
- Deduplication: cannot open a new incident for a component that has an active incident for the same failure type

---

### Aggregate 8: AlertRule Aggregate

**Root:** AlertRule

**Contains:**
- AlertRule (root)
- AlertCondition (embedded value object)
- CooldownPolicy (embedded value object)

**Boundary Justification:**  
An alert rule and its conditions are one atomic unit. Changing conditions without changing the rule makes no sense. Alert history is OUTSIDE because it's append-only and should not block rule evaluation.

**Invariants:**
- Must reference at least one notification channel
- Condition expression must be syntactically valid
- Cooldown period must be positive
- Cannot have two identical rules for the same scope

---

### Aggregate 9: AlertHistory Aggregate

**Root:** AlertHistory

**Contains:**
- AlertHistory (root, immutable)
- DeliveryAttempt[] (child entities, append-only)

**Boundary Justification:**  
Alert history tracks the full delivery lifecycle. Delivery attempts are appended as retries happen. Once all channels succeed or max retries exhausted, the record is finalized.

**Invariants:**
- Must reference a valid AlertRule
- DeliveryAttempts ordered by timestamp
- Maximum retry count per channel enforced
- Immutable once finalized (all channels succeeded or exhausted)

---

### Aggregate 10: NotificationChannel Aggregate

**Root:** NotificationChannel

**Contains:**
- NotificationChannel (root)
- ChannelConfig (embedded value object, type-specific)

**Boundary Justification:**  
Each channel is independently configured, verified, and managed. Its configuration must be atomically consistent.

**Invariants:**
- Must be verified before it can be used by alert rules
- Configuration must be valid for channel type
- Cannot delete a channel that is referenced by active alert rules

---

### Aggregate 11: MaintenanceWindow Aggregate

**Root:** MaintenanceWindow

**Contains:**
- MaintenanceWindow (root)

**Boundary Justification:**  
Simple lifecycle entity. Independent of other aggregates. Referenced by incident detection logic but not owned by it.

**Invariants:**
- End time must be after start time
- Cannot overlap with another window for the same scope
- Maximum duration enforced (prevent "forever maintenance" mistakes)

---

### Aggregate 12: HealthScore Aggregate

**Root:** HealthScore

**Contains:**
- HealthScore (root, immutable)
- ScoreFactor[] (embedded value objects)

**Boundary Justification:**  
Each score computation is an immutable record with its contributing factors. Factors must be atomically consistent with the score (you cannot have a score without knowing what produced it).

**Invariants:**
- Score value between 0 and 100
- Must have at least one contributing factor
- Factor weights must sum to 1.0
- Immutable after computation

---

## Value Objects

### Identity Value Objects

| Value Object | Purpose | Structure |
|-------------|---------|-----------|
| **TenantID** | Globally unique tenant identifier | UUID v7 (time-sortable) |
| **ProductID** | Product identifier, unique within tenant | UUID v7 |
| **EnvironmentID** | Environment identifier | UUID v7 |
| **ComponentID** | Component identifier | UUID v7 |
| **HealthCheckID** | Health check identifier | UUID v7 |
| **IncidentID** | Incident identifier, human-readable | Prefixed: `INC-{tenant_short}-{sequence}` |
| **AlertRuleID** | Alert rule identifier | UUID v7 |

### Domain-Specific Value Objects

| Value Object | Purpose | Possible Values / Structure |
|-------------|---------|----------------------------|
| **HealthStatus** | Current health classification of a component | `healthy`, `degraded`, `unhealthy`, `unknown` |
| **Severity** | Incident/alert severity classification | `critical`, `high`, `medium`, `low`, `info` |
| **IncidentStatus** | Current lifecycle state of an incident | `opened`, `acknowledged`, `investigating`, `resolved`, `closed` |
| **MetricType** | Classification of a metric | `availability`, `latency`, `error_rate`, `saturation`, `throughput`, `custom` |
| **CheckType** | Type of health check probe | `http`, `tcp`, `dns`, `tls_expiry`, `custom` |
| **EnvironmentType** | Classification of environment | `production`, `staging`, `development`, `custom` |
| **ChannelType** | Notification delivery mechanism | `telegram`, `email`, `webhook`, `slack` |
| **ComponentType** | Classification of monitored component | `database`, `cache`, `api`, `frontend`, `queue`, `worker`, `external_service` |
| **Score** | Bounded numeric score | Integer 0-100 with semantic meaning |
| **TimeWindow** | Duration for aggregation/evaluation | `{duration, unit}` e.g., 5 minutes, 1 hour, 24 hours |
| **Threshold** | Comparison value for alerting | `{operator, value, unit}` e.g., `> 500ms`, `< 99.9%` |
| **Schedule** | Cron-like execution schedule | `{interval_seconds, jitter_seconds, timezone}` |
| **HttpTarget** | HTTP check target specification | `{url, method, headers, expected_status, timeout, tls_verify}` |
| **Assertion** | Success criteria for a check | `{type, operator, expected_value}` e.g., status_code == 200 AND latency < 500ms |
| **CooldownPolicy** | Alert deduplication window | `{duration, reset_on_recovery}` |
| **RetryPolicy** | Failure retry specification | `{max_attempts, backoff_strategy, backoff_base}` |
| **DateRange** | Bounded time range | `{start: timestamp, end: timestamp}` |
| **GeoLocation** | Check execution origin (future) | `{region, datacenter, coordinates}` |
| **ScoreFactor** | One contributing factor to a health score | `{metric_type, weight, raw_value, normalized_value}` |
| **ContactInfo** | Alert recipient details | `{type, address, verified}` |
| **IncidentFingerprint** | Deduplication key for incidents | `{component_id, failure_type, check_id}` — used to prevent duplicate incidents |

### Why These Are Value Objects (Not Entities)

- They have **no identity** — two `Severity("critical")` values are interchangeable
- They are **immutable** — you don't update a severity, you replace it
- They are **self-validating** — construction enforces invariants (Score cannot be 101)
- They are **combinable** — Threshold + TimeWindow = AlertCondition

---

## Domain Events

### Execution Context Events

| Event | Produced By | Consumed By | Payload |
|-------|-------------|-------------|---------|
| `HealthCheckExecuted` | Check Executor | Intelligence, Incident Detection | `{check_id, component_id, tenant_id, result: {status, latency_ms, response_code, timestamp}}` |
| `HealthCheckFailed` | Check Executor | Incident Detection, Alerting | `{check_id, component_id, tenant_id, failure_reason, consecutive_failure_count, timestamp}` |
| `HealthCheckRecovered` | Check Executor | Incident Detection, Alerting | `{check_id, component_id, tenant_id, downtime_duration, timestamp}` |
| `HealthCheckTimeout` | Check Executor | Incident Detection | `{check_id, component_id, tenant_id, timeout_ms, timestamp}` |
| `CheckConfigurationChanged` | Component Aggregate | Check Executor (reschedule) | `{check_id, component_id, previous_schedule, new_schedule}` |

### Intelligence Context Events

| Event | Produced By | Consumed By | Payload |
|-------|-------------|-------------|---------|
| `HealthScoreComputed` | Scoring Engine | Dashboard, Alerting, Baseline | `{subject_type, subject_id, tenant_id, score, factors[], computed_at}` |
| `HealthDegraded` | Scoring Engine | Incident Detection, Alerting | `{component_id, tenant_id, previous_score, current_score, threshold_breached, timestamp}` |
| `HealthRestored` | Scoring Engine | Incident Detection | `{component_id, tenant_id, previous_score, current_score, timestamp}` |
| `AnomalyDetected` | Baseline Comparator | Incident Detection | `{component_id, tenant_id, metric_type, expected_range, actual_value, deviation_sigma, timestamp}` |
| `BaselineComputed` | Baseline Engine | Intelligence (internal) | `{component_id, metric_type, window, p50, p95, p99, stddev, computed_at}` |
| `MetricCollected` | Metric Processor | Intelligence, Storage | `{component_id, tenant_id, metric_type, value, unit, collected_at}` |

### Incident Context Events

| Event | Produced By | Consumed By | Payload |
|-------|-------------|-------------|---------|
| `IncidentOpened` | Incident Detector | Alerting, Dashboard, Audit | `{incident_id, tenant_id, component_id, severity, title, trigger_event, opened_at}` |
| `IncidentAcknowledged` | User Action / API | Alerting (stop escalation), Dashboard | `{incident_id, tenant_id, acknowledged_by, acknowledged_at}` |
| `IncidentEscalated` | Escalation Policy | Alerting | `{incident_id, tenant_id, previous_severity, new_severity, reason, escalated_at}` |
| `IncidentResolved` | Auto-detection / User | Alerting, Dashboard, Intelligence | `{incident_id, tenant_id, resolution_type, resolved_by, resolution_note, resolved_at, duration}` |
| `IncidentClosed` | User Action | Audit | `{incident_id, tenant_id, closed_by, postmortem_attached, closed_at}` |
| `IncidentMerged` | Correlation Engine | Dashboard | `{source_incident_id, target_incident_id, tenant_id, reason}` |
| `IncidentCorrelated` | Correlation Engine (future) | Dashboard, AI | `{incident_id, correlated_incidents[], common_root_component, confidence}` |

### Alerting Context Events

| Event | Produced By | Consumed By | Payload |
|-------|-------------|-------------|---------|
| `AlertRuleTriggered` | Rule Evaluator | Notification Dispatcher | `{rule_id, tenant_id, trigger_source, severity, channels[], timestamp}` |
| `AlertSent` | Notification Dispatcher | Audit, AlertHistory | `{alert_history_id, tenant_id, channel_type, recipient, delivery_status, sent_at}` |
| `AlertDeliveryFailed` | Notification Dispatcher | Retry Queue, Audit | `{alert_history_id, tenant_id, channel_type, error_reason, attempt_number, next_retry_at}` |
| `AlertSuppressed` | Rule Evaluator | Audit | `{rule_id, tenant_id, suppression_reason, timestamp}` — reasons: cooldown, maintenance window, muted |
| `AlertCooldownExpired` | Timer | Rule Evaluator | `{rule_id, tenant_id, timestamp}` |

### Tenant Context Events

| Event | Produced By | Consumed By | Payload |
|-------|-------------|-------------|---------|
| `TenantOnboarded` | Registration | Provisioning, Audit | `{tenant_id, name, plan, onboarded_at}` |
| `TenantPlanChanged` | Billing | Quota Enforcement, Execution | `{tenant_id, previous_plan, new_plan, effective_at}` |
| `TenantSuspended` | Billing / Admin | Execution (pause all), Alerting (pause) | `{tenant_id, reason, suspended_at}` |
| `TenantQuotaExceeded` | Quota Monitor | Alerting (notify tenant), Execution (throttle) | `{tenant_id, resource_type, limit, current_usage}` |

### Catalog Context Events

| Event | Produced By | Consumed By | Payload |
|-------|-------------|-------------|---------|
| `ComponentRegistered` | Catalog | Execution (start monitoring) | `{component_id, environment_id, product_id, tenant_id, component_type, health_checks[]}` |
| `ComponentDeregistered` | Catalog | Execution (stop monitoring), Intelligence (archive baseline) | `{component_id, tenant_id, deregistered_at}` |
| `DependencyDeclared` | Catalog | Intelligence (update scoring graph) | `{source_component_id, target_component_id, tenant_id, dependency_type}` |
| `EnvironmentCreated` | Catalog | Dashboard | `{environment_id, product_id, tenant_id, type, policy}` |

### System Events (Cross-Cutting)

| Event | Produced By | Consumed By | Payload |
|-------|-------------|-------------|---------|
| `MaintenanceWindowStarted` | Scheduler | Incident Detection (suppress), Alerting (mute) | `{window_id, tenant_id, scope_type, scope_id, started_at, ends_at}` |
| `MaintenanceWindowEnded` | Scheduler | Incident Detection (resume), Alerting (unmute) | `{window_id, tenant_id, ended_at}` |

---

## Tenant Hierarchy

### Option Analysis

#### Option A: Flat (Tenant → Product → Environment)

```
Tenant
└── Product
    └── Environment
        └── Component
```

**Pros:** Simple. Covers 90% of use cases. Easy to query. Easy to explain to users.  
**Cons:** No workspace isolation for teams within a tenant. No cross-product environment grouping.

#### Option B: With Workspace (Tenant → Workspace → Product → Environment)

```
Tenant
└── Workspace
    └── Product
        └── Environment
            └── Component
```

**Pros:** Supports team isolation within large tenants. Workspaces can have independent alert rules and users.  
**Cons:** Premature for Phase 1. Adds a layer that small tenants don't need. Complicates every query with an extra join.

#### Option C: With Organization Layer (Org → Tenant → Product → Environment)

```
Organization
└── Tenant
    └── Product
        └── Environment
            └── Component
```

**Pros:** Supports enterprise customers with multiple "tenants" (business units) under one billing entity.  
**Cons:** Extremely premature. Adds billing complexity. Only needed at 100+ enterprise customers.

### Decision: Option A with Future Migration Path to B

**Use Option A for Phase 1 through Phase 3.**

Rationale:
1. GymFlow Track and the next 10 customers are single-team organizations
2. Workspace adds query complexity to every single database read
3. If needed later, Workspace can be introduced as an optional grouping layer between Tenant and Product without breaking existing data (existing products get assigned to a "Default" workspace)

### Final Hierarchy

```
Tenant: "GymFlow"
├── Product: "GymFlow Track"
│   ├── Environment: production
│   │   ├── Component: backend-api (type: api)
│   │   ├── Component: postgresql (type: database)
│   │   ├── Component: redis (type: cache)
│   │   └── Component: frontend (type: frontend)
│   ├── Environment: staging
│   │   ├── Component: backend-api
│   │   └── Component: postgresql
│   └── Environment: development
│       └── Component: backend-api
└── Product: "GymFlow Admin" (future)
    └── Environment: production
        └── Component: admin-api
```

### Hierarchy Rules

1. **Tenant** is the billing and isolation boundary
2. **Product** is the organizational boundary (maps to a deployable system)
3. **Environment** is the operational boundary (controls monitoring behavior)
4. **Component** is the monitoring boundary (the smallest unit we check)
5. Health scores roll up: Component → Environment → Product → Tenant
6. Alerts scope down: Tenant-level rules → Product-level → Environment-level → Component-level (most specific wins)

---

## ER Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              TENANT CONTEXT                                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────────┐         ┌────────────────────┐                               │
│  │   TENANT     │ 1───* │ NOTIFICATION_CHANNEL │                               │
│  ├──────────────┤        ├────────────────────┤                                │
│  │ id (PK)      │        │ id (PK)            │                                │
│  │ name         │        │ tenant_id (FK)     │                                │
│  │ slug         │        │ type               │                                │
│  │ plan         │        │ config (encrypted) │                                │
│  │ status       │        │ verified           │                                │
│  │ created_at   │        │ status             │                                │
│  └──────┬───────┘        └────────────────────┘                                │
│         │                                                                       │
│         │ 1:*                                                                   │
│         ▼                                                                       │
│  ┌──────────────┐                                                              │
│  │   PRODUCT    │                                                              │
│  ├──────────────┤                                                              │
│  │ id (PK)      │                                                              │
│  │ tenant_id(FK)│                                                              │
│  │ name         │                                                              │
│  │ status       │                                                              │
│  └──────┬───────┘                                                              │
│         │                                                                       │
│         │ 1:*                                                                   │
│         ▼                                                                       │
│  ┌──────────────────┐                                                          │
│  │  ENVIRONMENT     │                                                          │
│  ├──────────────────┤                                                          │
│  │ id (PK)          │                                                          │
│  │ product_id (FK)  │                                                          │
│  │ tenant_id (FK)   │                                                          │
│  │ name             │                                                          │
│  │ type             │                                                          │
│  │ policy (JSONB)   │                                                          │
│  │ status           │                                                          │
│  └──────┬───────────┘                                                          │
│         │                                                                       │
│         │ 1:*                                                                   │
│         ▼                                                                       │
│  ┌──────────────────┐       ┌───────────────────────┐                          │
│  │   COMPONENT      │ 1──1 │   COMPONENT_HEALTH     │                          │
│  ├──────────────────┤       ├───────────────────────┤                          │
│  │ id (PK)          │       │ component_id (PK, FK) │                          │
│  │ environment_id   │       │ status                │                          │
│  │ tenant_id (FK)   │       │ score                 │                          │
│  │ name             │       │ last_check_at         │                          │
│  │ type             │       │ consecutive_failures  │                          │
│  │ status           │       │ consecutive_successes │                          │
│  │ created_at       │       │ updated_at            │                          │
│  └──┬───────┬───────┘       └───────────────────────┘                          │
│     │       │                                                                   │
│     │       │ 1:*                                                              │
│     │       ▼                                                                   │
│     │  ┌──────────────────┐      ┌──────────────────┐                          │
│     │  │  HEALTH_CHECK    │ 1──* │  CHECK_RESULT    │                          │
│     │  ├──────────────────┤      ├──────────────────┤                          │
│     │  │ id (PK)          │      │ id (PK)          │                          │
│     │  │ component_id(FK) │      │ health_check_id  │                          │
│     │  │ tenant_id (FK)   │      │ tenant_id (FK)   │                          │
│     │  │ type             │      │ status           │                          │
│     │  │ target (JSONB)   │      │ latency_ms       │                          │
│     │  │ assertion (JSONB)│      │ response_code    │                          │
│     │  │ schedule         │      │ error_message    │                          │
│     │  │ retry_policy     │      │ executed_at      │                          │
│     │  │ status           │      │ (PARTITIONED)    │                          │
│     │  └──────────────────┘      └──────────────────┘                          │
│     │                                                                           │
│     │ *:*                                                                       │
│     ▼                                                                           │
│  ┌────────────────────────┐                                                    │
│  │ COMPONENT_DEPENDENCY   │                                                    │
│  ├────────────────────────┤                                                    │
│  │ source_component_id    │                                                    │
│  │ target_component_id    │                                                    │
│  │ tenant_id (FK)         │                                                    │
│  │ dependency_type        │                                                    │
│  └────────────────────────┘                                                    │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                           INTELLIGENCE CONTEXT                                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────────────┐       ┌──────────────────┐                               │
│  │ METRIC_SNAPSHOT  │       │  HEALTH_SCORE    │                               │
│  ├──────────────────┤       ├──────────────────┤                               │
│  │ id (PK)          │       │ id (PK)          │                               │
│  │ component_id(FK) │       │ subject_type     │  (component|environment|      │
│  │ tenant_id (FK)   │       │ subject_id       │   product)                    │
│  │ metric_type      │       │ tenant_id (FK)   │                               │
│  │ value            │       │ score            │                               │
│  │ unit             │       │ factors (JSONB)  │                               │
│  │ collected_at     │       │ computed_at      │                               │
│  │ (PARTITIONED)    │       │ (PARTITIONED)    │                               │
│  └──────────────────┘       └──────────────────┘                               │
│                                                                                 │
│  ┌──────────────────┐                                                          │
│  │   BASELINE       │                                                          │
│  ├──────────────────┤                                                          │
│  │ id (PK)          │                                                          │
│  │ component_id(FK) │                                                          │
│  │ tenant_id (FK)   │                                                          │
│  │ metric_type      │                                                          │
│  │ window           │                                                          │
│  │ p50, p95, p99    │                                                          │
│  │ stddev           │                                                          │
│  │ sample_count     │                                                          │
│  │ computed_at      │                                                          │
│  │ status           │                                                          │
│  └──────────────────┘                                                          │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                            INCIDENT CONTEXT                                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────────────┐       ┌──────────────────┐                               │
│  │    INCIDENT      │ 1──* │ INCIDENT_EVENT    │                               │
│  ├──────────────────┤       ├──────────────────┤                               │
│  │ id (PK)          │       │ id (PK)          │                               │
│  │ tenant_id (FK)   │       │ incident_id (FK) │                               │
│  │ component_id(FK) │       │ event_type       │                               │
│  │ fingerprint      │       │ actor            │                               │
│  │ title            │       │ payload (JSONB)  │                               │
│  │ severity         │       │ occurred_at      │                               │
│  │ status           │       └──────────────────┘                               │
│  │ opened_at        │                                                          │
│  │ acknowledged_at  │                                                          │
│  │ resolved_at      │                                                          │
│  │ closed_at        │                                                          │
│  │ duration_seconds │                                                          │
│  └──────────────────┘                                                          │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                            ALERTING CONTEXT                                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────────────┐       ┌──────────────────┐                               │
│  │   ALERT_RULE     │ 1──* │  ALERT_HISTORY   │                               │
│  ├──────────────────┤       ├──────────────────┤                               │
│  │ id (PK)          │       │ id (PK)          │                               │
│  │ tenant_id (FK)   │       │ alert_rule_id    │                               │
│  │ scope_type       │       │ tenant_id (FK)   │                               │
│  │ scope_id         │       │ incident_id      │                               │
│  │ condition (JSONB)│       │ severity         │                               │
│  │ severity         │       │ channels[]       │                               │
│  │ channels[] (FK)  │       │ delivery_status  │                               │
│  │ cooldown         │       │ triggered_at     │                               │
│  │ status           │       │ (PARTITIONED)    │                               │
│  └──────────────────┘       └──────────────────┘                               │
│                                                                                 │
│  ┌────────────────────────┐                                                    │
│  │  MAINTENANCE_WINDOW    │                                                    │
│  ├────────────────────────┤                                                    │
│  │ id (PK)                │                                                    │
│  │ tenant_id (FK)         │                                                    │
│  │ scope_type             │                                                    │
│  │ scope_id               │                                                    │
│  │ starts_at              │                                                    │
│  │ ends_at                │                                                    │
│  │ reason                 │                                                    │
│  │ status                 │                                                    │
│  └────────────────────────┘                                                    │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                           SYSTEM / CROSS-CUTTING                                 │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────────────┐       ┌──────────────────┐                               │
│  │   AUDIT_ENTRY    │       │   EVENT_OUTBOX   │                               │
│  ├──────────────────┤       ├──────────────────┤                               │
│  │ id (PK)          │       │ id (PK)          │                               │
│  │ tenant_id (FK)   │       │ event_type       │                               │
│  │ actor            │       │ aggregate_type   │                               │
│  │ action           │       │ aggregate_id     │                               │
│  │ subject_type     │       │ tenant_id        │                               │
│  │ subject_id       │       │ payload (JSONB)  │                               │
│  │ metadata (JSONB) │       │ status           │                               │
│  │ occurred_at      │       │ created_at       │                               │
│  └──────────────────┘       │ processed_at     │                               │
│                             └──────────────────┘                               │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Cardinality Summary

| Relationship | Cardinality |
|-------------|-------------|
| Tenant → Product | 1:* |
| Product → Environment | 1:* |
| Environment → Component | 1:* |
| Component → HealthCheck | 1:* |
| HealthCheck → CheckResult | 1:* |
| Component → ComponentHealth | 1:1 |
| Component → MetricSnapshot | 1:* |
| Component → Baseline | 1:* (one per metric type) |
| Component → Incident | 1:* |
| Incident → IncidentEvent | 1:* |
| Tenant → NotificationChannel | 1:* |
| Tenant → AlertRule | 1:* |
| AlertRule → AlertHistory | 1:* |
| Component ↔ Component (dependency) | *:* |
| HealthScore → Subject (polymorphic) | *:1 |

### Partitioned Tables (Critical for Scale)

| Table | Partition Strategy | Partition Key |
|-------|-------------------|---------------|
| check_results | Range (time) | `(tenant_id, executed_at)` monthly |
| metric_snapshots | Range (time) | `(tenant_id, collected_at)` monthly |
| health_scores | Range (time) | `(tenant_id, computed_at)` monthly |
| alert_history | Range (time) | `(tenant_id, triggered_at)` monthly |
| event_outbox | Range (time) | `created_at` daily (short-lived) |

---

## Design Review

### Missing Entities Identified During Review

| Entity | Need | Phase |
|--------|------|-------|
| **User** | Authentication, authorization, incident assignment | Phase 2 |
| **Team** | Group users, scope permissions, on-call rotation | Phase 2 |
| **APIKey** | Machine authentication for REST API and push SDK | Phase 1 (add to Tenant aggregate) |
| **OnCallSchedule** | Route alerts to the right person at the right time | Phase 3 |
| **Runbook** | Document resolution steps. Link to incidents. Foundation for auto-remediation | Phase 3 |
| **StatusPage** | Public-facing component status | Phase 3 |
| **IntegrationConfig** | Third-party integrations (PagerDuty, Jira, etc.) | Phase 3 |
| **AIInsight** | Store AI-generated analysis linked to incidents or components | Phase 4 |
| **Postmortem** | Structured post-incident review linked to Incident | Phase 3 |

### Future AI Requirements (Design For Now, Build Later)

| Requirement | Domain Impact | What to Design Now |
|-------------|--------------|-------------------|
| Training data pipeline | MetricSnapshots and CheckResults must be queryable historically | Partitioning strategy. Retention policies. Don't delete raw data — archive it |
| Feature vectors per component | Need: latency p50/p95/p99, availability %, error patterns, time-of-day patterns | Store percentiles in MetricSnapshot, not just averages |
| Incident similarity | Incidents need fingerprinting that's consistent and semantic | IncidentFingerprint value object. Same failure = same fingerprint |
| Root cause graph | ComponentDependency must be accurate and complete | Make dependency declaration easy and eventually auto-discoverable |
| Natural language context | Every incident, check, and score needs human-readable context | Store `title`, `description`, `resolution_note` as text — not just codes |
| Feedback loop | AI recommendations need accept/reject signals to improve | Design a generic `Feedback` entity: `{subject_type, subject_id, signal: accept|reject|ignore}` |

### Scaling Concerns

| Concern | Impact | Design Decision |
|---------|--------|-----------------|
| Check result volume at 1000 tenants × 50 components × 30s interval = 1.44M writes/hour | PostgreSQL single-node ceiling | Partition by tenant_id + time. Batch inserts. Consider TimescaleDB extension as escape hatch |
| Health score computation at scale | CPU-bound aggregation over millions of rows | Pre-aggregate incrementally. Store rolling windows. Don't recompute from raw data every time |
| Event outbox throughput | At scale, outbox becomes a bottleneck | Partition outbox by tenant. Allow multiple workers to process different partitions |
| Dashboard read load | Thousands of concurrent dashboard sessions requesting live data | ComponentHealth table as a materialized "current state" view. Polling interval minimum 10s |
| Tenant isolation at query level | One missing WHERE clause = data breach | Repository pattern. PostgreSQL Row-Level Security as defense-in-depth. Integration test suite that verifies isolation |

### Security Concerns

| Concern | Domain Entity Affected | Mitigation |
|---------|----------------------|-----------|
| Notification channel credentials (webhook URLs, API tokens) | NotificationChannel.config | Encrypt at rest with per-tenant encryption key. Never return in API responses. Mask in UI |
| Health check target URLs may contain credentials | HealthCheck.target | Validate target. Strip embedded credentials. Encourage header-based auth. Encrypt sensitive headers |
| Audit log tampering | AuditEntry | Append-only table. No UPDATE/DELETE permissions at database level for audit table |
| Cross-tenant data in shared reporting | HealthScore, MetricSnapshot | Every materialized view MUST include tenant_id in its key |
| API key brute force | APIKey | Hash stored keys. Rate-limit auth failures. Key rotation support |

---

## Recommendations

### Critical (Before Writing First Migration)

1. **Add APIKey as a first-class entity in the Tenant aggregate** — Required for Phase 1 API access
2. **Define the IncidentFingerprint algorithm** — `hash(component_id + check_id + failure_type)`. This prevents duplicate incidents for the same problem
3. **Design the polymorphic scope reference pattern** — AlertRule, HealthScore, and MaintenanceWindow all have `scope_type + scope_id`. Standardize this as a value object: `MonitoringScope{type: component|environment|product, id: UUID}`
4. **Choose UUID v7** — Time-sortable UUIDs eliminate the need for separate `created_at` indexes in many cases
5. **Decide: Is ComponentHealth a materialized view or a managed table?** — Recommendation: Managed table updated by event handlers. Materialized views have refresh lag that's unacceptable for real-time dashboards

### Important (During Development)

6. **Implement the "consecutive failures" pattern correctly** — Don't count failures globally. Count per-check. A component with 5 checks failing on 1 check is different from all 5 failing
7. **Design health scores as composable** — Environment score = weighted average of component scores. Product score = weighted average of environment scores (production weighted higher). This must be explicitly defined, not assumed
8. **Event schema versioning from Day 1** — Every event payload includes a `schema_version` field. When payloads change, consumers must handle both old and new versions
9. **Idempotency keys on CheckResult and AlertHistory** — Prevent duplicates from retry scenarios. Use `{check_id}_{scheduled_at}` as natural idempotency key

### Score Composition Rules

```
Component Score = weighted_average(
  availability_factor × 0.4,    # % of checks passing in last window
  latency_factor × 0.3,         # p95 vs threshold
  consistency_factor × 0.2,     # coefficient of variation
  freshness_factor × 0.1        # time since last successful check
)

Environment Score = weighted_average(
  component_scores[],
  weights based on component_type:
    database: 1.5x
    api: 1.2x
    cache: 0.8x
    frontend: 0.7x
)

Product Score = weighted_average(
  environment_scores[],
  weights based on environment_type:
    production: 3.0x
    staging: 0.5x
    development: 0.1x
)
```

### Aggregate Design Validation Checklist

| Check | Status |
|-------|--------|
| Every aggregate root has TenantID | ✅ |
| No aggregate references another aggregate's internals | ✅ |
| High-volume writes (CheckResult, MetricSnapshot) are separate aggregates | ✅ |
| Immutable data (results, events, history) cannot be modified after creation | ✅ |
| Incident state machine is enforced by the aggregate, not by callers | ✅ |
| Component invariant (≥1 active check) is enforced at aggregate level | ✅ |
| No circular dependency allowed (validated on ComponentDependency add) | ✅ |

---

*End of Domain Model*
