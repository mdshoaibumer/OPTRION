# OPTRION — Implementation Strategy

**Author:** Engineering Director  
**Date:** 2026-05-29  
**Version:** 1.0  
**Status:** Execution Plan — Ready to Begin

---

## Critical Path

### Dependencies Graph

```
[Config + Logging]
       │
       ▼
[Database + Migrations]
       │
       ▼
[Shared Kernel (TenantID, Events, Clock)]
       │
       ├──────────────────────────────────┐
       ▼                                  ▼
[Tenant Context]                   [Event Bus + Outbox]
       │
       ▼
[Catalog Context]
       │
       ▼
[Execution Context (Health Checks)]
       │
       ▼
[Intelligence Context (Scoring)]
       │
       ▼
[Incident Context (Detection)]
       │
       ▼
[Alerting + Notification Context]
       │
       ▼
[REST API Layer]
       │
       ▼
[Embedded Dashboard]
       │
       ▼
[GymFlow Track Integration]
```

### The Critical Path (Longest Dependency Chain)

```
Config → DB → Tenant → Catalog → Execution → Intelligence → Incident → Alerting → API → Dashboard
```

**10 sequential steps.** Nothing can be parallelized until the foundation is solid.

### Key Risks on the Critical Path

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Database schema changes mid-build | Cascading rework across all contexts | Finalize migrations for core tables BEFORE writing any repository code |
| Event bus design thrash | Every context depends on it | Implement simplest possible bus first (sync dispatch). Outbox comes second |
| Over-engineering early contexts | Delays everything downstream | Tenant and Catalog are CRUD. Don't overthink them. Intelligence and Incident are where complexity belongs |
| Burnout from after-work development | Project abandonment | Milestones must deliver visible value. Never more than 2 weeks without a working demo |
| Scope creep from GymFlow needs | Delays MVP | GymFlow gets what the platform provides. No custom features |

### Solo Engineer Constraints

| Constraint | Impact on Plan |
|-----------|---------------|
| ~2-3 hours/day after work | Each "milestone" = 1-2 weeks calendar time |
| No code review partner | Must rely on linters, tests, and architecture docs |
| No ops team | Deployment must be trivially simple (Docker Compose on VPS) |
| Learning Go concurrently | Early milestones will be slower. Factor 1.5x for learning curve |
| Single person = single thread | Cannot parallelize. Strict sequential order matters |

---

## Milestones

### Milestone 1: Foundation

**Goal:** Bootable Go application with configuration, logging, database connectivity, and HTTP server skeleton.

**Deliverables:**
- Go module initialized (`go.mod`)
- Configuration loading (YAML + env vars)
- Structured logging (`slog`)
- PostgreSQL connection with pool
- HTTP server starts and responds to `/healthz`
- Docker Compose for local PostgreSQL + Redis
- Makefile with common commands
- Linter configuration (`.golangci.yml`)

**Dependencies:** None (this is the root)

**Acceptance Criteria:**
- `make run` starts the application
- `GET /healthz` returns 200
- Application connects to PostgreSQL and logs success
- Application loads config from `config/local.yaml`
- `make lint` passes with zero warnings
- `make test` runs (even if no tests yet)

---

### Milestone 2: Database + Migrations

**Goal:** All Phase 1 database tables created with proper partitioning, RLS policies, and seed data.

**Deliverables:**
- Migration tool integrated (`golang-migrate`)
- All core schema migrations written and applied
- RLS policies on all tenant-scoped tables
- Seed script for development (1 tenant, 1 product, 1 environment, 2 components)
- Partition creation for `check_results` (first 3 months)

**Dependencies:** Milestone 1 (database connection)

**Acceptance Criteria:**
- `make migrate-up` applies all migrations cleanly
- `make migrate-down` rolls back cleanly
- RLS prevents cross-tenant access (verified manually in psql)
- Seed data creates a working GymFlow tenant
- Partitioned tables exist and accept inserts

---

### Milestone 3: Shared Kernel + Event Bus

**Goal:** Cross-cutting types and in-process event bus operational.

**Deliverables:**
- `TenantID` value object
- `DomainEvent` interface
- `AggregateRoot` base type
- Common value objects (`Score`, `Severity`, `HealthStatus`)
- UUID v7 generation utility
- `Clock` interface (real + fake for tests)
- In-process event bus (synchronous dispatch)
- Event bus subscriber registration
- Outbox table writer (write events to outbox within transaction)
- Outbox poller (background worker that reads and dispatches)

**Dependencies:** Milestone 2 (outbox table exists)

**Acceptance Criteria:**
- Event published → handler receives it synchronously
- Event published within transaction → appears in outbox table
- Outbox poller picks up events and dispatches to handlers
- Fake clock allows deterministic time in tests
- All shared types compile and are importable by other packages

---

### Milestone 4: Tenant Context

**Goal:** Tenant registration, API key creation, and authentication middleware working end-to-end.

**Deliverables:**
- Tenant domain entity + repository interface
- Tenant PostgreSQL repository implementation
- API key generation (with hash storage)
- API key authentication middleware
- Tenant resolution middleware (sets `tenant_id` in context + RLS session variable)
- Rate limiting middleware (per-tenant, using in-memory counter for now)
- REST endpoints: create tenant, create API key, get tenant info
- Integration tests verifying tenant isolation

**Dependencies:** Milestone 3 (TenantID, event bus)

**Acceptance Criteria:**
- Can create a tenant via API
- Can generate an API key and authenticate with it
- Authenticated requests have `tenant_id` in context
- Cross-tenant data access returns 404 (verified by integration test)
- Rate limiter rejects requests beyond threshold
- Scoped keys can only access permitted endpoints

---

### Milestone 5: Catalog Context

**Goal:** Full CRUD for products, environments, and components. The "what are we monitoring?" context.

