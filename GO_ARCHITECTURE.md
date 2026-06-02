# OPTRION — Go Codebase Architecture

**Author:** Principal Go Architect  
**Date:** 2026-05-29 (Updated: 2026-06-02)  
**Version:** 2.0  
**Status:** Implementation Complete — All Contexts Operational

---

## Bounded Contexts

### Context Map

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          OPTRION PLATFORM                                │
│                                                                         │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────────────────┐  │
│  │   TENANT    │     │   CATALOG   │     │      EXECUTION          │  │
│  │             │     │             │     │                         │  │
│  │ Onboarding  │────▶│  Products   │────▶│  Health Check Scheduling│  │
│  │ Plans       │     │  Envs       │     │  Health Check Running   │  │
│  │ API Keys    │     │  Components │     │  Result Recording       │  │
│  │ Quotas      │     │  Deps       │     │                         │  │
│  └─────────────┘     └─────────────┘     └───────────┬─────────────┘  │
│                                                       │                 │
│                                                       ▼ events          │
│  ┌─────────────────────────┐     ┌─────────────────────────────────┐  │
│  │      INTELLIGENCE       │◀────│         INCIDENT                │  │
│  │                         │     │                                 │  │
│  │  Health Scoring         │────▶│  Detection                     │  │
│  │  Baseline Management    │     │  Lifecycle (open→resolve)      │  │
│  │  Anomaly Detection      │     │  Deduplication                 │  │
│  │  Trend Analysis         │     │  Escalation                    │  │
│  └─────────────────────────┘     └───────────┬─────────────────────┘  │
│                                               │                         │
│                                               ▼ events                  │
│  ┌─────────────────────────┐     ┌─────────────────────────────────┐  │
│  │     NOTIFICATION        │◀────│         ALERTING                │  │
│  │                         │     │                                 │  │
│  │  Channel Management     │     │  Rule Evaluation               │  │
│  │  Delivery               │     │  Cooldown Management           │  │
│  │  Retry Logic            │     │  Suppression (maintenance)     │  │
│  │  Templating             │     │  History                       │  │
│  └─────────────────────────┘     └─────────────────────────────────┘  │
│                                                                         │
│  ┌─────────────────────────┐  ┌─────────────────────────────────────┐  │
│  │  AI ROOT CAUSE ANALYSIS │  │     RECOMMENDATION ENGINE           │  │
│  │                         │  │                                     │  │
│  │  Multi-Provider (4)     │  │  Evidence-Based                     │  │
│  │  Context Builder        │  │  Safety Validation                  │  │
│  │  Confidence Scoring     │  │  Priority Ranking                   │  │
│  │  Investigation Hints    │  │  Hallucination Controls             │  │
│  └─────────────────────────┘  └─────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Context Definitions

| Context | Responsibility | Owner | Communication Style |
|---------|---------------|-------|-------------------|
| **Tenant** | Identity, authentication, authorization, plans, quotas | Platform team | Synchronous (queried by all others) |
| **Catalog** | Service registry — products, environments, components, dependencies | Platform team | Synchronous commands, publishes events on change |
| **Execution** | Scheduling and running health checks, recording raw results | Monitoring team | High-throughput writes, publishes events |
| **Intelligence** | Scoring, baselines, anomaly detection, trend analysis | Data team | Event-driven consumer, publishes derived events |
| **Incident** | Detecting, deduplicating, lifecycle-managing incidents | Reliability team | Event-driven consumer, publishes state changes |
| **Alerting** | Rule evaluation, cooldown, suppression, maintenance windows | Reliability team | Event-driven consumer, commands notification |
| **Notification** | Channel management, delivery, retries, templating | Platform team | Command-driven (receives "send" commands) |
| **AI** | Root cause analysis, confidence scoring, investigation hints | AI team | Event-driven consumer, multi-provider (Gemini, OpenAI, Anthropic, Ollama) |
| **Recommendation** | Evidence-based recommendations, safety validation, ranking | AI team | Triggered by AI analysis results |
| **Registration** | Bulk plug-and-play onboarding | Platform team | Synchronous (single-request setup) |
| **AutoDiscovery** | Detect infrastructure (PostgreSQL, Redis, HTTP) | Platform team | CLI-triggered, environment variable scanning |

### Context Communication Rules

| From | To | Mechanism |
|------|-----|-----------|
| Execution → Intelligence | Domain Event (`HealthCheckExecuted`) |
| Intelligence → Incident | Domain Event (`AnomalyDetected`, `HealthDegraded`) |
| Incident → Alerting | Domain Event (`IncidentOpened`, `IncidentEscalated`) |
| Alerting → Notification | Command (`SendNotification`) |
| Catalog → Execution | Domain Event (`ComponentRegistered`) |
| Any → Tenant | Synchronous query (TenantID validation, quota check) |

---

## Monorepo Structure

