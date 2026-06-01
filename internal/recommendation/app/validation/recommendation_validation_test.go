package validation

import (
	"strings"
	"testing"
)

func TestDefaultRecommendationValidator_ValidRecommendation(t *testing.T) {
	v := NewDefaultRecommendationValidator()

	err := v.Validate(
		"Increase connection pool size from 20 to 50 to handle peak load",
		[]string{"Connection pool utilization at 95% during peak hours (14:00-16:00)"},
	)
	if err != nil {
		t.Fatalf("expected valid recommendation, got error: %v", err)
	}
}

func TestDefaultRecommendationValidator_EmptyRecommendation(t *testing.T) {
	v := NewDefaultRecommendationValidator()

	err := v.Validate("", []string{"some evidence"})
	if err == nil {
		t.Fatal("expected error for empty recommendation")
	}
}

func TestDefaultRecommendationValidator_NoEvidence(t *testing.T) {
	v := NewDefaultRecommendationValidator()

	err := v.Validate("Increase pool size", []string{})
	if err == nil {
		t.Fatal("expected error for no evidence")
	}
	if !strings.Contains(err.Error(), "evidence") {
		t.Fatalf("expected evidence error, got: %v", err)
	}
}

func TestDefaultRecommendationValidator_TrivialEvidence(t *testing.T) {
	v := NewDefaultRecommendationValidator()

	err := v.Validate("Increase pool size", []string{"yes"})
	if err == nil {
		t.Fatal("expected error for trivial evidence")
	}
}

func TestDefaultRecommendationValidator_DangerousPatterns(t *testing.T) {
	v := NewDefaultRecommendationValidator()

	dangerous := []struct {
		rec     string
		pattern string
	}{
		{"Run DROP DATABASE production to reset", "DROP DATABASE"},
		{"Execute rm -rf /tmp to clean old files", "RM -RF"},
		{"Set chmod 777 on all files for access", "CHMOD 777"},
		{"DISABLE SECURITY features to improve performance", "DISABLE SECURITY"},
		{"Use GRANT ALL PRIVILEGES to fix permissions", "GRANT ALL PRIVILEGES"},
		{"DISABLE SSL to reduce latency", "DISABLE SSL"},
	}

	for _, tt := range dangerous {
		t.Run(tt.pattern, func(t *testing.T) {
			err := v.Validate(tt.rec, []string{"System is slow, need improvement action taken"})
			if err == nil {
				t.Fatalf("expected error for dangerous pattern: %s", tt.pattern)
			}
			if !strings.Contains(err.Error(), "dangerous") {
				t.Fatalf("expected dangerous operation error, got: %v", err)
			}
		})
	}
}

func TestDefaultRecommendationValidator_MaxLength(t *testing.T) {
	v := NewDefaultRecommendationValidator()

	longRec := strings.Repeat("a", 2001)
	err := v.Validate(longRec, []string{"Valid evidence with enough content"})
	if err == nil {
		t.Fatal("expected error for oversized recommendation")
	}
	if !strings.Contains(err.Error(), "maximum length") {
		t.Fatalf("expected max length error, got: %v", err)
	}
}

func TestDefaultRecommendationValidator_HallucinationDetection(t *testing.T) {
	v := NewDefaultRecommendationValidator()

	// Reference a CVE that doesn't appear in evidence
	err := v.Validate(
		"Patch CVE-2024-12345 to fix the vulnerability",
		[]string{"Server response time is elevated above normal"},
	)
	if err == nil {
		t.Fatal("expected hallucination detection error")
	}
	if !strings.Contains(err.Error(), "CVE-") {
		t.Fatalf("expected error about CVE reference, got: %v", err)
	}
}

func TestDefaultRecommendationValidator_ValidCVEReference(t *testing.T) {
	v := NewDefaultRecommendationValidator()

	// CVE appears in both recommendation and evidence
	err := v.Validate(
		"Patch CVE-2024-12345 to fix the vulnerability",
		[]string{"Vulnerability scan detected CVE-2024-12345 in openssl package"},
	)
	if err != nil {
		t.Fatalf("expected valid recommendation with CVE in evidence, got: %v", err)
	}
}
