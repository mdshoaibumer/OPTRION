# OPTRION — Database Architecture

**Author:** Principal Database Architect  
**Date:** 2026-05-29  
**Version:** 1.0  
**Status:** Design Complete — Pre-Migration Review

---

## Database Architecture Review

### Design Philosophy

OPTRION's database must serve two masters simultaneously:

1. **High-throughput time-series writes** — Check results and metrics arrive every 30 seconds per component
2. **Low-latency tenant-scoped reads** — Dashboards, incident lists, and health scores must respond in <50ms

These are fundamentally different workloads. The architecture must optimize for both without compromise.

### Key Constraints

| Constraint | Impact |
|-----------|--------|
| Multi-tenant shared database | Every query must filter by tenant_id. No exceptions |
| Append-only metrics | Immutable writes dominate. Updates are rare |
| Real-time dashboards | Current state must be instantly queryable |
| Event-driven architecture | Outbox table must handle sustained write bursts |
| Future AI training | Historical data must remain queryable for months |
| SaaS commercialization | Per-tenant resource accounting required |

---

## TimescaleDB Decision

### Option A: PostgreSQL Only

**Pros:**
- Zero additional operational complexity
- Native partitioning (PARTITION BY RANGE) covers basic time-based retention
- No extension dependency in managed cloud databases
- Simpler backup/restore
- Full compatibility with all PostgreSQL hosting providers

**Cons:**
- Manual partition management (create monthly, drop expired)
- No built-in continuous aggregates (must build materialized views + refresh logic)
- No built-in compression for old chunks
- No native downsampling (must write custom aggregation jobs)
- Query planning less optimized for time-range scans across many partitions

**Operational Complexity:** Low  
**Break Point:** ~5M metrics/day without significant optimization effort

### Option B: PostgreSQL + TimescaleDB

**Pros:**
- Automatic chunk management (no manual partition creation/dropping)
- Built-in compression (10-20x for old metrics — critical at scale)
- Continuous aggregates (real-time materialized rollups)
- Native retention policies (`drop_chunks` with one policy)
- Optimized query planning for time-range + tenant_id queries
- Downsampling built-in (1-minute raw → 5-minute → 1-hour → 1-day rollups)
- Hyperfunctions for percentile calculations (approx_percentile, time_bucket)

**Cons:**
- Extension dependency (limits hosting options to TimescaleDB Cloud, self-hosted, or Aiven)
- Additional operational knowledge required
- Licensing consideration (Apache 2.0 for core, Timescale License for some features)
- Slightly more complex schema migrations
- Not available on all managed PostgreSQL services (AWS RDS does NOT support it; Aurora does NOT)

**Operational Complexity:** Medium  
**Break Point:** Handles 100M+ metrics/day with proper chunk sizing

### Alternative Architecture: PostgreSQL + ClickHouse

| Aspect | PostgreSQL + TimescaleDB | PostgreSQL + ClickHouse |
|--------|------------------------|------------------------|
| Operational cost | One database engine | Two database engines |
| Query language | SQL (unified) | SQL (two dialects) |
| Write throughput | Good (millions/day) | Excellent (billions/day) |
| Point queries | Excellent | Poor |
| Aggregation | Good | Excellent |
| Maturity | High | High |
| Team knowledge needed | PostgreSQL + extension | PostgreSQL + column store |

### Recommendation: Option B — PostgreSQL + TimescaleDB

**Phase 1:** PostgreSQL with native RANGE partitioning. Do NOT add TimescaleDB yet.  
**Phase 2 trigger:** When daily metric volume exceeds 2M rows OR when retention management becomes a maintenance burden, migrate time-series tables to TimescaleDB hypertables.

**Rationale:**
1. At 10 tenants / 100K metrics per day, native partitioning is sufficient
2. TimescaleDB migration is non-destructive (convert existing partitioned table to hypertable)
3. Starting with plain PostgreSQL avoids hosting lock-in during early customer acquisition
4. The schema design below is TimescaleDB-ready — partition keys and chunk strategies are pre-designed

**Critical Rule:** Design the schema as if TimescaleDB will be added. Use `timestamp` + `tenant_id` as the natural key for all time-series tables. This makes the future migration a one-line operation per table.

---

## Schema Design

### Schema Organization

```
optrion (database)
├── public              -- Shared types, extensions, functions
├── core                -- Tenant hierarchy (tenants, products, environments, components)
├── monitoring          -- Health checks, results, metrics, scores
├── incidents           -- Incidents, events, timeline
├── alerting            -- Rules, history, channels, maintenance windows
├── intelligence        -- Baselines, AI analyses, predictions (future)
├── system              -- Outbox, audit, jobs
```

**Why schemas instead of one flat namespace:**
- Logical grouping improves developer navigation
- Allows per-schema GRANT policies (read-only access to `monitoring` for dashboards)
- Makes future service extraction cleaner (each schema = potential service boundary)
- Prevents table name collisions as the platform grows

---

### Core Tables

#### `core.tenants`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK, time-sortable |
| name | text | Display name |
| slug | text | URL-safe unique identifier |
| plan | text | Plan tier (free, starter, professional, enterprise) |
| status | text | active, suspended, terminated |
| settings | jsonb | Tenant-level configuration |
| created_at | timestamptz | |
| updated_at | timestamptz | |

**Purpose:** Root identity for all tenant-scoped data  
**Expected Volume:** 10 → 100 → 1000 rows  
**Retention:** Permanent (soft delete only)  
**Growth Risk:** None

---

#### `core.api_keys`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | FK → tenants. NOT NULL |
| name | text | Human label ("Production API Key") |
| key_prefix | text | First 8 chars for identification |
| key_hash | text | SHA-256 hash of full key |
| scopes | text[] | Array of allowed scopes |
| last_used_at | timestamptz | |
| expires_at | timestamptz | Nullable (null = no expiry) |
| status | text | active, revoked |
| created_at | timestamptz | |

**Purpose:** Machine authentication for REST API  
**Expected Volume:** 2-5 per tenant  
**Retention:** Permanent (revoked keys kept for audit)  
**Growth Risk:** None

---

#### `core.products`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | FK → tenants. NOT NULL |
| name | text | |
| description | text | |
| status | text | active, archived |
| created_at | timestamptz | |
| updated_at | timestamptz | |

**Purpose:** Groups environments under a logical product  
**Expected Volume:** 1-5 per tenant  
**Retention:** Permanent (soft delete)  
**Growth Risk:** None