**Deliverables:**
- Product domain entity + repository + REST CRUD
- Environment domain entity + repository + REST CRUD
- Component domain entity (aggregate root with health checks) + repository + REST CRUD
- Component dependency declaration
- Health check configuration within component
- Domain events: `ComponentRegistered`, `ComponentDeregistered`
- Validation: unique names within scope, valid component types

**Dependencies:** Milestone 4 (tenant auth, tenant_id in context)

**Acceptance Criteria:**
- Can create Product → Environment → Component hierarchy via API
- Can configure health checks (HTTP type) on a component
- Can declare dependencies between components
- All operations are tenant-scoped
- `ComponentRegistered` event fires when a component is created
- Invalid configurations are rejected with proper error responses

---

### Milestone 6: Execution Context (Health Checks)

**Goal:** OPTRION actively executes HTTP health checks on configured components and stores results.

**Deliverables:**
- HTTP health checker implementation
- Scheduler that triggers checks at configured intervals
- SSRF protection (deny-list for private IPs, localhost, link-local)
- Check result recording (append to `check_results` table)
- `HealthCheckExecuted` event publication
- `HealthCheckFailed` / `HealthCheckRecovered` events
- Graceful handling of timeouts and DNS failures
- Worker pool with bounded concurrency

**Dependencies:** Milestone 5 (component + health check config exists)

**Acceptance Criteria:**
- Configured health checks execute automatically at scheduled intervals
- Results are stored in partitioned `check_results` table
- Events fire on success, failure, and recovery transitions
- SSRF protection rejects checks targeting private IPs
- Check execution does not block the main HTTP server
- `GET /healthz` still responds while checks run
- Handles 50+ concurrent checks without goroutine leak

---

### Milestone 7: Intelligence Context (Health Scoring)

**Goal:** Compute and store health scores for components. Maintain current component health state.

**Deliverables:**
- Health score computation logic (weighted average of availability, latency, consistency)
- `ComponentHealth` upsert (current state table)
- `HealthScore` recording (historical time-series)
- Score computation triggered by `HealthCheckExecuted` events
- Handles: `HealthScoreComputed` event publication
- Handles: `HealthDegraded` event when score drops below threshold
- Basic baseline (last 24h average as "normal")

**Dependencies:** Milestone 6 (check results being produced)

**Acceptance Criteria:**
- After health checks run, component_health table reflects current state
- Health scores are computed and stored historically
- Score formula produces correct results (verified by unit tests with fake clock)
- `HealthDegraded` event fires when score crosses threshold
- Dashboard query (`GET /api/v1/health/summary`) returns live data
- Score history endpoint returns time-series for graphing

---

### Milestone 8: Incident Context

**Goal:** Detect incidents from health degradation events, manage lifecycle, prevent duplicates.

**Deliverables:**
- Incident aggregate (state machine: opened → acknowledged → resolved → closed)
- Incident detection handler (subscribes to `HealthDegraded`, `HealthCheckFailed`)
- Fingerprint-based deduplication (no duplicate incidents for same failure)
- Auto-resolution when health recovers (stable for cooldown period)
- Incident event timeline (append-only history)
- REST endpoints: list incidents, get incident, acknowledge, resolve, add note
- `IncidentOpened`, `IncidentResolved` events

**Dependencies:** Milestone 7 (health degradation events)

**Acceptance Criteria:**
- Health degradation automatically creates an incident
- Same failure does NOT create duplicate incidents (fingerprint dedup)
- Incident can be acknowledged and resolved via API
- Auto-resolution works when component recovers
- Incident timeline shows all state transitions
- `IncidentOpened` event is published (consumed by alerting)

---

### Milestone 9: Alerting + Notification

**Goal:** Alert rules evaluate incidents and send Telegram notifications.

**Deliverables:**
- Alert rule domain entity + CRUD API
- Notification channel (Telegram) domain entity + CRUD API
- Alert rule evaluator (triggered by `IncidentOpened`, `IncidentEscalated`)
- Cooldown tracking (prevent alert spam)
- Maintenance window support (suppress alerts during planned work)
- Telegram sender implementation
- Alert history recording
- Channel test endpoint (send test message)

**Dependencies:** Milestone 8 (incidents being created)

**Acceptance Criteria:**
- When incident opens, matching alert rule triggers Telegram notification
- Cooldown prevents repeated alerts for same incident within window
- Maintenance window suppresses all alerts for scoped components
- Alert history shows what was sent and delivery status
- Test endpoint successfully sends a Telegram message
- Muting a rule prevents it from firing

---

### Milestone 10: REST API Completion + Push Ingestion

**Goal:** All API endpoints functional. Push-based ingestion endpoint operational.

**Deliverables:**
- Ingestion endpoint (`POST /api/v1/ingest`) accepting metrics + health
- Batch ingestion endpoint (`POST /api/v1/ingest/batch`)
- Auto-registration of unknown components (configurable)
- Analytics endpoints (uptime, trends, reliability)
- Complete error model with standard error responses
- API documentation (OpenAPI spec or similar)
- Request validation on all endpoints
- Pagination on all list endpoints

**Dependencies:** Milestones 4-9 (all contexts operational)

**Acceptance Criteria:**
- SDK can push metrics to `/api/v1/ingest` and they appear in health scores
- All endpoints from API contracts document are functional
- Error responses follow the standard error model
- Pagination works correctly on all list endpoints
- API is self-consistent (same patterns everywhere)

---

### Milestone 11: Embedded Dashboard

**Goal:** Next.js dashboard showing real-time health, incidents, and alerts for GymFlow Track.

**Deliverables:**
- Next.js project setup
- Health summary page (product → environment → component tree)
- Component detail page (health score graph, recent checks, active incidents)
- Incident list page (filter by status, severity)
- Incident detail page (timeline view)
- Alert rules management page
- Embeddable mode (iframe-friendly, scoped token auth)
- Auto-refresh (polling every 10s)

