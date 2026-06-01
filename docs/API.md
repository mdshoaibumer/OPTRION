# OPTRION API Documentation

## Overview

OPTRION provides a comprehensive REST API for platform integration, registration, monitoring, and validation.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All API requests require an API key in the `Authorization` header:

```
Authorization: Bearer optrion_key_xxxxx
```

---

## Endpoints

### Registration

#### POST /register

Register a new application with OPTRION platform atomically. This single request registers tenant, product, environment, and components.

**Request:**
```json
{
  "tenant": {
    "name": "GymFlow",
    "slug": "gymflow",
    "plan": "starter"
  },
  "product": {
    "name": "Backend API",
    "slug": "backend",
    "description": "Main backend service"
  },
  "environment": {
    "name": "Production",
    "tier": "production"
  },
  "components": [
    {
      "name": "PostgreSQL",
      "kind": "database",
      "description": "Main database",
      "endpoint": "postgresql://localhost:5432/gymflow",
      "port": 5432
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "environment_id": "550e8400-e29b-41d4-a716-446655440002",
  "component_ids": [
    "550e8400-e29b-41d4-a716-446655440003"
  ],
  "api_key": "optrion_key_9e7b8c6f5d4a3b2c1f0e9d8c7b6a5f4e",
  "endpoint": "http://localhost:8080",
  "message": "Successfully registered Backend API with 1 components"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request data
- `409 Conflict` - Slug already exists
- `500 Internal Server Error` - Server error

---

### Metrics

#### POST /metrics

Submit application metrics to OPTRION platform.

**Request:**
```json
{
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "environment_id": "550e8400-e29b-41d4-a716-446655440002",
  "timestamp": "2026-05-31T12:00:00Z",
  "metrics": {
    "memory_alloc": 134217728,
    "memory_total": 1073741824,
    "goroutines": 42,
    "gc_runs": 15
  }
}
```

**Response (202 Accepted):**
```json
{
  "status": "accepted",
  "metric_id": "metric-uuid"
}
```

---

#### GET /metrics?tenant_id=xxx&limit=100

Retrieve historical metrics for a tenant.

**Response (200 OK):**
```json
{
  "metrics": [
    {
      "id": "metric-uuid-1",
      "timestamp": "2026-05-31T12:00:30Z",
      "memory_alloc": 134217728,
      "memory_total": 1073741824,
      "goroutines": 42
    }
  ],
  "count": 1,
  "total": 150
}
```

---

### Health Check

#### GET /health/tenant/{tenant_id}

Get comprehensive health status for a tenant and all components.

**Response (200 OK):**
```json
{
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "healthy",
  "timestamp": "2026-05-31T12:00:00Z",
  "overall_health_score": 95,
  "components": [
    {
      "id": "comp-uuid-1",
      "name": "PostgreSQL",
      "kind": "database",
      "status": "healthy",
      "health_score": 98,
      "response_time_ms": 15,
      "last_check": "2026-05-31T12:00:00Z"
    },
    {
      "id": "comp-uuid-2",
      "name": "Redis",
      "kind": "cache",
      "status": "healthy",
      "health_score": 96,
      "response_time_ms": 8,
      "last_check": "2026-05-31T12:00:00Z"
    },
    {
      "id": "comp-uuid-3",
      "name": "Backend",
      "kind": "api",
      "status": "healthy",
      "health_score": 90,
      "response_time_ms": 120,
      "last_check": "2026-05-31T12:00:00Z"
    }
  ]
}
```

---

#### GET /health/component/{component_id}

Get detailed health status for a specific component.

**Response (200 OK):**
```json
{
  "component_id": "comp-uuid-1",
  "name": "PostgreSQL",
  "kind": "database",
  "status": "healthy",
  "health_score": 98,
  "response_time_ms": 15,
  "uptime_hours": 720,
  "last_check": "2026-05-31T12:00:00Z",
  "last_failure": "2026-05-28T14:30:00Z",
  "consecutive_healthy": 432,
  "error_rate": 0.001
}
```

---

### Validation

#### POST /validate

Validate OPTRION integration for a tenant.

**Request:**
```json
{
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "api_key": "optrion_key_xxxxx"
}
```

**Response (200 OK):**
```json
{
  "status": "valid",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "integration_score": 95,
  "components_registered": 3,
  "components_healthy": 3,
  "metrics_flowing": true,
  "last_metrics_received": "2026-05-31T12:00:15Z",
  "issues": [],
  "recommendations": []
}
```

---

#### GET /validate/report/{tenant_id}

Get detailed validation report for a tenant.

**Response (200 OK):**
```json
{
  "report_id": "report-uuid",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "generated_at": "2026-05-31T12:00:00Z",
  "integration_status": "healthy",
  "overall_score": 95,
  "section_scores": {
    "registration": 100,
    "connectivity": 98,
    "metrics": 92,
    "health_checks": 95,
    "performance": 88
  },
  "checks": [
    {
      "name": "Components Registered",
      "status": "passed",
      "value": "3/3",
      "details": "All components successfully registered"
    },
    {
      "name": "Components Healthy",
      "status": "passed",
      "value": "3/3",
      "details": "All components reporting healthy status"
    },
    {
      "name": "Metrics Flowing",
      "status": "passed",
      "value": "Yes",
      "details": "Last metrics received 15 seconds ago"
    }
  ],
  "issues": [],
  "recommendations": [
    "Consider enabling deep diagnostics for better insights"
  ]
}
```

---

### Components

#### GET /components/{environment_id}

List all components in an environment.

**Response (200 OK):**
```json
{
  "environment_id": "env-uuid",
  "components": [
    {
      "id": "comp-uuid-1",
      "name": "PostgreSQL",
      "kind": "database",
      "endpoint": "postgresql://localhost:5432/gymflow",
      "port": 5432,
      "status": "healthy",
      "created_at": "2026-05-31T10:00:00Z"
    }
  ],
  "count": 1
}
```

---

#### PUT /components/{component_id}

Update component configuration.

**Request:**
```json
{
  "name": "PostgreSQL Primary",
  "endpoint": "postgresql://new-host:5432/gymflow",
  "settings": {
    "max_connections": 150
  }
}
```

**Response (200 OK):**
```json
{
  "id": "comp-uuid-1",
  "name": "PostgreSQL Primary",
  "endpoint": "postgresql://new-host:5432/gymflow",
  "updated_at": "2026-05-31T12:00:00Z"
}
```

---

### Incidents

#### GET /incidents?tenant_id=xxx&status=open

List incidents for a tenant.

**Response (200 OK):**
```json
{
  "incidents": [
    {
      "id": "incident-uuid-1",
      "title": "High Memory Usage on Backend API",
      "status": "open",
      "severity": "warning",
      "component_id": "comp-uuid-3",
      "detected_at": "2026-05-31T11:45:00Z",
      "updated_at": "2026-05-31T12:00:00Z",
      "details": {
        "current_value": 85,
        "threshold": 80,
        "unit": "percent"
      }
    }
  ],
  "count": 1,
  "total": 5
}
```

---

### Alerts

#### POST /alerts/{incident_id}/action

Take action on an incident (acknowledge, resolve, escalate).

**Request:**
```json
{
  "action": "acknowledge",
  "message": "Investigating memory leak",
  "user": "engineer@company.com"
}
```

**Response (200 OK):**
```json
{
  "incident_id": "incident-uuid-1",
  "status": "acknowledged",
  "updated_at": "2026-05-31T12:00:00Z"
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": "Additional details if available",
  "timestamp": "2026-05-31T12:00:00Z"
}
```

### Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `invalid_request` | 400 | Request validation failed |
| `unauthorized` | 401 | Invalid or missing API key |
| `forbidden` | 403 | Insufficient permissions |
| `not_found` | 404 | Resource not found |
| `conflict` | 409 | Resource already exists |
| `rate_limited` | 429 | Too many requests |
| `server_error` | 500 | Internal server error |

---

## Rate Limiting

API requests are rate limited per API key:

- **Standard Plan**: 1,000 requests/hour
- **Professional Plan**: 10,000 requests/hour
- **Enterprise Plan**: Unlimited

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1622470800
```