---

#### `core.environments`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | FK → tenants. NOT NULL (denormalized for query performance) |
| product_id | uuid | FK → products |
| name | text | |
| type | text | production, staging, development, custom |
| policy | jsonb | Check frequency, retention, alerting behavior |
| status | text | active, decommissioned |
| created_at | timestamptz | |
| updated_at | timestamptz | |

**Purpose:** Deployment context for components  
**Expected Volume:** 2-4 per product  
**Retention:** Permanent  
**Growth Risk:** None

**Design Note:** `tenant_id` is denormalized here (could be derived via product). This is intentional — it eliminates a JOIN on every tenant-scoped query touching environments.

---

#### `core.components`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | FK → tenants. NOT NULL |
| environment_id | uuid | FK → environments |
| name | text | |
| type | text | database, cache, api, frontend, queue, worker, external_service |
| description | text | |
| metadata | jsonb | Component-specific attributes |
| status | text | active, degraded, unhealthy, deregistered |
| created_at | timestamptz | |
| updated_at | timestamptz | |

**Purpose:** The atomic unit of monitoring  
**Expected Volume:** 3-20 per environment  
**Retention:** Permanent (deregistered = soft delete)  
**Growth Risk:** Low. ~50K at 1000 tenants

---

#### `core.component_dependencies`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | FK → tenants |
| source_component_id | uuid | FK → components (the dependent) |
| target_component_id | uuid | FK → components (the dependency) |
| dependency_type | text | hard, soft |
| created_at | timestamptz | |

**Purpose:** Dependency graph for root cause analysis  
**Expected Volume:** 1-5 per component  
**Retention:** Permanent  
**Growth Risk:** None

---

### Monitoring Tables

#### `monitoring.health_checks`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | FK → tenants |
| component_id | uuid | FK → components |
| name | text | Human-readable check name |
| type | text | http, tcp, dns, tls_expiry, custom |
| target | jsonb | Type-specific: {url, method, headers, expected_status, timeout_ms} |
| assertion | jsonb | Success criteria: {conditions: [{field, operator, value}]} |
| schedule_seconds | integer | Check interval |
| retry_policy | jsonb | {max_attempts, backoff_ms} |
| status | text | active, paused, disabled |
| created_at | timestamptz | |
| updated_at | timestamptz | |

**Purpose:** Configuration for what to check and how  
**Expected Volume:** 1-5 per component  
**Retention:** Permanent  
**Growth Risk:** None

---

#### `monitoring.check_results`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL — partition key |
| health_check_id | uuid | FK → health_checks |
| component_id | uuid | Denormalized for direct component queries |
| status | text | healthy, unhealthy, timeout, error |
| latency_ms | integer | Response time |
| response_code | smallint | HTTP status code (nullable for non-HTTP) |
| error_message | text | Nullable — only on failure |
| metadata | jsonb | Response headers, body excerpt, TLS info |
| executed_at | timestamptz | NOT NULL — partition key, time of execution |

**Purpose:** Immutable record of every check execution  
**Expected Volume:** HIGHEST. ~2,880 per check per day (30s interval). At 1000 tenants × 50 components × 3 checks = 432M rows/day  
**Retention:** 7 days raw (development), 30 days (staging), 90 days (production). Configurable per environment policy  
**Growth Risk:** CRITICAL. This is the table that will kill you if not partitioned correctly

---

#### `monitoring.metric_snapshots`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL — partition key |
| component_id | uuid | FK → components |
| metric_type | text | availability, latency_p50, latency_p95, latency_p99, error_rate, throughput |
| value | double precision | Metric value |
| unit | text | ms, percent, requests_per_second |
| collected_at | timestamptz | NOT NULL — partition key |

**Purpose:** Derived/aggregated metrics computed from check results. Lower volume than raw results  
**Expected Volume:** HIGH. ~96 per component per day (15-min aggregation windows)  
**Retention:** 30 days granular, 1 year downsampled (hourly), forever downsampled (daily)  
**Growth Risk:** High but manageable with downsampling

---

#### `monitoring.health_scores`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL — partition key |
| subject_type | text | component, environment, product |
| subject_id | uuid | References the scored entity |
| score | smallint | 0-100 |
| factors | jsonb | Contributing factors with weights and raw values |
| computed_at | timestamptz | NOT NULL — partition key |

**Purpose:** Historical health score records for trending and baseline comparison  
**Expected Volume:** Medium-High. ~96 per subject per day (15-min computation)  
**Retention:** 90 days granular, 1 year daily averages  
**Growth Risk:** Medium. Manageable with downsampling

---

#### `monitoring.component_health`

| Column | Type | Notes |
|--------|------|-------|
| component_id | uuid | PK, FK → components |
| tenant_id | uuid | NOT NULL |
| status | text | healthy, degraded, unhealthy, unknown |
| score | smallint | Current score 0-100 |
| last_check_at | timestamptz | |
| last_success_at | timestamptz | |
| last_failure_at | timestamptz | |
| consecutive_successes | integer | |
| consecutive_failures | integer | |
| updated_at | timestamptz | |

**Purpose:** Current state snapshot for real-time dashboard. One row per component. Updated on every check result  
**Expected Volume:** 1 row per component. ~50K at 1000 tenants  
**Retention:** Permanent (row is updated, not appended)  
**Growth Risk:** None. This is a hot-update table — monitor for bloat

---

### Incident Tables

#### `incidents.incidents`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL — partition key |
| component_id | uuid | FK → components (primary affected) |
| environment_id | uuid | Denormalized |
| fingerprint | text | Deduplication key |
| title | text | Human-readable incident title |
| description | text | |
| severity | text | critical, high, medium, low |
| status | text | opened, acknowledged, investigating, resolved, closed |
| trigger_type | text | threshold, anomaly, manual |
| trigger_details | jsonb | What condition triggered this |
| opened_at | timestamptz | NOT NULL |
| acknowledged_at | timestamptz | |
| resolved_at | timestamptz | |
| closed_at | timestamptz | |
| duration_seconds | integer | Computed on resolution |
| metadata | jsonb | |

**Purpose:** Tracks detected problems through their lifecycle  
**Expected Volume:** 1-10 per component per month (depends on stability)  
**Retention:** 1 year active, archive forever (incidents are business records)  
**Growth Risk:** Low

---

