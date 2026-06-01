package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/optrion/optrion/test/testutil"
)

func TestE2E_TenantIsolation(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Helper to create tenant
	createTenant := func(name, slug string) string {
		body := map[string]string{
			"name": name,
			"slug": slug,
			"plan": "free",
		}
		jsonBody, _ := json.Marshal(body)
		resp, err := env.Client.Post(env.Server.URL+"/api/v1/tenants", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create tenant: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected 201 Created, got %d", resp.StatusCode)
		}

		var respData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&respData)
		return respData["id"].(string)
	}

	tenantAID := createTenant("Tenant A", "tenant-a")
	tenantBID := createTenant("Tenant B", "tenant-b")

	// Helper to create product
	createProduct := func(tenantID, name, slug string, expectStatus int) string {
		body := map[string]string{
			"tenant_id": tenantID,
			"name":      name,
			"slug":      slug,
		}
		jsonBody, _ := json.Marshal(body)
		resp, err := env.Client.Post(env.Server.URL+"/api/v1/products", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != expectStatus {
			t.Fatalf("Expected %d, got %d", expectStatus, resp.StatusCode)
		}

		if expectStatus == http.StatusCreated {
			var respData map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&respData)
			return respData["id"].(string)
		}
		return ""
	}

	// 1. Tenant A creates Product A
	productAID := createProduct(tenantAID, "Product A", "product-a", http.StatusCreated)

	// 2. Tenant B tries to create a product under Tenant A's ID
	// Since Tenant B is doing this, we verify that the service checks tenant validity
	// If Tenant B passes Tenant A's ID, it succeeds for Tenant A, but if they try to access/link cross-tenant:
	// Let's create a product for Tenant B:
	productBID := createProduct(tenantBID, "Product B", "product-b", http.StatusCreated)

	// Helper to create environment
	createEnv := func(tenantID, productID, name, slug string, expectStatus int) string {
		body := map[string]string{
			"tenant_id":  tenantID,
			"product_id": productID,
			"name":       name,
			"slug":       slug,
			"tier":       "production",
		}
		jsonBody, _ := json.Marshal(body)
		resp, err := env.Client.Post(env.Server.URL+"/api/v1/environments", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create environment: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != expectStatus {
			t.Fatalf("Expected %d, got %d", expectStatus, resp.StatusCode)
		}

		if expectStatus == http.StatusCreated {
			var respData map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&respData)
			return respData["id"].(string)
		}
		return ""
	}

	// 3. Tenant B tries to create an environment under Tenant A's Product A (Cross-tenant mapping)
	// Should fail with 404 (product not found under Tenant B)
	createEnv(tenantBID, productAID, "Env B", "env-b", http.StatusNotFound)

	// Tenant A creates Environment A under Product A (succeeds)
	envAID := createEnv(tenantAID, productAID, "Env A", "env-a", http.StatusCreated)
	// Tenant B creates Environment B under Product B (succeeds)
	_ = createEnv(tenantBID, productBID, "Env B", "env-b", http.StatusCreated)

	// Helper to register component
	registerComponent := func(tenantID, productID, envID, name, slug, kind string, expectStatus int) string {
		body := map[string]string{
			"tenant_id":      tenantID,
			"product_id":     productID,
			"environment_id": envID,
			"name":           name,
			"slug":           slug,
			"kind":           kind,
			"endpoint_url":   "http://localhost/health",
		}
		jsonBody, _ := json.Marshal(body)
		resp, err := env.Client.Post(env.Server.URL+"/api/v1/components", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to register component: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != expectStatus {
			t.Fatalf("Expected %d, got %d", expectStatus, resp.StatusCode)
		}

		if expectStatus == http.StatusCreated {
			var respData map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&respData)
			return respData["id"].(string)
		}
		return ""
	}

	// 4. Tenant B tries to register Component B under Tenant A's Environment A (Cross-tenant mapping)
	// Should fail with 404 (environment not found under Product B / Tenant B)
	registerComponent(tenantBID, productBID, envAID, "Comp B", "comp-b", "database", http.StatusNotFound)

	// Tenant A registers Component A under Environment A (succeeds)
	registerComponent(tenantAID, productAID, envAID, "Comp A", "comp-a", "database", http.StatusCreated)

	// 5. Verify product listing isolation
	listProducts := func(tenantID string) []string {
		resp, err := env.Client.Get(fmt.Sprintf("%s/api/v1/products?tenant_id=%s", env.Server.URL, tenantID))
		if err != nil {
			t.Fatalf("Failed to list products: %v", err)
		}
		defer resp.Body.Close()

		var respData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&respData)
		data := respData["data"].([]interface{})

		var slugs []string
		for _, item := range data {
			slugs = append(slugs, item.(map[string]interface{})["slug"].(string))
		}
		return slugs
	}

	aProducts := listProducts(tenantAID)
	bProducts := listProducts(tenantBID)

	if len(aProducts) != 1 || aProducts[0] != "product-a" {
		t.Errorf("Expected Tenant A to have exactly 'product-a', got: %v", aProducts)
	}

	if len(bProducts) != 1 || bProducts[0] != "product-b" {
		t.Errorf("Expected Tenant B to have exactly 'product-b', got: %v", bProducts)
	}
}
