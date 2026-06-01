package validation

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaultOutputValidator_ValidOutput(t *testing.T) {
	v := NewDefaultOutputValidator()

	output := map[string]interface{}{
		"root_cause":          "Database connection pool exhausted due to slow query execution",
		"confidence":          0.85,
		"affected_components": []string{"postgres-primary"},
		"investigation_hints": []string{"Check pg_stat_activity for long-running queries"},
	}
	data, _ := json.Marshal(output)

	if err := v.Validate(data); err != nil {
		t.Fatalf("expected valid output, got error: %v", err)
	}
}

func TestDefaultOutputValidator_EmptyOutput(t *testing.T) {
	v := NewDefaultOutputValidator()

	if err := v.Validate([]byte{}); err == nil {
		t.Fatal("expected error for empty output")
	}
}

func TestDefaultOutputValidator_InvalidJSON(t *testing.T) {
	v := NewDefaultOutputValidator()

	if err := v.Validate([]byte("not json")); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDefaultOutputValidator_MissingRequiredField(t *testing.T) {
	v := NewDefaultOutputValidator()

	output := map[string]interface{}{
		"confidence": 0.5,
		// missing "root_cause"
	}
	data, _ := json.Marshal(output)

	err := v.Validate(data)
	if err == nil {
		t.Fatal("expected error for missing required field")
	}
	if !strings.Contains(err.Error(), "root_cause") {
		t.Fatalf("expected error about root_cause, got: %v", err)
	}
}

func TestDefaultOutputValidator_ConfidenceOutOfBounds(t *testing.T) {
	v := NewDefaultOutputValidator()

	tests := []struct {
		name       string
		confidence float64
		wantErr    bool
	}{
		{"negative", -0.1, true},
		{"above_one", 1.5, true},
		{"zero", 0.0, false},
		{"one", 1.0, false},
		{"valid", 0.75, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := map[string]interface{}{
				"root_cause": "Some root cause with enough detail here",
				"confidence": tt.confidence,
			}
			data, _ := json.Marshal(output)
			err := v.Validate(data)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for confidence %f", tt.confidence)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for confidence %f: %v", tt.confidence, err)
			}
		})
	}
}

func TestDefaultOutputValidator_ForbiddenPatterns(t *testing.T) {
	v := NewDefaultOutputValidator()

	dangerous := []string{
		"DROP DATABASE",
		"rm -rf",
		"<script>",
		"eval(",
	}

	for _, pattern := range dangerous {
		t.Run(pattern, func(t *testing.T) {
			output := map[string]interface{}{
				"root_cause": "Recommend: " + pattern + " to fix the issue",
				"confidence": 0.5,
			}
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(false)
			_ = enc.Encode(output)
			err := v.Validate(buf.Bytes())
			if err == nil {
				t.Fatalf("expected error for forbidden pattern: %s", pattern)
			}
			if !strings.Contains(err.Error(), "forbidden pattern") {
				t.Fatalf("expected forbidden pattern error, got: %v", err)
			}
		})
	}
}

func TestDefaultOutputValidator_HallucinationDetection(t *testing.T) {
	v := NewDefaultOutputValidator()

	// High confidence with very short root cause = suspicious
	output := map[string]interface{}{
		"root_cause": "unknown",
		"confidence": 0.95,
	}
	data, _ := json.Marshal(output)

	err := v.Validate(data)
	if err == nil {
		t.Fatal("expected hallucination detection error")
	}
	if !strings.Contains(err.Error(), "hallucination") {
		t.Fatalf("expected hallucination error, got: %v", err)
	}
}

func TestDefaultOutputValidator_OversizedOutput(t *testing.T) {
	v := NewDefaultOutputValidator()

	// Create output larger than 50KB
	huge := strings.Repeat("x", 60000)
	output := map[string]interface{}{
		"root_cause": huge,
		"confidence": 0.5,
	}
	data, _ := json.Marshal(output)

	err := v.Validate(data)
	if err == nil {
		t.Fatal("expected error for oversized output")
	}
}
