package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/optrion/optrion/internal/incident/domain"
	"github.com/optrion/optrion/internal/incident/port"
)

// IncidentService orchestrates incident management use cases.
type IncidentService struct {
	incidents port.IncidentRepository
	events    port.IncidentEventRepository
	rules     port.IncidentRuleRepository
	comments  port.IncidentCommentRepository
	timeline  port.IncidentTimelineRepository
	logger    *slog.Logger
}

// NewIncidentService creates a new incident service with all dependencies.
func NewIncidentService(
	incidents port.IncidentRepository,
	events port.IncidentEventRepository,
	rules port.IncidentRuleRepository,
	comments port.IncidentCommentRepository,
	timeline port.IncidentTimelineRepository,
	logger *slog.Logger,
) *IncidentService {
	return &IncidentService{
		incidents: incidents,
		events:    events,
		rules:     rules,
		comments:  comments,
		timeline:  timeline,
		logger:    logger,
	}
}

// --- Rule Evaluation (STEP 5 + 6 + 7 + 8) ---

// MetricInput represents a metric reading to evaluate against rules.
type MetricInput struct {
	TenantID      string
	ComponentID   string
	CollectorType string
	MetricType    string
	Value         float64
	HealthScore   int
}

// EvaluateMetric checks a metric reading against all enabled rules for the tenant.
// If a rule fires, it creates or updates an incident (with deduplication and correlation).
func (s *IncidentService) EvaluateMetric(ctx context.Context, input MetricInput) {
	rules, err := s.rules.ListEnabled(ctx, input.TenantID)
	if err != nil {
		s.logger.Error("failed to load incident rules", "tenant_id", input.TenantID, "error", err)
		return
	}

	for _, rule := range rules {
		if !s.ruleMatchesInput(rule, input) {
			continue
		}

		if !rule.Condition.Matches(input.Value) {
			continue
		}

		// Check cooldown
		if !s.isCooldownExpired(ctx, rule, input.ComponentID) {
			continue
		}

		// Deduplicate: check for existing active incident from same rule + component
		existing, err := s.incidents.FindActiveByRuleAndComponent(ctx, rule.ID, input.ComponentID)
		if err != nil {
			s.logger.Error("failed to check for existing incident", "rule_id", rule.ID, "error", err)
			continue
		}

		if existing != nil {
			// Update existing: escalate severity if needed
			newSeverity := s.computeSeverity(input, rule)
			if newSeverity.NumericLevel() > existing.Severity.NumericLevel() {
				existing.ChangeSeverity(newSeverity)
				if err := s.incidents.Update(ctx, existing); err != nil {
					s.logger.Error("failed to update incident severity", "incident_id", existing.ID, "error", err)
				}
				s.persistEvents(ctx, existing)
			}
			continue
		}

		// Create new incident
		severity := s.computeSeverity(input, rule)
		title := s.generateTitle(rule, input)
		description := s.generateDescription(rule, input)

		incident, err := domain.NewIncident(input.TenantID, input.ComponentID, rule.ID, title, description, severity)
		if err != nil {
			s.logger.Error("failed to create incident", "rule_id", rule.ID, "error", err)
			continue
		}

		// Correlate: check for related incidents on same component
		s.correlateIncident(ctx, incident)

		if err := s.incidents.Create(ctx, incident); err != nil {
			s.logger.Error("failed to persist incident", "incident_id", incident.ID, "error", err)
			continue
		}

		s.persistEvents(ctx, incident)
		s.addTimelineEntry(ctx, incident, domain.TimelineEvent, "Incident opened", fmt.Sprintf("Rule '%s' triggered: %s", rule.Name, title))

		// Record rule firing for cooldown tracking
		if err := s.rules.RecordFired(ctx, rule.ID, input.ComponentID, time.Now().UTC()); err != nil {
			s.logger.Error("failed to record rule firing", "rule_id", rule.ID, "error", err)
		}

		s.logger.Warn("incident created",
			"incident_id", incident.ID,
			"title", incident.Title,
			"severity", incident.Severity,
			"component_id", input.ComponentID,
		)
	}
}

