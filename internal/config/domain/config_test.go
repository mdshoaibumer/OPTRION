package domain_test

import (
	"testing"

	"github.com/optrion/optrion/internal/config/domain"
)

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := &domain.Config{
		Tenant:      domain.TenantConfig{Name: "ACME Corp", Slug: "acme-corp", Plan: "professional"},
		Product:     domain.ProductConfig{Name: "Web App", Slug: "web-app"},
		Environment: domain.EnvironmentConfig{Name: "Production", Tier: "production"},
		Components: []domain.ComponentConfig{
			{Name: "API Server", Kind: "api", Endpoint: "localhost", Port: 8080},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestConfig_Validate_MissingTenantName(t *testing.T) {
	cfg := &domain.Config{
		Tenant:      domain.TenantConfig{Name: "", Slug: "acme", Plan: "free"},
		Product:     domain.ProductConfig{Name: "App", Slug: "app"},
		Environment: domain.EnvironmentConfig{Name: "Dev", Tier: "development"},
		Components:  []domain.ComponentConfig{{Name: "API", Kind: "api"}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing tenant name")
	}
}

func TestConfig_Validate_MissingTenantSlug(t *testing.T) {
	cfg := &domain.Config{
		Tenant:      domain.TenantConfig{Name: "ACME", Slug: "", Plan: "free"},
		Product:     domain.ProductConfig{Name: "App", Slug: "app"},
		Environment: domain.EnvironmentConfig{Name: "Dev", Tier: "development"},
		Components:  []domain.ComponentConfig{{Name: "API", Kind: "api"}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing tenant slug")
	}
}

func TestConfig_Validate_MissingPlan(t *testing.T) {
	cfg := &domain.Config{
		Tenant:      domain.TenantConfig{Name: "ACME", Slug: "acme", Plan: ""},
		Product:     domain.ProductConfig{Name: "App", Slug: "app"},
		Environment: domain.EnvironmentConfig{Name: "Dev", Tier: "development"},
		Components:  []domain.ComponentConfig{{Name: "API", Kind: "api"}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing plan")
	}
}

func TestConfig_Validate_MissingProductName(t *testing.T) {
	cfg := &domain.Config{
		Tenant:      domain.TenantConfig{Name: "ACME", Slug: "acme", Plan: "free"},
		Product:     domain.ProductConfig{Name: "", Slug: "app"},
		Environment: domain.EnvironmentConfig{Name: "Dev", Tier: "development"},
		Components:  []domain.ComponentConfig{{Name: "API", Kind: "api"}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing product name")
	}
}

func TestConfig_Validate_MissingEnvironmentTier(t *testing.T) {
	cfg := &domain.Config{
		Tenant:      domain.TenantConfig{Name: "ACME", Slug: "acme", Plan: "free"},
		Product:     domain.ProductConfig{Name: "App", Slug: "app"},
		Environment: domain.EnvironmentConfig{Name: "Dev", Tier: ""},
		Components:  []domain.ComponentConfig{{Name: "API", Kind: "api"}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing environment tier")
	}
}

func TestConfig_Validate_NoComponents(t *testing.T) {
	cfg := &domain.Config{
		Tenant:      domain.TenantConfig{Name: "ACME", Slug: "acme", Plan: "free"},
		Product:     domain.ProductConfig{Name: "App", Slug: "app"},
		Environment: domain.EnvironmentConfig{Name: "Dev", Tier: "development"},
		Components:  []domain.ComponentConfig{},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty components")
	}
}

func TestConfig_Validate_ComponentMissingName(t *testing.T) {
	cfg := &domain.Config{
		Tenant:      domain.TenantConfig{Name: "ACME", Slug: "acme", Plan: "free"},
		Product:     domain.ProductConfig{Name: "App", Slug: "app"},
		Environment: domain.EnvironmentConfig{Name: "Dev", Tier: "development"},
		Components:  []domain.ComponentConfig{{Name: "", Kind: "api"}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for component missing name")
	}
}

func TestConfig_Validate_ComponentMissingKind(t *testing.T) {
	cfg := &domain.Config{
		Tenant:      domain.TenantConfig{Name: "ACME", Slug: "acme", Plan: "free"},
		Product:     domain.ProductConfig{Name: "App", Slug: "app"},
		Environment: domain.EnvironmentConfig{Name: "Dev", Tier: "development"},
		Components:  []domain.ComponentConfig{{Name: "API", Kind: ""}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for component missing kind")
	}
}

func TestDefaultMonitoringConfig(t *testing.T) {
	cfg := domain.DefaultMonitoringConfig()

	if !cfg.Enabled {
		t.Error("expected monitoring enabled by default")
	}
	if cfg.Interval != 30 {
		t.Errorf("expected interval 30, got %d", cfg.Interval)
	}
	if cfg.HealthCheckPath != "/health" {
		t.Errorf("expected health check path /health, got %s", cfg.HealthCheckPath)
	}
	if len(cfg.MetricsCollectors) != 3 {
		t.Errorf("expected 3 default collectors, got %d", len(cfg.MetricsCollectors))
	}
}
