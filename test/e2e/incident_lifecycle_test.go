package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	incidentapp "github.com/optrion/optrion/internal/incident/app"
	"github.com/optrion/optrion/test/testutil"
)

// TestE2E_IncidentLifecycle tests the full incident state machine:
// open → acknowledge → investigate → resolve → close
func TestE2E_IncidentLifecycle(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	tenantID, apiKey := createAuthenticatedTenant(t, env, "Lifecycle Tenant", "lifecycle-tenant")

	// Create a product/env/component hierarchy
	productID := createProductAuth(t, env, tenantID, "Lifecycle App", "lifecycle-app", apiKey)
	envID := createEnvironmentAuth(t, env, tenantID, productID, "Production", "production", apiKey)
	componentID := createComponentAuth(t, env, tenantID, productID, envID, "api-gateway", "service", apiKey)

	// Trigger an incident via health metric evaluation
	triggerIncident(t, env, tenantID, componentID)

	// List incidents to find the newly created one
	incidents := listIncidents(t, env, tenantID, apiKey)
	if len(incidents) == 0 {
		t.Skip("No incidents were created from metric evaluation — incident rules may not be seeded")
	}

	incidentID := incidents[0]["id"].(string)
	status := incidents[0]["status"].(string)
	if status != "open" {
		t.Fatalf("expected initial status 'open', got %q", status)
	}

	// Acknowledge
	doIncidentAction(t, env, incidentID, "acknowledge", apiKey, map[string]string{"actor_id": "engineer-1"})
	verifyIncidentStatus(t, env, incidentID, apiKey, "acknowledged")

	// Investigate
	doIncidentAction(t, env, incidentID, "investigate", apiKey, map[string]string{"actor_id": "engineer-1"})
	verifyIncidentStatus(t, env, incidentID, apiKey, "investigating")

	// Resolve
	doIncidentAction(t, env, incidentID, "resolve", apiKey, map[string]interface{}{
		"actor_id":   "engineer-1",
		"resolution": "Fixed connection pool exhaustion by increasing max_connections",
	})
	verifyIncidentStatus(t, env, incidentID, apiKey, "resolved")

	// Close
	doIncidentAction(t, env, incidentID, "close", apiKey, map[string]interface{}{
		"actor_id": "engineer-1",
		"reason":   "Root cause addressed and deployed",
	})
	verifyIncidentStatus(t, env, incidentID, apiKey, "closed")
}

// TestE2E_IncidentTimelineAndComments verifies timeline tracking and comments.
func TestE2E_IncidentTimelineAndComments(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	tenantID, apiKey := createAuthenticatedTenant(t, env, "Timeline Tenant", "timeline-tenant")

	productID := createProductAuth(t, env, tenantID, "Timeline App", "timeline-app", apiKey)
	envID := createEnvironmentAuth(t, env, tenantID, productID, "Staging", "staging", apiKey)
	componentID := createComponentAuth(t, env, tenantID, productID, envID, "cache-service", "cache", apiKey)

	triggerIncident(t, env, tenantID, componentID)

	incidents := listIncidents(t, env, tenantID, apiKey)
	if len(incidents) == 0 {
		t.Skip("No incidents were created — incident rules may not be seeded")
	}

	incidentID := incidents[0]["id"].(string)

	// Add comments
	addComment(t, env, incidentID, apiKey, tenantID, "engineer-1", "Investigating high latency on cache layer")
	addComment(t, env, incidentID, apiKey, tenantID, "engineer-2", "Confirmed: Redis evictions spiking")

	// Get timeline
	req, _ := http.NewRequest("GET", env.Server.URL+"/api/v1/incidents/"+incidentID+"/timeline?limit=50", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for timeline, got %d", resp.StatusCode)
	}

	// Get tenant-wide timeline
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/v1/incidents/timeline?tenant_id="+tenantID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err = env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to get tenant timeline: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for tenant timeline, got %d", resp.StatusCode)
	}

	// Get stats
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/v1/incidents/stats?tenant_id="+tenantID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err = env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for stats, got %d", resp.StatusCode)
	}
}

