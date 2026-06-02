package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/optrion/optrion/test/testutil"
)

func runDockerCommand(t *testing.T, action, container string) {
	t.Helper()
	cmd := exec.Command("docker", action, container)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run docker %s %s: %v; output=%s", action, container, err, string(output))
	}
	// Give it a moment to take effect
	t.Logf("docker %s %s succeeded: %s", action, container, string(output))
	time.Sleep(1 * time.Second)
}

func TestE2E_OutageResilience(t *testing.T) {
	if os.Getenv("SKIP_DOCKER") == "true" {
		t.Skip("Skipping outage resilience test — requires Docker")
	}

	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	// Verify initially healthy
	resp, err := env.Client.Get(env.Server.URL + "/readyz")
	if err != nil {
		t.Fatalf("Failed to call /readyz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// 1. Simulate PostgreSQL Outage
	t.Run("PostgreSQL Outage and Recovery", func(t *testing.T) {
		runDockerCommand(t, "stop", "optrion_e2e-postgres-1")
		defer runDockerCommand(t, "start", "optrion_e2e-postgres-1")

		// Verify readyz returns unhealthy (503)
		respOutage, err := env.Client.Get(env.Server.URL + "/readyz")
		if err != nil {
			t.Fatalf("Failed to call /readyz during outage: %v", err)
		}
		defer respOutage.Body.Close()

		var healthResp map[string]interface{}
		_ = json.NewDecoder(respOutage.Body).Decode(&healthResp)

		if respOutage.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("Expected 503 Service Unavailable, got %d. Response: %v", respOutage.StatusCode, healthResp)
		}

		// Restore PostgreSQL
		runDockerCommand(t, "start", "optrion_e2e-postgres-1")

		// Wait for PG to be ready and connection to recover
		var recoverResp *http.Response
		var recoverErr error
		for i := 0; i < 45; i++ {
			time.Sleep(1 * time.Second)
			recoverResp, recoverErr = env.Client.Get(env.Server.URL + "/readyz")
			if recoverErr == nil && recoverResp.StatusCode == http.StatusOK {
				break
			}
			if recoverResp != nil {
				recoverResp.Body.Close()
			}
		}

		if recoverErr != nil || recoverResp == nil || recoverResp.StatusCode != http.StatusOK {
			t.Fatalf("System failed to recover after PostgreSQL came back: status %v, err: %v", recoverResp, recoverErr)
		}
		defer recoverResp.Body.Close()
	})

	// 2. Simulate Redis Outage
	t.Run("Redis Outage and Recovery", func(t *testing.T) {
		runDockerCommand(t, "stop", "optrion_e2e-redis-1")
		defer runDockerCommand(t, "start", "optrion_e2e-redis-1")

		// Verify readyz returns unhealthy (503)
		respOutage, err := env.Client.Get(env.Server.URL + "/readyz")
		if err != nil {
			t.Fatalf("Failed to call /readyz during Redis outage: %v", err)
		}
		defer respOutage.Body.Close()

		if respOutage.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("Expected 503 Service Unavailable, got %d", respOutage.StatusCode)
		}

		// Restore Redis
		runDockerCommand(t, "start", "optrion_e2e-redis-1")

		// Wait for recovery
		var recoverResp *http.Response
		var recoverErr error
		for i := 0; i < 30; i++ {
			time.Sleep(1 * time.Second)
			recoverResp, recoverErr = env.Client.Get(env.Server.URL + "/readyz")
			if recoverErr == nil && recoverResp.StatusCode == http.StatusOK {
				break
			}
			if recoverResp != nil {
				recoverResp.Body.Close()
			}
		}

		if recoverErr != nil || recoverResp == nil || recoverResp.StatusCode != http.StatusOK {
			t.Fatalf("System failed to recover after Redis came back: status %v, err: %v", recoverResp, recoverErr)
		}
		defer recoverResp.Body.Close()
	})
}