**Dependencies:** Milestone 10 (all APIs functional)

**Acceptance Criteria:**
- Dashboard shows live health status for GymFlow Track
- Health score graph renders 24h history
- Incidents are visible and can be acknowledged from UI
- Dashboard embeds cleanly in an iframe
- Responsive on mobile (GymFlow team checks from phones)
- Performance: page load < 2s, refresh < 500ms

---

### Milestone 12: GymFlow Track Integration

**Goal:** OPTRION monitoring GymFlow Track's production system end-to-end.

**Deliverables:**
- GymFlow tenant configured in OPTRION
- Health checks configured for: Backend API, PostgreSQL, Redis, Frontend
- Alert rules configured for: API latency, database health, error rates
- Telegram channel connected to GymFlow ops group
- Go SDK integrated into GymFlow Track backend (push metrics)
- Verify: incident detection → Telegram alert → acknowledge → resolve flow
- Documentation: GymFlow onboarding guide

**Dependencies:** Milestone 11 (dashboard working)

**Acceptance Criteria:**
- GymFlow Track's production health is visible in OPTRION dashboard
- If GymFlow's API goes down, incident is created within 2 minutes
- Telegram alert arrives within 30 seconds of incident creation
- GymFlow team can acknowledge and resolve from dashboard
- System runs stable for 72 hours without manual intervention
- No false positive alerts in 72-hour test period

---

## Repository Plan

### Create Immediately (Milestone 1)

```
optrion/
├── cmd/
│   └── optrion/
│       └── main.go
├── internal/
│   └── shared/
│       ├── domain/
│       └── id/
├── config/
│   ├── config.go
│   └── local.yaml
├── migrations/
├── deploy/
│   └── docker/
│       └── docker-compose.yaml
├── go.mod
├── go.sum
├── Makefile
├── .golangci.yml
├── .gitignore
└── README.md
```

### Create at Milestone 2

```
migrations/
├── 000001_create_schemas.up.sql
├── 000001_create_schemas.down.sql
├── 000002_core_tenants.up.sql
├── 000002_core_tenants.down.sql
├── ... (all Phase 1 migrations)
```

### Create at Milestone 3

```
internal/
├── shared/
│   ├── domain/
│   │   ├── tenantid.go
│   │   ├── event.go
│   │   ├── aggregate.go
│   │   ├── errors.go
│   │   └── clock.go
│   ├── types/
│   │   ├── severity.go
│   │   ├── status.go
│   │   └── score.go
│   └── id/
│       └── uuid.go
├── platform/
│   ├── eventbus/
│   │   ├── bus.go
│   │   └── outbox.go
│   ├── database/
│   │   ├── postgres.go
│   │   └── tx.go
│   ├── observability/
│   │   └── logger.go
│   └── server/
│       └── http.go
```

### Create at Milestone 4

```
internal/
├── tenant/
│   ├── domain/
│   │   ├── tenant.go
│   │   ├── apikey.go
│   │   └── errors.go
│   ├── app/
│   │   └── service.go
│   ├── port/
│   │   └── repository.go
│   └── adapter/
│       ├── postgres/
│       │   └── repository.go
│       └── rest/
│           ├── handler.go
│           └── request.go
├── platform/
│   └── auth/
│       ├── middleware.go
│       └── apikey.go
```

### Create at Milestones 5-9 (One Context at a Time)

Each context follows the same structure. Create one context per milestone:

```
internal/{context}/
├── domain/
├── app/
├── port/
└── adapter/
    ├── postgres/
    └── rest/
```

### What Should NOT Exist Yet

| Folder/File | When to Create | Why Not Now |
|------------|----------------|-------------|
| `internal/ai/` | Phase 4 | No AI in Phase 1 |
| `pkg/sdk/` | Milestone 10 | SDK is built after API is stable |
| `deploy/k8s/` | Phase 2+ | VPS + Docker Compose for Phase 1 |
| `internal/platform/cache/redis.go` | Milestone 9 (cooldown tracking) | Redis only needed for alert cooldowns |
| `web/` or `frontend/` | Milestone 11 | Separate repo or folder when dashboard work begins |
| `internal/notification/adapter/webhook/` | Phase 2 | Telegram only in Phase 1 |
| `internal/notification/adapter/email/` | Phase 2 | Telegram only in Phase 1 |
| `internal/execution/adapter/checker/tcp.go` | Phase 2 | HTTP only in Phase 1 |
| `internal/execution/adapter/checker/dns.go` | Phase 2 | HTTP only in Phase 1 |

### Empty Placeholder Convention

When creating the structure, do NOT create empty files. Only create files when you're implementing that layer. Empty `package context` declarations are acceptable for compilation, but don't create placeholder domain entities.

---

## Migration Plan

### Migration Order (Sequential, Each Depends on Previous)

