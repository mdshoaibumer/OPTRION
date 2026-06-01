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
	"github.com/optrion/optrion/internal/alert/domain/alertrule"
	"github.com/optrion/optrion/internal/platform/server"
)

// mockAlertRuleRepo implements repository.AlertRuleRepository for testing.
type mockAlertRuleRepo struct {
	rules []*alertrule.AlertRule
}

func (m *mockAlertRuleRepo) Create(_ context.Context, r *alertrule.AlertRule) error {
	m.rules = append(m.rules, r)
	return nil
}

func (m *mockAlertRuleRepo) Update(_ context.Context, r *alertrule.AlertRule) error {
	for i, existing := range m.rules {
		if existing.ID == r.ID {
			m.rules[i] = r
			return nil
		}
	}
	return nil
}

func (m *mockAlertRuleRepo) FindByID(_ context.Context, id string) (*alertrule.AlertRule, error) {
	for _, r := range m.rules {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockAlertRuleRepo) ListByTenant(_ context.Context, tenantID string) ([]*alertrule.AlertRule, error) {
	var result []*alertrule.AlertRule
	for _, r := range m.rules {
		if r.TenantID == tenantID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockAlertRuleRepo) ListEnabledByTenant(_ context.Context, tenantID string) ([]*alertrule.AlertRule, error) {
	var result []*alertrule.AlertRule
	for _, r := range m.rules {
		if r.TenantID == tenantID && r.Enabled {
			result = append(result, r)
		}
	}
	return result, nil
}

func TestAlertRuleHandler_GetAlertRules(t *testing.T) {
	repo := &mockAlertRuleRepo{
		rules: []*alertrule.AlertRule{
			{ID: "r1", TenantID: "t1", Name: "High CPU", Severity: "critical", Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "r2", TenantID: "t1", Name: "Low Disk", Severity: "warning", Enabled: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	handler := v1.NewAlertRuleHandler(repo, nil)

	t.Run("returns rules for tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/alert-rules", nil)
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.GetAlertRules(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp map[string]interface{}
		json.NewDecoder(rec.Body).Decode(&resp)
		if int(resp["total"].(float64)) != 2 {
			t.Fatalf("expected 2 rules, got %v", resp["total"])
		}
	})
}

func TestAlertRuleHandler_PostAlertRule(t *testing.T) {
	repo := &mockAlertRuleRepo{}
	handler := v1.NewAlertRuleHandler(repo, nil)

	t.Run("creates alert rule", func(t *testing.T) {
		body := `{"name":"New Rule","severity":"critical","enabled":true,"channels":["ch1"]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/alert-rules", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.PostAlertRule(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		if len(repo.rules) != 1 {
			t.Fatalf("expected 1 rule in repo, got %d", len(repo.rules))
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		body := `{"name":"","severity":"critical"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/alert-rules", strings.NewReader(body))
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.PostAlertRule(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}

func TestAlertRuleHandler_PatchAlertRule(t *testing.T) {
	repo := &mockAlertRuleRepo{
		rules: []*alertrule.AlertRule{
			{ID: "r1", TenantID: "t1", Name: "Old Name", Severity: "warning", Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	handler := v1.NewAlertRuleHandler(repo, nil)

	t.Run("updates alert rule fields", func(t *testing.T) {
		body := `{"name":"New Name","enabled":false}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/alert-rules/r1", strings.NewReader(body))
		req.SetPathValue("id", "r1")
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.PatchAlertRule(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("returns 404 for wrong tenant", func(t *testing.T) {
		body := `{"name":"New Name"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/alert-rules/r1", strings.NewReader(body))
		req.SetPathValue("id", "r1")
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t-other")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.PatchAlertRule(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}