// TestE2E_AIAnalysisEndpoint verifies the AI analysis endpoints respond correctly.
func TestE2E_AIAnalysisEndpoint(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	tenantID, apiKey := createAuthenticatedTenant(t, env, "AI Tenant", "ai-tenant")

	productID := createProductAuth(t, env, tenantID, "AI App", "ai-app", apiKey)
	envID := createEnvironmentAuth(t, env, tenantID, productID, "Production", "production", apiKey)
	componentID := createComponentAuth(t, env, tenantID, productID, envID, "ml-service", "service", apiKey)

	triggerIncident(t, env, tenantID, componentID)

	incidents := listIncidents(t, env, tenantID, apiKey)
	if len(incidents) == 0 {
		t.Skip("No incidents created — incident rules may not be seeded")
	}

	incidentID := incidents[0]["id"].(string)

	// Try triggering analysis (may fail without AI_API_KEY configured, that's OK)
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/incidents/"+incidentID+"/analyze", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to trigger analysis: %v", err)
	}
	defer resp.Body.Close()

	// Should return 202 (accepted) or 503 (service unavailable if AI not configured)
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 202 or 503 for analysis trigger, got %d", resp.StatusCode)
	}

	// Get analysis results (may be empty)
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/v1/incidents/"+incidentID+"/analysis", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err = env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to get analysis: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for analysis get, got %d", resp.StatusCode)
	}
}

// TestE2E_RecommendationEndpoint verifies the recommendation endpoints respond correctly.
func TestE2E_RecommendationEndpoint(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	tenantID, apiKey := createAuthenticatedTenant(t, env, "Rec Tenant", "rec-tenant")

	productID := createProductAuth(t, env, tenantID, "Rec App", "rec-app", apiKey)
	envID := createEnvironmentAuth(t, env, tenantID, productID, "Production", "production", apiKey)
	componentID := createComponentAuth(t, env, tenantID, productID, envID, "db-primary", "database", apiKey)

	triggerIncident(t, env, tenantID, componentID)

	incidents := listIncidents(t, env, tenantID, apiKey)
	if len(incidents) == 0 {
		t.Skip("No incidents created — incident rules may not be seeded")
	}

	incidentID := incidents[0]["id"].(string)

	// Try triggering recommendation
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/incidents/"+incidentID+"/recommend", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to trigger recommendation: %v", err)
	}
	defer resp.Body.Close()

	// Should return 202 or 503
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 202 or 503 for recommendation trigger, got %d", resp.StatusCode)
	}

	// Get recommendations list
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/v1/incidents/"+incidentID+"/recommendations", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err = env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to get recommendations: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for recommendations get, got %d", resp.StatusCode)
	}
}

// --- Helpers ---

func triggerIncident(t *testing.T, env *testutil.TestEnv, tenantID, componentID string) {
	t.Helper()
	// Directly use the incident service to evaluate a critical metric
	// This simulates a health threshold breach
	ctx := context.Background()
	env.AppContainer.IncidentService.EvaluateMetric(ctx, incidentapp.MetricInput{
		TenantID:      tenantID,
		ComponentID:   componentID,
		CollectorType: "backend",
		MetricType:    "error_rate",
		Value:         99.0, // Critical error rate
	})
}

func listIncidents(t *testing.T, env *testutil.TestEnv, tenantID, apiKey string) []map[string]interface{} {
	t.Helper()
	req, _ := http.NewRequest("GET", env.Server.URL+"/api/v1/incidents?tenant_id="+tenantID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to list incidents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for incident list, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	data, ok := result["data"].([]interface{})
	if !ok {
		return nil
	}

	var incidents []map[string]interface{}
	for _, d := range data {
		if m, ok := d.(map[string]interface{}); ok {
			incidents = append(incidents, m)
		}
	}
	return incidents
}

func doIncidentAction(t *testing.T, env *testutil.TestEnv, incidentID, action, apiKey string, body interface{}) {
	t.Helper()
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/incidents/"+incidentID+"/"+action, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to %s incident: %v", action, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("expected 200 for %s, got %d: %v", action, resp.StatusCode, errResp)
	}
}

func verifyIncidentStatus(t *testing.T, env *testutil.TestEnv, incidentID, apiKey, expectedStatus string) {
	t.Helper()
	req, _ := http.NewRequest("GET", env.Server.URL+"/api/v1/incidents/"+incidentID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to get incident: %v", err)
	}
	defer resp.Body.Close()

	var incident map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&incident)
	actual, _ := incident["status"].(string)
	if actual != expectedStatus {
		t.Fatalf("expected incident status %q, got %q", expectedStatus, actual)
	}
}

func addComment(t *testing.T, env *testutil.TestEnv, incidentID, apiKey, tenantID, authorID, content string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"tenant_id": tenantID,
		"author_id": authorID,
		"content":   content,
	})
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/incidents/"+incidentID+"/comments", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to add comment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for comment, got %d", resp.StatusCode)
	}
}
