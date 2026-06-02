# Optrion Playwright E2E Tests

## Overview

End-to-end tests for the Optrion API using Playwright's API testing capabilities, with scaffolding for future browser-based UI tests.

## Prerequisites

- Node.js 18+
- Optrion API server running (or let Playwright start it via `webServer` config)
- PostgreSQL and Redis available

## Setup

```bash
cd test/e2e-playwright
npm install
npx playwright install
```

## Running Tests

### API Tests (headless, no browser)

```bash
npm run test:api
```

### All Tests (headless)

```bash
npm test
```

### Headed Mode (browser visible)

```bash
npm run test:headed
```

### UI Tests Only (when frontend is available)

```bash
npm run test:ui
npm run test:ui:headed
```

### Debug Mode (step through tests)

```bash
npm run test:debug
```

### View HTML Report

```bash
npm run report
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_URL` | `http://localhost:8080` | API server URL |
| `UI_BASE_URL` | `http://localhost:3000` | Frontend URL |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `optrion_e2e` | PostgreSQL database |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `CI` | - | Set in CI for stricter mode |

## Test Structure

```
tests/
├── fixtures.ts          # Shared test fixtures (auth, registration)
├── api/                 # API endpoint tests
│   ├── health.spec.ts   # Health/readiness endpoints
│   ├── registration.spec.ts  # Tenant registration
│   ├── tenant.spec.ts   # Tenant CRUD operations
│   ├── alerts.spec.ts   # Alert management
│   ├── ai-recommendations.spec.ts  # AI analysis
│   └── security.spec.ts # Security & input validation
└── ui/                  # Browser-based UI tests (future)
    └── dashboard.spec.ts  # Dashboard placeholder tests
```

## Adding New Tests

1. For API tests, add files in `tests/api/` importing from `../fixtures`
2. For UI tests, add files in `tests/ui/` using standard Playwright `test`
3. Use the `registeredTenant` fixture for tests needing an authenticated context