| # | Migration | Creates | Rationale for Order |
|---|-----------|---------|-------------------|
| 1 | `000001_create_schemas` | `core`, `monitoring`, `incidents`, `alerting`, `intelligence`, `system` schemas | Namespaces must exist before tables |
| 2 | `000002_extensions` | `uuid-ossp` or `pgcrypto` extension | UUID generation needed by all tables |
| 3 | `000003_tenants` | `core.tenants`, `core.api_keys` | Root entity. Everything else references it |
| 4 | `000004_products` | `core.products` | References tenants |
| 5 | `000005_environments` | `core.environments` | References products (and denormalized tenant_id) |
| 6 | `000006_components` | `core.components`, `core.component_dependencies` | References environments |
| 7 | `000007_health_checks` | `monitoring.health_checks` | References components |
| 8 | `000008_check_results` | `monitoring.check_results` (PARTITIONED) | References health_checks. Partitioned by month |
| 9 | `000009_component_health` | `monitoring.component_health` | 1:1 with components. Current state table |
| 10 | `000010_metric_snapshots` | `monitoring.metric_snapshots` (PARTITIONED) | References components. Time-series |
| 11 | `000011_health_scores` | `monitoring.health_scores` (PARTITIONED) | Polymorphic subject (component/env/product) |
| 12 | `000012_incidents` | `incidents.incidents`, `incidents.incident_events` | References components |
| 13 | `000013_alert_rules` | `alerting.alert_rules` | References tenants, scoped to components/envs |
| 14 | `000014_notification_channels` | `alerting.notification_channels` | References tenants |
| 15 | `000015_alert_history` | `alerting.alert_history` (PARTITIONED) | References rules and incidents |
| 16 | `000016_maintenance_windows` | `alerting.maintenance_windows` | References tenants, scoped |
| 17 | `000017_event_outbox` | `system.event_outbox` (PARTITIONED daily) | Standalone. Used by event bus |
| 18 | `000018_audit_log` | `system.audit_log` (PARTITIONED) | Cross-cutting. Last because it's non-critical |
| 19 | `000019_baselines` | `intelligence.baselines` | References components. Intelligence context |
| 20 | `000020_rls_policies` | Row-Level Security policies on ALL tables | Must be last — references all tables |
| 21 | `000021_seed_gymflow` | Development seed data | Only for `local` environment |

### Migration Principles

1. **One table per migration** — Makes rollback granular
2. **Always write both up and down** — Even if down is `DROP TABLE`
3. **Partitioned tables get partitions created in the same migration** — Create 3 months of partitions
4. **Indexes in same migration as table** — Don't separate (makes rollback messy)
5. **RLS policies in dedicated migration** — Easier to audit and test separately
6. **Seed data is a separate migration** — Only applied in development. Tagged with environment guard

---

## Implementation Order

### Layer-by-Layer Within Each Context

For each bounded context, implement in this order:

```
1. Domain entities + value objects (pure logic, tested immediately)
2. Port interfaces (define the contract)
3. Application service (use cases, tested with mocked ports)
4. PostgreSQL adapter (repository implementation)
5. REST adapter (HTTP handlers)
6. Integration test (end-to-end through the stack)
```

**Never skip ahead.** Don't write the HTTP handler before the domain is solid.

### Global Implementation Sequence

| Step | What | Why This Order |
|------|------|---------------|
| 1 | Configuration loading | Everything needs config |
| 2 | Structured logging | Everything should log |
| 3 | Database connection pool | Required for any persistence |
| 4 | HTTP server skeleton | Need a way to verify things work |
| 5 | Migration runner | Tables must exist |
| 6 | Shared kernel types | Used by all contexts |
| 7 | Transaction manager | Used by all repositories |
| 8 | Event bus (in-process, sync) | Used by all contexts for communication |
| 9 | Tenant domain + repository | First real business entity |
| 10 | API key auth + middleware | Protects all subsequent endpoints |
| 11 | Tenant REST handlers | First working API endpoints |
| 12 | Catalog domain (product/env/component) | What we monitor |
| 13 | Catalog repositories | Persistence for catalog |
| 14 | Catalog REST handlers | API for managing monitored services |
| 15 | Health check domain | Execution logic |
| 16 | HTTP checker adapter | Actually calls external URLs |
| 17 | Scheduler | Triggers checks on schedule |
| 18 | Check result repository | Stores results |
| 19 | SSRF protection | Security requirement |
| 20 | Intelligence domain (scoring) | Processes check results |
| 21 | Component health repository | Current state |
| 22 | Health score repository | Historical scores |
| 23 | Intelligence event handler | Subscribes to check executed events |
| 24 | Health REST handlers | Dashboard data APIs |
| 25 | Incident domain (state machine) | Incident lifecycle |
| 26 | Incident detection handler | Subscribes to health degraded events |
| 27 | Incident repository | Persistence |
| 28 | Incident REST handlers | Manage incidents via API |
| 29 | Alert rule domain | Rule configuration |
| 30 | Alert rule evaluator | Triggered by incident events |
| 31 | Notification channel domain | Telegram configuration |
| 32 | Telegram sender | Actually sends messages |
| 33 | Cooldown tracking (Redis or in-memory) | Prevent alert spam |
| 34 | Alert REST handlers | Manage rules and channels |
| 35 | Outbox poller (background worker) | Guaranteed event delivery |
| 36 | Ingestion endpoint | Push-based metrics |
| 37 | Analytics endpoints | Uptime, trends, reliability |
| 38 | OpenAPI documentation | Contract verification |

---

## Testing Plan

### Milestone 1: Foundation

| Test Type | What to Test |
|-----------|-------------|
| Unit | Config loading with various sources |
| Unit | Log level configuration |
| Integration | PostgreSQL connection succeeds |
| Integration | HTTP server responds to `/healthz` |
| Manual | `docker-compose up` starts all services |

### Milestone 2: Database

| Test Type | What to Test |
|-----------|-------------|
| Integration | All migrations apply cleanly (up) |
| Integration | All migrations roll back cleanly (down) |
| Integration | Idempotent re-run doesn't fail |
| Manual | `psql` inspection of tables, partitions, RLS policies |
| Manual | Verify RLS blocks cross-tenant SELECT |

### Milestone 3: Shared Kernel + Event Bus

| Test Type | What to Test |
|-----------|-------------|
| Unit | TenantID validation (reject empty, nil) |
| Unit | Score value object (reject <0, >100) |
| Unit | UUID v7 generation is time-sortable |
| Unit | Fake clock produces deterministic time |
| Integration | Event bus delivers events to registered handlers |
| Integration | Outbox writer persists events in transaction |
| Integration | Outbox poller dispatches persisted events |

