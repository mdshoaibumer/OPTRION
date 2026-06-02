package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/optrion/optrion/test/testutil"
)

type apiInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

type listResponse struct {
	Data  []map[string]interface{} `json:"data"`
	Count int                      `json:"count"`
}

func TestE2E_APIContract(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Basic readiness and info endpoints.
	for _, endpoint := range []struct {
		path string
		code int
	}{
		{"/healthz", http.StatusOK},
		{"/readyz", http.StatusOK},
		{"/api/v1/info", http.StatusOK},
	} {
		resp, err := env.Client.Get(env.Server.URL + endpoint.path)
		if err != nil {
			t.Fatalf("failed to call %s: %v", endpoint.path, err)
		}
		if resp.StatusCode != endpoint.code {
			t.Fatalf("expected %d from %s, got %d", endpoint.code, endpoint.path, resp.StatusCode)
		}
		resp.Body.Close()
	}

	resp, err := env.Client.Get(env.Server.URL + "/api/v1/info")
	if err != nil {
		t.Fatalf("failed to call /api/v1/info: %v", err)
	}
	defer resp.Body.Close()
	var info apiInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		t.Fatalf("failed to decode /api/v1/info response: %v", err)
	}
	if info.Name == "" || info.Version == "" || info.Environment == "" {
		t.Fatalf("api info response missing required fields: %+v", info)
	}

	// Tenant lifecycle and list contract — use registration endpoint (public, no auth).
	tenantID, apiKey := createAuthenticatedTenant(t, env, "Contract Tenant", "contract-tenant")

	// Create product and verify list contract.
	createProduct := map[string]string{"tenant_id": tenantID, "name": "Contract Product", "slug": "contract-product"}
	productBody, _ := json.Marshal(createProduct)
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/v1/products", bytes.NewBuffer(productBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err = env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created for product creation, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	req, _ = http.NewRequest("GET", env.Server.URL+fmt.Sprintf("/api/v1/products?tenant_id=%s", tenantID), nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	productListResp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	defer productListResp.Body.Close()
	if productListResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for product list, got %d", productListResp.StatusCode)
	}
	var products listResponse
	if err := json.NewDecoder(productListResp.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode product list response: %v", err)
	}
	if products.Count != len(products.Data) || products.Count < 1 {
		t.Fatalf("unexpected product list contract: %+v", products)
	}
	// Registration creates a product automatically; verify our explicit product exists.
	found := false
	for _, p := range products.Data {
		if p["slug"] == "contract-product" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected product slug contract-product in list, got %+v", products)
	}

	// Health summary contract should include overall_score and overall_status.
	req, _ = http.NewRequest("GET", env.Server.URL+fmt.Sprintf("/api/v1/health/summary?tenant_id=%s", tenantID), nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	healthResp, err := env.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to query health summary: %v", err)
	}
	defer healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for health summary, got %d", healthResp.StatusCode)
	}
	var healthData map[string]interface{}
	if err := json.NewDecoder(healthResp.Body).Decode(&healthData); err != nil {
		t.Fatalf("failed to decode health summary: %v", err)
	}
	if _, ok := healthData["overall_score"]; !ok {
		t.Fatalf("health summary response missing overall_score: %+v", healthData)
	}
	if _, ok := healthData["overall_status"]; !ok {
		t.Fatalf("health summary response missing overall_status: %+v", healthData)
	}
}
