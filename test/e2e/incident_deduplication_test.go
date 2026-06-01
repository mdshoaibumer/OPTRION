package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/incident/domain"
	"github.com/optrion/optrion/test/testutil"
)

func TestE2E_IncidentDeduplication(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	ctx := context.Background()

	// 1. Onboard tenant, product, env, component
	tenantCmd := map[string]string{
		"name": "GymFlow Deduplication",
		"slug": "gym-flow-dedup",
		"plan": "free",
	}
	jsonBody, _ := json.Marshal(tenantCmd)
	resp, err := env.Client.Post(env.Server.URL+"/api/v1/tenants", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}
	defer resp.Body.Close()

	var tenantData map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&tenantData)
	tenantID := tenantData["id"].(string)

	productCmd := map[string]string{
		"tenant_id": tenantID,
		"name":      "Deduplication API",
		"slug":      "dedup-api",
	}
	jsonBody, _ = json.Marshal(productCmd)
	resp, err = env.Client.Post(env.Server.URL+"/api/v1/products", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	resp.Body.Close()

	var productData map[string]interface{}
	_ = json.Unmarshal(jsonBody, &productData) // wait, no, let's get the ID
	// Let's list products for this tenant to get product ID
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
	resp, err = env.Client.Post(env.Server.URL+"/api/v1/environments", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create env: %v", err)
	}
	defer resp.Body.Close()
	var envData map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&envData)
	envID := envData["id"].(string)

	compCmd := map[string]string{
		"tenant_id":      tenantID,
		"product_id":     productID,
		"environment_id": envID,
		"name":           "PostgreSQL Main",
		"slug":           "postgres-main",
		"kind":           "database",
		"endpoint_url":   "postgres://localhost",
	}
	jsonBody, _ = json.Marshal(compCmd)
	resp, err = env.Client.Post(env.Server.URL+"/api/v1/components", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}
	defer resp.Body.Close()
	var compData map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&compData)
	componentID := compData["id"].(string)

	// 2. Register an incident rule directly using IncidentService / Postgres repos
	rule, err := domain.NewIncidentRule(
		tenantID,
		"High Query Latency",
		"Triggers when postgres query latency exceeds 500ms",
		componentID,
		"postgres",
		domain.RuleCondition{
			MetricType: "query_latency",
			Operator:   domain.OperatorGT,
			Threshold:  500.0,
		},
		domain.SeverityMajor,
		1*time.Second, // minimal cooldown for testing
	)
	if err != nil {
		t.Fatalf("Failed to create rule domain object: %v", err)
	}

	// We can access container database pool directly to run insert on the `incident_rules` table.
	// This is standard, robust, and completely decoupling from internal unexported struct fields.
	conditionJSON, _ := json.Marshal(rule.Condition)
	_, err = env.AppContainer.Database.Pool().Exec(ctx, `
		INSERT INTO incident_rules (id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, rule.ID, rule.TenantID, rule.Name, rule.Description,
		rule.ComponentID, rule.CollectorType, conditionJSON, rule.Severity,
		int(rule.Cooldown.Seconds()), rule.Enabled, rule.CreatedAt, rule.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to insert rule: %v", err)
	}

	// 3. Evaluate metric values
	// Simulate degraded query latency: 600ms
	// Run first evaluation
	env.AppContainer.IncidentService.EvaluateMetric(ctx, struct {
		TenantID      string
		ComponentID   string
		CollectorType string
		MetricType    string
		Value         float64
		HealthScore   int
	}{
		TenantID:      tenantID,
		ComponentID:   componentID,
		CollectorType: "postgres",
		MetricType:    "query_latency",
		Value:         600.0,
		HealthScore:   85,
	})

	// Fetch incidents via REST API
	listIncidents := func() []map[string]interface{} {
		r, err := env.Client.Get(fmt.Sprintf("%s/api/v1/incidents?tenant_id=%s", env.Server.URL, tenantID))
		if err != nil {
			t.Fatalf("Failed to list incidents: %v", err)
		}
		defer r.Body.Close()
		var respData map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&respData)
		data := respData["data"].([]interface{})
		var res []map[string]interface{}
		for _, item := range data {
			res = append(res, item.(map[string]interface{}))
		}
		return res
	}

	incidents := listIncidents()
	if len(incidents) != 1 {
		t.Fatalf("Expected exactly 1 incident to be created, got %d", len(incidents))
	}
	firstIncidentID := incidents[0]["id"].(string)

	// Evaluate again with another degraded value (700ms)
	env.AppContainer.IncidentService.EvaluateMetric(ctx, struct {
		TenantID      string
		ComponentID   string
		CollectorType string
		MetricType    string
		Value         float64
		HealthScore   int
	}{
		TenantID:      tenantID,
		ComponentID:   componentID,
		CollectorType: "postgres",
		MetricType:    "query_latency",
		Value:         700.0,
		HealthScore:   80,
	})

	// Deduplication check: verify no new incident was created
	incidents = listIncidents()
	if len(incidents) != 1 {
		t.Fatalf("Deduplication failed: expected exactly 1 incident, got %d", len(incidents))
	}

	// Verify correlation: Acknowledge the incident first
	ackBody := map[string]string{"actor_id": "test-user"}
	ackJson, _ := json.Marshal(ackBody)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/incidents/%s/acknowledge", env.Server.URL, firstIncidentID), bytes.NewBuffer(ackJson))
	req.Header.Set("Content-Type", "application/json")
	ackResp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("Failed to acknowledge incident: %v", err)
	}
	ackResp.Body.Close()

	if ackResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", ackResp.StatusCode)
	}

	// Verify status updated to Acknowledged
	incidents = listIncidents()
	if incidents[0]["status"].(string) != "acknowledged" {
		t.Errorf("Expected status to be acknowledged, got %s", incidents[0]["status"])
	}
}
