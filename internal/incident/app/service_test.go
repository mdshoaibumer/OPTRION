package app_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/incident/app"
	"github.com/optrion/optrion/internal/incident/domain"
	"github.com/optrion/optrion/internal/incident/port"
	"github.com/optrion/optrion/internal/shared/id"
)

// --- In-memory test doubles ---

type memIncidentRepo struct {
	incidents []*domain.Incident
}

func (m *memIncidentRepo) Create(_ context.Context, inc *domain.Incident) error {
	m.incidents = append(m.incidents, inc)
	return nil
}

func (m *memIncidentRepo) GetByID(_ context.Context, incID string) (*domain.Incident, error) {
	for _, inc := range m.incidents {
		if inc.ID == incID {
			return inc, nil
		}
	}
	return nil, nil
}

func (m *memIncidentRepo) Update(_ context.Context, inc *domain.Incident) error {
	for i, existing := range m.incidents {
		if existing.ID == inc.ID {
			m.incidents[i] = inc
			return nil
		}
	}
	return nil
}

func (m *memIncidentRepo) ListByTenant(_ context.Context, tenantID string, filter port.IncidentFilter) ([]*domain.Incident, error) {
	var result []*domain.Incident
	for _, inc := range m.incidents {
		if inc.TenantID != tenantID {
			continue
		}
		if filter.Status != nil && inc.Status != *filter.Status {
			continue
		}
		if filter.Severity != nil && inc.Severity != *filter.Severity {
			continue
		}
		if filter.ComponentID != nil && inc.ComponentID != *filter.ComponentID {
			continue
		}
		result = append(result, inc)
	}
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (m *memIncidentRepo) CountByTenant(_ context.Context, tenantID string, filter port.IncidentFilter) (int, error) {
	// Count without applying limit/offset
	countFilter := filter
	countFilter.Limit = 0
	countFilter.Offset = 0
	list, _ := m.ListByTenant(context.Background(), tenantID, countFilter)
	return len(list), nil
}

func (m *memIncidentRepo) FindActiveByRuleAndComponent(_ context.Context, ruleID, componentID string) (*domain.Incident, error) {
	for _, inc := range m.incidents {
		if inc.RuleID == ruleID && inc.ComponentID == componentID && inc.Status.IsActive() {
			return inc, nil
		}
	}
	return nil, nil
}

func (m *memIncidentRepo) FindByCorrelation(_ context.Context, corrID string) ([]*domain.Incident, error) {
	var result []*domain.Incident
	for _, inc := range m.incidents {
		if inc.CorrelationID == corrID {
			result = append(result, inc)
		}
	}
	return result, nil
}

type memEventRepo struct {
	events []domain.IncidentEvent
}

func (m *memEventRepo) Append(_ context.Context, events []domain.IncidentEvent) error {
	m.events = append(m.events, events...)
	return nil
}

func (m *memEventRepo) ListByIncident(_ context.Context, incID string) ([]domain.IncidentEvent, error) {
	var result []domain.IncidentEvent
	for _, e := range m.events {
		if e.IncidentID == incID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *memEventRepo) ListByTenant(_ context.Context, tenantID string, from, to time.Time, limit int) ([]domain.IncidentEvent, error) {
	var result []domain.IncidentEvent
	for _, e := range m.events {
		if e.TenantID == tenantID && !e.OccurredAt.Before(from) && !e.OccurredAt.After(to) {
			result = append(result, e)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

type memRuleRepo struct {
	rules   []*domain.IncidentRule
	firings map[string]time.Time
}

func newMemRuleRepo() *memRuleRepo {
	return &memRuleRepo{firings: make(map[string]time.Time)}
}

func (m *memRuleRepo) Create(_ context.Context, rule *domain.IncidentRule) error {
	m.rules = append(m.rules, rule)
	return nil
}

func (m *memRuleRepo) GetByID(_ context.Context, ruleID string) (*domain.IncidentRule, error) {
	for _, r := range m.rules {
		if r.ID == ruleID {
			return r, nil
		}
	}
	return nil, nil
}

func (m *memRuleRepo) Update(_ context.Context, rule *domain.IncidentRule) error {
	for i, r := range m.rules {
		if r.ID == rule.ID {
			m.rules[i] = rule
			return nil
		}
	}
	return nil
}

func (m *memRuleRepo) ListEnabled(_ context.Context, tenantID string) ([]*domain.IncidentRule, error) {
	var result []*domain.IncidentRule
	for _, r := range m.rules {
		if r.TenantID == tenantID && r.Enabled {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *memRuleRepo) ListByTenant(_ context.Context, tenantID string) ([]*domain.IncidentRule, error) {
	var result []*domain.IncidentRule
	for _, r := range m.rules {
		if r.TenantID == tenantID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *memRuleRepo) GetLastFiredAt(_ context.Context, ruleID, componentID string) (*time.Time, error) {
	key := ruleID + ":" + componentID
	if t, ok := m.firings[key]; ok {
		return &t, nil
	}
	return nil, nil
}

func (m *memRuleRepo) RecordFired(_ context.Context, ruleID, componentID string, firedAt time.Time) error {
	m.firings[ruleID+":"+componentID] = firedAt
	return nil
}

type memCommentRepo struct {
	comments []*domain.IncidentComment
}

func (m *memCommentRepo) Create(_ context.Context, c *domain.IncidentComment) error {
	m.comments = append(m.comments, c)
	return nil
}

func (m *memCommentRepo) ListByIncident(_ context.Context, incID string) ([]*domain.IncidentComment, error) {
	var result []*domain.IncidentComment
	for _, c := range m.comments {
		if c.IncidentID == incID {
			result = append(result, c)
		}
	}
	return result, nil
}

type memTimelineRepo struct {
	entries []*domain.IncidentTimeline
}

func (m *memTimelineRepo) Create(_ context.Context, entry *domain.IncidentTimeline) error {
	if entry.ID == "" {
		entry.ID = id.New()
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *memTimelineRepo) ListByIncident(_ context.Context, incID string, limit int) ([]*domain.IncidentTimeline, error) {
	var result []*domain.IncidentTimeline
	for _, e := range m.entries {
		if e.IncidentID == incID {
			result = append(result, e)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *memTimelineRepo) ListByTenant(_ context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.IncidentTimeline, error) {
	var result []*domain.IncidentTimeline
	for _, e := range m.entries {
		if e.TenantID == tenantID && !e.OccurredAt.Before(from) && !e.OccurredAt.After(to) {
			result = append(result, e)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// --- Test Helpers ---

func newTestService() (*app.IncidentService, *memIncidentRepo, *memEventRepo, *memRuleRepo) {
	incRepo := &memIncidentRepo{}
	eventRepo := &memEventRepo{}
	ruleRepo := newMemRuleRepo()
	commentRepo := &memCommentRepo{}
	timelineRepo := &memTimelineRepo{}
	logger := slog.Default()

	svc := app.NewIncidentService(incRepo, eventRepo, ruleRepo, commentRepo, timelineRepo, logger)
	return svc, incRepo, eventRepo, ruleRepo
}

const testTenantID = "01961234-5678-7000-8000-000000000001"
const testComponentID = "01961234-5678-7000-8000-000000000002"

// --- Tests ---

func TestEvaluateMetric_CreatesIncident(t *testing.T) {
	svc, incRepo, eventRepo, ruleRepo := newTestService()

	rule, _ := domain.NewIncidentRule(
		testTenantID, "High CPU Rule", "CPU > 90%",
		"", "server", // applies to all components
		domain.RuleCondition{MetricType: "cpu_percent", Operator: domain.OperatorGT, Threshold: 90},
		domain.SeverityMajor, 5*time.Minute,
	)
	ruleRepo.rules = append(ruleRepo.rules, rule)

	input := app.MetricInput{
		TenantID:      testTenantID,
		ComponentID:   testComponentID,
		CollectorType: "server",
		MetricType:    "cpu_percent",
		Value:         95.5,
		HealthScore:   0,
	}

	svc.EvaluateMetric(context.Background(), input)

	if len(incRepo.incidents) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(incRepo.incidents))
	}

	inc := incRepo.incidents[0]
	if inc.Status != domain.IncidentStatusOpen {
		t.Errorf("expected open, got %s", inc.Status)
	}
	if inc.Severity != domain.SeverityMajor {
		t.Errorf("expected major severity, got %s", inc.Severity)
	}
	if inc.ComponentID != testComponentID {
		t.Errorf("expected component %s, got %s", testComponentID, inc.ComponentID)
	}
	if inc.RuleID != rule.ID {
		t.Errorf("expected rule ID %s, got %s", rule.ID, inc.RuleID)
	}
	if len(eventRepo.events) == 0 {
		t.Error("expected events to be persisted")
	}
}

func TestEvaluateMetric_Deduplication(t *testing.T) {
	svc, incRepo, _, ruleRepo := newTestService()

	rule, _ := domain.NewIncidentRule(
		testTenantID, "High CPU Rule", "CPU > 90%",
		"", "server",
		domain.RuleCondition{MetricType: "cpu_percent", Operator: domain.OperatorGT, Threshold: 90},
		domain.SeverityMajor, 0, // No cooldown
	)
	ruleRepo.rules = append(ruleRepo.rules, rule)

	input := app.MetricInput{
		TenantID:      testTenantID,
		ComponentID:   testComponentID,
		CollectorType: "server",
		MetricType:    "cpu_percent",
		Value:         95.5,
	}

	// First evaluation creates incident
	svc.EvaluateMetric(context.Background(), input)
	// Second evaluation should NOT create a duplicate
	svc.EvaluateMetric(context.Background(), input)

	if len(incRepo.incidents) != 1 {
		t.Errorf("expected 1 incident (dedup), got %d", len(incRepo.incidents))
	}
}

func TestEvaluateMetric_SeverityFromHealthScore(t *testing.T) {
	svc, incRepo, _, ruleRepo := newTestService()

	rule, _ := domain.NewIncidentRule(
		testTenantID, "Health Score Check", "",
		"", "server",
		domain.RuleCondition{MetricType: "cpu_percent", Operator: domain.OperatorGT, Threshold: 80},
		domain.SeverityWarning, 0,
	)
	ruleRepo.rules = append(ruleRepo.rules, rule)

	input := app.MetricInput{
		TenantID:      testTenantID,
		ComponentID:   testComponentID,
		CollectorType: "server",
		MetricType:    "cpu_percent",
		Value:         95.5,
		HealthScore:   25, // < 30 → critical
	}

	svc.EvaluateMetric(context.Background(), input)

	if len(incRepo.incidents) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(incRepo.incidents))
	}
	if incRepo.incidents[0].Severity != domain.SeverityCritical {
		t.Errorf("expected critical (health score 25), got %s", incRepo.incidents[0].Severity)
	}
}

func TestEvaluateMetric_Cooldown(t *testing.T) {
	svc, incRepo, _, ruleRepo := newTestService()

	rule, _ := domain.NewIncidentRule(
		testTenantID, "CPU Rule", "",
		"", "server",
		domain.RuleCondition{MetricType: "cpu_percent", Operator: domain.OperatorGT, Threshold: 90},
		domain.SeverityMajor, 1*time.Hour, // 1 hour cooldown
	)
	ruleRepo.rules = append(ruleRepo.rules, rule)

	// Record a recent firing
	ruleRepo.firings[rule.ID+":"+testComponentID] = time.Now().Add(-5 * time.Minute) // 5 min ago

	input := app.MetricInput{
		TenantID:      testTenantID,
		ComponentID:   testComponentID,
		CollectorType: "server",
		MetricType:    "cpu_percent",
		Value:         95.5,
	}

	svc.EvaluateMetric(context.Background(), input)

	if len(incRepo.incidents) != 0 {
		t.Errorf("expected 0 incidents (cooldown active), got %d", len(incRepo.incidents))
	}
}

func TestEvaluateMetric_RuleDoesNotMatch(t *testing.T) {
	svc, incRepo, _, ruleRepo := newTestService()

	rule, _ := domain.NewIncidentRule(
		testTenantID, "CPU Rule", "",
		"", "server",
		domain.RuleCondition{MetricType: "cpu_percent", Operator: domain.OperatorGT, Threshold: 90},
		domain.SeverityMajor, 0,
	)
	ruleRepo.rules = append(ruleRepo.rules, rule)

	input := app.MetricInput{
		TenantID:      testTenantID,
		ComponentID:   testComponentID,
		CollectorType: "server",
		MetricType:    "cpu_percent",
		Value:         80.0, // Below threshold
	}

	svc.EvaluateMetric(context.Background(), input)

	if len(incRepo.incidents) != 0 {
		t.Errorf("expected 0 incidents (below threshold), got %d", len(incRepo.incidents))
	}
}

func TestEvaluateMetric_WrongCollectorType(t *testing.T) {
	svc, incRepo, _, ruleRepo := newTestService()

	rule, _ := domain.NewIncidentRule(
		testTenantID, "CPU Rule", "",
		"", "server",
		domain.RuleCondition{MetricType: "cpu_percent", Operator: domain.OperatorGT, Threshold: 90},
		domain.SeverityMajor, 0,
	)
	ruleRepo.rules = append(ruleRepo.rules, rule)

	input := app.MetricInput{
		TenantID:      testTenantID,
		ComponentID:   testComponentID,
		CollectorType: "database", // Different collector
		MetricType:    "cpu_percent",
		Value:         95.0,
	}

	svc.EvaluateMetric(context.Background(), input)

	if len(incRepo.incidents) != 0 {
		t.Errorf("expected 0 incidents (wrong collector), got %d", len(incRepo.incidents))
	}
}

func TestEvaluateMetric_Correlation(t *testing.T) {
	svc, incRepo, _, ruleRepo := newTestService()

	// Pre-existing active incident on same component
	existing, _ := domain.NewIncident(testTenantID, testComponentID, "old-rule", "Existing issue", "", domain.SeverityMinor)
	incRepo.incidents = append(incRepo.incidents, existing)

	rule, _ := domain.NewIncidentRule(
		testTenantID, "New Rule", "",
		"", "server",
		domain.RuleCondition{MetricType: "memory_percent", Operator: domain.OperatorGT, Threshold: 85},
		domain.SeverityMajor, 0,
	)
	ruleRepo.rules = append(ruleRepo.rules, rule)

	input := app.MetricInput{
		TenantID:      testTenantID,
		ComponentID:   testComponentID,
		CollectorType: "server",
		MetricType:    "memory_percent",
		Value:         92.0,
	}

	svc.EvaluateMetric(context.Background(), input)

	if len(incRepo.incidents) != 2 {
		t.Fatalf("expected 2 incidents, got %d", len(incRepo.incidents))
	}

	newInc := incRepo.incidents[1]
	if newInc.CorrelationID != existing.CorrelationID {
		t.Errorf("expected correlation with existing incident (%s), got %s", existing.CorrelationID, newInc.CorrelationID)
	}
}

func TestServiceAcknowledge(t *testing.T) {
	svc, incRepo, _, _ := newTestService()

	inc, _ := domain.NewIncident(testTenantID, testComponentID, "", "Test", "desc", domain.SeverityMajor)
	inc.ClearUncommittedEvents()
	incRepo.incidents = append(incRepo.incidents, inc)

	err := svc.Acknowledge(context.Background(), inc.ID, "user-1")
	if err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}

	updated, _ := incRepo.GetByID(context.Background(), inc.ID)
	if updated.Status != domain.IncidentStatusAcknowledged {
		t.Errorf("expected acknowledged, got %s", updated.Status)
	}
}

func TestServiceResolve(t *testing.T) {
	svc, incRepo, _, _ := newTestService()

	inc, _ := domain.NewIncident(testTenantID, testComponentID, "", "Test", "desc", domain.SeverityMajor)
	inc.ClearUncommittedEvents()
	incRepo.incidents = append(incRepo.incidents, inc)

	err := svc.Resolve(context.Background(), inc.ID, "user-1", "Fixed the issue")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	updated, _ := incRepo.GetByID(context.Background(), inc.ID)
	if updated.Status != domain.IncidentStatusResolved {
		t.Errorf("expected resolved, got %s", updated.Status)
	}
}

func TestServiceAddComment(t *testing.T) {
	svc, incRepo, _, _ := newTestService()

	inc, _ := domain.NewIncident(testTenantID, testComponentID, "", "Test", "desc", domain.SeverityMajor)
	inc.ClearUncommittedEvents()
	incRepo.incidents = append(incRepo.incidents, inc)

	comment, err := svc.AddComment(context.Background(), inc.ID, testTenantID, "user-1", "Investigating network latency")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}
	if comment.Content != "Investigating network latency" {
		t.Errorf("unexpected content: %s", comment.Content)
	}
}

func TestServiceGetStats(t *testing.T) {
	svc, incRepo, _, _ := newTestService()

	// Create incidents in various states
	inc1, _ := domain.NewIncident(testTenantID, testComponentID, "", "Open 1", "", domain.SeverityMajor)
	inc1.ClearUncommittedEvents()
	incRepo.incidents = append(incRepo.incidents, inc1)

	inc2, _ := domain.NewIncident(testTenantID, testComponentID, "", "Resolved", "", domain.SeverityMinor)
	inc2.ClearUncommittedEvents()
	_ = inc2.Resolve("user-1", "fixed")
	inc2.ClearUncommittedEvents()
	incRepo.incidents = append(incRepo.incidents, inc2)

	stats, err := svc.GetStats(context.Background(), testTenantID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalOpen != 1 {
		t.Errorf("expected 1 open, got %d", stats.TotalOpen)
	}
	if stats.TotalResolved != 1 {
		t.Errorf("expected 1 resolved, got %d", stats.TotalResolved)
	}
	if stats.TotalActive != 1 {
		t.Errorf("expected 1 active, got %d", stats.TotalActive)
	}
}

func TestServiceListIncidents(t *testing.T) {
	svc, incRepo, _, _ := newTestService()

	for i := 0; i < 5; i++ {
		inc, _ := domain.NewIncident(testTenantID, testComponentID, "", "Test", "desc", domain.SeverityMajor)
		inc.ClearUncommittedEvents()
		incRepo.incidents = append(incRepo.incidents, inc)
	}

	incidents, total, err := svc.ListIncidents(context.Background(), testTenantID, port.IncidentFilter{Limit: 3})
	if err != nil {
		t.Fatalf("ListIncidents: %v", err)
	}
	if len(incidents) != 3 {
		t.Errorf("expected 3 incidents (limit), got %d", len(incidents))
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
}
