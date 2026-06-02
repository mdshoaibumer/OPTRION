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

	// Create two separate tenants with their own API keys
	tenantAID, apiKeyA := createAuthenticatedTenant(t, env, "Tenant A", "tenant-a")
	tenantBID, apiKeyB := createAuthenticatedTenant(t, env, "Tenant B", "tenant-b")

	// Helper to create product with auth
	createProduct := func(tenantID, name, slug, apiKey string, expectStatus int) string {
		body := map[string]string{
			"tenant_id": tenantID,
			"name":      name,
			"slug":      slug,
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/products", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := env.Client.Do(req)
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
	productAID := createProduct(tenantAID, "Product A", "product-a", apiKeyA, http.StatusCreated)

	// Let's create a product for Tenant B:
	productBID := createProduct(tenantBID, "Product B", "product-b", apiKeyB, http.StatusCreated)

	// Helper to create environment with auth
	createEnv := func(tenantID, productID, name, slug, apiKey string, expectStatus int) string {
		body := map[string]string{
			"tenant_id":  tenantID,
			"product_id": productID,
			"name":       name,
			"slug":       slug,
			"tier":       "production",
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/environments", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := env.Client.Do(req)
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
	createEnv(tenantBID, productAID, "Env B", "env-b", apiKeyB, http.StatusNotFound)

	// Tenant A creates Environment A under Product A (succeeds)
	envAID := createEnv(tenantAID, productAID, "Env A", "env-a", apiKeyA, http.StatusCreated)
	// Tenant B creates Environment B under Product B (succeeds)
	_ = createEnv(tenantBID, productBID, "Env B", "env-b", apiKeyB, http.StatusCreated)

	// Helper to register component with auth
	registerComponent := func(tenantID, productID, envID, name, slug, kind, apiKey string, expectStatus int) string {
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
		req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/components", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := env.Client.Do(req)
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
	registerComponent(tenantBID, productBID, envAID, "Comp B", "comp-b", "database", apiKeyB, http.StatusNotFound)

	// Tenant A registers Component A under Environment A (succeeds)
	registerComponent(tenantAID, productAID, envAID, "Comp A", "comp-a", "database", apiKeyA, http.StatusCreated)

	// 5. Verify product listing isolation
	listProducts := func(tenantID, apiKey string) []string {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/products?tenant_id=%s", env.Server.URL, tenantID), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)
		resp, err := env.Client.Do(req)
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

	aProducts := listProducts(tenantAID, apiKeyA)
	bProducts := listProducts(tenantBID, apiKeyB)

	// Registration creates a product automatically, so each tenant has 2 products.
	// The key isolation check: products from A should NOT appear in B's list and vice versa.
	hasSlug := func(slugs []string, target string) bool {
		for _, s := range slugs {
			if s == target {
				return true
			}
		}
		return false
	}

	if !hasSlug(aProducts, "product-a") {
		t.Errorf("Expected Tenant A products to contain 'product-a', got: %v", aProducts)
	}
	if hasSlug(aProducts, "product-b") {
		t.Errorf("Tenant A should NOT see Tenant B's 'product-b', got: %v", aProducts)
	}

	if !hasSlug(bProducts, "product-b") {
		t.Errorf("Expected Tenant B products to contain 'product-b', got: %v", bProducts)
	}
	if hasSlug(bProducts, "product-a") {
		t.Errorf("Tenant B should NOT see Tenant A's 'product-a', got: %v", bProducts)
	}
}