// ruleMatchesInput checks if a rule is applicable to the given metric input.
func (s *IncidentService) ruleMatchesInput(rule *domain.IncidentRule, input MetricInput) bool {
	// Rule must match collector type
	if rule.CollectorType != input.CollectorType {
		return false
	}

	// Rule must match metric type
	if rule.Condition.MetricType != input.MetricType {
		return false
	}

	// If rule is scoped to a specific component, it must match
	if rule.ComponentID != "" && rule.ComponentID != input.ComponentID {
		return false
	}

	return true
}

// isCooldownExpired checks if enough time has passed since the last firing.
func (s *IncidentService) isCooldownExpired(ctx context.Context, rule *domain.IncidentRule, componentID string) bool {
	lastFired, err := s.rules.GetLastFiredAt(ctx, rule.ID, componentID)
	if err != nil || lastFired == nil {
		return true // No previous firing or error = allow
	}

	return time.Since(*lastFired) >= rule.Cooldown
}

// computeSeverity determines the appropriate severity based on the metric and rule.
func (s *IncidentService) computeSeverity(input MetricInput, rule *domain.IncidentRule) domain.IncidentSeverity {
	// If health score is available, use it to escalate
	if input.HealthScore > 0 {
		switch {
		case input.HealthScore < 30:
			return domain.SeverityCritical
		case input.HealthScore < 50:
			return domain.SeverityMajor
		case input.HealthScore < 70:
			return domain.SeverityMinor
		case input.HealthScore < 90:
			return domain.SeverityWarning
		}
	}

	// Fall back to rule-defined severity
	return rule.Severity
}

// generateTitle creates a human-readable incident title.
func (s *IncidentService) generateTitle(rule *domain.IncidentRule, input MetricInput) string {
	return fmt.Sprintf("%s: %s %s %.2f",
		rule.Name,
		input.MetricType,
		rule.Condition.Operator,
		input.Value,
	)
}

// generateDescription creates detailed incident context.
func (s *IncidentService) generateDescription(rule *domain.IncidentRule, input MetricInput) string {
	return fmt.Sprintf(
		"Rule '%s' triggered.\nComponent: %s\nMetric: %s\nValue: %.2f (threshold: %s %.2f)\nCollector: %s",
		rule.Name,
		input.ComponentID,
		input.MetricType,
		input.Value,
		rule.Condition.Operator,
		rule.Condition.Threshold,
		input.CollectorType,
	)
}

// correlateIncident links an incident to existing related incidents on the same component.
func (s *IncidentService) correlateIncident(ctx context.Context, incident *domain.Incident) {
	// Find active incidents for the same component
	filter := port.IncidentFilter{ComponentID: &incident.ComponentID, Limit: 10}
	related, err := s.incidents.ListByTenant(ctx, incident.TenantID, filter)
	if err != nil {
		return
	}

	for _, r := range related {
		if r.Status.IsActive() && r.ID != incident.ID {
			// Use the existing correlation group
			incident.Correlate(r.CorrelationID)
			return
		}
	}
}

// --- Commands ---

// Acknowledge marks an incident as acknowledged.
func (s *IncidentService) Acknowledge(ctx context.Context, incidentID, actorID string) error {
	incident, err := s.incidents.GetByID(ctx, incidentID)
	if err != nil {
		return err
	}
	if incident == nil {
		return fmt.Errorf("incident not found: %s", incidentID)
	}

	if err := incident.Acknowledge(actorID); err != nil {
		return err
	}

	if err := s.incidents.Update(ctx, incident); err != nil {
		return fmt.Errorf("updating incident: %w", err)
	}

	s.persistEvents(ctx, incident)
	s.addTimelineEntry(ctx, incident, domain.TimelineEvent, "Incident acknowledged", fmt.Sprintf("Acknowledged by %s", actorID))
	return nil
}

