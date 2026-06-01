package notificationtemplate

import (
	"testing"
)

func TestTemplateEngine_RenderSimple(t *testing.T) {
	engine := NewTemplateEngine()
	result, err := engine.Render("Hello {{.Name}}", map[string]interface{}{
		"Name": "World",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", result)
	}
}

func TestTemplateEngine_RenderMultipleVars(t *testing.T) {
	engine := NewTemplateEngine()
	tmpl := "🚨 *{{.Severity}}*: {{.Title}} on component {{.Component}}"
	result, err := engine.Render(tmpl, map[string]interface{}{
		"Severity":  "CRITICAL",
		"Title":     "Database connection pool exhausted",
		"Component": "postgres-primary",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "🚨 *CRITICAL*: Database connection pool exhausted on component postgres-primary"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestTemplateEngine_RenderInvalidTemplate(t *testing.T) {
	engine := NewTemplateEngine()
	_, err := engine.Render("Hello {{.Name", nil)
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestTemplateEngine_RenderMissingVar(t *testing.T) {
	engine := NewTemplateEngine()
	result, err := engine.Render("Hello {{.Name}}", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Go templates render missing keys as "<no value>"
	if result != "Hello <no value>" {
		t.Fatalf("expected 'Hello <no value>', got %q", result)
	}
}