### Milestone 4: Tenant Context

| Test Type | What to Test |
|-----------|-------------|
| Unit | Tenant domain: creation validation, status transitions |
| Unit | API key hash generation |
| Unit | Application service: orchestration with mocked repo |
| Integration | Repository: save, find, find-by-slug, find-by-key-hash |
| Integration | **Cross-tenant isolation test** (CRITICAL) |
| Integration | Auth middleware: valid key accepted, invalid rejected |
| Integration | Rate limiter: excess requests get 429 |
| Manual | Create tenant via API, get API key, use key for auth |

### Milestone 5: Catalog Context

| Test Type | What to Test |
|-----------|-------------|
| Unit | Component aggregate: add health check, validate, reject invalid |
| Unit | Circular dependency detection |
| Integration | Product/Environment/Component CRUD repositories |
| Integration | Cascade behavior (archive product archives children?) |
| Integration | ComponentRegistered event fires on creation |
| Manual | Create full hierarchy via API. Verify in database |

### Milestone 6: Execution Context

| Test Type | What to Test |
|-----------|-------------|
| Unit | HTTP checker: parse response, determine health status |
| Unit | SSRF validator: reject private IPs, localhost, link-local |
| Unit | Scheduler: correct interval calculation |
| Integration | HTTP checker against test server (httptest) |
| Integration | Check results stored in partitioned table |
| Integration | Events published on check completion |
| Manual | Configure check against real URL. Watch results accumulate |
| Manual | Configure check against internal IP. Verify rejection |

### Milestone 7: Intelligence Context

| Test Type | What to Test |
|-----------|-------------|
| Unit | Score formula: known inputs → expected score |
| Unit | Score with all-healthy checks = 100 |
| Unit | Score with mixed results = weighted correctly |
| Unit | Degradation detection threshold logic |
| Integration | Score computed when HealthCheckExecuted event arrives |
| Integration | ComponentHealth updated after scoring |
| Integration | HealthDegraded event fires when score drops |
| Manual | Watch scores update in real-time as checks run |

### Milestone 8: Incident Context

| Test Type | What to Test |
|-----------|-------------|
| Unit | Incident state machine: valid transitions only |
| Unit | Fingerprint generation: same input = same fingerprint |
| Unit | Deduplication: reject duplicate incident for active fingerprint |
| Unit | Auto-resolution logic: recover + stable = resolved |
| Integration | Incident created when HealthDegraded event received |
| Integration | Duplicate not created for same failure |
| Integration | Incident timeline records all events |
| Integration | API: acknowledge, resolve, add note |
| Manual | Trigger a real failure. Verify incident appears |

### Milestone 9: Alerting + Notification

| Test Type | What to Test |
|-----------|-------------|
| Unit | Rule evaluator: condition matching logic |
| Unit | Cooldown: within window = suppressed |
| Unit | Maintenance window: active window = suppressed |
| Integration | Telegram sender: mock HTTP server verifies request shape |
| Integration | Alert fires when incident matches rule |
| Integration | Alert suppressed when in cooldown |
| Integration | Alert history recorded |
| Manual | Trigger incident → Telegram message arrives in test group |
| Manual | Verify cooldown prevents spam |

### Testing Infrastructure

| Tool | Purpose | Setup Cost |
|------|---------|-----------|
| `go test` | Test runner | Zero (built-in) |
| `testify/assert` | Assertions | `go get` |
| `testcontainers-go` | PostgreSQL + Redis containers for integration tests | Medium (Docker required) |
| `httptest` | Mock HTTP servers for checker tests | Zero (stdlib) |
| `goleak` | Goroutine leak detection | `go get` |
| `-race` flag | Race condition detection | Zero (compiler flag) |

### Test Commands

```
make test-unit          # Fast, no external deps
make test-integration   # Requires Docker
make test-all           # Both
make test-race          # With race detector
make coverage           # Generate coverage report
```

---

## Deployment Plan

### Phase 1 Infrastructure

```
┌─────────────────────────────────────────────┐
│              Single VPS (4GB RAM)             │
│                                              │
│  ┌─────────────────────────────────────┐    │
│  │         Docker Compose               │    │
│  │                                     │    │
│  │  ┌───────────┐  ┌──────────────┐   │    │
│  │  │  optrion  │  │  postgresql  │   │    │
│  │  │  (Go app) │  │  (15.x)     │   │    │
│  │  │  :8080    │  │  :5432      │   │    │
│  │  └───────────┘  └──────────────┘   │    │
│  │                                     │    │
│  │  ┌───────────┐  ┌──────────────┐   │    │
│  │  │   redis   │  │   caddy      │   │    │
│  │  │  :6379    │  │  (reverse    │   │    │
│  │  │           │  │   proxy+TLS) │   │    │
│  │  └───────────┘  └──────────────┘   │    │
│  │                                     │    │
│  │  ┌───────────┐                      │    │
│  │  │  next.js  │                      │    │
│  │  │  (dash)   │                      │    │
│  │  │  :3000    │                      │    │
│  │  └───────────┘                      │    │
│  └─────────────────────────────────────┘    │
│                                              │
│  Caddy auto-TLS: api.optrion.dev            │
│                  dash.optrion.dev           │
│                                              │
└─────────────────────────────────────────────┘
```

### Deployment Process (Phase 1)

1. Push to `main` branch
2. SSH into VPS (or GitHub Action SSHes)
3. `git pull` latest code
4. `docker compose build optrion`
5. `docker compose up -d --no-deps optrion` (zero-downtime single instance)
6. Health check passes → deployment complete
7. Health check fails → `docker compose rollback` to previous image

### Why NOT Kubernetes

