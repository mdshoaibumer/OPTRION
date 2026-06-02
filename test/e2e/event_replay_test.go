package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/optrion/optrion/internal/incident/domain"
	"github.com/optrion/optrion/internal/shared/id"
	"github.com/optrion/optrion/test/testutil"
)

func TestE2E_EventReplay(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	ctx := context.Background()

	// 1. Setup tenant via registration endpoint (public, no auth needed)
	tenantID, apiKey := createAuthenticatedTenant(t, env, "Event Replay Tenant", "event-replay")

	productID := createProductAuth(t, env, tenantID, "App", "app", apiKey)
	envID := createEnvironmentAuth(t, env, tenantID, productID, "Production", "production", apiKey)
	componentID := createComponentAuth(t, env, tenantID, productID, envID, "frontend-web", "web", apiKey)

	// 2. Generate an Incident directly so we can execute REST operations on it
	inc, err := domain.NewIncident(tenantID, componentID, id.New(), "Frontend response latency high", "Description info", domain.SeverityWarning)
	if err != nil {
		t.Fatalf("Failed to create incident domain object: %v", err)
	}

	// Persist incident using container postgres connection directly or repository
	// Let's check incident repository queries. We can run Exec directly into `incidents` table.
	// incidents: id, tenant_id, component_id, title, description, status, severity, rule_id, correlation_id, occurred_at, created_at, updated_at
	_, err = env.AppContainer.Database.Pool().Exec(ctx, `
		INSERT INTO incidents (id, tenant_id, component_id, title, description, status, severity, rule_id, correlation_id, occurred_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, inc.ID, inc.TenantID, inc.ComponentID, inc.Title, inc.Description,
		inc.Status, inc.Severity, inc.RuleID, inc.CorrelationID, inc.OccurredAt, inc.CreatedAt, inc.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to insert incident: %v", err)
	}

	// Persist initial opened event too
	for _, evt := range inc.UncommittedEvents() {
		metadataJSON, _ := json.Marshal(evt.Metadata)
		_, err = env.AppContainer.Database.Pool().Exec(ctx, `
			INSERT INTO incident_events (id, tenant_id, incident_id, event_type, metadata, occurred_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, evt.ID, evt.TenantID, evt.IncidentID, evt.EventType, metadataJSON, evt.OccurredAt)
		if err != nil {
			t.Fatalf("Failed to insert initial event: %v", err)
		}
	}
	inc.ClearUncommittedEvents()

	incidentID := inc.ID

	// 3. Perform REST operations to trigger subsequent lifecycle events
	callAPI := func(action string, reqBody map[string]string) {
		jsonReq, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/incidents/%s/%s", env.Server.URL, incidentID, action), bytes.NewBuffer(jsonReq))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		r, err := env.Client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute REST API %s: %v", action, err)
		}
		r.Body.Close()
		if r.StatusCode != http.StatusOK {
			t.Fatalf("REST API %s returned status %d", action, r.StatusCode)
		}
	}

	// Acknowledge incident
	callAPI("acknowledge", map[string]string{"actor_id": "operator-bob"})
	// Investigate incident
	callAPI("investigate", map[string]string{"actor_id": "operator-bob"})
	// Resolve incident
	callAPI("resolve", map[string]string{"actor_id": "operator-bob", "resolution": "Scaled up Kubernetes replica counts"})

	// 4. Fetch the events from database and replay
	events, err := env.AppContainer.IncidentService.GetEvents(ctx, incidentID)
	if err != nil {
		t.Fatalf("Failed to fetch events from incident service: %v", err)
	}

	// Verify events length (should have opened, acknowledged, investigating, resolved = 4 events)
	if len(events) != 4 {
		t.Fatalf("Expected exactly 4 events in stream, got %d", len(events))
	}

	// 5. Replay using RebuildFromEvents
	rebuiltInc := domain.RebuildFromEvents(events)
	if rebuiltInc == nil {
		t.Fatal("Rebuilt incident is nil")
	}

	// Fetch current live incident details
	liveInc, err := env.AppContainer.IncidentService.GetIncident(ctx, incidentID)
	if err != nil {
		t.Fatalf("Failed to fetch live incident: %v", err)
	}

	// Assert reconstructed state matches live database state
	if rebuiltInc.ID != liveInc.ID {
		t.Errorf("ID mismatch: rebuilt %s, live %s", rebuiltInc.ID, liveInc.ID)
	}
	if rebuiltInc.Status != liveInc.Status {
		t.Errorf("Status mismatch: rebuilt %s, live %s", rebuiltInc.Status, liveInc.Status)
	}
	if rebuiltInc.Status != domain.IncidentStatusResolved {
		t.Errorf("Expected status to be resolved, got %s", rebuiltInc.Status)
	}
	if rebuiltInc.TenantID != liveInc.TenantID {
		t.Errorf("TenantID mismatch: rebuilt %s, live %s", rebuiltInc.TenantID, liveInc.TenantID)
	}
	if rebuiltInc.ComponentID != liveInc.ComponentID {
		t.Errorf("ComponentID mismatch: rebuilt %s, live %s", rebuiltInc.ComponentID, liveInc.ComponentID)
	}
	if rebuiltInc.Severity != liveInc.Severity {
		t.Errorf("Severity mismatch: rebuilt %s, live %s", rebuiltInc.Severity, liveInc.Severity)
	}
}