```
optrion/
├── cmd/
│   └── optrion/
│       └── main.go                    # Single binary entry point
│
├── internal/
│   ├── tenant/                        # Tenant bounded context
│   │   ├── domain/
│   │   │   ├── tenant.go             # Aggregate root
│   │   │   ├── apikey.go             # Entity
│   │   │   ├── plan.go               # Value object
│   │   │   ├── events.go             # Domain events
│   │   │   └── errors.go             # Domain errors
│   │   ├── app/
│   │   │   ├── service.go            # Application service (use cases)
│   │   │   ├── commands.go           # Command definitions
│   │   │   └── queries.go            # Query definitions
│   │   ├── port/
│   │   │   ├── repository.go         # Repository interface
│   │   │   ├── hasher.go             # Key hasher interface
│   │   │   └── quota.go              # Quota checker interface
│   │   └── adapter/
│   │       ├── postgres/
│   │       │   └── repository.go     # PostgreSQL implementation
│   │       └── rest/
│   │           ├── handler.go         # HTTP handlers
│   │           ├── request.go         # Request DTOs
│   │           └── response.go        # Response DTOs
│   │
│   ├── catalog/                       # Catalog bounded context
│   │   ├── domain/
│   │   │   ├── product.go
│   │   │   ├── environment.go
│   │   │   ├── component.go          # Aggregate root (includes health checks)
│   │   │   ├── dependency.go
│   │   │   ├── healthcheck.go        # Child entity within component aggregate
│   │   │   ├── events.go
│   │   │   └── errors.go
│   │   ├── app/
│   │   │   ├── service.go
│   │   │   ├── commands.go
│   │   │   └── queries.go
│   │   ├── port/
│   │   │   ├── repository.go
│   │   │   └── event_publisher.go
│   │   └── adapter/
│   │       ├── postgres/
│   │       │   └── repository.go
│   │       └── rest/
│   │           ├── handler.go
│   │           ├── request.go
│   │           └── response.go
│   │
│   ├── execution/                     # Execution bounded context
│   │   ├── domain/
│   │   │   ├── checkresult.go        # Aggregate root (immutable)
│   │   │   ├── checker.go            # Health check execution logic
│   │   │   ├── events.go
│   │   │   └── errors.go
│   │   ├── app/
│   │   │   ├── service.go            # Orchestrates check scheduling + execution
│   │   │   ├── scheduler.go          # Schedule management use case
│   │   │   └── commands.go
│   │   ├── port/
│   │   │   ├── repository.go         # CheckResult repository
│   │   │   ├── checker.go            # HealthChecker interface (HTTP, TCP, etc.)
│   │   │   ├── scheduler.go          # Scheduler interface
│   │   │   └── event_publisher.go
│   │   └── adapter/
│   │       ├── postgres/
│   │       │   └── repository.go
│   │       ├── checker/
│   │       │   ├── http.go           # HTTP health checker
│   │       │   ├── tcp.go            # TCP health checker
│   │       │   └── dns.go            # DNS health checker
│   │       └── scheduler/
│   │           └── ticker.go         # In-process scheduler
│   │
│   ├── intelligence/                  # Intelligence bounded context
│   │   ├── domain/
│   │   │   ├── healthscore.go        # Aggregate root (immutable)
│   │   │   ├── componenthealth.go    # Current state entity
│   │   │   ├── baseline.go           # Baseline entity
│   │   │   ├── scorer.go             # Scoring domain logic
│   │   │   ├── events.go
│   │   │   └── errors.go
│   │   ├── app/
│   │   │   ├── service.go
│   │   │   ├── scorer.go             # Scoring use case
│   │   │   ├── baseline.go           # Baseline computation use case
│   │   │   └── queries.go
│   │   ├── port/
│   │   │   ├── repository.go
│   │   │   ├── metric_store.go       # MetricSnapshot storage
│   │   │   └── event_publisher.go
│   │   └── adapter/
│   │       ├── postgres/
│   │       │   └── repository.go
│   │       └── rest/
│   │           ├── handler.go
│   │           └── response.go
│   │
│   ├── incident/                      # Incident bounded context
│   │   ├── domain/
│   │   │   ├── incident.go           # Aggregate root
│   │   │   ├── event.go              # IncidentEvent child entity
│   │   │   ├── detector.go           # Detection domain logic
│   │   │   ├── fingerprint.go        # Deduplication logic
│   │   │   ├── statemachine.go       # Status transition rules
│   │   │   ├── events.go
│   │   │   └── errors.go
│   │   ├── app/
│   │   │   ├── service.go
│   │   │   ├── detector.go           # Detection use case
│   │   │   ├── commands.go
│   │   │   └── queries.go
│   │   ├── port/
│   │   │   ├── repository.go
│   │   │   └── event_publisher.go
│   │   └── adapter/
│   │       ├── postgres/
│   │       │   └── repository.go
│   │       └── rest/
│   │           ├── handler.go
│   │           ├── request.go
│   │           └── response.go
│   │
│   ├── alerting/                      # Alerting bounded context
│   │   ├── domain/
│   │   │   ├── rule.go               # AlertRule aggregate root
│   │   │   ├── history.go            # AlertHistory aggregate
│   │   │   ├── evaluator.go          # Rule evaluation logic
│   │   │   ├── cooldown.go           # Cooldown management
│   │   │   ├── maintenance.go        # MaintenanceWindow entity
│   │   │   ├── events.go
│   │   │   └── errors.go
│   │   ├── app/
│   │   │   ├── service.go
│   │   │   ├── evaluator.go          # Rule evaluation use case
│   │   │   ├── commands.go
│   │   │   └── queries.go
│   │   ├── port/
│   │   │   ├── repository.go
│   │   │   ├── cooldown_store.go     # Cooldown state (Redis)
│   │   │   └── event_publisher.go
│   │   └── adapter/
│   │       ├── postgres/
│   │       │   └── repository.go
│   │       ├── redis/
│   │       │   └── cooldown.go       # Redis cooldown tracking
│   │       └── rest/
│   │           ├── handler.go
│   │           ├── request.go
│   │           └── response.go
│   │
│   ├── notification/                  # Notification bounded context
│   │   ├── domain/
│   │   │   ├── channel.go            # NotificationChannel aggregate
│   │   │   ├── delivery.go           # Delivery attempt entity
│   │   │   ├── events.go
│   │   │   └── errors.go
│   │   ├── app/
│   │   │   ├── service.go
│   │   │   ├── dispatcher.go         # Notification dispatch use case
│   │   │   └── commands.go
│   │   ├── port/
│   │   │   ├── repository.go
│   │   │   ├── sender.go             # Channel sender interface
│   │   │   └── template_renderer.go
│   │   └── adapter/
│   │       ├── postgres/
│   │       │   └── repository.go
│   │       ├── telegram/
│   │       │   └── sender.go         # Telegram delivery
│   │       ├── webhook/
│   │       │   └── sender.go         # Webhook delivery
│   │       └── rest/
│   │           ├── handler.go
│   │           └── request.go
│   │
│   ├── ai/                            # AI bounded context (IMPLEMENTED)
│   │   ├── domain/
│   │   │   ├── aianalysis/
│   │   │   ├── aicontext/
│   │   │   ├── rootcausereport/
│   │   │   ├── confidencescore/
│   │   │   └── investigationhint/
│   │   ├── app/
│   │   │   ├── service/
│   │   │   ├── contextbuilder/
│   │   │   ├── prompt/
│   │   │   └── validation/
│   │   ├── port/
│   │   │   └── repository/
│   │   └── adapter/
│   │       ├── provider/              # Gemini, OpenAI, Anthropic, Ollama + resilient wrapper
│   │       ├── repository/
│   │       └── rest/v1/
│   │
│   ├── platform/                      # Cross-cutting platform services
│   │   ├── eventbus/
│   │   │   ├── bus.go                # Event bus interface + in-process impl
│   │   │   ├── outbox.go            # Outbox pattern implementation
│   │   │   └── subscriber.go        # Subscriber registry
│   │   ├── auth/
│   │   │   ├── middleware.go         # Authentication middleware
│   │   │   ├── apikey.go            # API key validation
│   │   │   └── context.go           # Tenant context extraction
│   │   ├── multitenancy/
│   │   │   ├── context.go           # TenantID in context
│   │   │   ├── middleware.go        # Tenant resolution middleware
│   │   │   └── rls.go              # RLS session variable setting
│   │   ├── observability/
│   │   │   ├── logger.go           # Structured logging setup
│   │   │   ├── metrics.go          # Prometheus metrics
│   │   │   ├── tracing.go          # OpenTelemetry tracing
│   │   │   └── health.go           # Self health checks
│   │   ├── database/
│   │   │   ├── postgres.go         # Connection pool setup
│   │   │   ├── tx.go               # Transaction manager
│   │   │   └── migrate.go          # Migration runner
│   │   ├── cache/
│   │   │   ├── redis.go            # Redis client setup
│   │   │   └── ratelimit.go        # Rate limiter
│   │   └── server/
│   │       ├── http.go             # HTTP server setup
│   │       ├── router.go           # Route registration
│   │       └── middleware.go       # Common middleware chain
│   │
│   └── shared/                        # Shared kernel (minimal)
│       ├── domain/
│       │   ├── tenantid.go           # TenantID value object
│       │   ├── event.go             # Base domain event interface
│       │   ├── aggregate.go         # Base aggregate root
│       │   ├── errors.go           # Common domain errors
│       │   ├── clock.go            # Clock interface (for testing)
│       │   └── pagination.go       # Pagination value object
│       ├── types/
│       │   ├── severity.go         # Severity enum
│       │   ├── status.go           # Common status types
│       │   ├── score.go            # Score value object (0-100)
│       │   └── timewindow.go       # TimeWindow value object
│       └── id/
│           └── uuid.go             # UUID v7 generation
│
├── pkg/                               # Public packages (importable by external code)
│   └── sdk/
│       ├── client.go                 # OPTRION API client
│       ├── types.go                  # Public API types
│       └── push.go                   # Push-based metric reporting (future)
│
├── migrations/
│   ├── 001_create_schemas.sql
│   ├── 002_core_tables.sql
│   ├── 003_monitoring_tables.sql
│   ├── 004_incident_tables.sql
│   ├── 005_alerting_tables.sql
│   └── 006_system_tables.sql
│
├── config/
│   ├── config.go                     # Configuration struct definitions
│   ├── local.yaml                    # Local development
│   ├── development.yaml              # Shared development
│   ├── staging.yaml                  # Staging
│   └── production.yaml               # Production (secrets from env/vault)
│
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile                # Multi-stage build
│   │   └── docker-compose.yaml       # Local dev dependencies
│   └── k8s/                          # Kubernetes manifests (future)
│       ├── deployment.yaml
│       ├── service.yaml
│       └── configmap.yaml
│
├── scripts/
│   ├── migrate.sh                    # Run migrations
│   ├── seed.sh                       # Seed development data
│   └── lint.sh                       # Lint checks
│
├── docs/
│   ├── ARCHITECTURE_REVIEW.md
│   ├── DOMAIN_MODEL.md
│   ├── DATABASE_ARCHITECTURE.md
│   └── GO_ARCHITECTURE.md            # This document
│
├── tools/
│   └── tools.go                      # Tool dependencies (golangci-lint, etc.)
│
├── go.mod
├── go.sum
├── Makefile
└── .golangci.yml                     # Linter configuration
```