| Reason |
|--------|
| Solo engineer. Kubernetes operational overhead is 10x Docker Compose |
| Single VPS handles 10 tenants easily |
| No horizontal scaling needed in Phase 1 |
| Caddy provides auto-TLS for free |
| Total monthly cost: $20-40 for a VPS that handles the entire MVP |

### Deployment Checklist (Pre-Launch)

- [ ] VPS provisioned (Ubuntu 22.04+, 4GB RAM, 80GB SSD)
- [ ] Docker + Docker Compose installed
- [ ] Firewall: only 80, 443, 22 open
- [ ] DNS: `api.optrion.dev` → VPS IP
- [ ] DNS: `dash.optrion.dev` → VPS IP
- [ ] PostgreSQL data volume on persistent storage
- [ ] Automated daily backups (pg_dump to object storage)
- [ ] Uptime monitoring on OPTRION's own `/healthz` (external service like Uptime Robot)
- [ ] Log rotation configured
- [ ] Fail2ban on SSH

### Backup Strategy

| What | How | Frequency | Retention |
|------|-----|-----------|-----------|
| PostgreSQL full | `pg_dump` to compressed file | Daily at 03:00 UTC | 30 days |
| PostgreSQL WAL | Streaming to object storage | Continuous | 7 days |
| Application config | Git (already versioned) | Every commit | Forever |
| Docker volumes | Not backed up (reproducible) | N/A | N/A |

---

## MVP Definition

### The Smallest Thing That Provides Value to GymFlow Track

**MVP = "Can I see if my system is healthy and get a Telegram alert when it's not?"**

#### MVP Features (All Must Work)

1. GymFlow registers as a tenant and gets an API key
2. GymFlow configures 4 components: Backend API, PostgreSQL, Redis, Frontend
3. OPTRION checks each component's health endpoint every 30 seconds
4. Dashboard shows: all components, their status (green/yellow/red), health score
5. When Backend API goes down: incident auto-created, Telegram alert sent within 60 seconds
6. GymFlow acknowledges the incident from the dashboard
7. When Backend API recovers: incident auto-resolved, recovery Telegram sent
8. Dashboard shows uptime history for last 24 hours

#### MVP Does NOT Include

- Push-based SDK ingestion (pull-only)
- Business metrics/KPIs
- Analytics (trends, MTTR, reliability stats)
- Multiple alert channels (Telegram only)
- User management (single API key per tenant)
- Billing
- AI

#### MVP Success Criteria

| Metric | Target |
|--------|--------|
| Time from failure to Telegram alert | < 90 seconds |
| Time from recovery to resolution alert | < 90 seconds |
| Dashboard load time | < 2 seconds |
| False positive rate (72h test) | 0 |
| Uptime of OPTRION itself (72h test) | 100% |
| Manual intervention required | 0 (fully automated detection + alerting) |

---

## Anti-Scope List

### What Must NOT Be Built Now

| Feature | Why Not | Earliest Phase |
|---------|---------|---------------|
| **AI / Gemini integration** | Zero historical data exists. AI without data is theater | Phase 4 (6+ months of data needed) |
| **Auto remediation** | Requires high confidence, runbooks, and liability model. One wrong auto-fix and you lose trust forever | Phase 5 |
| **Kubernetes deployment** | Solo engineer. K8s operational cost > Docker Compose cost by 10x. No horizontal scaling needed at <10 tenants | Phase 2+ (when ops team exists) |
| **Complex analytics** | MTTR/MTBF/SLO burn rates are valuable but not MVP. GymFlow needs "is it up?" not "what's my SLO burn rate?" | Phase 2 |
| **Multi-region check execution** | Single VPS checks from one location. Multi-region adds coordination complexity | Phase 3 |
| **Custom check types (TCP, gRPC, DNS)** | HTTP health endpoints cover 95% of GymFlow's needs | Phase 2 |
| **User management / RBAC** | GymFlow has one developer. One API key is sufficient | Phase 2 |
| **Billing / Stripe integration** | First 5 customers billed manually. Building billing before revenue is premature optimization | Phase 2 |
| **Webhooks (outgoing)** | Telegram covers GymFlow's needs. Webhooks are for enterprise integrations | Phase 2 |
| **Public status pages** | Different product surface. Not core to health intelligence | Phase 3 |
| **Push-based SDK** | Pull-based (OPTRION polls) is simpler and sufficient for Phase 1. SDK adds auth complexity | Phase 2 (Milestone 10 is borderline — defer if behind schedule) |
| **Email notifications** | Telegram is immediate. Email requires SMTP setup, deliverability concerns | Phase 2 |
| **Mobile app** | Responsive web dashboard + Telegram is mobile access for Phase 1 | Phase 3+ |
| **Plugin system** | Hexagonal architecture provides extensibility without a plugin framework | Never (architecture handles it) |
| **GraphQL** | REST is sufficient. GraphQL adds parsing complexity and security surface | Evaluate Phase 3 |
| **Real-time WebSocket dashboard** | 10s polling is acceptable for Phase 1. WebSocket adds connection management | Phase 2 |
| **Internationalization** | English only. GymFlow is English-speaking team | Phase 3 |
| **Dark mode** | Ship light mode. Dark mode is CSS work, not architecture | Whenever |
| **Terraform / Infrastructure as Code** | One VPS. `docker-compose.yaml` IS the infrastructure code | Phase 2+ |

### The "Not Yet" Principle

> If a feature does not directly contribute to: **"GymFlow sees their system health and gets alerted when something breaks"** — it does not belong in Phase 1.

---

## Effort Estimate

### Assumptions

- Solo engineer, after-work development
- ~2.5 hours/day average (some days 1h, weekends 4-5h)
- ~15-18 hours/week effective development time
- Go proficiency: intermediate (learning on the job adds ~30% overhead)
- PostgreSQL proficiency: strong
- No context-switching tax (single project focus)

