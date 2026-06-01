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
	"github.com/optrion/optrion/internal/alert/domain/alert"
	"github.com/optrion/optrion/internal/platform/server"
)

// mockAlertRepo implements repository.AlertRepository for testing.
type mockAlertRepo struct {
	alerts []*alert.Alert
}

func (m *mockAlertRepo) Create(_ context.Context, a *alert.Alert) error {
	m.alerts = append(m.alerts, a)
	return nil
}

func (m *mockAlertRepo) Update(_ context.Context, a *alert.Alert) error {
	for i, existing := range m.alerts {
		if existing.ID == a.ID {
			m.alerts[i] = a
			return nil
		}
	}
	return nil
}

func (m *mockAlertRepo) FindByID(_ context.Context, id string) (*alert.Alert, error) {
	for _, a := range m.alerts {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAlertRepo) ListByTenant(_ context.Context, tenantID string) ([]*alert.Alert, error) {
	var result []*alert.Alert
	for _, a := range m.alerts {
		if a.TenantID == tenantID {
			result = append(result, a)
		}
	}
	return result, nil
}

func TestAlertHandler_GetAlerts(t *testing.T) {
	repo := &mockAlertRepo{
		alerts: []*alert.Alert{
			{ID: "a1", TenantID: "t1", Severity: "critical", Status: alert.AlertStatusPending, Message: "test", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "a2", TenantID: "t1", Severity: "warning", Status: alert.AlertStatusSent, Message: "test2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "a3", TenantID: "t2", Severity: "info", Status: alert.AlertStatusPending, Message: "other tenant", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	handler := v1.NewAlertHandler(repo, nil)

	t.Run("returns alerts for authenticated tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.GetAlerts(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp map[string]interface{}
		json.NewDecoder(rec.Body).Decode(&resp)
		if int(resp["total"].(float64)) != 2 {
			t.Fatalf("expected 2 alerts, got %v", resp["total"])
		}
	})

	t.Run("returns 401 without tenant context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
		rec := httptest.NewRecorder()

		handler.GetAlerts(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}

func TestAlertHandler_GetAlertByID(t *testing.T) {
	repo := &mockAlertRepo{
		alerts: []*alert.Alert{
			{ID: "a1", TenantID: "t1", Severity: "critical", Status: alert.AlertStatusPending, Message: "test", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	handler := v1.NewAlertHandler(repo, nil)

	t.Run("returns alert by ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/a1", nil)
		req.SetPathValue("id", "a1")
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.GetAlertByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("returns 404 for wrong tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/a1", nil)
		req.SetPathValue("id", "a1")
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t-other")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.GetAlertByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("returns 404 for non-existent alert", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/does-not-exist", nil)
		req.SetPathValue("id", "does-not-exist")
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.GetAlertByID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}

func TestAlertHandler_UpdateAlertStatus(t *testing.T) {
	repo := &mockAlertRepo{
		alerts: []*alert.Alert{
			{ID: "a1", TenantID: "t1", Severity: "critical", Status: alert.AlertStatusPending, Message: "test", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	handler := v1.NewAlertHandler(repo, nil)

	t.Run("updates alert status", func(t *testing.T) {
		body := `{"status":"delivered"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/alerts/a1", strings.NewReader(body))
		req.SetPathValue("id", "a1")
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.UpdateAlertStatus(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		body := `{"status":"invalid_status"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/alerts/a1", strings.NewReader(body))
		req.SetPathValue("id", "a1")
		ctx := context.WithValue(req.Context(), server.ContextKeyTenantID, "t1")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.UpdateAlertStatus(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}
