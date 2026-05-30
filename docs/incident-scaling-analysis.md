# Incident Intelligence Engine — Scaling Analysis

## Capacity Model

| Scenario | Incidents/Day | Events/Day | Timeline Entries/Day | Notes |
|----------|--------------|------------|---------------------|-------|
| Small (startup) | 10-50 | 50-250 | 50-250 | Single team, few components |
| Medium (scale-up) | 100-500 | 500-2,500 | 500-2,500 | Multiple products, ~50 components |
| Large (enterprise) | 1,000-10,000 | 5,000-50,000 | 5,000-50,000 | Multi-tenant, hundreds of components |

## Current Architecture Strengths

### Deduplication
- `FindActiveByRuleAndComponent` prevents runaway incident creation during sustained outages
- Without this, a single failing collector could generate 1 incident/second = 86,400/day

### Cooldown Mechanism
- Per-rule cooldown (configurable, default 5 min) prevents alert storms
- Rate-limits incident creation to `(24h / cooldown)` per rule per component
- With 5-min cooldown: max 288 incidents/rule/component/day

### Event Sourcing
- Append-only event store: INSERT-only pattern scales linearly
- No UPDATE contention on the event table
- Batch INSERT via pgx.Batch reduces round-trips

### Optimistic Concurrency
- Version-based conflict detection avoids row-level locks on hot incidents
- Retry at the application layer (not implemented yet — recommended for high contention)

### Correlation
- Groups related incidents automatically (same component, active window)
- Reduces cognitive load in war rooms

## Bottleneck Analysis at 10,000 Incidents/Day

### Database Write Path
- **Incidents table**: ~10K INSERTs + ~20K UPDATEs (state transitions) = ~30K writes/day = 0.35 TPS
- **Event store**: ~50K INSERTs/day = 0.6 TPS
- **Timeline**: ~50K INSERTs/day = 0.6 TPS
- **Total write TPS**: ~1.5 sustained — well within single-node PostgreSQL capacity

### Database Read Path
- `ListByTenant` with filters: Indexed on `(tenant_id, status, occurred_at DESC)`
- `FindActiveByRuleAndComponent`: Indexed on `(rule_id, component_id, status)`
- `CountByTenant`: Same indexes, COUNT(*) on filtered subset
- At 10K incidents/day with 1-week active window = 70K rows max in hot set

### Query Performance Estimates (Cold)
| Query | Rows Scanned | Est. Time |
|-------|-------------|-----------|
| ListByTenant (50 limit) | Index seek + 50 rows | <5ms |
| CountByTenant (active) | Index count | <10ms |
| FindActiveByRuleAndComponent | Index seek + 1 row | <2ms |
| ListByIncident (events) | Index seek + ~5-10 rows | <3ms |

## Recommended Scaling Improvements (When Needed)

### At 1,000 incidents/day (Phase 1)
- ✅ Current architecture sufficient
- Add: Connection pool monitoring for the incident DB queries
- Add: pg_stat_statements monitoring

### At 5,000 incidents/day (Phase 2)
- **Add read replicas** for query endpoints (ListByTenant, GetStats)
- **Implement retry with backoff** on optimistic concurrency conflicts
- **Add Redis caching** for GetStats (30s TTL, invalidated on state change)
- **Partition incident_events** by month: `incident_events_2026_01`, etc.

### At 10,000+ incidents/day (Phase 3)
- **Table partitioning** on incidents by `occurred_at` (monthly)
- **Archive cold incidents** (closed > 90 days) to separate schema
- **Async event processing** via Redis Streams or NATS
- **Denormalized stats table** updated via triggers or background jobs
- **Rule evaluation parallelism**: Process rules concurrently per tenant

## Index Coverage

```sql
-- Primary lookup paths (already created in migrations)
CREATE INDEX idx_incidents_tenant_status ON incidents (tenant_id, status, occurred_at DESC);
CREATE INDEX idx_incidents_rule_component ON incidents (rule_id, component_id) WHERE status IN ('open','acknowledged','investigating');
CREATE INDEX idx_incidents_correlation ON incidents (correlation_id);
CREATE INDEX idx_incident_events_incident ON incident_events (incident_id, occurred_at);
CREATE INDEX idx_incident_events_tenant ON incident_events (tenant_id, occurred_at DESC);
CREATE INDEX idx_incident_timeline_incident ON incident_timeline (incident_id, occurred_at);
CREATE INDEX idx_incident_rules_tenant ON incident_rules (tenant_id, enabled);
```

## Memory Footprint

Per incident in memory (Go struct):
- Base fields: ~350 bytes
- Event slice (5 events avg): ~800 bytes
- Map allocations: ~200 bytes
- **Total**: ~1.4 KB per loaded incident

Loading 1000 incidents simultaneously: ~1.4 MB — negligible.

## Conclusion

The current design handles 10,000 incidents/day on a single PostgreSQL node with minimal resource usage. The primary scaling vector is **read load** (dashboards, API queries), which can be addressed with read replicas and caching without modifying the write path. The event-sourced architecture ensures the write path remains simple and fast regardless of aggregate complexity.