### Per-Milestone Estimates

| Milestone | Best Case | Realistic | Worst Case | Notes |
|-----------|-----------|-----------|------------|-------|
| 1. Foundation | 3 days | 5 days | 8 days | Go module setup, Docker, config. Learning curve |
| 2. Database | 3 days | 5 days | 7 days | SQL is known. Partitioning needs research |
| 3. Shared Kernel + Events | 4 days | 7 days | 10 days | Event bus is the hardest part. Must get right |
| 4. Tenant Context | 4 days | 6 days | 9 days | First full vertical slice. Learning the pattern |
| 5. Catalog Context | 3 days | 5 days | 7 days | Similar to tenant. Faster now that pattern is known |
| 6. Execution (Health Checks) | 5 days | 8 days | 12 days | Scheduler + concurrency + SSRF. Most complex milestone |
| 7. Intelligence (Scoring) | 4 days | 6 days | 9 days | Scoring logic + event handling |
| 8. Incident | 4 days | 7 days | 10 days | State machine + dedup. Domain complexity |
| 9. Alerting + Notification | 4 days | 6 days | 9 days | Telegram integration + cooldown logic |
| 10. API Completion | 3 days | 5 days | 8 days | Ingestion + analytics. Pattern is established |
| 11. Dashboard | 5 days | 8 days | 12 days | Next.js. Different tech stack. Context switch |
| 12. GymFlow Integration | 2 days | 4 days | 6 days | Configuration + testing. Low code |

### Total Estimates

| Scenario | Calendar Days | Calendar Weeks | Wall Clock (at 2.5h/day) |
|----------|--------------|----------------|--------------------------|
| **Best case** | 44 days | ~6.5 weeks | ~8 weeks |
| **Realistic** | 72 days | ~10 weeks | ~12 weeks |
| **Worst case** | 107 days | ~15 weeks | ~18 weeks |

### Realistic Timeline: 10-12 Weeks

With buffer for:
- Debugging production issues discovered during integration
- Learning Go idioms and fixing non-idiomatic code
- Unexpected database performance issues
- Life events (sick days, busy weeks at day job)
- The inevitable "I need to refactor this" moment at Milestone 6

### Velocity Adjustment Factors

| Factor | Impact |
|--------|--------|
| First 2 milestones are slower (setting up patterns) | +30% time |
| Milestones 5-9 accelerate (pattern is established) | -20% time |
| Milestone 11 (dashboard) is a tech-stack switch | +40% time |
| Weekend deep-work sessions | -15% total (compensates for short weekday sessions) |

---

## Weekly Roadmap

### Week 1: Foundation + Database

**Focus:** Get a running application that connects to PostgreSQL.

| Day | Task |
|-----|------|
| Mon | Initialize Go module. Set up Makefile. Docker Compose with PostgreSQL |
| Tue | Configuration loading (viper/koanf). Structured logging (slog) |
| Wed | Database connection pool (pgx). HTTP server skeleton (/healthz) |
| Thu | Migration tool setup. Write migrations 1-10 (schemas through health scores) |
| Fri | Write migrations 11-21. Run all. Verify partitions. Seed data |
| Sat | Integration tests: migrations apply/rollback. RLS verification |
| Sun | Buffer / catch up. README. Document what you've built |

**Deliverable:** Running app. All tables exist. Seed data loaded.

---

### Week 2: Shared Kernel + Tenant Context

**Focus:** Cross-cutting types and first full vertical slice (tenant CRUD + auth).

| Day | Task |
|-----|------|
| Mon | Shared kernel: TenantID, DomainEvent interface, Clock, UUID v7 |
| Tue | Value objects: Score, Severity, HealthStatus. Unit tests for all |
| Wed | Event bus: in-process sync dispatch. Outbox writer. Tests |
| Thu | Tenant domain: entity, creation logic, validation. Unit tests |
| Fri | Tenant repository (PostgreSQL). API key generation + hash storage |
| Sat | Auth middleware. Rate limiter. Integration tests for auth flow |
| Sun | Tenant REST handlers: create tenant, get tenant, create key. Manual test via curl |

**Deliverable:** Can create tenant, get API key, authenticate requests.

---

### Week 3: Catalog Context

**Focus:** Product → Environment → Component CRUD. Health check configuration.

| Day | Task |
|-----|------|
| Mon | Product domain + repository + REST handler |
| Tue | Environment domain + repository + REST handler |
| Wed | Component domain (aggregate with health checks). Invariant validation |
| Thu | Component repository. Health check configuration within component |
| Fri | Component REST handlers. Dependency declaration API |
| Sat | Integration tests. Cross-tenant isolation verification |
| Sun | ComponentRegistered event. Wire event publication. Manual test: create full hierarchy |

**Deliverable:** Full CRUD for monitoring hierarchy. Events fire on component changes.

---

### Week 4: Execution Context (Health Checks)

**Focus:** OPTRION actively checks configured endpoints. This is the core engine.

| Day | Task |
|-----|------|
| Mon | HTTP checker: make request, parse response, determine status. Unit tests |
| Tue | SSRF protection: IP deny-list, DNS validation. Unit tests |
| Wed | Scheduler: manages per-check tickers. Start/stop/reschedule |
| Thu | Wire scheduler to component creation (ComponentRegistered → schedule check) |
| Fri | Check result recording. Repository implementation |
| Sat | HealthCheckExecuted event publication. HealthCheckFailed/Recovered transition logic |
| Sun | Integration test: configure component → checks run automatically → results stored |

**Deliverable:** Health checks run on schedule. Results accumulate. Events fire.

---

### Week 5: Intelligence + Incident Contexts

**Focus:** Scoring engine and incident detection. The "intelligence" part.

