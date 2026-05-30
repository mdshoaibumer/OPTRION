package domain

import (
	"fmt"
	"time"

	"github.com/optrion/optrion/internal/shared/id"
)

// IncidentStatus represents the lifecycle state of an incident.
type IncidentStatus string

const (
	IncidentStatusOpen          IncidentStatus = "open"
	IncidentStatusAcknowledged  IncidentStatus = "acknowledged"
	IncidentStatusInvestigating IncidentStatus = "investigating"
	IncidentStatusResolved      IncidentStatus = "resolved"
	IncidentStatusClosed        IncidentStatus = "closed"
)

// IsValid checks if the status is a recognized value.
func (s IncidentStatus) IsValid() bool {
	switch s {
	case IncidentStatusOpen, IncidentStatusAcknowledged,
		IncidentStatusInvestigating, IncidentStatusResolved, IncidentStatusClosed:
		return true
	}
	return false
}

// IsActive returns true if the incident requires attention.
func (s IncidentStatus) IsActive() bool {
	return s == IncidentStatusOpen || s == IncidentStatusAcknowledged || s == IncidentStatusInvestigating
}

// IncidentSeverity represents the severity level of an incident.
type IncidentSeverity string

const (
	SeverityInfo     IncidentSeverity = "info"
	SeverityWarning  IncidentSeverity = "warning"
	SeverityMinor    IncidentSeverity = "minor"
	SeverityMajor    IncidentSeverity = "major"
	SeverityCritical IncidentSeverity = "critical"
)

// IsValid checks if the severity is recognized.
func (s IncidentSeverity) IsValid() bool {
	switch s {
	case SeverityInfo, SeverityWarning, SeverityMinor, SeverityMajor, SeverityCritical:
		return true
	}
	return false
}

// NumericLevel returns an integer for comparison (higher = more severe).
func (s IncidentSeverity) NumericLevel() int {
	switch s {
	case SeverityInfo:
		return 1
	case SeverityWarning:
		return 2
	case SeverityMinor:
		return 3
	case SeverityMajor:
		return 4
	case SeverityCritical:
		return 5
	default:
		return 0
	}
}

// EventType identifies what happened in the incident lifecycle.
type EventType string

const (
	EventIncidentOpened        EventType = "incident.opened"
	EventIncidentAcknowledged  EventType = "incident.acknowledged"
	EventIncidentInvestigating EventType = "incident.investigating"
	EventIncidentResolved      EventType = "incident.resolved"
	EventIncidentClosed        EventType = "incident.closed"
	EventCommentAdded          EventType = "incident.comment_added"
	EventSeverityChanged       EventType = "incident.severity_changed"
	EventCorrelated            EventType = "incident.correlated"
)

// --- State Machine ---

// validTransitions defines allowed state transitions.
var validTransitions = map[IncidentStatus][]IncidentStatus{
	IncidentStatusOpen:          {IncidentStatusAcknowledged, IncidentStatusInvestigating, IncidentStatusResolved, IncidentStatusClosed},
	IncidentStatusAcknowledged:  {IncidentStatusInvestigating, IncidentStatusResolved, IncidentStatusClosed},
	IncidentStatusInvestigating: {IncidentStatusResolved, IncidentStatusClosed},
	IncidentStatusResolved:      {IncidentStatusClosed, IncidentStatusOpen}, // Reopen allowed
	IncidentStatusClosed:        {},                                         // Terminal state
}