// Investigate marks an incident as under investigation.
func (s *IncidentService) Investigate(ctx context.Context, incidentID, actorID string) error {
	incident, err := s.incidents.GetByID(ctx, incidentID)
	if err != nil {
		return err
	}
	if incident == nil {
		return fmt.Errorf("incident not found: %s", incidentID)
	}

	if err := incident.Investigate(actorID); err != nil {
		return err
	}

	if err := s.incidents.Update(ctx, incident); err != nil {
		return fmt.Errorf("updating incident: %w", err)
	}

	s.persistEvents(ctx, incident)
	s.addTimelineEntry(ctx, incident, domain.TimelineEvent, "Investigation started", fmt.Sprintf("Investigation started by %s", actorID))
	return nil
}

// Resolve marks an incident as resolved.
func (s *IncidentService) Resolve(ctx context.Context, incidentID, actorID, resolution string) error {
	incident, err := s.incidents.GetByID(ctx, incidentID)
	if err != nil {
		return err
	}
	if incident == nil {
		return fmt.Errorf("incident not found: %s", incidentID)
	}

	if err := incident.Resolve(actorID, resolution); err != nil {
		return err
	}

	if err := s.incidents.Update(ctx, incident); err != nil {
		return fmt.Errorf("updating incident: %w", err)
	}

	s.persistEvents(ctx, incident)
	s.addTimelineEntry(ctx, incident, domain.TimelineEvent, "Incident resolved", fmt.Sprintf("Resolved by %s: %s", actorID, resolution))
	return nil
}

// Close marks an incident as closed (terminal state).
func (s *IncidentService) Close(ctx context.Context, incidentID, actorID, reason string) error {
	incident, err := s.incidents.GetByID(ctx, incidentID)
	if err != nil {
		return err
	}
	if incident == nil {
		return fmt.Errorf("incident not found: %s", incidentID)
	}

	if err := incident.Close(actorID, reason); err != nil {
		return err
	}

	if err := s.incidents.Update(ctx, incident); err != nil {
		return fmt.Errorf("updating incident: %w", err)
	}

	s.persistEvents(ctx, incident)
	s.addTimelineEntry(ctx, incident, domain.TimelineEvent, "Incident closed", fmt.Sprintf("Closed by %s: %s", actorID, reason))
	return nil
}

// AddComment adds a human comment to an incident.
func (s *IncidentService) AddComment(ctx context.Context, incidentID, tenantID, authorID, content string) (*domain.IncidentComment, error) {
	incident, err := s.incidents.GetByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, fmt.Errorf("incident not found: %s", incidentID)
	}

	comment, err := domain.NewIncidentComment(tenantID, incidentID, authorID, content)
	if err != nil {
		return nil, err
	}

	if err := s.comments.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("creating comment: %w", err)
	}

	s.addTimelineEntry(ctx, incident, domain.TimelineComment, "Comment added", content)
	return comment, nil
}

// --- Queries ---

// GetIncident returns a single incident by ID.
func (s *IncidentService) GetIncident(ctx context.Context, id string) (*domain.Incident, error) {
	return s.incidents.GetByID(ctx, id)
}

// ListIncidents returns incidents for a tenant with optional filtering.
func (s *IncidentService) ListIncidents(ctx context.Context, tenantID string, filter port.IncidentFilter) ([]*domain.Incident, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}

	incidents, err := s.incidents.ListByTenant(ctx, tenantID, filter)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.incidents.CountByTenant(ctx, tenantID, filter)
	if err != nil {
		return nil, 0, err
	}

	return incidents, count, nil
}

// GetTimeline returns the timeline for an incident.
func (s *IncidentService) GetTimeline(ctx context.Context, incidentID string, limit int) ([]*domain.IncidentTimeline, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.timeline.ListByIncident(ctx, incidentID, limit)
}

// GetTenantTimeline returns a cross-incident timeline for a tenant.
func (s *IncidentService) GetTenantTimeline(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.IncidentTimeline, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	return s.timeline.ListByTenant(ctx, tenantID, from, to, limit)
}

// GetEvents returns the event stream for an incident (event sourcing replay).
func (s *IncidentService) GetEvents(ctx context.Context, incidentID string) ([]domain.IncidentEvent, error) {
	return s.events.ListByIncident(ctx, incidentID)
}

// GetComments returns comments for an incident.
func (s *IncidentService) GetComments(ctx context.Context, incidentID string) ([]*domain.IncidentComment, error) {
	return s.comments.ListByIncident(ctx, incidentID)
}

