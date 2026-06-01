package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/optrion/optrion/internal/registration/domain"
)

// mockTenantService stubs out the tenant service for testing.
type mockAPIKeyGenerator struct {
	lastTenantID string
	key          string
	err          error
}

func (m *mockAPIKeyGenerator) Generate(tenantID string) (string, error) {
	m.lastTenantID = tenantID
	if m.err != nil {
		return "", m.err
	}
	return m.key, nil
}

type mockRegistrationRepository struct {
	audits []*domain.RegistrationAudit
}

func (m *mockRegistrationRepository) CreateAudit(_ context.Context, audit *domain.RegistrationAudit) error {
	m.audits = append(m.audits, audit)
	return nil
}

func (m *mockRegistrationRepository) GetAuditByID(_ context.Context, id string) (*domain.RegistrationAudit, error) {
	for _, a := range m.audits {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockRegistrationRepository) ListAuditsByTenant(_ context.Context, tenantID string) ([]*domain.RegistrationAudit, error) {
	var result []*domain.RegistrationAudit
	for _, a := range m.audits {
		if a.TenantID == tenantID {
			result = append(result, a)
		}
	}
	return result, nil
}

func TestRegistrationRequest_Validate_InvalidTenant(t *testing.T) {
	tests := []struct {
		name string
		req  domain.TenantRegistration
	}{
		{"empty_name", domain.TenantRegistration{Name: "", Slug: "valid-slug", Plan: "free"}},
		{"empty_slug", domain.TenantRegistration{Name: "Test", Slug: "", Plan: "free"}},
		{"short_slug", domain.TenantRegistration{Name: "Test", Slug: "ab", Plan: "free"}},
		{"invalid_slug", domain.TenantRegistration{Name: "Test", Slug: "INVALID", Plan: "free"}},
		{"empty_plan", domain.TenantRegistration{Name: "Test", Slug: "valid-slug", Plan: ""}},
		{"invalid_plan", domain.TenantRegistration{Name: "Test", Slug: "valid-slug", Plan: "unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestRegistrationRequest_Validate_ValidTenant(t *testing.T) {
	req := domain.TenantRegistration{
		Name: "My Company",
		Slug: "my-company",
		Plan: "free",
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid tenant, got: %v", err)
	}
}

func TestRegistrationRequest_Validate_InvalidProduct(t *testing.T) {
	tests := []struct {
		name string
		req  domain.ProductRegistration
	}{
		{"empty_name", domain.ProductRegistration{Name: "", Slug: "valid-slug"}},
		{"empty_slug", domain.ProductRegistration{Name: "Test", Slug: ""}},
		{"short_slug", domain.ProductRegistration{Name: "Test", Slug: "ab"}},
		{"invalid_slug", domain.ProductRegistration{Name: "Test", Slug: "INVALID!"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestRegistrationRequest_Validate_InvalidEnvironment(t *testing.T) {
	tests := []struct {
		name string
		req  domain.EnvironmentRegistration
	}{
		{"empty_name", domain.EnvironmentRegistration{Name: "", Tier: "production"}},
		{"empty_tier", domain.EnvironmentRegistration{Name: "Prod", Tier: ""}},
		{"invalid_tier", domain.EnvironmentRegistration{Name: "Prod", Tier: "unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestRegistrationRequest_Validate_InvalidComponent(t *testing.T) {
	tests := []struct {
		name string
		req  domain.ComponentRegistration
	}{
		{"empty_name", domain.ComponentRegistration{Name: "", Kind: "database"}},
		{"empty_kind", domain.ComponentRegistration{Name: "PG", Kind: ""}},
		{"invalid_kind", domain.ComponentRegistration{Name: "PG", Kind: "unknown-type"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestAPIKeyGenerator_RecordsTenantID(t *testing.T) {
	gen := &mockAPIKeyGenerator{key: "opk_testkey123"}

	key, err := gen.Generate("tenant-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "opk_testkey123" {
		t.Fatalf("expected key opk_testkey123, got %s", key)
	}
	if gen.lastTenantID != "tenant-abc" {
		t.Fatalf("expected tenant-abc, got %s", gen.lastTenantID)
	}
}

func TestAPIKeyGenerator_Error(t *testing.T) {
	gen := &mockAPIKeyGenerator{err: fmt.Errorf("key generation failed")}

	_, err := gen.Generate("tenant-abc")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistrationAudit_MarkSuccess(t *testing.T) {
	audit := domain.NewRegistrationAudit("tenant-1", "bulk", nil)
	if audit.Status != "pending" {
		t.Fatalf("expected pending, got %s", audit.Status)
	}

	audit.MarkSuccess("response-data")
	if audit.Status != "success" {
		t.Fatalf("expected success, got %s", audit.Status)
	}
}

func TestRegistrationAudit_MarkFailed(t *testing.T) {
	audit := domain.NewRegistrationAudit("tenant-1", "bulk", nil)
	audit.MarkFailed(fmt.Errorf("something went wrong"))

	if audit.Status != "failed" {
		t.Fatalf("expected failed, got %s", audit.Status)
	}
	if audit.Error != "something went wrong" {
		t.Fatalf("expected error message, got %s", audit.Error)
	}
}

func TestRegistrationRepository_CreateAndRetrieve(t *testing.T) {
	repo := &mockRegistrationRepository{}

	audit := domain.NewRegistrationAudit("tenant-1", "bulk", nil)
	if err := repo.CreateAudit(context.Background(), audit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := repo.GetAuditByID(context.Background(), audit.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected audit to be found")
	}
	if retrieved.ID != audit.ID {
		t.Fatalf("expected ID %s, got %s", audit.ID, retrieved.ID)
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}
