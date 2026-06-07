# Contributing to Optrion

Thank you for your interest in contributing to Optrion! This guide will help you get started.

## Prerequisites

- **Go 1.26+** — [Install Go](https://go.dev/dl/)
- **Node.js 20+** — [Install Node.js](https://nodejs.org/)
- **PostgreSQL 16** — Required for the backend
- **Redis 7** — Required for caching and rate limiting
- **Docker & Docker Compose** — For running infrastructure locally

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/optrion/optrion.git
cd optrion
```

### 2. Start Infrastructure

```bash
# Start PostgreSQL + Redis via Docker Compose
make infra-up
```

This runs PostgreSQL 16 on port 5432 and Redis 7 on port 6379.

### 3. Run the Backend

```bash
# Run with default configuration (connects to local PostgreSQL + Redis)
make run
```

The API server starts at `http://localhost:8080`.

### 4. Run the Dashboard

```bash
cd dashboard
npm install
npm run dev
```

The dashboard starts at `http://localhost:3000`.

### 5. Register Your First Application

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"name": "My App", "slug": "my-app"}'
```

Save the returned API key — you'll need it for the dashboard.

## Development Workflow

### Running Tests

```bash
# Unit tests (fast, no external dependencies)
make test

# Integration tests (requires PostgreSQL + Redis)
make test-integration

# All tests with race detection
make test-all

# Run a specific test
go test ./internal/health/domain/... -run TestHealthCheckConfig
```

### Code Style

- We use `gofmt` and `golangci-lint` for Go code
- Run `make lint` before submitting a PR
- Follow the existing hexagonal architecture patterns

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the server binary |
| `make run` | Run the server locally |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests |
| `make lint` | Run linter |
| `make docker-up` | Start full stack with Docker Compose |
| `make docker-down` | Stop Docker Compose |
| `make infra-up` | Start PostgreSQL + Redis only |
| `make infra-down` | Stop infrastructure |

## Architecture

Optrion uses **hexagonal architecture** (ports & adapters) with **bounded contexts**:

```
internal/
  {context}/
    domain/       # Domain models, business rules (no dependencies)
    port/         # Interfaces (repositories, services)
    app/          # Application services (use case orchestration)
    adapter/
      rest/       # HTTP handlers (inbound adapter)
      postgres/   # PostgreSQL repositories (outbound adapter)
```

### Key Principles

1. **Domain logic has zero infrastructure dependencies** — domain packages import only stdlib and shared packages
2. **Ports define contracts** — all persistence and external calls go through interfaces
3. **The container is the composition root** — `internal/app/container.go` wires everything
4. **Events for cross-context communication** — bounded contexts communicate via the event bus, never by direct import

### Bounded Contexts

| Context | Package | Responsibility |
|---------|---------|----------------|
| Tenant | `internal/tenant` | Multi-tenancy, products, environments, components, audit |
| Health | `internal/health` | Health checks, scoring, anomaly detection |
| Incident | `internal/incident` | Incident lifecycle, state machine, timeline |
| Alert | `internal/alert` | Alert rules, channels, delivery, escalation |
| AI | `internal/ai` | Root cause analysis via LLM providers |
| Recommendation | `internal/recommendation` | Evidence-based recommendations |
| Registration | `internal/registration` | Plug-and-play onboarding |

## Adding a New Feature

### Adding a New API Endpoint

1. Define the domain model in `internal/{context}/domain/`
2. Add the port interface in `internal/{context}/port/`
3. Implement the application service in `internal/{context}/app/`
4. Create the HTTP handler in `internal/{context}/adapter/rest/`
5. Implement the PostgreSQL repository in `internal/{context}/adapter/postgres/`
6. Wire everything in `internal/app/container.go`
7. Add tests at each layer

### Adding a Database Migration

Create migration files in `migrations/`:

```
migrations/000029_create_my_table.up.sql
migrations/000029_create_my_table.down.sql
```

Migrations are embedded in the binary and run automatically on startup.

### Adding a Notification Channel

1. Create `internal/alert/adapter/{channel}/{channel}_channel.go`
2. Implement the `ChannelSender` interface
3. Add tests in `{channel}_channel_test.go`
4. Wire in the alert engine in `container.go`

## Pull Request Guidelines

1. **One PR per feature/fix** — keep changes focused
2. **Write tests** — aim for 80%+ coverage on new code
3. **Update the OpenAPI spec** — if you add/change API endpoints, update `docs/openapi.yaml`
4. **Don't break existing tests** — run `make test` before pushing
5. **Follow naming conventions** — match existing patterns in the codebase

## API Documentation

- **OpenAPI Spec**: [docs/openapi.yaml](docs/openapi.yaml)
- **Error Codes**: `GET /api/v1/error-codes` returns the full error code catalog
- **Health Probes**: `GET /healthz` (liveness), `GET /readyz` (readiness)
- **Metrics**: `GET /metrics` (Prometheus format)

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `local` | Environment (local, staging, production) |
| `HTTP_PORT` | `8080` | HTTP server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `optrion` | Database user |
| `DB_PASSWORD` | — | Database password |
| `DB_NAME` | `optrion` | Database name |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `AUTH_ENABLED` | `true` | Enable API key authentication |
| `AI_ENABLED` | `false` | Enable AI services |
| `AI_PROVIDER` | `gemini` | AI provider (gemini, openai, anthropic, ollama) |
| `AI_API_KEY` | — | AI provider API key |
| `RATE_LIMIT_RPS` | `100` | Requests per second per tenant |

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
