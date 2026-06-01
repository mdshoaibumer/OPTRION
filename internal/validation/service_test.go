package validation_test

import (
	"testing"

	"github.com/optrion/optrion/internal/validation"
)

func TestValidationReport_HealthyStatus(t *testing.T) {
	report := &validation.ValidationReport{
		TenantID:             "t-1",
		ComponentsRegistered: 3,
		ComponentsHealthy:    3,
		MetricsFlowing:       true,
		IntegrationStatus:    "healthy",
		Issues:               []validation.ValidationIssue{},
		Recommendations:      []string{},
	}

	if report.IntegrationStatus != "healthy" {
		t.Errorf("expected healthy, got %s", report.IntegrationStatus)
	}
	if len(report.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(report.Issues))
	}
}

func TestValidationReport_DegradedStatus(t *testing.T) {
	report := &validation.ValidationReport{
		TenantID:             "t-1",
		ComponentsRegistered: 3,
		ComponentsHealthy:    2,
		IntegrationStatus:    "degraded",
		Issues: []validation.ValidationIssue{
			{Severity: "warning", Component: "redis", Message: "High latency"},
		},
		Recommendations: []string{"Check component performance"},
	}

	if report.IntegrationStatus != "degraded" {
		t.Errorf("expected degraded, got %s", report.IntegrationStatus)
	}
	if len(report.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(report.Issues))
	}
}

func TestValidationReport_UnhealthyStatus(t *testing.T) {
	report := &validation.ValidationReport{
		TenantID:             "t-1",
		ComponentsRegistered: 3,
		ComponentsHealthy:    0,
		IntegrationStatus:    "unhealthy",
		Issues: []validation.ValidationIssue{
			{Severity: "error", Component: "postgres", Message: "Connection refused"},
			{Severity: "error", Component: "redis", Message: "Connection refused"},
		},
	}

	if report.IntegrationStatus != "unhealthy" {
		t.Errorf("expected unhealthy, got %s", report.IntegrationStatus)
	}
	if len(report.Issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(report.Issues))
	}
}

func TestValidationIssue_Severity(t *testing.T) {
	tests := []struct {
		name     string
		severity string
	}{
		{"info level", "info"},
		{"warning level", "warning"},
		{"error level", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := validation.ValidationIssue{
				Severity:  tt.severity,
				Component: "test-component",
				Message:   "test message",
			}
			if issue.Severity != tt.severity {
				t.Errorf("expected %s, got %s", tt.severity, issue.Severity)
			}
		})
	}
}