---

## Webhooks

Register webhooks to receive notifications about health changes, incidents, and metrics.

### Supported Events

- `component.health_changed` - Component health status changed
- `incident.created` - New incident detected
- `incident.resolved` - Incident resolved
- `metrics.threshold_exceeded` - Metric exceeded threshold
- `registration.completed` - Registration completed

### Webhook Payload

```json
{
  "event": "component.health_changed",
  "timestamp": "2026-05-31T12:00:00Z",
  "data": {
    "component_id": "comp-uuid-1",
    "previous_status": "healthy",
    "current_status": "degraded",
    "details": {}
  }
}
```

---

## SDK Integration

See the SDK documentation for language-specific implementation:

- [Go SDK](../sdk/go-sdk/README.md)
- [JavaScript SDK](../sdk/js-sdk/README.md)

---

## Examples

### Complete Registration Flow

```bash
# 1. Initialize configuration
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "tenant": {"name": "MyApp", "slug": "myapp", "plan": "starter"},
    "product": {"name": "API", "slug": "api", "description": "Main API"},
    "environment": {"name": "Production", "tier": "production"},
    "components": [
      {"name": "DB", "kind": "database", "endpoint": "localhost", "port": 5432}
    ]
  }'

# Response contains api_key and ids

# 2. Send metrics
curl -X POST http://localhost:8080/api/v1/metrics \
  -H "Authorization: Bearer optrion_key_xxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "...",
    "product_id": "...",
    "environment_id": "...",
    "timestamp": "2026-05-31T12:00:00Z",
    "metrics": {"memory_alloc": 134217728}
  }'

# 3. Check health
curl -X GET http://localhost:8080/api/v1/health/tenant/xxx \
  -H "Authorization: Bearer optrion_key_xxxxx"

# 4. Validate integration
curl -X POST http://localhost:8080/api/v1/validate \
  -H "Authorization: Bearer optrion_key_xxxxx" \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": "...", "api_key": "optrion_key_xxxxx"}'
```

---

## Support

For API support and questions:
- Documentation: https://optrion.dev/docs
- GitHub Issues: https://github.com/optrion/optrion/issues
- Email: support@optrion.dev
