# OPTRION — Engineering Intelligence Platform

[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**OPTRION** is an engineering intelligence platform that provides real-time health monitoring, intelligent scoring, incident detection, and automated alerting for distributed software systems.

Built for engineering teams who need visibility into the health of their services without the complexity of enterprise observability tools.

---

## What It Does

| Capability | Description |
|-----------|-------------|
| **Health Monitoring** | Pull-based HTTP/TCP/DNS health checks with configurable intervals |
| **Intelligent Scoring** | Deterministic health scoring (0-100) with hierarchical composition |
| **Incident Detection** | Automatic incident creation from degraded health with deduplication |
| **Smart Alerting** | Threshold-based alerts with cooldown, suppression, and escalation |
| **Multi-Tenant** | Tenant isolation via PostgreSQL Row-Level Security |
| **API-First** | RESTful API with scoped API key authentication |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        OPTRION PLATFORM                          │
│                                                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │  Tenant  │  │ Catalog  │  │Execution │  │ Intelligence │  │
│  │          │  │          │  │          │  │              │  │
│  │ API Keys │  │ Products │  │ Health   │  │ Scoring      │  │
│  │ Plans    │  │ Envs     │  │ Checks   │  │ Baselines    │  │
│  │ Quotas   │  │ Comps    │  │ Results  │  │ Anomalies    │  │
│  └──────────┘  └──────────┘  └────┬─────┘  └──────┬───────┘  │
│                                    │ events         │ events   │
│                                    ▼                ▼          │
│  ┌──────────────────┐  ┌────────────────────────────────────┐ │
│  │   Notification   │◀─│           Alerting / Incident       │ │
│  │                  │  │                                    │ │
│  │  Telegram        │  │  Detection, Dedup, Escalation     │ │
│  │  Webhooks        │  │  Cooldowns, Maintenance Windows   │ │
│  └──────────────────┘  └────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

**Design Principles:**
- **Modular Monolith** — Single deployable binary with bounded context isolation
- **Hexagonal Architecture** — Domain logic has zero infrastructure dependencies
- **Domain-Driven Design** — Aggregates, value objects, domain events
- **Event-Driven** — In-process event bus with PostgreSQL outbox pattern

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.23 |
| Database | PostgreSQL 16 (pgx, connection pooling, RLS) |
| Cache | Redis 7 (rate limiting, cooldowns) |
| HTTP | Standard library `net/http` + custom middleware |
| Logging | `log/slog` (structured JSON, correlation IDs) |
| Config | Environment variables with validation |
| Deployment | Docker Compose (single VPS) |
| Testing | Standard `testing` package, table-driven tests |

---

## Project Structure

```
optrion/
├── cmd/optrion/              # Application entry point
│   └── main.go              # Startup + graceful shutdown
├── internal/
│   ├── app/                 # DI container (composition root)
│   ├── health/              # Health Monitoring bounded context
│   │   ├── domain/          # Entities, value objects, types
│   │   ├── port/            # Repository + collector interfaces
│   │   ├── app/             # Application service (orchestration)
│   │   ├── collector/       # Backend, Postgres, Redis, Server collectors
│   │   ├── scoring/         # Rule-based health scoring engine
│   │   ├── anomaly/         # Statistical anomaly detection (z-score)
│   │   ├── scheduler/       # Per-collector periodic scheduling
│   │   └── adapter/
│   │       ├── postgres/    # Repository implementations (pgx)
│   │       └── rest/        # HTTP handlers + response DTOs
│   ├── tenant/              # Tenant bounded context
│   │   ├── domain/          # Tenant, Product, Environment, Component
│   │   ├── port/            # Repository interfaces
│   │   ├── app/             # Application service
│   │   └── adapter/
│   │       ├── postgres/    # Repository implementations
│   │       └── rest/        # HTTP handlers
│   ├── platform/            # Cross-cutting infrastructure
│   │   ├── config/          # Configuration loading + validation
│   │   ├── logger/          # Structured logging with context
│   │   ├── database/        # PostgreSQL pool + health checks
│   │   ├── cache/           # Redis client + health checks
│   │   └── server/          # HTTP server, router, middleware
│   └── shared/              # Shared kernel
│       ├── id/              # UUID v7 generation
│       └── domain/          # Shared domain errors
├── migrations/              # Versioned SQL migrations (embedded)
├── scripts/
│   ├── seed/                # Tenant hierarchy seed script
│   └── seed-health/         # Health metric definitions seed
├── deploy/docker/           # Dockerfile + docker-compose
├── .env.local               # Local development config
├── .env.development         # Docker development config
├── Makefile                 # Build automation
└── go.mod                   # Go module definition
```

---

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- Make (optional)

### 1. Clone

```bash
git clone https://github.com/mdshoaibumer/OPTRION.git
cd OPTRION
```

### 2. Start Infrastructure

```bash
make infra-up
# Starts PostgreSQL on :5432 and Redis on :6379
```

### 3. Run the Application

```bash
# Option A: Using Make
make run

# Option B: Direct
source .env.local  # or manually export env vars on Windows
go run ./cmd/optrion
```

### 4. Verify

```bash
# Liveness probe
curl http://localhost:8080/healthz

# Readiness probe (checks DB + Redis)
curl http://localhost:8080/readyz

# Application info
curl http://localhost:8080/api/v1/info
```

### 5. Run Tests

```bash
make test
# or
go test ./... -v -count=1
```

### 6. Full Docker Stack

```bash
make docker-up
# Builds and starts: app + postgres + redis
# App available at http://localhost:8080
```

---

## Available Make Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the binary |
| `make run` | Build and run locally |
| `make test` | Run all tests |
| `make test-cover` | Run tests with coverage report |
| `make lint` | Run golangci-lint |
| `make docker-up` | Start full Docker stack |
| `make docker-down` | Stop Docker stack |
| `make infra-up` | Start only PostgreSQL + Redis |
| `make clean` | Remove build artifacts |

---

## API Endpoints

### Platform

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/healthz` | Liveness probe (always 200) |
| `GET` | `/readyz` | Readiness probe (checks all deps) |
| `GET` | `/api/v1/info` | Application version + environment |

### Tenant Management

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/tenants` | Create tenant |
| `GET` | `/api/v1/tenants/{id}` | Get tenant by ID |
| `GET` | `/api/v1/tenants` | List tenants |
| `POST` | `/api/v1/products` | Create product |
| `POST` | `/api/v1/environments` | Create environment |
| `POST` | `/api/v1/components` | Register component |

### Health Monitoring

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/health/summary?tenant_id=` | Overall health summary |
| `GET` | `/api/v1/health/components?tenant_id=` | Component health statuses |
| `GET` | `/api/v1/health/history?tenant_id=` | Historical health scores |
| `GET` | `/api/v1/health/anomalies?tenant_id=` | Detected anomalies |

---

## Health Monitoring

OPTRION's Health Monitoring Engine provides continuous, automated assessment of system health.

### How It Works

1. **Collectors** pull metrics from each component (HTTP endpoints, PostgreSQL, Redis, server resources)
2. **Scheduler** runs collectors at configurable intervals (30s-60s per collector type)
3. **Scoring Engine** applies rule-based evaluation to compute a 0-100 health score
4. **Anomaly Detector** uses statistical z-score analysis (3σ threshold) to identify unusual behavior
5. **REST API** exposes real-time summaries, component statuses, and anomaly history

### Collector Types

| Collector | Metrics | Interval |
|-----------|---------|----------|
| **Backend** | availability, response_time, error_rate, throughput, uptime | 30s |
| **PostgreSQL** | connection_status, query_latency, active_connections, slow_queries, deadlocks, index_usage, pool_health | 60s |
| **Redis** | availability, memory_usage, hit_ratio, evictions, connected_clients | 60s |
| **Server** | cpu, ram, disk, load_average, network | 60s |

### Scoring

Health scores are computed per-component using configurable rules:
- **100** = Perfectly healthy
- **≥90** = Healthy (no action needed)
- **70-89** = Degraded (investigate)
- **<70** = Critical (immediate action)

### Anomaly Detection

Uses a rolling window (60 samples) with z-score analysis:
- **>3σ** deviation = Medium anomaly
- **>4σ** deviation = High anomaly
- **>5σ** deviation = Critical anomaly

### Example Response

```json
// GET /api/v1/health/summary?tenant_id=<uuid>
{
  "tenant_id": "0194ffa0-...",
  "overall_score": 92,
  "overall_status": "healthy",
  "components": 4,
  "healthy": 3,
  "degraded": 1,
  "critical": 0,
  "reasons": ["Query latency 120ms exceeds 100ms threshold"],
  "last_updated_at": "2026-01-15T10:30:00Z"
}
```

---

## Seeding Test Data

```bash
# 1. Start infrastructure
make infra-up

# 2. Run the application
make run

# 3. In another terminal, seed tenant hierarchy
go run scripts/seed/seed.go

# 4. Seed health metric definitions (requires PostgreSQL access)
go run scripts/seed-health/main.go
```

---

## Configuration

All configuration is done via environment variables. See [`.env.local`](.env.local) for the full list.

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `local` | Environment (local/development/production) |
| `HTTP_PORT` | `8080` | HTTP server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_NAME` | `optrion` | Database name |
| `REDIS_HOST` | `localhost` | Redis host |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Log format (json/text) |

Production requires `DB_SSL_MODE != disable` and `DB_PASSWORD` to be set.

---

## Implementation Roadmap

| Phase | Status | Milestone |
|-------|--------|-----------|
| 1 | ✅ Complete | Foundation (config, logging, DB, Redis, HTTP, Docker) |
| 2 | ✅ Complete | Tenant bounded context (multi-tenant, products, envs, components) |
| 3 | ✅ Complete | Health Monitoring Engine (collectors, scoring, anomaly detection) |
| 4 | 🔲 Planned | Incident detection |
| 5 | 🔲 Planned | Alerting + Notifications |
| 6 | 🔲 Planned | API authentication |
| 7 | 🔲 Planned | Dashboard (Next.js) |

---

## Design Documents

Detailed architecture documentation available in the repository:

- [`ARCHITECTURE_REVIEW.md`](ARCHITECTURE_REVIEW.md) — System architecture, risks, and decisions
- [`DOMAIN_MODEL.md`](DOMAIN_MODEL.md) — DDD domain model (entities, aggregates, events)
- [`DATABASE_ARCHITECTURE.md`](DATABASE_ARCHITECTURE.md) — Schema design, indexes, partitioning
- [`GO_ARCHITECTURE.md`](GO_ARCHITECTURE.md) — Go codebase structure, dependency rules
- [`API_CONTRACTS.md`](API_CONTRACTS.md) — REST API contracts, error model
- [`IMPLEMENTATION_STRATEGY.md`](IMPLEMENTATION_STRATEGY.md) — Execution roadmap

---

## First Customer

**GymFlow Track** — A gym management platform. OPTRION monitors their services (payment processing, booking system, member portal) and alerts the team via Telegram when things degrade.

---

## License

MIT

---

## Author

Built by a solo engineer, after work hours, one bounded context at a time.