#### `incidents.incident_events`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL |
| incident_id | uuid | FK → incidents |
| event_type | text | opened, acknowledged, severity_changed, note_added, escalated, resolved, closed |
| actor | text | system, user:{id}, automation:{name} |
| payload | jsonb | Event-specific data |
| occurred_at | timestamptz | NOT NULL |

**Purpose:** Immutable timeline of everything that happened during an incident  
**Expected Volume:** 5-20 per incident  
**Retention:** Same as parent incident  
**Growth Risk:** Low

---

### Alerting Tables

#### `alerting.notification_channels`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | FK → tenants |
| name | text | Human label |
| type | text | telegram, email, webhook, slack |
| config | bytea | ENCRYPTED. Channel-specific configuration |
| verified | boolean | Has connectivity been confirmed |
| status | text | active, disabled |
| created_at | timestamptz | |
| updated_at | timestamptz | |

**Purpose:** Delivery endpoints for alerts  
**Expected Volume:** 1-5 per tenant  
**Retention:** Permanent  
**Growth Risk:** None

**Security Note:** `config` is bytea (encrypted) not jsonb. Contains secrets (webhook URLs, API tokens, chat IDs). Encrypted with per-tenant key using envelope encryption. Application decrypts at delivery time only.

---

#### `alerting.alert_rules`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | FK → tenants |
| name | text | |
| scope_type | text | component, environment, product |
| scope_id | uuid | References the scoped entity |
| condition | jsonb | {metric, operator, threshold, window, consecutive_count} |
| severity | text | Severity to assign when triggered |
| channel_ids | uuid[] | Notification channels to use |
| cooldown_seconds | integer | Minimum time between repeated alerts |
| status | text | active, muted, disabled |
| created_at | timestamptz | |
| updated_at | timestamptz | |

**Purpose:** Defines when and how to alert  
**Expected Volume:** 5-20 per tenant  
**Retention:** Permanent  
**Growth Risk:** None

---

#### `alerting.alert_history`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL — partition key |
| alert_rule_id | uuid | FK → alert_rules |
| incident_id | uuid | FK → incidents (nullable) |
| severity | text | |
| trigger_value | double precision | The value that triggered the alert |
| channels_targeted | text[] | Which channels were targeted |
| delivery_status | jsonb | Per-channel delivery result |
| suppressed | boolean | Was this alert suppressed (cooldown/maintenance) |
| suppression_reason | text | |
| triggered_at | timestamptz | NOT NULL — partition key |
| delivered_at | timestamptz | |

**Purpose:** Audit trail of every alert evaluation  
**Expected Volume:** 10-50 per tenant per day  
**Retention:** 90 days  
**Growth Risk:** Low-Medium

---

#### `alerting.maintenance_windows`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | FK → tenants |
| scope_type | text | component, environment, product |
| scope_id | uuid | |
| reason | text | |
| starts_at | timestamptz | |
| ends_at | timestamptz | |
| status | text | scheduled, active, completed, cancelled |
| created_by | text | |
| created_at | timestamptz | |

**Purpose:** Suppress alerts during planned maintenance  
**Expected Volume:** 1-5 per tenant per week  
**Retention:** 1 year  
**Growth Risk:** None

---

### Audit Tables

#### `system.audit_log`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL — partition key |
| actor | text | user:{id}, system, api_key:{prefix} |
| action | text | created, updated, deleted, triggered, acknowledged |
| resource_type | text | component, health_check, alert_rule, incident, etc. |
| resource_id | uuid | |
| changes | jsonb | Before/after diff for updates |
| ip_address | inet | |
| user_agent | text | |
| occurred_at | timestamptz | NOT NULL — partition key |

**Purpose:** Immutable record of all significant actions for compliance and debugging  
**Expected Volume:** 50-200 per tenant per day  
**Retention:** 1 year minimum. 7 years for enterprise (compliance)  
**Growth Risk:** Medium at scale. Partition aggressively

---

#### `system.event_outbox`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL |
| event_type | text | Domain event type name |
| aggregate_type | text | Source aggregate |
| aggregate_id | uuid | Source aggregate root ID |
| payload | jsonb | Full event payload |
| status | text | pending, processing, processed, failed |
| created_at | timestamptz | NOT NULL |
| processed_at | timestamptz | |
| retry_count | smallint | Default 0 |
| next_retry_at | timestamptz | For exponential backoff |

**Purpose:** Transactional outbox for reliable event delivery  
**Expected Volume:** Matches total domain event volume. ~10K-100K per day  
**Retention:** 24 hours after processing (aggressively purged)  
**Growth Risk:** HIGH if processing stalls. Must monitor queue depth. Dead letter after 5 retries

---

### Intelligence Tables (Future — Design Now, Build Later)

#### `intelligence.baselines`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL |
| component_id | uuid | FK → components |
| metric_type | text | |
| window_hours | integer | Baseline computation window (e.g., 168 = 1 week) |
| p50 | double precision | |
| p95 | double precision | |
| p99 | double precision | |
| mean | double precision | |
| stddev | double precision | |
| sample_count | integer | |
| status | text | computing, active, stale |
| computed_at | timestamptz | |

**Purpose:** Statistical normal for anomaly detection  
**Expected Volume:** 1 per metric_type per component (updated weekly)  
**Retention:** Keep last 4 versions per component+metric  
**Growth Risk:** None

---

#### `intelligence.ai_analyses`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL |
| subject_type | text | incident, component, environment |
| subject_id | uuid | |
| analysis_type | text | root_cause, impact_assessment, trend_explanation |
| provider | text | gemini, openai, local |
| prompt_hash | text | For deduplication and caching |
| input_context | jsonb | What data was provided to the AI |
| output | jsonb | Structured AI response |
| confidence | double precision | 0.0-1.0 |
| feedback | text | accepted, rejected, ignored (nullable) |
| model_version | text | |
| tokens_used | integer | For cost tracking |
| created_at | timestamptz | |

**Purpose:** Store AI-generated insights for display and feedback loop  
**Expected Volume:** 1-5 per incident  
**Retention:** 1 year  
**Growth Risk:** Low

---

#### `intelligence.ai_recommendations`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL |
| component_id | uuid | |
| recommendation_type | text | add_check, adjust_threshold, investigate, scale_resource |
| title | text | |
| description | text | |
| priority | text | high, medium, low |
| status | text | pending, accepted, dismissed, expired |
| evidence | jsonb | Data points supporting the recommendation |
| created_at | timestamptz | |
| acted_at | timestamptz | |

