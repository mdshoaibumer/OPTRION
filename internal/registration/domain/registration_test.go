package domain

import (
	"strings"
	"testing"
)

func TestTenantRegistration_Validate_Valid(t *testing.T) {
	tr := TenantRegistration{
		Name: "Acme Corp",
		Slug: "acme-corp",
		Plan: "free",
	}
	if err := tr.Validate(); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestTenantRegistration_Validate_EmptyName(t *testing.T) {
	tr := TenantRegistration{Name: "", Slug: "acme", Plan: "free"}
	if err := tr.Validate(); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestTenantRegistration_Validate_EmptySlug(t *testing.T) {
	tr := TenantRegistration{Name: "Acme", Slug: "", Plan: "free"}
	if err := tr.Validate(); err == nil {
		t.Fatal("expected error for empty slug")
	}
}

func TestTenantRegistration_Validate_SlugTooShort(t *testing.T) {
	tr := TenantRegistration{Name: "Acme", Slug: "ab", Plan: "free"}
	if err := tr.Validate(); err == nil {
		t.Fatal("expected error for slug too short")
	}
}

func TestTenantRegistration_Validate_SlugInvalidChars(t *testing.T) {
	invalid := []string{
		"UPPERCASE",
		"has spaces",
		"has_underscore",
		"has.dot",
		"has/slash",
		"-starts-with-hyphen",
		"ends-with-hyphen-",
	}
	for _, slug := range invalid {
		t.Run(slug, func(t *testing.T) {
			tr := TenantRegistration{Name: "Acme", Slug: slug, Plan: "free"}
			if err := tr.Validate(); err == nil {
				t.Fatalf("expected error for invalid slug: %s", slug)
			}
		})
	}
}

func TestTenantRegistration_Validate_SlugValidFormats(t *testing.T) {
	valid := []string{
		"acme",
		"acme-corp",
		"my-company-123",
		"abc",
	}
	for _, slug := range valid {
		t.Run(slug, func(t *testing.T) {
			tr := TenantRegistration{Name: "Acme", Slug: slug, Plan: "free"}
			if err := tr.Validate(); err != nil {
				t.Fatalf("expected valid slug %q, got error: %v", slug, err)
			}
		})
	}
}

func TestTenantRegistration_Validate_NameTooLong(t *testing.T) {
	tr := TenantRegistration{
		Name: strings.Repeat("a", 129),
		Slug: "acme",
		Plan: "free",
	}
	if err := tr.Validate(); err == nil {
		t.Fatal("expected error for name too long")
	}
}

func TestTenantRegistration_Validate_SlugTooLong(t *testing.T) {
	tr := TenantRegistration{
		Name: "Acme",
		Slug: strings.Repeat("a", 65),
		Plan: "free",
	}
	if err := tr.Validate(); err == nil {
		t.Fatal("expected error for slug too long")
	}
}

func TestTenantRegistration_Validate_InvalidPlan(t *testing.T) {
	tr := TenantRegistration{Name: "Acme", Slug: "acme", Plan: "invalid"}
	if err := tr.Validate(); err == nil {
		t.Fatal("expected error for invalid plan")
	}
}

func TestProductRegistration_Validate_Valid(t *testing.T) {
	pr := ProductRegistration{
		Name:        "My Product",
		Slug:        "my-product",
		Description: "A great product",
	}
	if err := pr.Validate(); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestProductRegistration_Validate_SlugFormat(t *testing.T) {
	pr := ProductRegistration{Name: "Product", Slug: "INVALID_SLUG!", Description: ""}
	if err := pr.Validate(); err == nil {
		t.Fatal("expected error for invalid slug format")
	}
}

func TestProductRegistration_Validate_DescriptionTooLong(t *testing.T) {
	pr := ProductRegistration{
		Name:        "Product",
		Slug:        "product",
		Description: strings.Repeat("x", 513),
	}
	if err := pr.Validate(); err == nil {
		t.Fatal("expected error for description too long")
	}
}

func TestEnvironmentRegistration_Validate_Valid(t *testing.T) {
	er := EnvironmentRegistration{Name: "Production", Tier: "production"}
	if err := er.Validate(); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestEnvironmentRegistration_Validate_InvalidTier(t *testing.T) {
	er := EnvironmentRegistration{Name: "Dev", Tier: "invalid"}
	if err := er.Validate(); err == nil {
		t.Fatal("expected error for invalid tier")
	}
}

func TestEnvironmentRegistration_Validate_NameTooLong(t *testing.T) {
	er := EnvironmentRegistration{
		Name: strings.Repeat("x", 129),
		Tier: "production",
	}
	if err := er.Validate(); err == nil {
		t.Fatal("expected error for name too long")
	}
}

func TestComponentRegistration_Validate_Valid(t *testing.T) {
	cr := ComponentRegistration{
		Name:        "PostgreSQL",
		Kind:        "database",
		Description: "Primary database",
		Endpoint:    "localhost",
		Port:        5432,
	}
	if err := cr.Validate(); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestComponentRegistration_Validate_EmptyKind(t *testing.T) {
	cr := ComponentRegistration{Name: "DB", Kind: ""}
	if err := cr.Validate(); err == nil {
		t.Fatal("expected error for empty kind")
	}
}

func TestComponentRegistration_Validate_InvalidKind(t *testing.T) {
	cr := ComponentRegistration{Name: "DB", Kind: "invalid_kind"}
	if err := cr.Validate(); err == nil {
		t.Fatal("expected error for invalid kind")
	}
}

func TestComponentRegistration_Validate_DescriptionTooLong(t *testing.T) {
	cr := ComponentRegistration{
		Name:        "DB",
		Kind:        "database",
		Description: strings.Repeat("x", 513),
	}
	if err := cr.Validate(); err == nil {
		t.Fatal("expected error for description too long")
	}
}
