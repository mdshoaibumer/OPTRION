package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
	"github.com/optrion/optrion/test/testutil"
)

func TestE2E_HealthRegression(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	ctx := context.Background()

	// 1. Setup tenant, product, environment, and component
	tenantCmd := map[string]string{
		"name": "Health GymFlow",
		"slug": "health-gym-flow",
		"plan": "free",
	}
	jsonBody, _ := json.Marshal(tenantCmd)
	resp, err := env.Client.Post(env.Server.URL+"/api/v1/tenants", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}
	var tenantData map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&tenantData)
	resp.Body.Close()
	tenantID := tenantData["id"].(string)

	productCmd := map[string]string{
		"tenant_id": tenantID,
		"name":      "Backend",
		"slug":      "backend",
	}
	jsonBody, _ = json.Marshal(productCmd)
	resp, _ = env.Client.Post(env.Server.URL+"/api/v1/products", "application/json", bytes.NewBuffer(jsonBody))
	resp.Body.Close()

	resp, _ = env.Client.Get(env.Server.URL + "/api/v1/products?tenant_id=" + tenantID)
	var productsList map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&productsList)
	resp.Body.Close()
	productID := productsList["data"].([]interface{})[0].(map[string]interface{})["id"].(string)

	envCmd := map[string]string{
		"tenant_id":  tenantID,
		"product_id": productID,
		"name":       "Production",
		"slug":       "production",
		"tier":       "production",
	}
	jsonBody, _ = json.Marshal(envCmd)
	resp, _ = env.Client.Post(env.Server.URL+"/api/v1/environments", "application/json", bytes.NewBuffer(jsonBody))
	var envData map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&envData)
	resp.Body.Close()
	envID := envData["id"].(string)

	compCmd := map[string]string{
		"tenant_id":      tenantID,
		"product_id":     productID,
		"environment_id": envID,
		"name":           "Billing API",
		"slug":           "billing-api",
		"kind":           "api",
		"endpoint_url":   "http://billing/health",
	}
	jsonBody, _ = json.Marshal(compCmd)
	resp, _ = env.Client.Post(env.Server.URL+"/api/v1/components", "application/json", bytes.NewBuffer(jsonBody))
	var compData map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&compData)
	resp.Body.Close()
	componentID := compData["id"].(string)

	// 2. Define and seed health metrics configuration directly in PostgreSQL
	// We'll create definition for availability, response_time, error_rate
	seedMetric := func(mType domain.MetricType, name, unit string) {
		m, err := domain.NewHealthMetric(tenantID, componentID, mType, domain.CollectorBackend, name, unit, domain.Thresholds{})
		if err != nil {
			t.Fatalf("Failed to create domain HealthMetric: %v", err)
		}
		thresholdsJSON, _ := json.Marshal(m.Thresholds)
		_, err = env.AppContainer.Database.Pool().Exec(ctx, `
			INSERT INTO health_metrics (id, tenant_id, component_id, metric_type, collector_type, name, unit, thresholds, enabled, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, m.ID, m.TenantID, m.ComponentID, m.MetricType, m.CollectorType, m.Name, m.Unit, thresholdsJSON, m.Enabled, m.CreatedAt, m.UpdatedAt)
		if err != nil {
			t.Fatalf("Failed to seed health metric %s: %v", mType, err)
		}
	}

	seedMetric(domain.MetricAvailability, "Availability", "percent")
	seedMetric(domain.MetricResponseTime, "Response Time", "ms")
	seedMetric(domain.MetricErrorRate, "Error Rate", "percent")

	// Helper to get health score from API
	getScore := func() int {
		r, err := env.Client.Get(fmt.Sprintf("%s/api/v1/health/summary?tenant_id=%s", env.Server.URL, tenantID))
		if err != nil {
			t.Fatalf("Failed to call health summary: %v", err)
		}
		defer r.Body.Close()
		var summary map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&summary)
		return int(summary["overall_score"].(float64))
	}

	// Step 1: Healthy metrics (expected score: 100)
	t.Run("Step 1: Healthy status", func(t *testing.T) {
		res := &port.CollectorResult{
			ComponentID:   componentID,
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricAvailability, Value: 1.0},
				{MetricType: domain.MetricResponseTime, Value: 150.0},
				{MetricType: domain.MetricErrorRate, Value: 0.0},
			},
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
		}
		env.AppContainer.HealthService.ProcessCollectorResult(ctx, res)
		score := getScore()
		if score != 100 {
			t.Errorf("Expected health score 100, got %d", score)
		}
	})

	// Step 2: Slightly degraded response time (expected score: 90)
	t.Run("Step 2: Degraded Response Time", func(t *testing.T) {
		res := &port.CollectorResult{
			ComponentID:   componentID,
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricAvailability, Value: 1.0},
				{MetricType: domain.MetricResponseTime, Value: 1200.0},
				{MetricType: domain.MetricErrorRate, Value: 0.0},
			},
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
		}
		env.AppContainer.HealthService.ProcessCollectorResult(ctx, res)
		score := getScore()
		if score != 90 {
			t.Errorf("Expected health score 90, got %d", score)
		}
	})

	// Step 3: Critical failure (expected score: 25)
	t.Run("Step 3: Critical failure status", func(t *testing.T) {
		res := &port.CollectorResult{
			ComponentID:   componentID,
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricAvailability, Value: 0.0},
				{MetricType: domain.MetricResponseTime, Value: 2500.0},
				{MetricType: domain.MetricErrorRate, Value: 0.05},
			},
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
		}
		env.AppContainer.HealthService.ProcessCollectorResult(ctx, res)
		score := getScore()
		if score != 25 {
			t.Errorf("Expected health score 25, got %d", score)
		}
	})

	// Step 4: Recovery (expected score: 100)
	t.Run("Step 4: Recovery back to healthy", func(t *testing.T) {
		res := &port.CollectorResult{
			ComponentID:   componentID,
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricAvailability, Value: 1.0},
				{MetricType: domain.MetricResponseTime, Value: 150.0},
				{MetricType: domain.MetricErrorRate, Value: 0.0},
			},
			CollectedAt: time.Now().UTC().Format(time.RFC3339),
		}
		env.AppContainer.HealthService.ProcessCollectorResult(ctx, res)
		score := getScore()
		if score != 100 {
			t.Errorf("Expected health score 100, got %d", score)
		}
	})
}