**Purpose:** Proactive suggestions from AI analysis  
**Expected Volume:** 1-10 per tenant per week  
**Retention:** 6 months  
**Growth Risk:** None

---

#### `intelligence.ai_predictions`

| Column | Type | Notes |
|--------|------|-------|
| id | uuid (v7) | PK |
| tenant_id | uuid | NOT NULL |
| component_id | uuid | |
| prediction_type | text | degradation, capacity_exhaustion, incident_recurrence |
| predicted_at | timestamptz | When the prediction was made |
| predicted_for | timestamptz | When the predicted event is expected |
| confidence | double precision | |
| input_features | jsonb | Feature vector used |
| outcome | text | correct, incorrect, pending (evaluated after predicted_for passes) |
| created_at | timestamptz | |

**Purpose:** Predictive intelligence with outcome tracking for model improvement  
**Expected Volume:** 1-5 per component per week  
**Retention:** 1 year (needed for model evaluation)  
**Growth Risk:** None

---

## Index Strategy

### Core Table Indexes

| Table | Index | Columns | Type | Purpose |
|-------|-------|---------|------|---------|
| tenants | pk | id | B-tree (PK) | Primary lookup |
| tenants | idx_tenants_slug | slug | B-tree UNIQUE | Slug-based lookup |
| products | pk | id | B-tree (PK) | Primary lookup |
| products | idx_products_tenant | tenant_id, status | B-tree | List products for tenant |
| environments | pk | id | B-tree (PK) | Primary lookup |
| environments | idx_envs_tenant_product | tenant_id, product_id, status | B-tree | List environments for product |
| components | pk | id | B-tree (PK) | Primary lookup |
| components | idx_components_tenant_env | tenant_id, environment_id, status | B-tree | List components for environment |
| components | idx_components_tenant_type | tenant_id, type | B-tree | Filter by component type |

### Monitoring Table Indexes

| Table | Index | Columns | Type | Purpose |
|-------|-------|---------|------|---------|
| health_checks | pk | id | B-tree (PK) | Primary lookup |
| health_checks | idx_checks_component | tenant_id, component_id, status | B-tree | List active checks for component |
| health_checks | idx_checks_schedule | status, schedule_seconds | B-tree | Scheduler finds due checks |
| **check_results** | pk | id, tenant_id, executed_at | B-tree (PK, partition key) | Primary lookup |
| **check_results** | idx_results_check_time | tenant_id, health_check_id, executed_at DESC | B-tree | Latest results for a check |
| **check_results** | idx_results_component_time | tenant_id, component_id, executed_at DESC | B-tree | Latest results for a component (dashboard) |
| **check_results** | idx_results_status | tenant_id, status, executed_at DESC | B-tree (partial: WHERE status != 'healthy') | Find failures quickly |
| component_health | pk | component_id | B-tree (PK) | Direct lookup |
| component_health | idx_health_tenant_status | tenant_id, status | B-tree | Dashboard: "show me unhealthy components" |
| component_health | idx_health_tenant_score | tenant_id, score | B-tree | Dashboard: "sort by health score" |
| **metric_snapshots** | pk | id, tenant_id, collected_at | B-tree (PK, partition key) | Primary lookup |
| **metric_snapshots** | idx_metrics_component_type_time | tenant_id, component_id, metric_type, collected_at DESC | B-tree | Time-series query for specific metric |
| **health_scores** | pk | id, tenant_id, computed_at | B-tree (PK, partition key) | Primary lookup |
| **health_scores** | idx_scores_subject_time | tenant_id, subject_type, subject_id, computed_at DESC | B-tree | Score history for dashboard graphs |

### Incident Table Indexes

| Table | Index | Columns | Type | Purpose |
|-------|-------|---------|------|---------|
| incidents | pk | id | B-tree (PK) | Primary lookup |
| incidents | idx_incidents_tenant_status | tenant_id, status, opened_at DESC | B-tree | Active incidents dashboard |
| incidents | idx_incidents_component | tenant_id, component_id, status | B-tree | Incidents for a specific component |
| incidents | idx_incidents_fingerprint | tenant_id, fingerprint, status | B-tree UNIQUE (partial: WHERE status IN ('opened','acknowledged','investigating')) | Deduplication: only one active incident per fingerprint |
| incidents | idx_incidents_severity | tenant_id, severity, status | B-tree | Filter by severity |
| incident_events | pk | id | B-tree (PK) | Primary lookup |
| incident_events | idx_events_incident | incident_id, occurred_at | B-tree | Timeline for an incident |

### Alerting Table Indexes

| Table | Index | Columns | Type | Purpose |
|-------|-------|---------|------|---------|
| alert_rules | pk | id | B-tree (PK) | Primary lookup |
| alert_rules | idx_rules_tenant_scope | tenant_id, scope_type, scope_id, status | B-tree | Find rules for a scope |
| alert_rules | idx_rules_active | tenant_id, status | B-tree (partial: WHERE status = 'active') | Rule evaluation: only active rules |
| **alert_history** | pk | id, tenant_id, triggered_at | B-tree (PK, partition key) | Primary lookup |
| **alert_history** | idx_history_tenant_time | tenant_id, triggered_at DESC | B-tree | Alert history dashboard |
| **alert_history** | idx_history_rule_time | tenant_id, alert_rule_id, triggered_at DESC | B-tree | History for a specific rule |
| maintenance_windows | idx_windows_active | tenant_id, status, starts_at, ends_at | B-tree | Check if a window is active for a scope |

### System Table Indexes

| Table | Index | Columns | Type | Purpose |
|-------|-------|---------|------|---------|
| event_outbox | idx_outbox_pending | status, created_at | B-tree (partial: WHERE status = 'pending') | Worker finds unprocessed events |
| event_outbox | idx_outbox_retry | status, next_retry_at | B-tree (partial: WHERE status = 'failed') | Retry worker finds due retries |
| **audit_log** | pk | id, tenant_id, occurred_at | B-tree (PK, partition key) | Primary lookup |
| **audit_log** | idx_audit_tenant_time | tenant_id, occurred_at DESC | B-tree | Audit log viewer |
| **audit_log** | idx_audit_resource | tenant_id, resource_type, resource_id | B-tree | "What happened to this entity?" |

### Index Design Principles