### Folder Explanations

| Folder | Purpose | Rules |
|--------|---------|-------|
| `cmd/optrion/` | Single binary entry point. Wires dependencies, starts server | Zero business logic. Only dependency injection and lifecycle |
| `internal/{context}/domain/` | Domain entities, value objects, domain events, domain errors | NO imports from infrastructure. NO imports from other contexts. Pure Go |
| `internal/{context}/app/` | Application services (use cases). Orchestrates domain + ports | May import own `domain/` and `port/`. Never imports adapters directly |
| `internal/{context}/port/` | Interface definitions (driven + driving ports) | Only interfaces and types. No implementations |
| `internal/{context}/adapter/` | Implementations of ports (PostgreSQL, Redis, HTTP, external APIs) | Implements interfaces from `port/`. May import external libraries |
| `internal/platform/` | Cross-cutting infrastructure shared across all contexts | Logging, tracing, database connection, auth, middleware |
| `internal/shared/` | Shared kernel — minimal types used across bounded contexts | TenantID, base Event interface, common value objects. Keep SMALL |
| `pkg/sdk/` | Public Go SDK for OPTRION API consumers | Stable API. Versioned. No internal dependencies |
| `migrations/` | PostgreSQL migration files in order | Sequential numbering. Never modify applied migrations |
| `config/` | Environment-specific configuration files | No secrets in files. Secrets from environment variables or vault |
| `deploy/` | Deployment artifacts (Docker, Kubernetes) | Infrastructure-as-code |
| `scripts/` | Developer convenience scripts | Not for production use |
| `tools/` | Go tool dependencies (linters, generators) | Pinned versions in go.mod |

---

## Dependency Rules

### The Dependency Rule (Inward-Only)

```
                    ┌──────────────────┐
                    │     DOMAIN       │  ← Depends on NOTHING
                    │                  │
                    │  Entities        │
                    │  Value Objects   │
                    │  Domain Events   │
                    │  Domain Errors   │
                    └────────▲─────────┘
                             │
                    ┌────────┴─────────┐
                    │      PORTS       │  ← Depends on Domain only
                    │                  │
                    │  Repository I/F  │
                    │  Service I/F     │
                    │  Publisher I/F   │
                    └────────▲─────────┘
                             │
                    ┌────────┴─────────┐
                    │   APPLICATION    │  ← Depends on Domain + Ports
                    │                  │
                    │  Use Cases       │
                    │  Commands        │
                    │  Queries         │
                    └────────▲─────────┘
                             │
                    ┌────────┴─────────┐
                    │    ADAPTERS      │  ← Depends on Domain + Ports + External libs
                    │                  │
                    │  PostgreSQL      │
                    │  Redis           │
                    │  HTTP Handlers   │
                    │  Telegram Client │
                    └──────────────────┘
```

### Import Rules Matrix

| Package | May Import | Must NEVER Import |
|---------|-----------|-------------------|
| `internal/{ctx}/domain/` | `internal/shared/domain/`, `internal/shared/types/`, `internal/shared/id/` | Any `adapter/`, any `app/`, any `port/`, any `platform/`, any other context |
| `internal/{ctx}/port/` | Own `domain/`, `internal/shared/` | Any `adapter/`, any `app/`, any `platform/`, any other context |
| `internal/{ctx}/app/` | Own `domain/`, own `port/`, `internal/shared/` | Any `adapter/`, any other context's internals |
| `internal/{ctx}/adapter/` | Own `domain/`, own `port/`, `internal/platform/`, `internal/shared/`, external libraries | Other context's `adapter/` or `app/` |
| `internal/platform/` | `internal/shared/`, external libraries | Any context's `domain/`, `app/`, `port/` |
| `internal/shared/` | Standard library only | Everything else |
| `cmd/optrion/` | Everything (this is the composition root) | N/A |

### Cross-Context Communication Rules

| Rule | Mechanism |
|------|-----------|
| Context A needs data from Context B | Context A defines an interface in its own `port/`. The composition root (`cmd/`) wires Context B's adapter to satisfy Context A's port |
| Context A reacts to Context B's state change | Context B publishes a domain event. Context A subscribes to that event type. The event bus (in `platform/eventbus/`) routes it |
| Context A commands Context B | Context A publishes a command event. Context B's handler processes it. Never direct function calls across contexts |

### Circular Dependency Prevention

| Forbidden | Why | Solution |
|-----------|-----|----------|
| `execution` imports `catalog` | Tight coupling | Execution defines its own `ComponentInfo` port. Catalog adapter satisfies it |
| `incident` imports `intelligence` | Tight coupling | Intelligence publishes events. Incident subscribes |
| `alerting` imports `notification` | Tight coupling | Alerting publishes `SendNotification` command event. Notification subscribes |
| `shared` imports any context | Shared kernel must be dependency-free | Never. If shared needs context-specific logic, it belongs in that context |

### Enforcing Dependency Rules

Use `golangci-lint` with `depguard` configuration:

```yaml
# .golangci.yml (conceptual, not implementation)
linters:
  depguard:
    rules:
      domain-isolation:
        deny:
          - pkg: "database/sql"
            desc: "Domain layer must not import database packages"
          - pkg: "net/http"
            desc: "Domain layer must not import HTTP packages"
          - pkg: "github.com/optrion/optrion/internal/*/adapter"
            desc: "Domain and app layers must not import adapters"
```

---

## Layer Architecture

### Layer Definitions

