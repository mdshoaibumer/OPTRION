package domain_test

import (
	"testing"

	"github.com/optrion/optrion/internal/incident/domain"
)

func TestIncidentStateMachine(t *testing.T) {
	tests := []struct {
		name   string
		from   domain.IncidentStatus
		to     domain.IncidentStatus
		expect bool
	}{
		{"open→acknowledged", domain.IncidentStatusOpen, domain.IncidentStatusAcknowledged, true},
		{"open→investigating", domain.IncidentStatusOpen, domain.IncidentStatusInvestigating, true},
		{"open→resolved", domain.IncidentStatusOpen, domain.IncidentStatusResolved, true},
		{"open→closed", domain.IncidentStatusOpen, domain.IncidentStatusClosed, true},
		{"acknowledged→investigating", domain.IncidentStatusAcknowledged, domain.IncidentStatusInvestigating, true},
		{"acknowledged→resolved", domain.IncidentStatusAcknowledged, domain.IncidentStatusResolved, true},
		{"acknowledged→closed", domain.IncidentStatusAcknowledged, domain.IncidentStatusClosed, true},
		{"investigating→resolved", domain.IncidentStatusInvestigating, domain.IncidentStatusResolved, true},
		{"investigating→closed", domain.IncidentStatusInvestigating, domain.IncidentStatusClosed, true},
		{"resolved→closed", domain.IncidentStatusResolved, domain.IncidentStatusClosed, true},
		{"resolved→open (reopen)", domain.IncidentStatusResolved, domain.IncidentStatusOpen, true},
		// Invalid transitions
		{"closed→open", domain.IncidentStatusClosed, domain.IncidentStatusOpen, false},
		{"closed→acknowledged", domain.IncidentStatusClosed, domain.IncidentStatusAcknowledged, false},
		{"resolved→investigating", domain.IncidentStatusResolved, domain.IncidentStatusInvestigating, false},
		{"acknowledged→open", domain.IncidentStatusAcknowledged, domain.IncidentStatusOpen, false},
		{"investigating→acknowledged", domain.IncidentStatusInvestigating, domain.IncidentStatusAcknowledged, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.CanTransitionTo(tt.from, tt.to)
			if got != tt.expect {
				t.Errorf("CanTransitionTo(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.expect)
			}
		})
	}
}

