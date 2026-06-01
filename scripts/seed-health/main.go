package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// seed_health.go creates health metric definitions for all seeded components.
// Prerequisites: Run scripts/seed/seed.go first to create the tenant hierarchy.
// Usage: go run scripts/seed/seed_health.go [base_url] [database_url]
// Default base URL: http://localhost:8080
// Default DB URL: postgres://optrion:optrion@localhost:5432/optrion?sslmode=disable

func main() {
	baseURL := "http://localhost:8080"
	dbURL := "postgres://optrion:optrion@localhost:5432/optrion?sslmode=disable"

	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}
	if len(os.Args) > 2 {
		dbURL = os.Args[2]
	}

	fmt.Printf("Seeding health metrics\n")
	fmt.Printf("  API:      %s\n", baseURL)
	fmt.Printf("  Database: %s\n\n", dbURL)

	ctx := context.Background()

	// Connect to database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Get tenant by slug
	tenantID := getTenantID(baseURL, "gymflow-track")
	if tenantID == "" {
		fmt.Fprintf(os.Stderr, "ERROR: tenant 'gymflow-track' not found. Run scripts/seed/seed.go first.\n")
		os.Exit(1)
	}
	fmt.Printf("✓ Found tenant: %s\n", tenantID)

	// Get components directly from database
	rows, err := pool.Query(ctx, "SELECT id, name, kind FROM components WHERE tenant_id = $1", tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: querying components: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var components []component
	for rows.Next() {
		var c component
		if err := rows.Scan(&c.ID, &c.Name, &c.Kind); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: scanning component: %v\n", err)
			os.Exit(1)
		}
		components = append(components, c)
	}

	if len(components) == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: no components found. Run scripts/seed/seed.go first.\n")
		os.Exit(1)
	}

	// Seed health metrics for each component
	var count int
	for _, comp := range components {
		switch comp.Kind {
		case "api":
			count += seedBackendMetrics(ctx, pool, tenantID, comp.ID)
		case "database":
			count += seedPostgresMetrics(ctx, pool, tenantID, comp.ID)
		case "cache":
			count += seedRedisMetrics(ctx, pool, tenantID, comp.ID)
		}
	}

	// Also seed server metrics (self-monitoring)
	serverComponentID := seedServerComponent(ctx, pool, tenantID)
	count += seedServerMetrics(ctx, pool, tenantID, serverComponentID)

	fmt.Printf("\n✓ Health metrics seeded: %d metrics across %d components\n", count, len(components)+1)
}

type component struct {
	ID   string
	Name string
	Kind string
}

func getTenantID(baseURL, slug string) string {
	resp, err := http.Get(baseURL + "/api/v1/tenants") //nolint:gosec
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}

	for _, t := range result.Data {
		if t.Slug == slug {
			return t.ID
		}
	}
	return ""
}

func getComponents(baseURL, tenantID string) []component {
	resp, err := http.Get(baseURL + "/api/v1/components?tenant_id=" + tenantID) //nolint:gosec
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Kind string `json:"kind"`
		} `json:"data"`
	}
	json.Unmarshal(body, &result) //nolint:errcheck

	components := make([]component, 0, len(result.Data))
	for _, d := range result.Data {
		components = append(components, component{ID: d.ID, Name: d.Name, Kind: d.Kind})
	}
	return components
}

func seedBackendMetrics(ctx context.Context, pool *pgxpool.Pool, tenantID, componentID string) int {
	metrics := []struct {
		metricType    string
		collectorType string
		name          string
		unit          string
		warnMax       *float64
		critMax       *float64
		warnMin       *float64
		critMin       *float64
	}{
		{"availability", "backend", "Backend Availability", "bool", nil, nil, floatPtr(1), floatPtr(0)},
		{"response_time", "backend", "Response Time", "ms", floatPtr(1000), floatPtr(2000), nil, nil},
		{"error_rate", "backend", "Error Rate", "ratio", floatPtr(0.01), floatPtr(0.05), nil, nil},
		{"throughput", "backend", "Throughput", "req/s", nil, nil, floatPtr(10), floatPtr(1)},
		{"uptime", "backend", "Uptime", "seconds", nil, nil, nil, nil},
	}

	for _, m := range metrics {
		insertMetric(ctx, pool, tenantID, componentID, m.metricType, m.collectorType, m.name, m.unit, m.warnMin, m.warnMax, m.critMin, m.critMax)
	}
	fmt.Printf("✓ Backend metrics seeded for component %s\n", componentID)
	return len(metrics)
}

func seedPostgresMetrics(ctx context.Context, pool *pgxpool.Pool, tenantID, componentID string) int {
	metrics := []struct {
		metricType    string
		collectorType string
		name          string
		unit          string
		warnMax       *float64
		critMax       *float64
		warnMin       *float64
		critMin       *float64
	}{
		{"connection_status", "postgres", "Connection Status", "bool", nil, nil, floatPtr(1), floatPtr(0)},
		{"query_latency", "postgres", "Query Latency", "ms", floatPtr(100), floatPtr(500), nil, nil},
		{"active_connections", "postgres", "Active Connections", "count", floatPtr(80), floatPtr(95), nil, nil},
		{"slow_queries", "postgres", "Slow Queries", "count", floatPtr(1), floatPtr(5), nil, nil},
		{"deadlocks", "postgres", "Deadlocks", "count", floatPtr(1), floatPtr(3), nil, nil},
		{"index_usage", "postgres", "Index Usage Ratio", "%", nil, nil, floatPtr(90), floatPtr(80)},
		{"pool_health", "postgres", "Connection Pool Usage", "%", floatPtr(80), floatPtr(95), nil, nil},
	}

	for _, m := range metrics {
		insertMetric(ctx, pool, tenantID, componentID, m.metricType, m.collectorType, m.name, m.unit, m.warnMin, m.warnMax, m.critMin, m.critMax)
	}
	fmt.Printf("✓ PostgreSQL metrics seeded for component %s\n", componentID)
	return len(metrics)
}

