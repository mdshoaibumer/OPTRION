package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// seed.go creates the GymFlow Track tenant with its full hierarchy.
// Usage: go run scripts/seed/seed.go [base_url]
// Default base URL: http://localhost:8080

func main() {
	baseURL := "http://localhost:8080"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	fmt.Printf("Seeding OPTRION at %s\n\n", baseURL)

	// 1. Create GymFlow tenant
	tenantID := createTenant(baseURL, map[string]string{
		"name": "GymFlow Track",
		"slug": "gymflow-track",
		"plan": "starter",
	})
	fmt.Printf("✓ Tenant created: %s\n", tenantID)

	// 2. Create Frontend product
	frontendID := createProduct(baseURL, map[string]string{
		"tenant_id":   tenantID,
		"name":        "Frontend",
		"slug":        "frontend",
		"description": "GymFlow Track web application",
	})
	fmt.Printf("✓ Product 'Frontend' created: %s\n", frontendID)

	// 3. Create Backend product
	backendID := createProduct(baseURL, map[string]string{
		"tenant_id":   tenantID,
		"name":        "Backend",
		"slug":        "backend",
		"description": "GymFlow Track API services",
	})
	fmt.Printf("✓ Product 'Backend' created: %s\n", backendID)

	// 4. Create Production environment for Backend
	prodEnvID := createEnvironment(baseURL, map[string]string{
		"tenant_id":  tenantID,
		"product_id": backendID,
		"name":       "Production",
		"slug":       "production",
		"tier":       "production",
	})
	fmt.Printf("✓ Environment 'Production' created: %s\n", prodEnvID)

	// 5. Create Development environment for Backend
	devEnvID := createEnvironment(baseURL, map[string]string{
		"tenant_id":  tenantID,
		"product_id": backendID,
		"name":       "Development",
		"slug":       "development",
		"tier":       "development",
	})
	fmt.Printf("✓ Environment 'Development' created: %s\n", devEnvID)

	// 6. Register components in Production
	pgID := registerComponent(baseURL, map[string]string{
		"tenant_id":      tenantID,
		"product_id":     backendID,
		"environment_id": prodEnvID,
		"name":           "PostgreSQL",
		"slug":           "postgres-main",
		"kind":           "database",
		"endpoint_url":   "postgresql://prod-db.gymflow.internal:5432/gymflow",
	})
	fmt.Printf("✓ Component 'PostgreSQL' registered: %s\n", pgID)

	redisID := registerComponent(baseURL, map[string]string{
		"tenant_id":      tenantID,
		"product_id":     backendID,
		"environment_id": prodEnvID,
		"name":           "Redis",
		"slug":           "redis-cache",
		"kind":           "cache",
		"endpoint_url":   "redis://prod-redis.gymflow.internal:6379",
	})
	fmt.Printf("✓ Component 'Redis' registered: %s\n", redisID)

	apiID := registerComponent(baseURL, map[string]string{
		"tenant_id":      tenantID,
		"product_id":     backendID,
		"environment_id": prodEnvID,
		"name":           "Backend API",
		"slug":           "backend-api",
		"kind":           "api",
		"endpoint_url":   "https://api.gymflow.track/v1",
	})
	fmt.Printf("✓ Component 'Backend API' registered: %s\n", apiID)

	fmt.Printf("\n✓ Seed complete! GymFlow Track is ready.\n")
}

func createTenant(baseURL string, data map[string]string) string {
	return post(baseURL+"/api/v1/tenants", data)
}

func createProduct(baseURL string, data map[string]string) string {
	return post(baseURL+"/api/v1/products", data)
}

func createEnvironment(baseURL string, data map[string]string) string {
	return post(baseURL+"/api/v1/environments", data)
}

func registerComponent(baseURL string, data map[string]string) string {
	return post(baseURL+"/api/v1/components", data)
}

func post(url string, data map[string]string) string {
	body, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body)) //nolint:gosec // seed script uses trusted local URL
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: POST %s failed: %v\n", url, err)
		os.Exit(1) //nolint:gocritic // acceptable in CLI script
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		fmt.Fprintf(os.Stderr, "ERROR: POST %s returned %d: %s\n", url, resp.StatusCode, string(respBody))
		os.Exit(1) //nolint:gocritic // acceptable in CLI script
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to parse response: %v\n", err)
		os.Exit(1)
	}

	id, ok := result["id"].(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: response has no 'id' field\n")
		os.Exit(1)
	}
	return id
}