| Day | Task |
|-----|------|
| Mon | Health score formula implementation. Unit tests with various inputs |
| Tue | ComponentHealth repository (upsert current state). HealthScore repository (historical) |
| Wed | Score computation event handler (triggered by HealthCheckExecuted). Wire it up |
| Thu | HealthDegraded event detection. Threshold logic. Integration test |
| Fri | Incident domain: state machine, fingerprint, dedup logic. Thorough unit tests |
| Sat | Incident detection handler (triggered by HealthDegraded). Repository. Timeline |
| Sun | Incident REST handlers: list, get, acknowledge, resolve. Integration tests |

**Deliverable:** Checks → scores → incidents. Full automated pipeline.

---

### Week 6: Alerting + Notification

**Focus:** Alert rules trigger Telegram notifications. End-to-end alerting chain.

| Day | Task |
|-----|------|
| Mon | Alert rule domain. Condition evaluation logic. Unit tests |
| Tue | Alert rule repository + REST CRUD handlers |
| Wed | Notification channel domain. Telegram sender implementation |
| Thu | Rule evaluator event handler (triggered by IncidentOpened). Cooldown tracking |
| Fri | Maintenance window. Alert suppression logic. Alert history recording |
| Sat | Integration test: incident → rule matches → Telegram sent. Channel test endpoint |
| Sun | End-to-end test: component fails → incident created → Telegram arrives |

**Deliverable:** Full alerting pipeline. Telegram notifications working.

---

### Week 7: API Completion + Polish

**Focus:** Remaining endpoints. Error handling. Ingestion. Documentation.

| Day | Task |
|-----|------|
| Mon | Ingestion endpoint (POST /api/v1/ingest). Batch endpoint |
| Tue | Auto-registration logic. Ingestion validation |
| Wed | Health summary API. Component history API (time-series) |
| Thu | Analytics endpoints: uptime calculation, basic reliability stats |
| Fri | Standard error model. Consistent error responses across all endpoints |
| Sat | Pagination on all list endpoints. Request validation hardening |
| Sun | OpenAPI spec generation or manual documentation. API review |

**Deliverable:** All Phase 1 APIs functional. Error model consistent.

---

### Week 8: Outbox + Reliability

**Focus:** Guaranteed event delivery. Background workers. Operational stability.

| Day | Task |
|-----|------|
| Mon | Outbox poller implementation. Background worker lifecycle |
| Tue | Retry logic for failed events. Dead letter handling |
| Wed | Graceful shutdown: drain checks, flush outbox, close connections |
| Thu | Self-monitoring: log outbox depth, check execution stats, error rates |
| Fri | Prometheus metrics exposure (/metrics endpoint) |
| Sat | Load test: 50 components, 30s checks, verify stability over 2 hours |
| Sun | Fix issues found in load test. Memory profiling. Goroutine leak check |

**Deliverable:** Reliable, production-worthy backend. No resource leaks.

---

### Week 9: Dashboard (Next.js)

**Focus:** Visual interface for GymFlow Track.

| Day | Task |
|-----|------|
| Mon | Next.js project setup. API client utility. Authentication flow |
| Tue | Health summary page: product → environment → component tree with scores |
| Wed | Component detail page: health score graph (24h), recent check results |
| Thu | Incident list page: filter by status/severity. Incident detail with timeline |
| Fri | Alert rules management: list, create, edit, mute |
| Sat | Embeddable mode: scoped token auth, iframe-friendly headers |
| Sun | Responsive styling. Mobile-friendly. Polish |

**Deliverable:** Working dashboard showing live OPTRION data.

---

### Week 10: Integration + Deployment

**Focus:** Deploy to VPS. Connect GymFlow Track. Verify end-to-end.

| Day | Task |
|-----|------|
| Mon | VPS provisioning. Docker Compose production config. Caddy TLS setup |
| Tue | Deploy OPTRION to VPS. Verify /healthz reachable at api.optrion.dev |
| Wed | Automated backup setup. Uptime monitoring (external) on OPTRION itself |
| Thu | Configure GymFlow tenant. Set up components. Configure Telegram channel |
| Fri | Verify: checks running → scores computing → dashboard showing live data |
| Sat | Kill GymFlow API intentionally. Verify: incident created → Telegram alert → recovery |
| Sun | 72-hour stability test begins. Monitor for false positives |

**Deliverable:** OPTRION running in production, monitoring GymFlow Track.

---

### Week 11-12: Buffer + Hardening

**Purpose:** Every plan needs buffer. Real projects always slip.

| Task | Priority |
|------|----------|
| Fix bugs found during 72h stability test | P0 |
| Improve error messages based on actual usage | P1 |
| Add missing validation found during integration | P1 |
| Performance tune slow queries (if any) | P1 |
| Write GymFlow onboarding guide | P2 |
| Add any missing integration tests | P2 |
| Document operational runbook (restart, backup, recovery) | P2 |
| Retrospective: what worked, what to change for Phase 2 | P2 |

---

## Summary

| Artifact | Status |
|----------|--------|
| Critical path identified | ✅ |
| 12 milestones defined with acceptance criteria | ✅ |
| Repository creation plan (progressive, no premature structure) | ✅ |
| 21 migrations in dependency order | ✅ |
| 38-step implementation sequence | ✅ |
| Per-milestone test plan | ✅ |
| Deployment: Docker Compose on VPS with Caddy | ✅ |
| MVP: Pull-based checks + scoring + incidents + Telegram | ✅ |
| Anti-scope: 17 items explicitly deferred | ✅ |
| Effort: 10-12 weeks realistic for solo after-work engineer | ✅ |
| Weekly roadmap: 10 weeks + 2 weeks buffer | ✅ |

### The One Rule

> **Every week must end with something you can demo.** If you can't show progress in a browser or terminal, you've gone too deep into abstraction. Ship visible value continuously.

---

*End of Implementation Strategy*
