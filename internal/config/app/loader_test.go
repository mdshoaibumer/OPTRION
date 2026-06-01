package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/optrion/optrion/internal/config/app"
)

func TestConfigLoader_Load_ValidFile(t *testing.T) {
	content := `
tenant:
  name: "Test Corp"
  slug: "test-corp"
  plan: "starter"

product:
  name: "Test App"
  slug: "test-app"
  description: "A test application"

environment:
  name: "Development"
  tier: "development"

components:
  - name: "API Server"
    kind: "api"
    endpoint: "localhost"
    port: 8080
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "optrion.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	loader := app.NewConfigLoader(filePath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Tenant.Name != "Test Corp" {
		t.Errorf("expected tenant name 'Test Corp', got %s", cfg.Tenant.Name)
	}
	if cfg.Product.Slug != "test-app" {
		t.Errorf("expected product slug 'test-app', got %s", cfg.Product.Slug)
	}
	if cfg.Environment.Tier != "development" {
		t.Errorf("expected tier development, got %s", cfg.Environment.Tier)
	}
	if len(cfg.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(cfg.Components))
	}
}

func TestConfigLoader_Load_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(filePath, []byte(":::invalid yaml"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	loader := app.NewConfigLoader(filePath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestConfigLoader_Load_FileNotFound(t *testing.T) {
	loader := app.NewConfigLoader("/nonexistent/path/optrion.yaml")
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestConfigLoader_Load_ValidationFails(t *testing.T) {
	content := `
tenant:
  name: ""
  slug: "test"
  plan: "free"

product:
  name: "App"
  slug: "app"

environment:
  name: "Dev"
  tier: "development"

components:
  - name: "API"
    kind: "api"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid-config.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	loader := app.NewConfigLoader(filePath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConfigLoader_LoadFromString(t *testing.T) {
	content := `
tenant:
  name: "Inline Corp"
  slug: "inline"
  plan: "enterprise"

product:
  name: "Service"
  slug: "service"

environment:
  name: "Prod"
  tier: "production"

components:
  - name: "DB"
    kind: "database"
`
	loader := app.NewConfigLoader("")
	cfg, err := loader.LoadFromString(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Tenant.Plan != "enterprise" {
		t.Errorf("expected enterprise plan, got %s", cfg.Tenant.Plan)
	}
}

func TestGenerateTemplate(t *testing.T) {
	tmpl := app.GenerateTemplate()
	if tmpl == "" {
		t.Fatal("expected non-empty template")
	}
	if len(tmpl) < 100 {
		t.Error("template seems too short")
	}
}