1. **Every index starts with `tenant_id`** — PostgreSQL will use the index for tenant-scoped queries even if other columns are specified
2. **Partial indexes for status filters** — Most queries filter `WHERE status = 'active'` or `WHERE status != 'healthy'`. Partial indexes are smaller and faster
3. **DESC on time columns** — "Most recent first" is the default sort for dashboards. DESC index avoids reverse scan
4. **No GIN indexes on JSONB in Phase 1** — Only add if query patterns demand filtering inside JSONB. Expression indexes on specific JSONB paths are preferred
5. **Bold table names** indicate partitioned tables — indexes are per-partition (local indexes)

---

## Partition Strategy

### Phase 1: Native PostgreSQL Range Partitioning

#### `monitoring.check_results`

| Aspect | Design |
|--------|--------|
| Partition type | RANGE |
| Partition key | `(tenant_id, executed_at)` — composite for list+range when needed |
| Partition granularity | Monthly |
| Naming convention | `check_results_2026_06`, `check_results_2026_07` |
| Retention | Drop partitions older than environment policy (7/30/90 days) |
| Creation | Pre-create 3 months ahead via scheduled job |

**Why monthly not daily:** At Phase 1 volumes (100K/day), daily partitions create 365 child tables per year — excessive. Monthly gives 12 manageable partitions with sufficient pruning benefit.

**Phase 2 migration:** Convert to TimescaleDB hypertable with 1-day chunks when volume exceeds 2M/day.

#### `monitoring.metric_snapshots`

| Aspect | Design |
|--------|--------|
| Partition type | RANGE |
| Partition key | `collected_at` |
| Partition granularity | Monthly |
| Retention | 30 days raw, downsample then drop |

#### `monitoring.health_scores`

| Aspect | Design |
|--------|--------|
| Partition type | RANGE |
| Partition key | `computed_at` |
| Partition granularity | Monthly |
| Retention | 90 days granular, 1 year daily |

#### `alerting.alert_history`

| Aspect | Design |
|--------|--------|
| Partition type | RANGE |
| Partition key | `triggered_at` |
| Partition granularity | Monthly |
| Retention | 90 days |

#### `system.audit_log`

| Aspect | Design |
|--------|--------|
| Partition type | RANGE |
| Partition key | `occurred_at` |
| Partition granularity | Monthly |
| Retention | 1 year standard, 7 years enterprise |

#### `system.event_outbox`

| Aspect | Design |
|--------|--------|
| Partition type | RANGE |
| Partition key | `created_at` |
| Partition granularity | Daily |
| Retention | 48 hours (aggressively pruned) |

**Why daily for outbox:** Outbox rows are short-lived (<24h). Daily partitions allow fast `TRUNCATE` of yesterday's processed events without vacuum overhead.

### Downsampling Strategy (Phase 2+)

```
Raw (30s interval) ──[7-90 days]──→ DROP
        │
        ▼ aggregate
5-minute rollups ──[30 days]──→ DROP
        │
        ▼ aggregate
1-hour rollups ──[1 year]──→ DROP
        │
        ▼ aggregate
1-day rollups ──[forever]──→ ARCHIVE
```

Rollup tables:
- `monitoring.metric_snapshots_5min`
- `monitoring.metric_snapshots_1hour`
- `monitoring.metric_snapshots_1day`

Same schema as `metric_snapshots` but with additional `sample_count` and `min`/`max` columns. With TimescaleDB, these become continuous aggregates with automatic refresh.

---

## Tenant Isolation Strategy

### Option Analysis

#### Option A: Shared Schema, Row-Level Isolation (`tenant_id` in every table)

**How it works:**  
Every table has a `tenant_id` column. Every query includes `WHERE tenant_id = ?`. Application layer enforces this via repository pattern.

| Aspect | Assessment |
|--------|-----------|
| Implementation complexity | Low |
| Operational complexity | Low (one database, one schema, one set of migrations) |
| Tenant provisioning time | Instant (insert one row) |
| Cross-tenant query risk | Medium (requires discipline + RLS) |
| Connection pooling | Efficient (shared pool) |
| Backup/restore granularity | Whole database only (per-tenant export requires application logic) |
| Index efficiency | Good (tenant_id as prefix = partition-like pruning) |
| Scaling ceiling | ~1000 tenants before hot spots emerge |

#### Option B: Schema Per Tenant

**How it works:**  
Each tenant gets their own PostgreSQL schema (`tenant_abc.components`, `tenant_def.components`). Search path switches per request.

| Aspect | Assessment |
|--------|-----------|
| Implementation complexity | Medium |
| Operational complexity | HIGH (N schemas × M tables = N×M objects to manage migrations for) |
| Tenant provisioning time | Seconds (run full migration set) |
| Cross-tenant query risk | Low (schema boundary prevents accidental access) |
| Connection pooling | Complicated (must switch schema per request) |
| Backup/restore granularity | Per-schema export possible |
| Index efficiency | Smaller indexes per tenant (better cache hit rate) |
| Scaling ceiling | ~200 schemas before `pg_catalog` bloats and DDL becomes slow |

#### Option C: Database Per Tenant

**How it works:**  
Each tenant gets their own PostgreSQL database. Completely isolated.

| Aspect | Assessment |
|--------|-----------|
| Implementation complexity | High |
| Operational complexity | EXTREME (N databases to patch, backup, monitor, migrate) |
| Tenant provisioning time | 30+ seconds (create database, run migrations) |
| Cross-tenant query risk | None (complete isolation) |
| Connection pooling | Very complicated (connection per database) |
| Backup/restore granularity | Perfect (per-database backup) |
| Index efficiency | Best (dedicated resources) |
| Scaling ceiling | ~50 databases per PostgreSQL instance |

### Recommendation: Option A with Row-Level Security (RLS) as Defense-in-Depth

**Primary isolation:** Application-layer `tenant_id` filtering via repository pattern  
**Secondary isolation:** PostgreSQL Row-Level Security policies  
**Tertiary isolation:** Integration tests that verify no cross-tenant data leakage

**Rationale:**

1. At the expected scale (10 → 1000 tenants), Option A handles the load without operational burden
2. Schema-per-tenant (B) breaks at ~200 tenants and makes migrations a nightmare
3. Database-per-tenant (C) is only justified for enterprise customers with regulatory requirements (offer as "Dedicated" tier at 10x price)
4. RLS provides a safety net — even if application code has a bug, PostgreSQL enforces isolation at the query planner level
5. Partitioning by `tenant_id` prefix in composite partition keys gives performance benefits similar to schema isolation

### RLS Policy Design

Every table with `tenant_id` gets a policy:

```
Policy: tenant_isolation
  ON table_name
  FOR ALL
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::uuid)
```

