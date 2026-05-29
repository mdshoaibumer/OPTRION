package domain_test

import (
	"testing"

	"github.com/optrion/optrion/internal/tenant/domain"
)

// --- Slug Tests ---

func TestNewSlug_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"gym-flow", "gym-flow"},
		{"abc", "abc"},
		{"my-app-123", "my-app-123"},
		{"a2b", "a2b"},
		{"GYM-Flow", "gym-flow"},     // lowercased
		{"  gym-flow  ", "gym-flow"}, // trimmed
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			slug, err := domain.NewSlug(tt.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if slug.String() != tt.want {
				t.Errorf("got %q, want %q", slug.String(), tt.want)
			}
		})
	}
}

func TestNewSlug_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too short", "ab"},
		{"starts with hyphen", "-abc"},
		{"ends with hyphen", "abc-"},
		{"has space", "my slug"},
		{"has underscore", "my_slug"},
		{"empty", ""},
		{"too long", string(make([]byte, 101))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewSlug(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// --- Plan Tests ---

func TestPlan_IsValid(t *testing.T) {
	valid := []domain.Plan{domain.PlanFree, domain.PlanStarter, domain.PlanProfessional, domain.PlanEnterprise}
	for _, p := range valid {
		if !p.IsValid() {
			t.Errorf("expected %s to be valid", p)
		}
	}

	invalid := domain.Plan("ultimate")
	if invalid.IsValid() {
		t.Error("expected 'ultimate' to be invalid")
	}
}

// --- Tier Tests ---

func TestTier_IsValid(t *testing.T) {
	valid := []domain.Tier{domain.TierDevelopment, domain.TierStaging, domain.TierProduction}
	for _, tier := range valid {
		if !tier.IsValid() {
			t.Errorf("expected %s to be valid", tier)
		}
	}

	invalid := domain.Tier("canary")
	if invalid.IsValid() {
		t.Error("expected 'canary' to be invalid")
	}
}

// --- ComponentKind Tests ---

func TestComponentKind_IsValid(t *testing.T) {
	valid := []domain.ComponentKind{
		domain.KindDatabase, domain.KindCache, domain.KindAPI, domain.KindWeb,
		domain.KindQueue, domain.KindStorage, domain.KindService, domain.KindExternal,
	}
	for _, k := range valid {
		if !k.IsValid() {
			t.Errorf("expected %s to be valid", k)
		}
	}

	invalid := domain.ComponentKind("lambda")
	if invalid.IsValid() {
		t.Error("expected 'lambda' to be invalid")
	}
}

// --- Tenant Tests ---

func TestNewTenant_Success(t *testing.T) {
	slug, _ := domain.NewSlug("gym-flow")
	tenant, err := domain.NewTenant("GymFlow", slug, domain.PlanFree)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tenant.ID == "" {
		t.Error("expected non-empty ID")
	}
	if tenant.Name != "GymFlow" {
		t.Errorf("got name %q, want %q", tenant.Name, "GymFlow")
	}
	if tenant.Slug != slug {
		t.Errorf("got slug %q, want %q", tenant.Slug, slug)
	}
	if tenant.Plan != domain.PlanFree {
		t.Errorf("got plan %s, want %s", tenant.Plan, domain.PlanFree)
	}
	if tenant.Status != domain.StatusActive {
		t.Errorf("got status %s, want %s", tenant.Status, domain.StatusActive)
	}
	if tenant.Settings == nil {
		t.Error("expected non-nil settings map")
	}
	if tenant.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

func TestNewTenant_EmptyName(t *testing.T) {
	slug, _ := domain.NewSlug("test-slug")
	_, err := domain.NewTenant("", slug, domain.PlanFree)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestNewTenant_InvalidPlan(t *testing.T) {
	slug, _ := domain.NewSlug("test-slug")
	_, err := domain.NewTenant("Test", slug, domain.Plan("bogus"))
	if err == nil {
		t.Fatal("expected error for invalid plan")
	}
}

func TestTenant_Suspend(t *testing.T) {
	slug, _ := domain.NewSlug("test-slug")
	tenant, _ := domain.NewTenant("Test", slug, domain.PlanFree)

	err := tenant.Suspend()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tenant.Status != domain.StatusSuspended {
		t.Errorf("expected suspended, got %s", tenant.Status)
	}

	// Cannot suspend again
	err = tenant.Suspend()
	if err == nil {
		t.Fatal("expected error suspending already suspended tenant")
	}
}

func TestTenant_Activate(t *testing.T) {
	slug, _ := domain.NewSlug("test-slug")
	tenant, _ := domain.NewTenant("Test", slug, domain.PlanFree)
	_ = tenant.Suspend()

	err := tenant.Activate()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tenant.Status != domain.StatusActive {
		t.Errorf("expected active, got %s", tenant.Status)
	}

	// Cannot activate an active tenant
	err = tenant.Activate()
	if err == nil {
		t.Fatal("expected error activating already active tenant")
	}
}

// --- Product Tests ---

func TestNewProduct_Success(t *testing.T) {
	slug, _ := domain.NewSlug("backend-api")
	product, err := domain.NewProduct("019450c0-7e90-7a0f-8c1e-123456789abc", "Backend API", slug, "Main backend service")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product.ID == "" {
		t.Error("expected non-empty ID")
	}
	if product.Status != domain.StatusActive {
		t.Errorf("expected active status, got %s", product.Status)
	}
}

func TestNewProduct_InvalidTenantID(t *testing.T) {
	slug, _ := domain.NewSlug("backend-api")
	_, err := domain.NewProduct("not-a-uuid", "Backend", slug, "")
	if err == nil {
		t.Fatal("expected error for invalid tenant ID")
	}
}

func TestNewProduct_EmptyName(t *testing.T) {
	slug, _ := domain.NewSlug("backend-api")
	_, err := domain.NewProduct("019450c0-7e90-7a0f-8c1e-123456789abc", "", slug, "")
	if err == nil {
		t.Fatal("expected error for empty product name")
	}
}

// --- Environment Tests ---

func TestNewEnvironment_Success(t *testing.T) {
	slug, _ := domain.NewSlug("production")
	env, err := domain.NewEnvironment(
		"019450c0-7e90-7a0f-8c1e-123456789abc",
		"019450c0-7e90-7a0f-8c1e-123456789def",
		"Production",
		slug,
		domain.TierProduction,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if env.Tier != domain.TierProduction {
		t.Errorf("expected production tier, got %s", env.Tier)
	}
}

func TestNewEnvironment_InvalidTier(t *testing.T) {
	slug, _ := domain.NewSlug("canary-env")
	_, err := domain.NewEnvironment(
		"019450c0-7e90-7a0f-8c1e-123456789abc",
		"019450c0-7e90-7a0f-8c1e-123456789def",
		"Canary",
		slug,
		domain.Tier("canary"),
	)
	if err == nil {
		t.Fatal("expected error for invalid tier")
	}
}

// --- Component Tests ---

func TestNewComponent_Success(t *testing.T) {
	slug, _ := domain.NewSlug("postgres-main")
	comp, err := domain.NewComponent(
		"019450c0-7e90-7a0f-8c1e-123456789abc",
		"019450c0-7e90-7a0f-8c1e-123456789def",
		"019450c0-7e90-7a0f-8c1e-1234567890ab",
		"PostgreSQL Main",
		slug,
		domain.KindDatabase,
		"postgresql://localhost:5432/gymflow",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if comp.Kind != domain.KindDatabase {
		t.Errorf("expected database kind, got %s", comp.Kind)
	}
	if comp.Metadata == nil {
		t.Error("expected non-nil metadata map")
	}
}

func TestNewComponent_InvalidKind(t *testing.T) {
	slug, _ := domain.NewSlug("lambda-fn")
	_, err := domain.NewComponent(
		"019450c0-7e90-7a0f-8c1e-123456789abc",
		"019450c0-7e90-7a0f-8c1e-123456789def",
		"019450c0-7e90-7a0f-8c1e-1234567890ab",
		"Lambda Function",
		slug,
		domain.ComponentKind("lambda"),
		"",
	)
	if err == nil {
		t.Fatal("expected error for invalid kind")
	}
}

func TestNewComponent_EmptyName(t *testing.T) {
	slug, _ := domain.NewSlug("some-comp")
	_, err := domain.NewComponent(
		"019450c0-7e90-7a0f-8c1e-123456789abc",
		"019450c0-7e90-7a0f-8c1e-123456789def",
		"019450c0-7e90-7a0f-8c1e-1234567890ab",
		"",
		slug,
		domain.KindAPI,
		"",
	)
	if err == nil {
		t.Fatal("expected error for empty component name")
	}
}
