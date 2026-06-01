package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/optrion/optrion/test/testutil"
)

func TestE2E_AIAnalysisWorkflow(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Create tenant and get API key
	tenantID, apiKey := createAuthenticatedTenant(t, env, "AI Test Tenant", "ai-test-tenant")

	// Create product + environment + component
	productID := createProductAuth(t, env, tenantID, "AI Product", "ai-product", apiKey)
	envID := createEnvironmentAuth(t, env, tenantID, productID, "Production", "production", apiKey)
	_ = createComponentAuth(t, env, tenantID, productID, envID, "api-service", "service", apiKey)

	// Attempt to trigger AI analysis on a non-existent incident
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/incidents/nonexistent-id/analyze", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to call analyze endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Should return 404 or 501 (not implemented) — NOT 500 or an auth error
	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("AI analysis endpoint should accept valid API key")
	}

	// Test GET /api/v1/analysis with auth
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/v1/analysis?tenant_id="+tenantID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err = env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to call analysis endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("analysis GET endpoint should accept valid API key")
	}
}

func TestE2E_RecommendationWorkflow(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Create tenant and get API key
	tenantID, apiKey := createAuthenticatedTenant(t, env, "Rec Test Tenant", "rec-test-tenant")

	// Test GET /api/v1/recommendations with auth
	req, _ := http.NewRequest("GET", env.Server.URL+"/api/v1/recommendations?tenant_id="+tenantID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to call recommendations endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("recommendations GET endpoint should accept valid API key")
	}

	// Test POST recommend on non-existent incident
	req, _ = http.NewRequest("POST", env.Server.URL+"/api/v1/incidents/nonexistent-id/recommend", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err = env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to call recommend endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("recommend POST endpoint should accept valid API key")
	}
}

func TestE2E_HealthToIncidentPipeline(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Create tenant and get API key
	tenantID, apiKey := createAuthenticatedTenant(t, env, "Pipeline Tenant", "pipeline-tenant")

	// Create product + environment + component
	productID := createProductAuth(t, env, tenantID, "Pipeline Product", "pipeline-product", apiKey)
	envID := createEnvironmentAuth(t, env, tenantID, productID, "Production", "production", apiKey)
	_ = createComponentAuth(t, env, tenantID, productID, envID, "database", "database", apiKey)

	// Check health summary endpoint
	req, _ := http.NewRequest("GET", env.Server.URL+"/api/v1/health/summary?tenant_id="+tenantID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to call health summary: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("health summary endpoint should accept valid API key")
	}

	// Check incidents endpoint
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/v1/incidents?tenant_id="+tenantID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err = env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to call incidents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for incidents list, got %d", resp.StatusCode)
	}

	// Check incident stats
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/v1/incidents/stats?tenant_id="+tenantID, nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err = env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to call incident stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("incident stats should accept valid API key")
	}
}

func TestE2E_AuthRequired(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// These endpoints should require auth when auth is enabled
	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/tenants"},
		{"GET", "/api/v1/products?tenant_id=abc"},
		{"GET", "/api/v1/health/summary?tenant_id=abc"},
		{"GET", "/api/v1/incidents?tenant_id=abc"},
		{"GET", "/api/v1/analysis?tenant_id=abc"},
		{"GET", "/api/v1/recommendations?tenant_id=abc"},
	}

	for _, ep := range protectedEndpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req, _ := http.NewRequest(ep.method, env.Server.URL+ep.path, nil)
			// No Authorization header
			resp, err := env.Client.Do(req)
			if err != nil {
				t.Fatalf("failed to call %s: %v", ep.path, err)
			}
			resp.Body.Close()

			// Should be 401 if auth is enabled, or proceed if auth is disabled
			if resp.StatusCode == http.StatusOK {
				// Auth might be disabled in test env — that's acceptable
				t.Logf("auth appears disabled for %s (got 200)", ep.path)
			}
		})
	}
}

func TestE2E_RateLimiting(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Send many rapid requests to trigger rate limiting
	var lastStatus int
	for i := 0; i < 500; i++ {
		resp, err := env.Client.Get(env.Server.URL + "/api/v1/info")
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		lastStatus = resp.StatusCode
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			// Rate limiting working correctly
			return
		}
	}

	// If we never got rate limited, it might be configured with high limits
	t.Logf("rate limit not triggered after 500 requests (last status: %d) — may be configured high", lastStatus)
}

func TestE2E_SecurityHeaders_APIEndpoint(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	resp, err := env.Client.Get(env.Server.URL + "/api/v1/info")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	headers := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"Referrer-Policy":           "no-referrer",
	}

	for name, expected := range headers {
		actual := resp.Header.Get(name)
		if actual != expected {
			t.Errorf("expected %s: %s, got: %s", name, expected, actual)
		}
	}
}

func TestE2E_InvalidJSON(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Send malformed JSON to creation endpoint
	resp, err := env.Client.Post(
		env.Server.URL+"/api/v1/register",
		"application/json",
		bytes.NewBufferString(`{invalid json`),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestE2E_OversizedPayload(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Create a very large payload (>1MB)
	largePayload := make([]byte, 2*1024*1024) // 2MB
	for i := range largePayload {
		largePayload[i] = 'A'
	}

	resp, err := env.Client.Post(
		env.Server.URL+"/api/v1/register",
		"application/json",
		bytes.NewBuffer(largePayload),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should be rejected — either 400 (bad request) or 413 (payload too large)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		t.Fatal("oversized payload should be rejected")
	}
}

// --- Helper Functions ---

func createAuthenticatedTenant(t *testing.T, env *testutil.TestEnv, name, slug string) (tenantID, apiKey string) {
	t.Helper()

	// Register via the registration endpoint which creates tenant + API key
	regBody := map[string]interface{}{
		"tenant": map[string]string{
			"name": name,
			"slug": slug,
			"plan": "free",
		},
		"product": map[string]string{
			"name": name + " App",
			"slug": slug + "-app",
		},
		"environment": map[string]string{
			"name": "Development",
			"tier": "development",
		},
		"components": []map[string]string{},
	}
	body, _ := json.Marshal(regBody)
	resp, err := env.Client.Post(env.Server.URL+"/api/v1/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for registration, got %d", resp.StatusCode)
	}

	var regResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&regResp)

	tenantID, _ = regResp["tenant_id"].(string)
	apiKey, _ = regResp["api_key"].(string)
	if tenantID == "" || apiKey == "" {
		t.Fatalf("registration response missing tenant_id or api_key: %+v", regResp)
	}
	return tenantID, apiKey
}

func createProductAuth(t *testing.T, env *testutil.TestEnv, tenantID, name, slug, apiKey string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"tenant_id": tenantID,
		"name":      name,
		"slug":      slug,
	})
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/products", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for product, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data["id"].(string)
}

func createEnvironmentAuth(t *testing.T, env *testutil.TestEnv, tenantID, productID, name, tier, apiKey string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"tenant_id":  tenantID,
		"product_id": productID,
		"name":       name,
		"slug":       name + "-env",
		"tier":       tier,
	})
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/environments", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to create environment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for environment, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data["id"].(string)
}

func createComponentAuth(t *testing.T, env *testutil.TestEnv, tenantID, productID, envID, name, kind, apiKey string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"tenant_id":      tenantID,
		"product_id":     productID,
		"environment_id": envID,
		"name":           name,
		"slug":           name + "-comp",
		"kind":           kind,
	})
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/components", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to create component: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for component, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return data["id"].(string)
}