Application sets `app.current_tenant_id` session variable at connection checkout from pool. This variable is validated against the authenticated API key/JWT.

### Migration Path to Schema-Per-Tenant (If Ever Needed)

If a single tenant's data volume justifies isolation (e.g., enterprise customer with 10K components):
1. Create dedicated schema for that tenant
2. Migrate their data from shared tables to dedicated schema
3. Route their requests to the dedicated schema via tenant configuration
4. Shared tables retain data for all other tenants

This is an **escape hatch**, not the default path.

---

## Query Patterns

### Pattern 1: Health Dashboard (Real-Time)

**Query:** "Show me all components and their current health for environment X"

```
Path: component_health JOIN components
Filter: tenant_id = ? AND environment_id = ?
Sort: score ASC (worst first)
Expected rows: 5-20
Target latency: <10ms
```

**Index used:** `idx_health_tenant_status` or `idx_components_tenant_env`  
**Why fast:** `component_health` is a small table (1 row per component). Direct lookup, no aggregation needed.

---

### Pattern 2: Health Score Trend (Graph)

**Query:** "Show me health score over the last 24 hours for component X"

```
Path: health_scores
Filter: tenant_id = ? AND subject_type = 'component' AND subject_id = ? AND computed_at > now() - interval '24h'
Sort: computed_at ASC
Expected rows: ~96 (15-min intervals)
Target latency: <20ms
```

**Index used:** `idx_scores_subject_time`  
**Why fast:** Narrow time range + specific subject. Index seek + sequential scan of ~96 rows.

---

### Pattern 3: Latest Check Results (Component Detail)

**Query:** "Show me the last 20 check results for component X"

```
Path: check_results
Filter: tenant_id = ? AND component_id = ? AND executed_at > now() - interval '1h'
Sort: executed_at DESC
Limit: 20
Expected rows: 20
Target latency: <15ms
```

**Index used:** `idx_results_component_time`  
**Why fast:** Index provides exact ordering. Limit stops scan after 20 rows. Partition pruning eliminates old partitions.

---

### Pattern 4: Active Incidents Dashboard

**Query:** "Show me all open/acknowledged incidents for this tenant"

```
Path: incidents
Filter: tenant_id = ? AND status IN ('opened', 'acknowledged', 'investigating')
Sort: severity (critical first), opened_at DESC
Expected rows: 0-10 (hopefully)
Target latency: <10ms
```

**Index used:** `idx_incidents_tenant_status`  
**Why fast:** Very few active incidents at any time. Index seek with tiny result set.

---

### Pattern 5: Incident Timeline

**Query:** "Show me the full timeline for incident X"

```
Path: incident_events
Filter: incident_id = ?
Sort: occurred_at ASC
Expected rows: 5-50
Target latency: <10ms
```

**Index used:** `idx_events_incident`  
**Why fast:** Direct foreign key lookup. Small result set.

---

### Pattern 6: Alert History (Last 7 Days)

**Query:** "Show me all alerts fired in the last 7 days for this tenant"

```
Path: alert_history
Filter: tenant_id = ? AND triggered_at > now() - interval '7 days' AND suppressed = false
Sort: triggered_at DESC
Expected rows: 10-100
Target latency: <30ms
```

**Index used:** `idx_history_tenant_time` with partition pruning  
**Why fast:** Time-range filter prunes to 1-2 partitions. Index on tenant + time gives direct access.

---

### Pattern 7: Incident Deduplication Check

**Query:** "Is there already an active incident for this component with this fingerprint?"

```
Path: incidents
Filter: tenant_id = ? AND fingerprint = ? AND status IN ('opened', 'acknowledged', 'investigating')
Expected rows: 0 or 1
Target latency: <5ms
```

**Index used:** `idx_incidents_fingerprint` (partial unique index)  
**Why fast:** Unique index on fingerprint + status filter. At most 1 row returned.

---

### Pattern 8: Metric Aggregation for Health Scoring

**Query:** "Compute availability, p95 latency, and error rate for component X over last 15 minutes"

```
Path: check_results
Filter: tenant_id = ? AND component_id = ? AND executed_at > now() - interval '15 min'
Aggregation: COUNT(*), COUNT(WHERE status='healthy'), percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms)
Expected rows scanned: ~30 (one per 30-second check)
Target latency: <20ms
```

**Index used:** `idx_results_component_time`  
**Why fast:** Scans only ~30 rows in the latest partition. Aggregation over tiny dataset.

---

### Pattern 9: AI Analysis Input (Future)

**Query:** "Get all metrics and check results for component X over the last 7 days for AI root cause analysis"

```
Path: check_results + metric_snapshots
Filter: tenant_id = ? AND component_id = ? AND time > now() - interval '7 days'
Expected rows: ~20K check results + ~672 metric snapshots
Target latency: <500ms (acceptable for async AI pipeline)
```

**Index used:** `idx_results_component_time` + `idx_metrics_component_type_time`  
**Why acceptable latency:** This is a background operation, not user-facing. 500ms for 20K rows is fine.

---

### Pattern 10: Outbox Processing

**Query:** "Get next batch of unprocessed events"

```
Path: event_outbox
Filter: status = 'pending' AND created_at > now() - interval '24h'
Sort: created_at ASC
Limit: 100
Target latency: <5ms
```

**Index used:** `idx_outbox_pending` (partial index)  
**Why fast:** Partial index only contains pending rows. Very small. Sequential access to oldest first.

---

## Scalability Review

### What Breaks at 10 Tenants (Phase 1)

**Answer: Nothing breaks. Everything is comfortable.**

| Metric | Value | Headroom |
|--------|-------|----------|
| Total components | ~200 | 100x headroom |
| Check results/day | ~100K | Well within single-node PostgreSQL |
| Active incidents | ~5-10 | Negligible |
| Dashboard connections | ~20 | Trivial |

**Risks at this scale:**
- None. A single-core PostgreSQL instance handles this effortlessly.
- Risk is over-engineering, not under-capacity.

---

### What Breaks at 100 Tenants (Phase 2)

**Answer: Write throughput on `check_results` becomes the first concern.**

| Metric | Value | Concern Level |
|--------|-------|---------------|
| Total components | ~5,000 | Fine |
| Check results/day | ~5M | Approaching single-writer limits |
| Active incidents | ~50-100 | Fine |
| Dashboard connections | ~200 | Connection pooling needed |
| `component_health` updates/sec | ~170 | UPDATE contention emerging |
| Outbox depth | ~50K events/day | Must ensure processing keeps up |

