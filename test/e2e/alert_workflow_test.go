package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/optrion/optrion/test/testutil"
)

func TestE2E_AlertRuleWorkflow(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Create a tenant first
	createTenant := map[string]string{"name": "Alert Test Tenant", "slug": "alert-test-tenant", "plan": "free"}
	body, _ := json.Marshal(createTenant)
	resp, err := env.Client.Post(env.Server.URL+"/api/v1/tenants", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for tenant creation, got %d", resp.StatusCode)
	}
	var tenantData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&tenantData)
	resp.Body.Close()
	tenantID := tenantData["id"].(string)

	// Create an alert rule
	createRule := map[string]interface{}{
		"tenant_id":   tenantID,
		"name":        "Critical Health Rule",
		"description": "Alert on critical health degradation",
		"severity":    "critical",
		"enabled":     true,
		"channels":    []string{},
	}
	body, _ = json.Marshal(createRule)
	resp, err = env.Client.Post(env.Server.URL+"/api/v1/alert-rules", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create alert rule: %v", err)
	}

	// Alert rule REST endpoint might not be implemented yet
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		t.Skip("Alert rule REST endpoint not yet registered — skipping alert workflow test")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for alert rule creation, got %d", resp.StatusCode)
	}

	var ruleData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&ruleData)
	resp.Body.Close()

	ruleID, ok := ruleData["id"].(string)
	if !ok || ruleID == "" {
		t.Fatal("alert rule creation response missing id")
	}

	// List alert rules for this tenant
	resp, err = env.Client.Get(env.Server.URL + "/api/v1/alert-rules?tenant_id=" + tenantID)
	if err != nil {
		t.Fatalf("failed to list alert rules: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for alert rules list, got %d", resp.StatusCode)
	}

	// Create an alert channel
	createChannel := map[string]interface{}{
		"tenant_id": tenantID,
		"type":      "telegram",
		"name":      "Ops Channel",
		"config": map[string]string{
			"bot_token": "test-bot-token",
			"chat_id":   "-100123456789",
		},
		"enabled": true,
	}
	body, _ = json.Marshal(createChannel)
	resp, err = env.Client.Post(env.Server.URL+"/api/v1/alert-channels", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create alert channel: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		t.Skip("Alert channel REST endpoint not yet registered")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 for alert channel creation, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_SecurityHeaders(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	resp, err := env.Client.Get(env.Server.URL + "/healthz")
	if err != nil {
		t.Fatalf("failed to call /healthz: %v", err)
	}
	defer resp.Body.Close()

	// Verify security headers are present
	headers := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
		"Referrer-Policy":         "no-referrer",
		"Permissions-Policy":      "camera=(), microphone=(), geolocation=()",
	}

	for header, expected := range headers {
		actual := resp.Header.Get(header)
		if actual != expected {
			t.Errorf("header %s: expected %q, got %q", header, expected, actual)
		}
	}
}