func seedRedisMetrics(ctx context.Context, pool *pgxpool.Pool, tenantID, componentID string) int {
	metrics := []struct {
		metricType    string
		collectorType string
		name          string
		unit          string
		warnMax       *float64
		critMax       *float64
		warnMin       *float64
		critMin       *float64
	}{
		{"availability", "redis", "Redis Availability", "bool", nil, nil, floatPtr(1), floatPtr(0)},
		{"memory_usage", "redis", "Memory Usage", "%", floatPtr(75), floatPtr(90), nil, nil},
		{"hit_ratio", "redis", "Cache Hit Ratio", "%", nil, nil, floatPtr(80), floatPtr(60)},
		{"evictions", "redis", "Evictions", "count", floatPtr(100), floatPtr(1000), nil, nil},
		{"connected_clients", "redis", "Connected Clients", "count", floatPtr(100), floatPtr(500), nil, nil},
	}

	for _, m := range metrics {
		insertMetric(ctx, pool, tenantID, componentID, m.metricType, m.collectorType, m.name, m.unit, m.warnMin, m.warnMax, m.critMin, m.critMax)
	}
	fmt.Printf("✓ Redis metrics seeded for component %s\n", componentID)
	return len(metrics)
}

func seedServerComponent(ctx context.Context, pool *pgxpool.Pool, tenantID string) string {
	id := uuid.Must(uuid.NewV7()).String()
	now := time.Now().UTC()

	_, err := pool.Exec(ctx, `
		INSERT INTO components (id, tenant_id, product_id, environment_id, name, slug, kind, endpoint_url, created_at, updated_at)
		SELECT $1, $2, p.id, e.id, 'OPTRION Server', 'optrion-server', 'service', 'localhost:8080', $3, $4
		FROM products p
		JOIN environments e ON e.product_id = p.id
		WHERE p.tenant_id = $2
		LIMIT 1
		ON CONFLICT DO NOTHING
	`, id, tenantID, now, now)
	if err != nil {
		// Component might already exist, try to find it
		var existingID string
		pool.QueryRow(ctx, `SELECT id FROM components WHERE tenant_id = $1 AND slug = 'optrion-server'`, tenantID).Scan(&existingID) //nolint:errcheck
		if existingID != "" {
			return existingID
		}
		fmt.Fprintf(os.Stderr, "WARNING: could not create server component: %v\n", err)
		return id
	}
	fmt.Printf("✓ Server component created: %s\n", id)
	return id
}

func seedServerMetrics(ctx context.Context, pool *pgxpool.Pool, tenantID, componentID string) int {
	metrics := []struct {
		metricType    string
		collectorType string
		name          string
		unit          string
		warnMax       *float64
		critMax       *float64
		warnMin       *float64
		critMin       *float64
	}{
		{"cpu", "server", "CPU Usage", "%", floatPtr(75), floatPtr(90), nil, nil},
		{"ram", "server", "RAM Usage", "%", floatPtr(75), floatPtr(90), nil, nil},
		{"disk", "server", "Disk Usage", "%", floatPtr(80), floatPtr(95), nil, nil},
		{"load_average", "server", "Load Average", "ratio", floatPtr(5), floatPtr(10), nil, nil},
		{"network", "server", "Network Activity", "ratio", floatPtr(80), floatPtr(95), nil, nil},
	}

	for _, m := range metrics {
		insertMetric(ctx, pool, tenantID, componentID, m.metricType, m.collectorType, m.name, m.unit, m.warnMin, m.warnMax, m.critMin, m.critMax)
	}
	fmt.Printf("✓ Server metrics seeded for component %s\n", componentID)
	return len(metrics)
}

func insertMetric(ctx context.Context, pool *pgxpool.Pool, tenantID, componentID, metricType, collectorType, name, unit string, warnMin, warnMax, critMin, critMax *float64) {
	id := uuid.Must(uuid.NewV7()).String()
	now := time.Now().UTC()

	thresholds := map[string]*float64{
		"warning_min":  warnMin,
		"warning_max":  warnMax,
		"critical_min": critMin,
		"critical_max": critMax,
	}
	thresholdsJSON, err := json.Marshal(thresholds)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to marshal thresholds for metric %s/%s: %v\n", componentID, metricType, err)
		return
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO health_metrics (id, tenant_id, component_id, metric_type, collector_type, name, unit, thresholds, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, $9, $10)
		ON CONFLICT (component_id, metric_type) DO UPDATE SET
			name = EXCLUDED.name,
			unit = EXCLUDED.unit,
			thresholds = EXCLUDED.thresholds,
			updated_at = EXCLUDED.updated_at
	`, id, tenantID, componentID, metricType, collectorType, name, unit, thresholdsJSON, now, now)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to insert metric %s/%s: %v\n", componentID, metricType, err)
	}
}

func floatPtr(v float64) *float64 {
	return &v
}