**What breaks:**

1. **`component_health` hot rows** — 170 UPDATE/s on ~5000 rows. Vacuum pressure increases. Row-level locking starts showing latency spikes.

   **Mitigation:** Batch updates. Instead of updating on every check result, update every N results or every M seconds (30s debounce). Acceptable freshness delay for dashboards.

2. **Connection pool saturation** — 200 dashboard sessions + health check workers + API requests.

   **Mitigation:** PgBouncer in transaction mode. Separate pools for read (dashboard) and write (check results) workloads.

3. **Partition management** — Monthly partitions are sufficient but must be pre-created and old ones dropped automatically.

   **Mitigation:** Cron job or pg_cron extension for partition lifecycle.

---

### What Breaks at 1000 Tenants (Future)

**Answer: Multiple things break simultaneously.**

| Metric | Value | Concern Level |
|--------|-------|---------------|
| Total components | ~50,000 | Index sizes significant |
| Check results/day | ~100M+ | Single-node PostgreSQL ceiling |
| Active incidents | ~500-1000 | Fine (incidents are low-volume) |
| Dashboard connections | ~2000+ | Read replica required |
| `component_health` updates/sec | ~1,700 | Critical hot spot |
| Outbox depth | ~1M events/day | Single outbox table bottleneck |
| Total DB size (1 year) | ~5-10TB | Storage and backup concerns |

**What breaks:**

1. **Single-node write throughput** — 100M inserts/day on `check_results` = ~1,157 inserts/sec sustained. Single PostgreSQL handles this, but with degraded vacuum performance.

   **Mitigation options (in order of complexity):**
   - TimescaleDB with compression (10x storage reduction, better write batching)
   - Partitioned outbox (multiple workers process different tenant ranges)
   - Write-ahead buffer in Redis (batch inserts in 1-second windows)
   - Read replicas for all dashboard queries

2. **`component_health` contention** — 1,700 UPDATE/s on 50K rows. Vacuum cannot keep up. Table bloat grows.

   **Mitigation:**
   - Move to event-driven CQRS: write check results → event → async update component_health
   - Or: maintain component_health in Redis (write-through cache), persist to PostgreSQL every 60s
   - Phase 3+: This is where Redis becomes justified

3. **Index bloat on partitioned tables** — Monthly partitions of `check_results` at 100M/day = 3B rows/month. B-tree indexes on 3B rows become large.

   **Mitigation:**
   - Switch to weekly or daily partitions at this scale
   - TimescaleDB automatic chunk management
   - Aggressive index-only scan design (covering indexes)

4. **Backup duration** — 5-10TB database takes hours to backup with pg_dump. Point-in-time recovery (WAL archiving) becomes mandatory.

   **Mitigation:**
   - WAL-G for continuous archival
   - pg_basebackup for initial backup
   - Per-tenant logical export for tenant data portability

5. **Tenant hot spots** — One tenant with 10x more components than average dominates shared resources.

   **Mitigation:**
   - Per-tenant quotas enforced at application level
   - Query timeout per tenant (statement_timeout per session)
   - Promote hot tenants to dedicated schema or dedicated instance

---

## Operational Risks

### Risk 1: Index Bloat on `component_health`

**Cause:** Frequent UPDATEs create dead tuples. Autovacuum may not keep pace under sustained write load.

**Detection:** Monitor `pg_stat_user_tables.n_dead_tup` and table size vs live tuple ratio.

**Mitigation:**
- Aggressive autovacuum settings for this table: `autovacuum_vacuum_scale_factor = 0.01`
- Consider `FILLFACTOR = 70` to enable HOT updates (Heap-Only Tuple) when only score/status/timestamps change
- Phase 2+: Debounce updates to reduce frequency

### Risk 2: Outbox Table Growth During Processing Stalls

**Cause:** If the event worker stops (crash, deployment, bug), outbox grows unbounded.

**Detection:** Monitor `SELECT count(*) FROM event_outbox WHERE status = 'pending'` — alert if > 10,000.

**Mitigation:**
- Dead letter: after 5 retries, move to `event_outbox_dead_letter` table
- Worker health check: meta-monitoring (OPTRION monitors itself)
- Multiple workers with partition assignment (events for tenant_id % N processed by worker N)
- Daily partition with aggressive TRUNCATE of processed partitions

### Risk 3: `check_results` Partition Explosion

**Cause:** At daily partitioning with 3-month retention = 90 partitions. Query planner degrades with >100 partitions.

**Detection:** Monitor `SELECT count(*) FROM pg_inherits WHERE inhparent = 'check_results'::regclass`.

**Mitigation:**
- Monthly partitions in Phase 1 (max 12 active)
- Weekly partitions in Phase 2 (max 13-52 depending on retention)
- TimescaleDB manages this automatically with configurable chunk intervals
- Never exceed 200 child tables per partitioned table

### Risk 4: Large JSONB Columns Causing Table Bloat

**Cause:** `target`, `payload`, `metadata` columns in JSONB can be large. Toast storage helps but adds I/O.

**Detection:** Monitor `pg_total_relation_size` vs `pg_relation_size` (large difference = toast bloat).

**Mitigation:**
- Enforce maximum JSONB document size at application layer (reject >16KB)
- For `check_results.metadata`: store only essential fields, not full response bodies
- For `event_outbox.payload`: keep payloads lean, reference entities by ID instead of embedding

### Risk 5: Tenant Data Leak via Missing RLS or Application Bug

**Cause:** A code path that doesn't set `app.current_tenant_id` or joins without tenant_id filter.

**Detection:** Integration test suite that creates data for Tenant A and Tenant B, then verifies Tenant A cannot see Tenant B's data through any API endpoint.

**Mitigation:**
- RLS enabled on ALL tables (defense-in-depth)
- Repository pattern: every method signature takes TenantID as first parameter
- Static analysis/linter rule: reject any raw SQL that doesn't include `tenant_id`
- Periodic audit: query `pg_stat_statements` for queries missing tenant_id filter

### Risk 6: Connection Exhaustion

**Cause:** Each dashboard session, health check worker, and API request needs a database connection. Default PostgreSQL max_connections = 100.

**Detection:** Monitor `SELECT count(*) FROM pg_stat_activity`.

**Mitigation:**
- PgBouncer in transaction-mode pooling (pool size 20-50 actual connections serving 500+ application connections)
- Separate pools: `pool_write` (10 connections, for check results + outbox) and `pool_read` (30 connections, for dashboards + API)
- Phase 2+: Read replica for all dashboard queries