| Layer | Responsibility | Lifespan | Testability |
|-------|---------------|----------|-------------|
| **Domain** | Business rules, invariants, state transitions | Permanent (changes only when business rules change) | Unit tested. No mocks needed (pure logic) |
| **Port** | Contracts between layers. Interface definitions | Stable (changes when capabilities change) | Not tested directly (it's an interface) |
| **Application** | Orchestration of domain operations. Transaction boundaries. Event publishing | Moderate (changes when use cases change) | Unit tested with mocked ports |
| **Adapter** | Implementation of ports using specific technology | Volatile (changes when tech changes) | Integration tested against real infrastructure |
| **Transport** | HTTP/gRPC request handling, validation, serialization | Volatile (changes when API shape changes) | Integration tested with HTTP client |
| **Platform** | Infrastructure plumbing (DB pools, observability, auth) | Stable (changes when infrastructure changes) | Integration tested |

### Request Flow (Example: Create Health Check)

```
HTTP Request
    │
    ▼
[Transport Layer: rest/handler.go]
    │  - Parse HTTP request
    │  - Validate input
    │  - Extract TenantID from auth context
    │  - Construct command
    │
    ▼
[Application Layer: app/service.go]
    │  - Begin transaction (via port)
    │  - Load Component aggregate (via repository port)
    │  - Call domain method: component.AddHealthCheck(...)
    │  - Persist changes (via repository port)
    │  - Publish domain events (via event publisher port)
    │  - Commit transaction
    │
    ▼
[Domain Layer: domain/component.go]
    │  - Validate health check configuration
    │  - Enforce invariants (max checks per component, valid schedule)
    │  - Produce domain event: HealthCheckConfigured
    │  - Return result
    │
    ▼
[Port Layer: port/repository.go]
    │  - Interface: Save(ctx, component) error
    │
    ▼
[Adapter Layer: adapter/postgres/repository.go]
    │  - Implement Save() using PostgreSQL
    │  - Set tenant RLS context
    │  - Execute SQL within transaction
    │
    ▼
PostgreSQL
```

### Event Flow (Example: Health Check Executed → Score Updated → Incident Opened)

```
[Execution Context]
    │  CheckResult created
    │  Event: HealthCheckExecuted
    │
    ▼ (via EventBus)
[Intelligence Context - Event Handler]
    │  Receives HealthCheckExecuted
    │  Loads component's recent results
    │  Computes new health score
    │  Updates ComponentHealth (current state)
    │  Persists HealthScore (historical)
    │  If score < threshold: publishes HealthDegraded
    │
    ▼ (via EventBus)
[Incident Context - Event Handler]
    │  Receives HealthDegraded
    │  Computes fingerprint
    │  Checks for existing active incident (dedup)
    │  If no existing: opens new Incident
    │  Publishes IncidentOpened
    │
    ▼ (via EventBus)
[Alerting Context - Event Handler]
    │  Receives IncidentOpened
    │  Evaluates alert rules for this scope
    │  Checks cooldown state
    │  Checks maintenance windows
    │  If not suppressed: publishes AlertRuleTriggered
    │
    ▼ (via EventBus)
[Notification Context - Event Handler]
    │  Receives AlertRuleTriggered
    │  Loads channel configurations
    │  Renders notification template
    │  Sends via Telegram/Webhook
    │  Records AlertHistory
    │  Publishes AlertSent
```

### Transaction Boundaries

| Rule | Explanation |
|------|-------------|
| One aggregate per transaction | Never load and save two aggregate roots in one DB transaction |
| Events published within aggregate transaction | Domain event + aggregate save happen atomically (outbox pattern) |
| Cross-aggregate consistency is eventual | If Incident creation depends on HealthScore, it happens in a separate transaction triggered by an event |
| Read models (component_health) updated asynchronously | The "current state" table is updated by an event handler, not in the same transaction as check_results |

---

## Interface Catalog

### Shared Kernel Interfaces

```go
// internal/shared/domain/event.go
type DomainEvent interface {
    EventType() string
    OccurredAt() time.Time
    AggregateID() string
    TenantID() TenantID
}

// internal/shared/domain/aggregate.go
type AggregateRoot interface {
    ID() string
    TenantID() TenantID
    Events() []DomainEvent
    ClearEvents()
    Version() int
}

// internal/shared/domain/clock.go
type Clock interface {
    Now() time.Time
}
```

### Tenant Context Ports

```go
// internal/tenant/port/repository.go
type TenantRepository interface {
    FindByID(ctx context.Context, id TenantID) (*domain.Tenant, error)
    FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
    FindByAPIKeyHash(ctx context.Context, hash string) (*domain.Tenant, error)
    Save(ctx context.Context, tenant *domain.Tenant) error
}

type APIKeyRepository interface {
    FindByHash(ctx context.Context, hash string) (*domain.APIKey, error)
    Save(ctx context.Context, key *domain.APIKey) error
    ListByTenant(ctx context.Context, tenantID TenantID) ([]*domain.APIKey, error)
}

// internal/tenant/port/hasher.go
type KeyHasher interface {
    Hash(key string) string
    GenerateKey() (fullKey string, prefix string, err error)
}

// internal/tenant/port/quota.go
type QuotaChecker interface {
    CheckComponentQuota(ctx context.Context, tenantID TenantID) error
    CheckCheckQuota(ctx context.Context, tenantID TenantID) error
    IncrementUsage(ctx context.Context, tenantID TenantID, resource string) error
}
```

### Catalog Context Ports

```go
// internal/catalog/port/repository.go
type ProductRepository interface {
    FindByID(ctx context.Context, tenantID TenantID, id string) (*domain.Product, error)
    ListByTenant(ctx context.Context, tenantID TenantID) ([]*domain.Product, error)
    Save(ctx context.Context, product *domain.Product) error
}

type EnvironmentRepository interface {
    FindByID(ctx context.Context, tenantID TenantID, id string) (*domain.Environment, error)
    ListByProduct(ctx context.Context, tenantID TenantID, productID string) ([]*domain.Environment, error)
    Save(ctx context.Context, env *domain.Environment) error
}

type ComponentRepository interface {
    FindByID(ctx context.Context, tenantID TenantID, id string) (*domain.Component, error)
    ListByEnvironment(ctx context.Context, tenantID TenantID, envID string) ([]*domain.Component, error)
    Save(ctx context.Context, component *domain.Component) error
    Delete(ctx context.Context, tenantID TenantID, id string) error
}
```

### Execution Context Ports

```go
// internal/execution/port/checker.go
type HealthChecker interface {
    Check(ctx context.Context, target domain.CheckTarget) (*domain.CheckOutcome, error)
    Type() string  // "http", "tcp", "dns"
}

// internal/execution/port/repository.go
type CheckResultRepository interface {
    Save(ctx context.Context, result *domain.CheckResult) error
    SaveBatch(ctx context.Context, results []*domain.CheckResult) error
    FindRecent(ctx context.Context, tenantID TenantID, checkID string, limit int) ([]*domain.CheckResult, error)
    FindByComponentInWindow(ctx context.Context, tenantID TenantID, componentID string, window time.Duration) ([]*domain.CheckResult, error)
}

// internal/execution/port/scheduler.go
type Scheduler interface {
    Schedule(checkID string, interval time.Duration, fn func()) error
    Unschedule(checkID string) error
    Reschedule(checkID string, interval time.Duration) error
}

// internal/execution/port/event_publisher.go
type EventPublisher interface {
    Publish(ctx context.Context, events ...shared.DomainEvent) error
}
```

### Intelligence Context Ports

```go
// internal/intelligence/port/repository.go
type ComponentHealthRepository interface {
    FindByComponent(ctx context.Context, tenantID TenantID, componentID string) (*domain.ComponentHealth, error)
    FindByEnvironment(ctx context.Context, tenantID TenantID, envID string) ([]*domain.ComponentHealth, error)
    Upsert(ctx context.Context, health *domain.ComponentHealth) error
}

type HealthScoreRepository interface {
    Save(ctx context.Context, score *domain.HealthScore) error
    FindBySubject(ctx context.Context, tenantID TenantID, subjectType string, subjectID string, window time.Duration) ([]*domain.HealthScore, error)
}

type BaselineRepository interface {
    FindByComponent(ctx context.Context, tenantID TenantID, componentID string, metricType string) (*domain.Baseline, error)
    Save(ctx context.Context, baseline *domain.Baseline) error
}

// internal/intelligence/port/metric_store.go
type MetricStore interface {
    Save(ctx context.Context, snapshot *domain.MetricSnapshot) error
    FindByComponent(ctx context.Context, tenantID TenantID, componentID string, metricType string, window time.Duration) ([]*domain.MetricSnapshot, error)
    Aggregate(ctx context.Context, tenantID TenantID, componentID string, metricType string, window time.Duration) (*domain.MetricAggregate, error)
}
```

### Incident Context Ports

```go
// internal/incident/port/repository.go
type IncidentRepository interface {
    FindByID(ctx context.Context, tenantID TenantID, id string) (*domain.Incident, error)
    FindActiveByFingerprint(ctx context.Context, tenantID TenantID, fingerprint string) (*domain.Incident, error)
    FindActiveByTenant(ctx context.Context, tenantID TenantID, opts QueryOpts) ([]*domain.Incident, error)
    FindByComponent(ctx context.Context, tenantID TenantID, componentID string, opts QueryOpts) ([]*domain.Incident, error)
    Save(ctx context.Context, incident *domain.Incident) error
}

type IncidentEventRepository interface {
    FindByIncident(ctx context.Context, incidentID string) ([]*domain.IncidentEvent, error)
    Append(ctx context.Context, event *domain.IncidentEvent) error
}
```

### Alerting Context Ports

```go
// internal/alerting/port/repository.go
type AlertRuleRepository interface {
    FindByID(ctx context.Context, tenantID TenantID, id string) (*domain.AlertRule, error)
    FindActiveByScope(ctx context.Context, tenantID TenantID, scopeType string, scopeID string) ([]*domain.AlertRule, error)
    FindAllActive(ctx context.Context, tenantID TenantID) ([]*domain.AlertRule, error)
    Save(ctx context.Context, rule *domain.AlertRule) error
}

type AlertHistoryRepository interface {
    Save(ctx context.Context, history *domain.AlertHistory) error
    FindByTenant(ctx context.Context, tenantID TenantID, window time.Duration, opts QueryOpts) ([]*domain.AlertHistory, error)
}

type MaintenanceWindowRepository interface {
    FindActiveForScope(ctx context.Context, tenantID TenantID, scopeType string, scopeID string, at time.Time) (*domain.MaintenanceWindow, error)
    Save(ctx context.Context, window *domain.MaintenanceWindow) error
}

// internal/alerting/port/cooldown_store.go
type CooldownStore interface {
    IsInCooldown(ctx context.Context, ruleID string) (bool, error)
    SetCooldown(ctx context.Context, ruleID string, duration time.Duration) error
    ClearCooldown(ctx context.Context, ruleID string) error
}
```

### Notification Context Ports

```go
// internal/notification/port/sender.go
type NotificationSender interface {
    Send(ctx context.Context, channel *domain.Channel, message *domain.Message) error
    Type() string  // "telegram", "webhook", "email"
}

// internal/notification/port/repository.go
type ChannelRepository interface {
    FindByID(ctx context.Context, tenantID TenantID, id string) (*domain.Channel, error)
    FindByIDs(ctx context.Context, tenantID TenantID, ids []string) ([]*domain.Channel, error)
    ListByTenant(ctx context.Context, tenantID TenantID) ([]*domain.Channel, error)
    Save(ctx context.Context, channel *domain.Channel) error
}

// internal/notification/port/template_renderer.go
type TemplateRenderer interface {
    Render(templateName string, data map[string]any) (string, error)
}
```

### AI Context Ports (Future)

```go
// internal/ai/port/provider.go
type AIProvider interface {
    AnalyzeIncident(ctx context.Context, input *domain.IncidentContext) (*domain.Analysis, error)
    PredictDegradation(ctx context.Context, input *domain.PredictionInput) (*domain.Prediction, error)
    ExplainTrend(ctx context.Context, input *domain.TrendContext) (*domain.Explanation, error)
}

// internal/ai/port/feature_store.go
type FeatureStore interface {
    GetComponentFeatures(ctx context.Context, tenantID TenantID, componentID string, window time.Duration) (*domain.FeatureVector, error)
    GetIncidentFeatures(ctx context.Context, tenantID TenantID, incidentID string) (*domain.IncidentFeatureVector, error)
}

// internal/ai/port/repository.go
type AnalysisRepository interface {
    Save(ctx context.Context, analysis *domain.Analysis) error
    FindBySubject(ctx context.Context, tenantID TenantID, subjectType string, subjectID string) ([]*domain.Analysis, error)
}

type RecommendationRepository interface {
    Save(ctx context.Context, rec *domain.Recommendation) error
    FindPendingByTenant(ctx context.Context, tenantID TenantID) ([]*domain.Recommendation, error)
    UpdateFeedback(ctx context.Context, id string, feedback string) error
}
```

### Platform Interfaces

```go
// internal/platform/eventbus/bus.go
type EventBus interface {
    Publish(ctx context.Context, events ...shared.DomainEvent) error
    Subscribe(eventType string, handler EventHandler) error
    Unsubscribe(eventType string, handler EventHandler) error
}

type EventHandler interface {
    Handle(ctx context.Context, event shared.DomainEvent) error
    Name() string  // For logging/tracing
}

// internal/platform/database/tx.go
type TransactionManager interface {
    WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// internal/platform/auth/middleware.go
type Authenticator interface {
    Authenticate(ctx context.Context, token string) (*AuthResult, error)
}

type AuthResult struct {
    TenantID TenantID
    Scopes   []string
    KeyID    string
}

// internal/platform/cache/ratelimit.go
type RateLimiter interface {
    Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}
```

---

## Event Architecture

### Event Bus Design

```
┌──────────────────────────────────────────────────────────────────┐
│                        EVENT BUS                                   │
│                                                                    │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                    IN-PROCESS BUS                           │  │
│  │                                                            │  │
│  │  Publisher ──▶ [Event Router] ──▶ Handler 1                │  │
│  │                      │         ──▶ Handler 2                │  │
│  │                      │         ──▶ Handler 3                │  │
│  │                      │                                      │  │
│  │                      ▼                                      │  │
│  │               [Outbox Writer]                               │  │
│  │                      │                                      │  │
│  │                      ▼                                      │  │
│  │           ┌─────────────────────┐                          │  │
│  │           │  event_outbox table │                          │  │
│  │           └──────────┬──────────┘                          │  │
│  │                      │                                      │  │
│  │                      ▼                                      │  │
│  │              [Outbox Poller]                                │  │
│  │                      │                                      │  │
│  │                      ▼                                      │  │
│  │         [Guaranteed Delivery Handlers]                      │  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
└──────────────────────────────────────────────────────────────────┘
```

### Event Delivery Guarantees

| Handler Type | Delivery | Use Case |
|-------------|----------|----------|
| **Synchronous** | Same goroutine, same transaction | Updating read models that must be immediately consistent |
| **Async (in-process)** | Separate goroutine, best-effort | Dashboard updates, metric computations |
| **Async (outbox)** | At-least-once, survives crashes | Notifications, incident creation, cross-context commands |

### Event Routing Table

| Event | Published By | Sync Handlers | Async Handlers (Outbox) |
|-------|-------------|---------------|------------------------|
| `HealthCheckExecuted` | Execution | — | Intelligence (score), Intelligence (metric store) |
| `HealthCheckFailed` | Execution | — | Incident (detection), Intelligence (streak tracking) |
| `HealthCheckRecovered` | Execution | — | Incident (auto-resolve check), Intelligence (reset) |
| `HealthScoreComputed` | Intelligence | ComponentHealth (update current state) | Dashboard (push update) |
| `HealthDegraded` | Intelligence | — | Incident (detection) |
| `AnomalyDetected` | Intelligence | — | Incident (detection) |
| `IncidentOpened` | Incident | — | Alerting (rule evaluation), Audit (log) |
| `IncidentAcknowledged` | Incident | — | Alerting (stop escalation), Audit (log) |
| `IncidentResolved` | Incident | — | Alerting (recovery notification), Audit (log) |
| `AlertRuleTriggered` | Alerting | — | Notification (send) |
| `AlertSuppressed` | Alerting | — | Audit (log) |
| `NotificationSent` | Notification | — | AlertHistory (record delivery) |
| `NotificationFailed` | Notification | — | AlertHistory (record failure), Retry queue |
| `ComponentRegistered` | Catalog | — | Execution (start scheduling), Intelligence (init baseline) |
| `ComponentDeregistered` | Catalog | — | Execution (stop scheduling), Intelligence (archive) |

### Event Schema

```go
// Every event follows this envelope structure
type EventEnvelope struct {
    ID            string    // Unique event ID (UUID v7)
    Type          string    // "incident.opened", "health_check.executed"
    TenantID      string    // Tenant scope
    AggregateType string    // "incident", "check_result"
    AggregateID   string    // Aggregate root ID
    Payload       []byte    // JSON-encoded event-specific data
    SchemaVersion int       // For backward compatibility
    OccurredAt    time.Time // When the business event happened
    PublishedAt   time.Time // When it entered the outbox
}
```

### Event Naming Convention

```
{context}.{aggregate}.{past_tense_verb}

Examples:
  execution.check_result.recorded
  intelligence.health_score.computed
  intelligence.anomaly.detected
  incident.incident.opened
  incident.incident.resolved
  alerting.alert_rule.triggered
  notification.delivery.completed
  notification.delivery.failed
  catalog.component.registered
  tenant.tenant.onboarded
```

### Idempotency

Every event handler must be idempotent. The same event delivered twice must produce the same result.

| Strategy | When to Use |
|----------|-------------|
| Idempotency key check | Before creating new entities (check if already exists) |
| Upsert on natural key | When updating read models (component_health) |
| Event ID deduplication | Store processed event IDs in a set (Redis or DB) for 24h |

---

## Configuration Strategy

### Configuration Hierarchy (Lowest to Highest Priority)

```
1. Compiled defaults (in config struct tags)
2. config/{environment}.yaml
3. Environment variables (OPTRION_*)
4. Command-line flags (--port, --db-url)
```

Higher priority overrides lower.

### Configuration Structure

```go
// config/config.go (conceptual structure, not implementation)
type Config struct {
    App      AppConfig
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Auth     AuthConfig
    Alerting AlertingConfig
    Observe  ObservabilityConfig
}

type AppConfig struct {
    Name        string        // "optrion"
    Environment string        // "local", "development", "staging", "production"
    LogLevel    string        // "debug", "info", "warn", "error"
    ShutdownTimeout time.Duration
}

type ServerConfig struct {
    Host         string
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

type DatabaseConfig struct {
    Host            string
    Port            int
    Name            string
    User            string        // From env/vault in production
    Password        string        // NEVER in config file. Always from env/vault
    SSLMode         string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

type RedisConfig struct {
    Addr     string
    Password string            // From env/vault
    DB       int
    PoolSize int
}
```

### Environment-Specific Configuration

| Setting | Local | Development | Staging | Production |
|---------|-------|-------------|---------|-----------|
| Log level | debug | debug | info | warn |
| DB host | localhost | dev-db.internal | staging-db.internal | prod-db.internal |
| DB SSL | disable | require | require | verify-full |
| Redis | localhost:6379 | dev-redis.internal | staging-redis | prod-redis (cluster) |
| Check interval minimum | 5s | 15s | 30s | 30s |
| Rate limit | disabled | relaxed | production-like | enforced |
| Tracing sample rate | 100% | 100% | 10% | 1% |

### Secret Management

| Environment | Secret Source | Mechanism |
|-------------|-------------|-----------|
| Local | `.env` file (gitignored) | Loaded by config library |
| Development | Environment variables | Set in docker-compose |
| Staging | Cloud secret manager | Injected as env vars by orchestrator |
| Production | HashiCorp Vault or cloud-native (AWS Secrets Manager, GCP Secret Manager) | Sidecar or init container injects env vars |

### Secrets That Must NEVER Be in Config Files

- Database passwords
- Redis passwords
- API key signing secrets
- Telegram bot tokens
- Webhook signing keys
- Encryption keys for notification channel configs
- JWT signing keys

### Configuration Validation

Application MUST validate all configuration at startup and **fail fast** with clear error messages if:
- Required values are missing
- Values are out of valid range
- Database is unreachable
- Redis is unreachable (if enabled)

---

## Observability Strategy

### How OPTRION Monitors OPTRION

```
┌──────────────────────────────────────────┐
│            OPTRION PROCESS                │
│                                          │
│  [Business Logic] ──▶ [Structured Logs]  │
│         │                                │
│         ▼                                │
│  [Prometheus Metrics] ◀── [HTTP Server]  │
│         │                                │
│         ▼                                │
│  [OpenTelemetry Traces] ──▶ [Exporter]   │
│                                          │
│  [/healthz] [/readyz] [/metrics]         │
│                                          │
└──────────────────────────────────────────┘
         │              │             │
         ▼              ▼             ▼
    [Log Sink]    [Prometheus]    [Jaeger/OTLP]
```

### Structured Logging

| Aspect | Decision |
|--------|----------|
| Library | `log/slog` (standard library, Go 1.21+) |
| Format | JSON in production, text in local development |
| Level | Configurable per environment |
| Context fields | `tenant_id`, `request_id`, `trace_id`, `component_id` on every log line |
| Sensitive data | NEVER log API keys, passwords, PII, full request bodies |

**Standard log fields on every request:**

```
{
  "time": "2026-05-29T10:15:30Z",
  "level": "INFO",
  "msg": "health check executed",
  "tenant_id": "abc123",
  "request_id": "req-xyz",
  "trace_id": "trace-789",
  "component_id": "comp-456",
  "check_id": "chk-012",
  "latency_ms": 142,
  "status": "healthy"
}
```

### Metrics (Prometheus)

| Metric Name | Type | Labels | Purpose |
|-------------|------|--------|---------|
| `optrion_http_requests_total` | Counter | method, path, status_code, tenant_id | Request volume |
| `optrion_http_request_duration_seconds` | Histogram | method, path | Latency distribution |
| `optrion_health_checks_executed_total` | Counter | tenant_id, check_type, result | Check execution volume |
| `optrion_health_check_duration_seconds` | Histogram | check_type | Check execution latency |
| `optrion_health_scores_computed_total` | Counter | tenant_id, subject_type | Scoring volume |
| `optrion_incidents_opened_total` | Counter | tenant_id, severity | Incident creation rate |
| `optrion_alerts_sent_total` | Counter | tenant_id, channel_type, status | Alert delivery |
| `optrion_event_outbox_depth` | Gauge | status | Outbox queue depth |
| `optrion_event_outbox_processing_duration_seconds` | Histogram | event_type | Event processing latency |
| `optrion_db_connections_active` | Gauge | pool_name | Connection pool utilization |
| `optrion_db_query_duration_seconds` | Histogram | query_name | Query performance |
| `optrion_tenant_component_count` | Gauge | tenant_id | Per-tenant resource usage |

### Distributed Tracing (OpenTelemetry)

| Span | Created At | Key Attributes |
|------|-----------|----------------|
| `http.request` | HTTP middleware | method, path, status, tenant_id |
| `db.query` | Database adapter | query.name, table, tenant_id |
| `event.publish` | Event bus | event.type, aggregate.id |
| `event.handle` | Event handler | event.type, handler.name |
| `health_check.execute` | Checker adapter | check.type, target.host, result |
| `notification.send` | Notification sender | channel.type, delivery.status |
| `redis.command` | Redis adapter | command, key_prefix |

### Health Endpoints

| Endpoint | Purpose | Checks |
|----------|---------|--------|
| `GET /healthz` | Liveness probe. Is the process alive? | Always returns 200 if HTTP server is responding |
| `GET /readyz` | Readiness probe. Can it serve traffic? | PostgreSQL ping, Redis ping, outbox worker running |
| `GET /metrics` | Prometheus scrape endpoint | Exposes all registered metrics |

### Self-Monitoring Alert Rules (Meta-Monitoring)

| Condition | Alert |
|-----------|-------|
| `optrion_event_outbox_depth{status="pending"} > 10000` | Outbox backing up — processing stalled |
| `rate(optrion_health_checks_executed_total[5m]) == 0` | No checks executing — scheduler may be dead |
| `optrion_db_connections_active / optrion_db_connections_max > 0.8` | Connection pool near exhaustion |
| `rate(optrion_http_requests_total{status_code=~"5.."}[5m]) > 10` | Elevated 5xx errors |
| `histogram_quantile(0.99, optrion_http_request_duration_seconds) > 5` | P99 latency exceeding 5s |

---

## Testing Strategy

### Testing Pyramid

```
        ╱╲
       ╱  ╲        E2E Tests (few, slow, fragile)
      ╱────╲       Contract Tests
     ╱      ╲      Integration Tests (adapters, DB, Redis)
    ╱────────╲     Application Tests (use cases with mocked ports)
   ╱          ╲    Domain Tests (pure logic, no mocks)
  ╱────────────╲
```

### Domain Tests (Most Tests Live Here)

| What | How | Example |
|------|-----|---------|
| Aggregate invariants | Direct instantiation and method calls | `component.AddHealthCheck()` rejects invalid schedule |
| State machine transitions | Call transition methods, assert new state | `incident.Acknowledge()` only works from `opened` status |
| Value object validation | Constructor rejects invalid values | `NewScore(101)` returns error |
| Domain event emission | Call aggregate method, assert emitted events | `incident.Open()` emits `IncidentOpened` |
| Fingerprint computation | Deterministic input → deterministic output | Same component + check + failure → same fingerprint |

**Properties:**
- Zero external dependencies
- Zero mocks (pure domain logic needs no mocks)
- Sub-millisecond execution
- 100% of business rules covered

### Application Tests (Use Case Tests)

| What | How | Example |
|------|-----|---------|
| Use case orchestration | Mock all ports, verify interactions | `CreateComponent` calls repo.Save then publishes event |
| Command validation | Invalid commands rejected before domain logic | Missing tenant_id returns validation error |
| Transaction boundary | Verify transactional consistency | Save + publish happen atomically (or both fail) |
| Authorization | Verify tenant scoping | TenantA cannot access TenantB's components |

**Properties:**
- All ports are mocked (generated mocks from interfaces)
- Tests verify the orchestration logic, not infrastructure
- Fast (<10ms per test)

### Integration Tests (Adapter Tests)

| What | How | Example |
|------|-----|---------|
| PostgreSQL repositories | Real PostgreSQL (Docker) | Save component → FindByID returns same data |
| Redis adapters | Real Redis (Docker) | SetCooldown → IsInCooldown returns true before TTL |
| HTTP handlers | httptest server + real request | POST /api/v1/components returns 201 with valid body |
| Event bus | In-process bus with test handlers | Publish event → handler receives it |
| Telegram sender | Mock HTTP server mimicking Telegram API | Send message → correct HTTP request shape |

**Properties:**
- Require Docker (testcontainers-go)
- Run in parallel with isolated databases (random schema per test)
- Slower (100ms-1s per test)
- Tagged: `//go:build integration`

### Repository Tests (Specific Pattern)

Every repository implementation gets a standardized test suite:

| Test | Purpose |
|------|---------|
| `TestSave_NewEntity` | Insert works |
| `TestSave_UpdateEntity` | Upsert works |
| `TestFindByID_Exists` | Returns entity when exists |
| `TestFindByID_NotFound` | Returns specific error when not found |
| `TestFindByID_WrongTenant` | Returns not found for other tenant's data (CRITICAL) |
| `TestList_FiltersByTenant` | Only returns current tenant's data |
| `TestList_Pagination` | Pagination works correctly |
| `TestDelete_RemovesEntity` | Soft/hard delete works |
| `TestConcurrentSave` | Optimistic locking prevents lost updates |

**The most important test:**

> `TestFindByID_WrongTenant` — Create entity for TenantA. Query with TenantB's context. MUST return "not found". This test prevents the #1 security vulnerability in multi-tenant systems.

### Contract Tests (Between Contexts)

| What | How | Example |
|------|-----|---------|
| Event schema contracts | Producer publishes → consumer can deserialize | Intelligence can deserialize `HealthCheckExecuted` events from Execution |
| API response contracts | Handler returns → SDK client can parse | SDK client successfully parses `/api/v1/incidents` response |

**Properties:**
- Verify that bounded contexts agree on event shapes
- Catch breaking changes in event payloads before deployment
- Run as part of CI on every PR

### Test File Naming Convention

```
internal/incident/domain/incident.go           # Production code
internal/incident/domain/incident_test.go      # Domain unit tests

internal/incident/app/service.go               # Production code
internal/incident/app/service_test.go          # Application tests (mocked ports)

internal/incident/adapter/postgres/repository.go
internal/incident/adapter/postgres/repository_integration_test.go  # Integration test (tagged)
```

### Test Infrastructure

| Tool | Purpose |
|------|---------|
| `testing` (stdlib) | Test runner |
| `testify/assert` | Assertions (or go-cmp) |
| `testify/mock` or `moq` | Interface mock generation |
| `testcontainers-go` | Docker containers for integration tests |
| `goleak` | Goroutine leak detection |
| `go test -race` | Race condition detection (always enabled in CI) |

---

## Architecture Review

### Future Bottlenecks

| Bottleneck | When It Hits | Signals | Mitigation Path |
|-----------|-------------|---------|-----------------|
| **Event bus throughput** | 100+ tenants, high check frequency | Outbox depth growing, event latency increasing | Partition outbox by tenant_id range. Multiple outbox workers. Eventually: external message broker (NATS) |
| **Scheduler goroutine count** | 50K+ active health checks | Memory usage climbing, GC pressure | Worker pool pattern with bounded concurrency. Job queue instead of per-check goroutine |
| **Single binary memory** | All contexts loaded in one process at 1000 tenants | RSS > 4GB, GC pauses > 10ms | Extract high-throughput contexts (execution, intelligence) into separate processes. Communicate via event bus |
| **Database connection pressure** | Concurrent dashboard reads + check result writes | Connection pool saturation, query timeouts | Read replica for dashboard queries. PgBouncer. Eventually: CQRS with separate read DB |
| **Health score computation** | Thousands of components recomputing simultaneously | Score computation latency > 5s, stale dashboard | Stagger computations (jitter). Pre-compute and cache. Eventually: stream processing |

### Package Smells to Watch For

| Smell | Detection | Fix |
|-------|----------|-----|
| `shared/` package growing too large | More than 15 files in shared/ | Split into focused sub-packages. Question whether types truly need sharing |
| Domain importing infrastructure | `import "database/sql"` in domain package | Move the offending code to an adapter. Domain must remain pure |
| Circular dependency between contexts | Compiler error or linter warning | Introduce an event or invert the dependency through an interface |
| God service in app layer | Application service file > 500 lines | Split into focused use-case-specific services |
| Adapter logic leaking into ports | Port interface has method signatures that expose PostgreSQL details | Port interfaces should use domain types only |
| Fat event payloads | Event payload > 10KB | Events should reference entities by ID, not embed full state |
| Shared mutable state between handlers | Data race detected by `-race` | All shared state through channels or sync primitives. Prefer immutable event passing |

### Scalability Concerns

| Concern | Current Design | 100-Tenant Threshold | 1000-Tenant Threshold |
|---------|---------------|---------------------|----------------------|
| Check scheduling | In-process ticker per check | Worker pool with job queue | Distributed scheduler (separate service) |
| Event processing | Single outbox poller | Partitioned outbox (multiple pollers) | External event broker (NATS JetStream) |
| API serving | Single HTTP server | Read replica for queries | API gateway + multiple instances |
| Score computation | On every check result | Debounced (max once per 15s per component) | Stream processing (separate service) |
| Dashboard data | Direct PostgreSQL query | Cached in Redis (10s TTL) | Dedicated read model service |

### Architecture Decisions Record (ADR) Summary

| Decision | Status | Rationale |
|----------|--------|-----------|
| Modular monolith (single binary) | Accepted | Simplifies deployment, debugging, and local development. Extract later when justified |
| In-process event bus with outbox | Accepted | Provides reliability without external broker dependency |
| Interface-per-port (not one god interface) | Accepted | Enables focused mocking in tests. Clear responsibility |
| UUID v7 for all IDs | Accepted | Time-sortable, globally unique, no sequence contention |
| Shared kernel kept minimal | Accepted | Only TenantID, DomainEvent interface, and common value objects |
| No ORM | Accepted | Use `sqlx` or raw `database/sql`. ORMs hide query complexity |
| Separate read model (component_health) | Accepted | Avoids aggregation queries on hot dashboard path |
| Context-specific error types | Accepted | Each context defines its own errors. No global error catalog |

---

## Final Recommendations

### Before Writing First Line of Code

1. **Set up the monorepo skeleton** — Create all folder structures with empty `.go` files containing only `package` declarations. Verify compilation.

2. **Define shared kernel types first** — `TenantID`, `DomainEvent`, `AggregateRoot`, `Score`, `Severity`. These are imported everywhere.

3. **Define port interfaces before adapters** — Write all repository and service interfaces. This is the contract. Implementation comes second.

4. **Set up dependency linting** — Configure `depguard` in `golangci-lint` to enforce layer rules from Day 1. Violations caught in CI.

5. **Set up integration test infrastructure** — Docker Compose for PostgreSQL + Redis. `testcontainers-go` wrappers. Verify tests run before writing business logic.

### Implementation Order (Context by Context)

| Phase | Context | Why This Order |
|-------|---------|---------------|
| 1 | Shared kernel + Platform (DB, auth, event bus) | Foundation for everything |
| 2 | Tenant | Required by all other contexts (authentication, tenant_id) |
| 3 | Catalog | Defines what we monitor (components, environments) |
| 4 | Execution | Produces the raw data (check results) |
| 5 | Intelligence | Transforms raw data into scores |
| 6 | Incident | Detects problems from scores |
| 7 | Alerting + Notification | Notifies humans about problems |
| 8 | AI (future) | Enhances intelligence with ML |

### Critical Go Architecture Rules

| Rule | Enforcement |
|------|------------|
| No business logic in HTTP handlers | Code review + handler size limit (< 30 lines) |
| No SQL in domain layer | `depguard` linter rule |
| Every exported function has a context.Context as first param | `golangci-lint` custom rule |
| Every repository method takes TenantID | Interface definition enforces it |
| No `panic()` in production paths | `gocritic` linter rule |
| Errors are wrapped with context | `errwrap` linter |
| All goroutines tracked via errgroup or WaitGroup | `goleak` in tests |
| No init() functions (except for metric registration) | Code review |
| No global mutable state | Code review + `-race` in CI |
| All time from Clock interface (not time.Now() directly) | Required for deterministic testing |

### Dependency Choices (Libraries)

| Need | Recommended | Why |
|------|-------------|-----|
| HTTP router | `net/http` + `chi` or `gorilla/mux` | Standard-compatible. No magic |
| Database | `pgx/v5` + `scany` | Best PostgreSQL driver for Go. Connection pooling built-in |
| Migrations | `golang-migrate/migrate` | File-based, no ORM dependency |
| Redis | `go-redis/redis/v9` | Full-featured, connection pooling |
| Config | `viper` or `koanf` | Multi-source config with env var support |
| Logging | `log/slog` (stdlib) | Standard library. No external dependency |
| Metrics | `prometheus/client_golang` | Industry standard |
| Tracing | `go.opentelemetry.io/otel` | Vendor-neutral standard |
| Testing | `testify` + `testcontainers-go` | Assertions + Docker containers |
| Linting | `golangci-lint` | Aggregates all Go linters |
| Mocking | `vektra/mockery` or `matryer/moq` | Generate mocks from interfaces |
| UUID | `google/uuid` (with v7 support) | Standard UUID library |
| Validation | `go-playground/validator` | Struct tag validation for DTOs |

### What NOT to Use

| Don't Use | Why | Use Instead |
|-----------|-----|-------------|
| GORM / any ORM | Hides query complexity. Breaks at scale. N+1 problems | Raw SQL with `pgx` + `scany` |
| Gin / Echo (heavy frameworks) | Unnecessary abstraction over net/http | `chi` router or standard `http.ServeMux` |
| Cobra (for a single command) | Overkill. OPTRION has one entry point | Simple `flag` package or none |
| Wire / Uber Fx (DI frameworks) | Magic. Hard to debug. Unnecessary for this size | Manual dependency injection in `cmd/optrion/main.go` |
| Protocol Buffers (in Phase 1) | No microservices yet. No need for IDL | JSON over HTTP. Add proto when extracting services |
| Kafka / NATS (in Phase 1) | Operational complexity not justified | In-process event bus with PostgreSQL outbox |

---

*End of Go Codebase Architecture*
