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
│   ├── platform/            # Cross-cutting infrastructure
│   │   ├── config/          # Configuration loading + validation
│   │   ├── logger/          # Structured logging with context
│   │   ├── database/        # PostgreSQL pool + health checks
│   │   ├── cache/           # Redis client + health checks
│   │   └── server/          # HTTP server, router, middleware
│   └── shared/              # Shared kernel
│       ├── id/              # UUID v7 generation
│       └── domain/          # Shared domain errors
├── deploy/docker/           # Dockerfile + docker-compose
├── docs/                    # Architecture documentation
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

## API Endpoints (Phase 1 — Foundation)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/healthz` | Liveness probe (always 200) |
| `GET` | `/readyz` | Readiness probe (checks all deps) |
| `GET` | `/api/v1/info` | Application version + environment |

---

## Health Check Response

```json
// GET /readyz
{
  "status": "healthy",
  "timestamp": "2026-05-29T18:30:00Z",
  "version": "0.1.0",
  "checks": {
    "postgresql": { "status": "healthy" },
    "redis": { "status": "healthy" }
  }
}
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
| 2 | 🔲 Planned | Database migrations + Shared kernel |
| 3 | 🔲 Planned | Tenant bounded context |
| 4 | 🔲 Planned | Catalog bounded context |
| 5 | 🔲 Planned | Execution (health check scheduling) |
| 6 | 🔲 Planned | Intelligence (health scoring) |
| 7 | 🔲 Planned | Incident detection |
| 8 | 🔲 Planned | Alerting + Notifications |
| 9 | 🔲 Planned | API authentication |
| 10 | 🔲 Planned | Dashboard (Next.js) |

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