// CanTransitionTo checks if a transition from current status to target is valid.
func CanTransitionTo(current, target IncidentStatus) bool {
	allowed, ok := validTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

// ErrInvalidTransition is returned when a state transition is not permitted.
type ErrInvalidTransition struct {
	From IncidentStatus
	To   IncidentStatus
}

func (e ErrInvalidTransition) Error() string {
	return fmt.Sprintf("invalid transition from %s to %s", e.From, e.To)
}

// --- Incident Aggregate ---

// Incident is the aggregate root for the incident bounded context.
// State is rebuilt from IncidentEvents (event sourcing).
type Incident struct {
	ID             string
	TenantID       string
	ComponentID    string
	Title          string
	Description    string
	Status         IncidentStatus
	Severity       IncidentSeverity
	RuleID         string // Which rule triggered this incident
	CorrelationID  string // Groups related incidents
	OccurredAt     time.Time
	AcknowledgedAt *time.Time
	ResolvedAt     *time.Time
	ClosedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int // Optimistic concurrency control

	// Uncommitted events produced by commands on this aggregate
	uncommittedEvents []IncidentEvent
}

// NewIncident creates a new incident and raises the IncidentOpened event.
func NewIncident(tenantID, componentID, ruleID, title, description string, severity IncidentSeverity) (*Incident, error) {
	if !id.IsValid(tenantID) {
		return nil, fmt.Errorf("invalid tenant ID: %s", tenantID)
	}
	if !id.IsValid(componentID) {
		return nil, fmt.Errorf("invalid component ID: %s", componentID)
	}
	if title == "" {
		return nil, fmt.Errorf("incident title is required")
	}
	if !severity.IsValid() {
		return nil, fmt.Errorf("invalid severity: %s", severity)
	}

	now := time.Now().UTC()
	inc := &Incident{
		ID:            id.New(),
		TenantID:      tenantID,
		ComponentID:   componentID,
		Title:         title,
		Description:   description,
		Status:        IncidentStatusOpen,
		Severity:      severity,
		RuleID:        ruleID,
		CorrelationID: id.New(), // Each new incident starts its own correlation group
		OccurredAt:    now,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}

	inc.raiseEvent(EventIncidentOpened, map[string]string{
		"title":        title,
		"severity":     string(severity),
		"component_id": componentID,
		"rule_id":      ruleID,
	})

	return inc, nil
}

// Acknowledge transitions the incident to acknowledged status.
func (i *Incident) Acknowledge(actorID string) error {
	if err := i.transitionTo(IncidentStatusAcknowledged); err != nil {
		return err
	}
	now := time.Now().UTC()
	i.AcknowledgedAt = &now
	i.raiseEvent(EventIncidentAcknowledged, map[string]string{"actor_id": actorID})
	return nil
}

// Investigate transitions the incident to investigating status.
func (i *Incident) Investigate(actorID string) error {
	if err := i.transitionTo(IncidentStatusInvestigating); err != nil {
		return err
	}
	i.raiseEvent(EventIncidentInvestigating, map[string]string{"actor_id": actorID})
	return nil
}

// Resolve transitions the incident to resolved status.
func (i *Incident) Resolve(actorID, resolution string) error {
	if err := i.transitionTo(IncidentStatusResolved); err != nil {
		return err
	}
	now := time.Now().UTC()
	i.ResolvedAt = &now
	i.raiseEvent(EventIncidentResolved, map[string]string{
		"actor_id":   actorID,
		"resolution": resolution,
	})
	return nil
}

// Close transitions the incident to closed status (terminal).
func (i *Incident) Close(actorID, reason string) error {
	if err := i.transitionTo(IncidentStatusClosed); err != nil {
		return err
	}
	now := time.Now().UTC()
	i.ClosedAt = &now
	i.raiseEvent(EventIncidentClosed, map[string]string{
		"actor_id": actorID,
		"reason":   reason,
	})
	return nil
}

// ChangeSeverity updates the severity if it differs from current.
func (i *Incident) ChangeSeverity(newSeverity IncidentSeverity) {
	if i.Severity == newSeverity {
		return
	}
	old := i.Severity
	i.Severity = newSeverity
	i.UpdatedAt = time.Now().UTC()
	i.raiseEvent(EventSeverityChanged, map[string]string{
		"old_severity": string(old),
		"new_severity": string(newSeverity),
	})
}

// Correlate links this incident to a correlation group.
func (i *Incident) Correlate(correlationID string) {
	if i.CorrelationID == correlationID {
		return
	}
	i.CorrelationID = correlationID
	i.UpdatedAt = time.Now().UTC()
	i.raiseEvent(EventCorrelated, map[string]string{
		"correlation_id": correlationID,
	})
}

// transitionTo validates and applies a state transition.
func (i *Incident) transitionTo(target IncidentStatus) error {
	if !CanTransitionTo(i.Status, target) {
		return ErrInvalidTransition{From: i.Status, To: target}
	}
	i.Status = target
	i.UpdatedAt = time.Now().UTC()
	i.Version++
	return nil
}

// raiseEvent appends an uncommitted event to the aggregate.
func (i *Incident) raiseEvent(eventType EventType, metadata map[string]string) {
	event := IncidentEvent{
		ID:         id.New(),
		TenantID:   i.TenantID,
		IncidentID: i.ID,
		EventType:  eventType,
		Metadata:   metadata,
		OccurredAt: time.Now().UTC(),
	}
	i.uncommittedEvents = append(i.uncommittedEvents, event)
}

// UncommittedEvents returns events produced by recent commands.
func (i *Incident) UncommittedEvents() []IncidentEvent {
	return i.uncommittedEvents
}

// ClearUncommittedEvents removes pending events after persistence.
func (i *Incident) ClearUncommittedEvents() {
	i.uncommittedEvents = nil
}

// --- IncidentEvent (Event Sourcing) ---

// IncidentEvent records a single state change in the incident lifecycle.
type IncidentEvent struct {
	ID         string
	TenantID   string
	IncidentID string
	EventType  EventType
	Metadata   map[string]string
	OccurredAt time.Time
}

// --- IncidentComment ---

// IncidentComment represents a human-written note on an incident.
type IncidentComment struct {
	ID         string
	TenantID   string
	IncidentID string
	AuthorID   string
	Content    string
	CreatedAt  time.Time
}

// NewIncidentComment creates a new comment on an incident.
func NewIncidentComment(tenantID, incidentID, authorID, content string) (*IncidentComment, error) {
	if content == "" {
		return nil, fmt.Errorf("comment content is required")
	}
	if !id.IsValid(incidentID) {
		return nil, fmt.Errorf("invalid incident ID: %s", incidentID)
	}

	return &IncidentComment{
		ID:         id.New(),
		TenantID:   tenantID,
		IncidentID: incidentID,
		AuthorID:   authorID,
		Content:    content,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

// --- IncidentRule ---

// RuleConditionOperator defines how to compare values.
type RuleConditionOperator string

const (
	OperatorGT  RuleConditionOperator = "gt"
	OperatorLT  RuleConditionOperator = "lt"
	OperatorGTE RuleConditionOperator = "gte"
	OperatorLTE RuleConditionOperator = "lte"
	OperatorEQ  RuleConditionOperator = "eq"
)

// RuleCondition defines a threshold check.
type RuleCondition struct {
	MetricType string                `json:"metric_type"`
	Operator   RuleConditionOperator `json:"operator"`
	Threshold  float64               `json:"threshold"`
}

// Matches checks whether a metric value satisfies this condition.
func (rc RuleCondition) Matches(value float64) bool {
	switch rc.Operator {
	case OperatorGT:
		return value > rc.Threshold
	case OperatorGTE:
		return value >= rc.Threshold
	case OperatorLT:
		return value < rc.Threshold
	case OperatorLTE:
		return value <= rc.Threshold
	case OperatorEQ:
		return value == rc.Threshold
	default:
		return false
	}
}

// IncidentRule defines when an incident should be created.
type IncidentRule struct {
	ID            string
	TenantID      string
	Name          string
	Description   string
	ComponentID   string           // Empty = applies to all components
	CollectorType string           // Which collector's metrics to evaluate
	Condition     RuleCondition    // The threshold condition
	Severity      IncidentSeverity // Severity to assign
	Cooldown      time.Duration    // Min time between incidents from same rule
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewIncidentRule creates a new rule for incident detection.
func NewIncidentRule(tenantID, name, description, componentID, collectorType string, condition RuleCondition, severity IncidentSeverity, cooldown time.Duration) (*IncidentRule, error) {
	if !id.IsValid(tenantID) {
		return nil, fmt.Errorf("invalid tenant ID: %s", tenantID)
	}
	if name == "" {
		return nil, fmt.Errorf("rule name is required")
	}
	if !severity.IsValid() {
		return nil, fmt.Errorf("invalid severity: %s", severity)
	}
	if cooldown < 0 {
		cooldown = 5 * time.Minute
	}

	now := time.Now().UTC()
	return &IncidentRule{
		ID:            id.New(),
		TenantID:      tenantID,
		Name:          name,
		Description:   description,
		ComponentID:   componentID,
		CollectorType: collectorType,
		Condition:     condition,
		Severity:      severity,
		Cooldown:      cooldown,
		Enabled:       true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// --- IncidentTimeline ---

// TimelineEntryType categorizes timeline entries.
type TimelineEntryType string

const (
	TimelineEvent   TimelineEntryType = "event"
	TimelineComment TimelineEntryType = "comment"
	TimelineMetric  TimelineEntryType = "metric"
)

// IncidentTimeline is a unified view of everything that happened during an incident.
type IncidentTimeline struct {
	ID         string
	TenantID   string
	IncidentID string
	EntryType  TimelineEntryType
	Title      string
	Details    string
	ActorID    string
	OccurredAt time.Time
}

// --- Rebuilding State from Events ---

// RebuildFromEvents reconstructs an Incident's current state from its event history.
func RebuildFromEvents(events []IncidentEvent) *Incident {
	if len(events) == 0 {
		return nil
	}

	var inc *Incident
	for _, evt := range events {
		switch evt.EventType {
		case EventIncidentOpened:
			inc = &Incident{
				ID:            evt.IncidentID,
				TenantID:      evt.TenantID,
				ComponentID:   evt.Metadata["component_id"],
				Title:         evt.Metadata["title"],
				Status:        IncidentStatusOpen,
				Severity:      IncidentSeverity(evt.Metadata["severity"]),
				RuleID:        evt.Metadata["rule_id"],
				CorrelationID: evt.IncidentID, // Default to own ID
				OccurredAt:    evt.OccurredAt,
				CreatedAt:     evt.OccurredAt,
				UpdatedAt:     evt.OccurredAt,
				Version:       1,
			}
		case EventIncidentAcknowledged:
			if inc != nil {
				inc.Status = IncidentStatusAcknowledged
				t := evt.OccurredAt
				inc.AcknowledgedAt = &t
				inc.UpdatedAt = evt.OccurredAt
				inc.Version++
			}
		case EventIncidentInvestigating:
			if inc != nil {
				inc.Status = IncidentStatusInvestigating
				inc.UpdatedAt = evt.OccurredAt
				inc.Version++
			}
		case EventIncidentResolved:
			if inc != nil {
				inc.Status = IncidentStatusResolved
				t := evt.OccurredAt
				inc.ResolvedAt = &t
				inc.UpdatedAt = evt.OccurredAt
				inc.Version++
			}
		case EventIncidentClosed:
			if inc != nil {
				inc.Status = IncidentStatusClosed
				t := evt.OccurredAt
				inc.ClosedAt = &t
				inc.UpdatedAt = evt.OccurredAt
				inc.Version++
			}
		case EventSeverityChanged:
			if inc != nil {
				inc.Severity = IncidentSeverity(evt.Metadata["new_severity"])
				inc.UpdatedAt = evt.OccurredAt
			}
		case EventCorrelated:
			if inc != nil {
				inc.CorrelationID = evt.Metadata["correlation_id"]
				inc.UpdatedAt = evt.OccurredAt
			}
		}
	}

	return inc
}