// IncidentStats provides summary statistics for a tenant.
type IncidentStats struct {
	TenantID      string                          `json:"tenant_id"`
	TotalActive   int                             `json:"total_active"`
	TotalOpen     int                             `json:"total_open"`
	TotalResolved int                             `json:"total_resolved"`
	TotalClosed   int                             `json:"total_closed"`
	BySeverity    map[domain.IncidentSeverity]int `json:"by_severity"`
	MTTR          time.Duration                   `json:"mttr_seconds"` // Mean Time To Resolve
}

// GetStats returns incident statistics for a tenant.
func (s *IncidentService) GetStats(ctx context.Context, tenantID string) (*IncidentStats, error) {
	stats := &IncidentStats{
		TenantID:   tenantID,
		BySeverity: make(map[domain.IncidentSeverity]int),
	}

	// Count by status
	statuses := []domain.IncidentStatus{
		domain.IncidentStatusOpen,
		domain.IncidentStatusAcknowledged,
		domain.IncidentStatusInvestigating,
		domain.IncidentStatusResolved,
		domain.IncidentStatusClosed,
	}

	for _, status := range statuses {
		filter := port.IncidentFilter{Status: &status, Limit: 0}
		count, err := s.incidents.CountByTenant(ctx, tenantID, filter)
		if err != nil {
			return nil, err
		}

		switch status {
		case domain.IncidentStatusOpen:
			stats.TotalOpen = count
			stats.TotalActive += count
		case domain.IncidentStatusAcknowledged, domain.IncidentStatusInvestigating:
			stats.TotalActive += count
		case domain.IncidentStatusResolved:
			stats.TotalResolved = count
		case domain.IncidentStatusClosed:
			stats.TotalClosed = count
		}
	}

	// Count by severity (active only)
	severities := []domain.IncidentSeverity{
		domain.SeverityInfo, domain.SeverityWarning,
		domain.SeverityMinor, domain.SeverityMajor, domain.SeverityCritical,
	}
	for _, sev := range severities {
		filter := port.IncidentFilter{Severity: &sev, Limit: 0}
		count, err := s.incidents.CountByTenant(ctx, tenantID, filter)
		if err != nil {
			return nil, err
		}
		stats.BySeverity[sev] = count
	}

	// Calculate MTTR from resolved incidents in last 30 days
	from := time.Now().Add(-30 * 24 * time.Hour)
	to := time.Now()
	resolvedStatus := domain.IncidentStatusResolved
	resolvedFilter := port.IncidentFilter{Status: &resolvedStatus, From: &from, To: &to, Limit: 1000}
	resolved, err := s.incidents.ListByTenant(ctx, tenantID, resolvedFilter)
	if err == nil && len(resolved) > 0 {
		var totalDuration time.Duration
		for _, inc := range resolved {
			if inc.ResolvedAt != nil {
				totalDuration += inc.ResolvedAt.Sub(inc.OccurredAt)
			}
		}
		stats.MTTR = totalDuration / time.Duration(len(resolved))
	}

	return stats, nil
}

// --- Internal Helpers ---

func (s *IncidentService) persistEvents(ctx context.Context, incident *domain.Incident) {
	events := incident.UncommittedEvents()
	if len(events) == 0 {
		return
	}

	if err := s.events.Append(ctx, events); err != nil {
		s.logger.Error("failed to persist incident events", "incident_id", incident.ID, "error", err)
	}
	incident.ClearUncommittedEvents()
}

func (s *IncidentService) addTimelineEntry(ctx context.Context, incident *domain.Incident, entryType domain.TimelineEntryType, title, details string) {
	entry := &domain.IncidentTimeline{
		ID:         "", // Will be set by repo
		TenantID:   incident.TenantID,
		IncidentID: incident.ID,
		EntryType:  entryType,
		Title:      title,
		Details:    details,
		ActorID:    "system",
		OccurredAt: time.Now().UTC(),
	}

	if err := s.timeline.Create(ctx, entry); err != nil {
		s.logger.Error("failed to create timeline entry", "incident_id", incident.ID, "error", err)
	}
}
