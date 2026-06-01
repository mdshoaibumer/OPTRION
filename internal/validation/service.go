package validation

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/optrion/optrion/internal/tenant/domain"
	"github.com/optrion/optrion/internal/tenant/port"
)

// ValidationService validates OPTRION integration.
type ValidationService struct {
	components port.ComponentRepository
	health     port.HealthCheckRepository
	logger     *slog.Logger
}

// NewValidationService creates a new ValidationService.
func NewValidationService(
	components port.ComponentRepository,
	health port.HealthCheckRepository,
	logger *slog.Logger,
) *ValidationService {
	return &ValidationService{
		components: components,
		health:     health,
		logger:     logger,
	}
}

// ValidationReport represents the results of an integration validation.
type ValidationReport struct {
	TenantID              string
	RegistrationTime      time.Time
	ComponentsRegistered  int
	ComponentsHealthy     int
	MetricsFlowing        bool
	LastMetricsReceived   time.Time
	AverageResponseTime   time.Duration
	HealthChecksPassed    int
	HealthChecksFailed    int
	IntegrationStatus     string
	Issues                []ValidationIssue
	Recommendations       []string
}

// ValidationIssue represents a validation issue found during verification.
type ValidationIssue struct {
	Severity string // "info", "warning", "error"
	Component string
	Message  string
	Details  string
}

// ValidateIntegration performs a comprehensive integration validation.
func (vs *ValidationService) ValidateIntegration(ctx context.Context, tenantID string) (*ValidationReport, error) {
	report := &ValidationReport{
		TenantID:         tenantID,
		RegistrationTime: time.Now().UTC(),
		Issues:           make([]ValidationIssue, 0),
		Recommendations:  make([]string, 0),
	}

	// Get all components for tenant
	comps, err := vs.components.ListByTenant(ctx, tenantID)
	if err != nil {
		vs.logger.ErrorContext(ctx, "failed to list components", "error", err)
		return report, fmt.Errorf("failed to list components: %w", err)
	}

	report.ComponentsRegistered = len(comps)

	// Check each component
	for _, comp := range comps {
		// Check if component is healthy
		healthy, responseTime, err := vs.checkComponentHealth(ctx, comp)
		if err != nil {
			report.Issues = append(report.Issues, ValidationIssue{
				Severity:  "warning",
				Component: comp.Name,
				Message:   fmt.Sprintf("Failed to check health: %v", err),
			})
		} else if healthy {
			report.ComponentsHealthy++
			report.AverageResponseTime += responseTime
		} else {
			report.Issues = append(report.Issues, ValidationIssue{
				Severity:  "error",
				Component: comp.Name,
				Message:   "Component health check failed",
				Details:   fmt.Sprintf("Response time: %v", responseTime),
			})
		}

		report.HealthChecksPassed++
	}

	if report.ComponentsRegistered > 0 {
		report.AverageResponseTime /= time.Duration(report.ComponentsRegistered)
	}

	// Check if metrics are flowing
	metricsOK, lastReceived, err := vs.checkMetricsFlow(ctx, tenantID)
	if err == nil {
		report.MetricsFlowing = metricsOK
		report.LastMetricsReceived = lastReceived
	} else {
		report.Issues = append(report.Issues, ValidationIssue{
			Severity: "warning",
			Message:  "Failed to check metrics flow",
			Details:  err.Error(),
		})
	}

	// Determine overall status
	if len(report.Issues) == 0 {
		report.IntegrationStatus = "healthy"
	} else {
		hasErrors := false
		for _, issue := range report.Issues {
			if issue.Severity == "error" {
				hasErrors = true
				break
			}
		}
		if hasErrors {
			report.IntegrationStatus = "unhealthy"
		} else {
			report.IntegrationStatus = "degraded"
		}
	}

	// Generate recommendations
	if report.ComponentsHealthy < report.ComponentsRegistered {
		report.Recommendations = append(report.Recommendations,
			"Some components are not healthy. Check component endpoints and network connectivity.")
	}

	if !report.MetricsFlowing {
		report.Recommendations = append(report.Recommendations,
			"Metrics are not flowing. Verify SDK is running and API key is correct.")
	}

	if report.AverageResponseTime > 5*time.Second {
		report.Recommendations = append(report.Recommendations,
			"Average response time is high. Check component performance and network latency.")
	}

	vs.logger.InfoContext(ctx, "integration validation completed",
		"tenant_id", tenantID,
		"status", report.IntegrationStatus,
		"components_healthy", report.ComponentsHealthy,
		"components_total", report.ComponentsRegistered,
	)

	return report, nil
}

// checkComponentHealth performs a health check on a component.
func (vs *ValidationService) checkComponentHealth(ctx context.Context, comp *domain.Component) (bool, time.Duration, error) {
	// TODO: Implement actual health check based on component kind and endpoint
	// For now, return a placeholder
	return true, 100 * time.Millisecond, nil
}

// checkMetricsFlow checks if metrics are being received from the application.
func (vs *ValidationService) checkMetricsFlow(ctx context.Context, tenantID string) (bool, time.Time, error) {
	// TODO: Query metrics store to verify recent metrics from tenant
	// Return (metricsFlowing, lastReceived, error)
	return true, time.Now().UTC(), nil
}

// ValidateComponentConnectivity checks if a component endpoint is reachable.
func (vs *ValidationService) ValidateComponentConnectivity(ctx context.Context, endpoint string, port int) (bool, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://%s:%d/health", endpoint, port)

	resp, err := client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GenerateValidationReport creates a human-readable validation report.
func (vs *ValidationService) GenerateValidationReport(report *ValidationReport) string {
	output := fmt.Sprintf(`
OPTRION Integration Validation Report
======================================

Tenant ID: %s
Validation Time: %s
Status: %s

Component Health:
  - Registered: %d
  - Healthy: %d
  - Unhealthy: %d
  - Average Response Time: %v

Metrics Status:
  - Flowing: %v
  - Last Received: %s

Issues:
`, report.TenantID, report.RegistrationTime.Format(time.RFC3339), report.IntegrationStatus,
		report.ComponentsRegistered, report.ComponentsHealthy,
		report.ComponentsRegistered-report.ComponentsHealthy, report.AverageResponseTime,
		report.MetricsFlowing, report.LastMetricsReceived.Format(time.RFC3339))

	if len(report.Issues) == 0 {
		output += "  ✓ No issues detected\n"
	} else {
		for _, issue := range report.Issues {
			symbol := "ℹ"
			if issue.Severity == "error" {
				symbol = "✗"
			} else if issue.Severity == "warning" {
				symbol = "⚠"
			}
			output += fmt.Sprintf("  %s [%s] %s: %s\n", symbol, issue.Severity, issue.Component, issue.Message)
			if issue.Details != "" {
				output += fmt.Sprintf("      Details: %s\n", issue.Details)
			}
		}
	}

	if len(report.Recommendations) > 0 {
		output += "\nRecommendations:\n"
		for i, rec := range report.Recommendations {
			output += fmt.Sprintf("  %d. %s\n", i+1, rec)
		}
	}

	return output
}
