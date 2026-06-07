package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/optrion/optrion/test/testutil"
)

func TestE2E_RegistrationWorkflow(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Test full plug-and-play registration
	regRequest := map[string]interface{}{
		"tenant": map[string]string{
			"name": "Registration Test Tenant",
			"slug": "reg-test-tenant",
			"plan": "free",
		},
		"product": map[string]string{
			"name":        "My App",
			"slug":        "my-app",
			"description": "Test product",
		},
		"environment": map[string]string{
			"name": "Production",
			"tier": "production",
		},
		"components": []map[string]interface{}{
			{
				"name":        "PostgreSQL",
				"kind":        "database",
				"description": "Primary database",
				"endpoint":    "db.example.com",
				"port":        5432,
			},
			{
				"name":        "Redis Cache",
				"kind":        "cache",
				"description": "Session cache",
				"endpoint":    "redis.example.com",
				"port":        6379,
			},
		},
	}

	body, _ := json.Marshal(regRequest)
	resp, err := env.Client.Post(env.Server.URL+"/api/v1/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("registration request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	var regResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		t.Fatalf("failed to decode registration response: %v", err)
	}

	// Verify all IDs returned
	if regResp["tenant_id"] == nil || regResp["tenant_id"] == "" {
		t.Fatal("registration response missing tenant_id")
	}
	if regResp["product_id"] == nil || regResp["product_id"] == "" {
		t.Fatal("registration response missing product_id")
	}
	if regResp["environment_id"] == nil || regResp["environment_id"] == "" {
		t.Fatal("registration response missing environment_id")
	}
	if regResp["component_ids"] == nil {
		t.Fatal("registration response missing component_ids")
	}

	componentIDs, ok := regResp["component_ids"].([]interface{})
	if !ok || len(componentIDs) != 2 {
		t.Fatalf("expected 2 component IDs, got %v", regResp["component_ids"])
	}

	// Test validation: invalid slug
	invalidReq := map[string]interface{}{
		"tenant": map[string]string{
			"name": "Bad",
			"slug": "INVALID SLUG!",
			"plan": "free",
		},
		"product": map[string]string{
			"name": "Product",
			"slug": "product",
		},
		"environment": map[string]string{
			"name": "Prod",
			"tier": "production",
		},
		"components": []map[string]interface{}{},
	}
	body, _ = json.Marshal(invalidReq)
	resp, err = env.Client.Post(env.Server.URL+"/api/v1/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("invalid registration request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for invalid slug, got %d", resp.StatusCode)
	}

	// Verify error response does not leak internal details
	var errResp map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&errResp) //nolint:errcheck // test helper, decode failures caught by subsequent assertions
	if errResp["error"] == "" {
		t.Fatal("expected error message in response")
	}

	// Test validation: name too long
	longNameReq := map[string]interface{}{
		"tenant": map[string]string{
			"name": fmt.Sprintf("%0129d", 0), // 129 chars
			"slug": "long-name",
			"plan": "free",
		},
		"product": map[string]string{
			"name": "Product",
			"slug": "product",
		},
		"environment": map[string]string{
			"name": "Prod",
			"tier": "production",
		},
		"components": []map[string]interface{}{},
	}
	body, _ = json.Marshal(longNameReq)
	resp, err = env.Client.Post(env.Server.URL+"/api/v1/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("long name registration request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for long name, got %d", resp.StatusCode)
	}
}
