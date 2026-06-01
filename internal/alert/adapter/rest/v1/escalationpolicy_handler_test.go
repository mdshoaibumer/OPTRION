package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	v1 "github.com/optrion/optrion/internal/alert/adapter/rest/v1"
	"github.com/optrion/optrion/internal/alert/domain/escalationpolicy"
	"github.com/optrion/optrion/internal/platform/server"
)

// mockEscalationPolicyRepo implements repository.EscalationPolicyRepository for testing.
type mockEscalationPolicyRepo struct {
	policies []*escalationpolicy.EscalationPolicy
}

func (m *mockEscalationPolicyRepo) Create(_ context.Context, p *escalationpolicy.EscalationPolicy) error {
	m.policies = append(m.policies, p)
	return nil
}

func (m *mockEscalationPolicyRepo) Update(_ context.Context, p *escalationpolicy.EscalationPolicy) error {
	for i, existing := range m.policies {
		if existing.ID == p.ID {
			m.policies[i] = p
			return nil
		}
	}
	return nil
}

func (m *mockEscalationPolicyRepo) FindByID(_ context.Context, id string) (*escalationpolicy.EscalationPolicy, error) {
	for _, p := range m.policies {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockEscalationPolicyRepo) ListByTenant(_ context.Context, tenantID string) ([]*escalationpolicy.EscalationPolicy, error) {
	var result []*escalationpolicy.EscalationPolicy
	for _, p := range m.policies {
		if p.TenantID == tenantID {
			result = append(result, p)
		}
	}
	return result, nil
}

func TestEscalationPolicyHandler_GetEscalationPolicies(t *testing.T) {
	repo := &mockEscalationPolicyRepo{
		policies: []*escalationpolicy.EscalationPolicy{
			{
				ID: "p1", TenantID: "t1", Name: "Critical Path",
				Steps:     []escalationpolicy.EscalationStep{{DelayMinutes: 5, ChannelIDs: []string{"ch1"}}},
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			},
		},
	}

	handler := v1.NewEscalationPolicyHandler(repo, nil)

	t.Run("returns policies for tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/escalation-policies", nil)
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.GetEscalationPolicies(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp map[string]interface{}
		json.NewDecoder(rec.Body).Decode(&resp)
		if int(resp["total"].(float64)) != 1 {
			t.Fatalf("expected 1 policy, got %v", resp["total"])
		}
	})

	t.Run("returns 401 without tenant context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/escalation-policies", nil)
		rec := httptest.NewRecorder()

		handler.GetEscalationPolicies(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}

func TestEscalationPolicyHandler_PostEscalationPolicy(t *testing.T) {
	repo := &mockEscalationPolicyRepo{}
	handler := v1.NewEscalationPolicyHandler(repo, nil)

	t.Run("creates escalation policy", func(t *testing.T) {
		body := `{"name":"Critical Escalation","description":"Escalate critical alerts","steps":[{"delay_minutes":5,"channel_ids":["ch1"],"reminder":true}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/escalation-policies", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.PostEscalationPolicy(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		if len(repo.policies) != 1 {
			t.Fatalf("expected 1 policy in repo, got %d", len(repo.policies))
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		body := `{"name":"","steps":[{"delay_minutes":5}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/escalation-policies", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.PostEscalationPolicy(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("rejects empty steps", func(t *testing.T) {
		body := `{"name":"No Steps","steps":[]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/escalation-policies", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.PostEscalationPolicy(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}

func TestEscalationPolicyHandler_PatchEscalationPolicy(t *testing.T) {
	repo := &mockEscalationPolicyRepo{
		policies: []*escalationpolicy.EscalationPolicy{
			{
				ID: "p1", TenantID: "t1", Name: "Old Name",
				Steps:     []escalationpolicy.EscalationStep{{DelayMinutes: 5, ChannelIDs: []string{"ch1"}}},
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			},
		},
	}
	handler := v1.NewEscalationPolicyHandler(repo, nil)

	t.Run("updates policy name", func(t *testing.T) {
		body := `{"name":"New Name"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/escalation-policies/p1", strings.NewReader(body))
		req.SetPathValue("id", "p1")
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.PatchEscalationPolicy(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("returns 404 for wrong tenant", func(t *testing.T) {
		body := `{"name":"Attempt"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/escalation-policies/p1", strings.NewReader(body))
		req.SetPathValue("id", "p1")
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t-other")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.PatchEscalationPolicy(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}