### Risk 7: Vacuum Cannot Keep Pace

**Cause:** High-volume tables (check_results, component_health) generate dead tuples faster than autovacuum processes them.

**Detection:** Monitor `pg_stat_user_tables.last_autovacuum` and `n_dead_tup`. Alert if dead tuples exceed 20% of live tuples.

**Mitigation:**
- Per-table autovacuum tuning (lower thresholds for hot tables)
- `check_results` is mostly INSERT-only (append-only design) — vacuum pressure is low because there are no dead tuples from updates
- `component_health` is UPDATE-heavy — use FILLFACTOR and aggressive vacuum settings
- `event_outbox` has DELETE after processing — use daily partitions + TRUNCATE instead of DELETE

### Risk 8: Retention Job Failures

**Cause:** Scheduled partition drop job fails silently. Data grows unbounded.

**Detection:** Monitor total table size daily. Alert if growth exceeds expected rate by 2x.

**Mitigation:**
- Retention job writes to audit log on success and failure
- Separate alerting on retention job health (meta-monitoring)
- Manual fallback: documented runbook for emergency partition cleanup
- Never use DELETE for retention — always DROP PARTITION or TRUNCATE

---

## Final Recommendations

### Phase 1 Implementation Order

| Priority | Action | Reason |
|----------|--------|--------|
| 1 | Create schema structure (core, monitoring, incidents, alerting, system) | Namespace separation from Day 1 |
| 2 | Implement core tables with RLS policies | Foundation + security |
| 3 | Implement `monitoring.check_results` with monthly RANGE partitioning | Highest volume table, must be right from start |
| 4 | Implement `monitoring.component_health` with FILLFACTOR 70 | Real-time dashboard source |
| 5 | Implement `system.event_outbox` with daily partitions | Event architecture foundation |
| 6 | Implement incident + alerting tables (no partitioning needed yet) | Low volume, partition when needed |
| 7 | Create intelligence schema tables (empty, for future) | Reserved namespace |
| 8 | Configure PgBouncer | Connection management before load arrives |

### Database Configuration Recommendations

| Parameter | Phase 1 Value | Phase 2 Value | Purpose |
|-----------|--------------|---------------|---------|
| max_connections | 100 | 200 (behind PgBouncer) | Connection limit |
| shared_buffers | 25% RAM | 25% RAM | Buffer cache |
| effective_cache_size | 75% RAM | 75% RAM | Planner hint |
| work_mem | 64MB | 128MB | Sort/hash operations |
| maintenance_work_mem | 512MB | 1GB | Vacuum, index creation |
| random_page_cost | 1.1 (SSD) | 1.1 (SSD) | SSD-optimized planning |
| autovacuum_max_workers | 3 | 5 | Parallel vacuum |
| checkpoint_completion_target | 0.9 | 0.9 | Spread checkpoint I/O |
| wal_level | logical | logical | For future CDC/replication |
| max_wal_size | 2GB | 4GB | Reduce checkpoint frequency |

### UUID v7 Strategy

All primary keys use UUID v7 (RFC 9562):
- Time-sortable: `created_at` index often unnecessary
- Globally unique: safe for future service extraction
- No sequence contention: parallel writers don't block on next-value
- B-tree friendly: time-ordering means sequential inserts (low page splits)

### What NOT to Do

| Anti-Pattern | Why |
|--------------|-----|
| Don't use SERIAL/BIGSERIAL for PKs | Contention under concurrent writes. UUID v7 is better |
| Don't store raw HTTP response bodies in check_results | Bloats the table. Store a truncated excerpt (first 1KB) or hash only |
| Don't use NOTIFY/LISTEN for event bus at scale | Connection-bound. Outbox + polling is more reliable |
| Don't use GIN indexes on JSONB columns prematurely | Only add when you have specific JSON path queries with measured performance problems |
| Don't create covering indexes for rare queries | Only cover queries that execute >100 times/second |
| Don't use schema-per-tenant | Migration management becomes untenable at 100+ tenants |
| Don't skip RLS "because the app handles it" | Defense-in-depth is non-negotiable for multi-tenant |
| Don't use DELETE for retention on partitioned tables | DROP PARTITION is O(1). DELETE of millions of rows causes hours of vacuum work |

### Redis Role (Strictly Defined)

Redis is NOT a primary data store. Its role is:

| Use Case | Justification | Phase |
|----------|--------------|-------|
| Rate limiting | Token bucket per tenant per endpoint | Phase 1 |
| Session/token cache | Short-lived API session tokens | Phase 1 |
| Alert cooldown tracking | "Has this rule fired in the last N seconds?" | Phase 1 |
| Dashboard cache | Cache computed dashboard state for 10s TTL | Phase 2 |
| `component_health` write buffer | Batch updates every 30s instead of per-check | Phase 2 |
| Pub/sub for real-time dashboard | Push score updates to connected WebSocket clients | Phase 2 |

**Rule:** If Redis loses all data, the system continues functioning correctly (slower, but correct). PostgreSQL is the source of truth. Always.

### Monitoring the Monitor (Meta-Observability)

OPTRION must monitor its own database:

| Metric | Alert Threshold |
|--------|----------------|
| Connection pool utilization | > 80% |
| Replication lag (if replicas exist) | > 5 seconds |
| Outbox pending count | > 10,000 |
| Dead tuple ratio (any table) | > 20% |
| Partition count per table | > 150 |
| Disk usage growth rate | > 2x expected daily growth |
| Longest running query | > 30 seconds |
| Lock wait time | > 5 seconds |

---

### Summary Decision Matrix

| Decision | Choice | Confidence |
|----------|--------|-----------|
| TimescaleDB | Not Phase 1. Add at Phase 2 when volume demands it | High |
| Tenant isolation | Shared schema + RLS | High |
| Partitioning | Native RANGE (monthly) for time-series tables | High |
| Primary keys | UUID v7 everywhere | High |
| Connection pooling | PgBouncer in transaction mode | High |
| Event delivery | Outbox table with daily partitions | High |
| Redis role | Cache + rate limiting only. Not primary store | High |
| JSONB usage | Configuration and metadata only. Not for queryable analytics | High |
| Retention | DROP PARTITION, never bulk DELETE | High |
| Read scaling | Single node Phase 1. Read replica Phase 2 | High |

---

*End of Database Architecture*