func TestNewIncident(t *testing.T) {
	inc, err := domain.NewIncident(
		"01961234-5678-7000-8000-000000000001",
		"01961234-5678-7000-8000-000000000002",
		"01961234-5678-7000-8000-000000000003",
		"Database connection pool exhausted",
		"Connection pool at 100% utilization",
		domain.SeverityCritical,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inc.Status != domain.IncidentStatusOpen {
		t.Errorf("expected status open, got %s", inc.Status)
	}
	if inc.Severity != domain.SeverityCritical {
		t.Errorf("expected severity critical, got %s", inc.Severity)
	}
	if inc.Version != 1 {
		t.Errorf("expected version 1, got %d", inc.Version)
	}
	if len(inc.UncommittedEvents()) != 1 {
		t.Errorf("expected 1 uncommitted event, got %d", len(inc.UncommittedEvents()))
	}
	if inc.UncommittedEvents()[0].EventType != domain.EventIncidentOpened {
		t.Errorf("expected event type %s, got %s", domain.EventIncidentOpened, inc.UncommittedEvents()[0].EventType)
	}
}

func TestNewIncident_Validation(t *testing.T) {
	tests := []struct {
		name        string
		tenantID    string
		componentID string
		title       string
		severity    domain.IncidentSeverity
	}{
		{"empty tenant", "", "01961234-5678-7000-8000-000000000002", "title", domain.SeverityMajor},
		{"empty component", "01961234-5678-7000-8000-000000000001", "", "title", domain.SeverityMajor},
		{"empty title", "01961234-5678-7000-8000-000000000001", "01961234-5678-7000-8000-000000000002", "", domain.SeverityMajor},
		{"invalid severity", "01961234-5678-7000-8000-000000000001", "01961234-5678-7000-8000-000000000002", "title", domain.IncidentSeverity("invalid")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewIncident(tt.tenantID, tt.componentID, "", tt.title, "desc", tt.severity)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestIncidentLifecycle_FullFlow(t *testing.T) {
	inc, err := domain.NewIncident(
		"01961234-5678-7000-8000-000000000001",
		"01961234-5678-7000-8000-000000000002",
		"",
		"High CPU usage",
		"CPU at 95%",
		domain.SeverityMajor,
	)
	if err != nil {
		t.Fatalf("NewIncident: %v", err)
	}
	inc.ClearUncommittedEvents()

	// Acknowledge
	if err := inc.Acknowledge("user-1"); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	if inc.Status != domain.IncidentStatusAcknowledged {
		t.Errorf("expected acknowledged, got %s", inc.Status)
	}
	if inc.AcknowledgedAt == nil {
		t.Error("expected AcknowledgedAt to be set")
	}
	if inc.Version != 2 {
		t.Errorf("expected version 2, got %d", inc.Version)
	}

	// Investigate
	if err := inc.Investigate("user-1"); err != nil {
		t.Fatalf("Investigate: %v", err)
	}
	if inc.Status != domain.IncidentStatusInvestigating {
		t.Errorf("expected investigating, got %s", inc.Status)
	}

	// Resolve
	if err := inc.Resolve("user-1", "Scaled up instances"); err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if inc.Status != domain.IncidentStatusResolved {
		t.Errorf("expected resolved, got %s", inc.Status)
	}
	if inc.ResolvedAt == nil {
		t.Error("expected ResolvedAt to be set")
	}

	// Close
	if err := inc.Close("user-1", "RCA complete"); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if inc.Status != domain.IncidentStatusClosed {
		t.Errorf("expected closed, got %s", inc.Status)
	}
	if inc.ClosedAt == nil {
		t.Error("expected ClosedAt to be set")
	}

	// Verify events were raised
	events := inc.UncommittedEvents()
	if len(events) != 4 {
		t.Errorf("expected 4 events, got %d", len(events))
	}
}

func TestIncidentInvalidTransition(t *testing.T) {
	inc, _ := domain.NewIncident(
		"01961234-5678-7000-8000-000000000001",
		"01961234-5678-7000-8000-000000000002",
		"",
		"Test",
		"desc",
		domain.SeverityMinor,
	)

	// Close immediately (valid from open)
	_ = inc.Close("user-1", "false alarm")

	// Try to acknowledge after close (should fail)
	err := inc.Acknowledge("user-1")
	if err == nil {
		t.Error("expected error for invalid transition from closed→acknowledged")
	}
	if _, ok := err.(domain.ErrInvalidTransition); !ok {
		t.Errorf("expected ErrInvalidTransition, got %T: %v", err, err)
	}
}

func TestSeverityNumericLevel(t *testing.T) {
	if domain.SeverityCritical.NumericLevel() <= domain.SeverityMajor.NumericLevel() {
		t.Error("critical should have higher level than major")
	}
	if domain.SeverityMajor.NumericLevel() <= domain.SeverityMinor.NumericLevel() {
		t.Error("major should have higher level than minor")
	}
	if domain.SeverityMinor.NumericLevel() <= domain.SeverityWarning.NumericLevel() {
		t.Error("minor should have higher level than warning")
	}
	if domain.SeverityWarning.NumericLevel() <= domain.SeverityInfo.NumericLevel() {
		t.Error("warning should have higher level than info")
	}
}

func TestRuleConditionMatches(t *testing.T) {
	tests := []struct {
		name     string
		cond     domain.RuleCondition
		value    float64
		expected bool
	}{
		{"gt matches", domain.RuleCondition{Operator: domain.OperatorGT, Threshold: 90}, 95, true},
		{"gt no match", domain.RuleCondition{Operator: domain.OperatorGT, Threshold: 90}, 85, false},
		{"lt matches", domain.RuleCondition{Operator: domain.OperatorLT, Threshold: 30}, 25, true},
		{"lt no match", domain.RuleCondition{Operator: domain.OperatorLT, Threshold: 30}, 35, false},
		{"gte boundary", domain.RuleCondition{Operator: domain.OperatorGTE, Threshold: 90}, 90, true},
		{"lte boundary", domain.RuleCondition{Operator: domain.OperatorLTE, Threshold: 30}, 30, true},
		{"eq matches", domain.RuleCondition{Operator: domain.OperatorEQ, Threshold: 42}, 42, true},
		{"eq no match", domain.RuleCondition{Operator: domain.OperatorEQ, Threshold: 42}, 43, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cond.Matches(tt.value)
			if got != tt.expected {
				t.Errorf("Matches(%v) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestIncidentChangeSeverity(t *testing.T) {
	inc, _ := domain.NewIncident(
		"01961234-5678-7000-8000-000000000001",
		"01961234-5678-7000-8000-000000000002",
		"",
		"Test",
		"desc",
		domain.SeverityWarning,
	)
	inc.ClearUncommittedEvents()

	inc.ChangeSeverity(domain.SeverityCritical)
	if inc.Severity != domain.SeverityCritical {
		t.Errorf("expected critical, got %s", inc.Severity)
	}
	events := inc.UncommittedEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != domain.EventSeverityChanged {
		t.Errorf("expected severity_changed event, got %s", events[0].EventType)
	}
}

func TestIncidentChangeSeverity_NoOp(t *testing.T) {
	inc, _ := domain.NewIncident(
		"01961234-5678-7000-8000-000000000001",
		"01961234-5678-7000-8000-000000000002",
		"",
		"Test",
		"desc",
		domain.SeverityMajor,
	)
	inc.ClearUncommittedEvents()

	inc.ChangeSeverity(domain.SeverityMajor)
	if len(inc.UncommittedEvents()) != 0 {
		t.Error("expected no events when severity unchanged")
	}
}

func TestIncidentCorrelate(t *testing.T) {
	inc, _ := domain.NewIncident(
		"01961234-5678-7000-8000-000000000001",
		"01961234-5678-7000-8000-000000000002",
		"",
		"Test",
		"desc",
		domain.SeverityMinor,
	)
	inc.ClearUncommittedEvents()
	originalCorr := inc.CorrelationID

	newCorr := "01961234-5678-7000-8000-000000000099"
	inc.Correlate(newCorr)
	if inc.CorrelationID != newCorr {
		t.Errorf("expected correlation %s, got %s", newCorr, inc.CorrelationID)
	}

	events := inc.UncommittedEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != domain.EventCorrelated {
		t.Errorf("expected correlated event, got %s", events[0].EventType)
	}

	// Correlate again with same ID — should be no-op
	inc.ClearUncommittedEvents()
	inc.Correlate(newCorr)
	if len(inc.UncommittedEvents()) != 0 {
		t.Error("expected no events for same correlation ID")
	}

	_ = originalCorr
}

func TestRebuildFromEvents(t *testing.T) {
	inc, _ := domain.NewIncident(
		"01961234-5678-7000-8000-000000000001",
		"01961234-5678-7000-8000-000000000002",
		"01961234-5678-7000-8000-000000000003",
		"Test Incident",
		"desc",
		domain.SeverityMajor,
	)

	_ = inc.Acknowledge("user-1")
	_ = inc.Investigate("user-2")
	_ = inc.Resolve("user-2", "fixed")

	events := inc.UncommittedEvents()

	rebuilt := domain.RebuildFromEvents(events)
	if rebuilt == nil {
		t.Fatal("expected non-nil incident from events")
	}
	if rebuilt.Status != domain.IncidentStatusResolved {
		t.Errorf("expected resolved, got %s", rebuilt.Status)
	}
}

func TestRebuildFromEvents_Empty(t *testing.T) {
	rebuilt := domain.RebuildFromEvents(nil)
	if rebuilt != nil {
		t.Error("expected nil for empty events")
	}
}

func TestNewIncidentComment(t *testing.T) {
	c, err := domain.NewIncidentComment(
		"01961234-5678-7000-8000-000000000001",
		"01961234-5678-7000-8000-000000000002",
		"user-1",
		"This looks like a network issue",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Content != "This looks like a network issue" {
		t.Errorf("unexpected content: %s", c.Content)
	}
}

func TestNewIncidentComment_EmptyContent(t *testing.T) {
	_, err := domain.NewIncidentComment(
		"01961234-5678-7000-8000-000000000001",
		"01961234-5678-7000-8000-000000000002",
		"user-1",
		"",
	)
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestStatusIsActive(t *testing.T) {
	active := []domain.IncidentStatus{
		domain.IncidentStatusOpen,
		domain.IncidentStatusAcknowledged,
		domain.IncidentStatusInvestigating,
	}
	for _, s := range active {
		if !s.IsActive() {
			t.Errorf("expected %s to be active", s)
		}
	}

	inactive := []domain.IncidentStatus{
		domain.IncidentStatusResolved,
		domain.IncidentStatusClosed,
	}
	for _, s := range inactive {
		if s.IsActive() {
			t.Errorf("expected %s to be inactive", s)
		}
	}
}
