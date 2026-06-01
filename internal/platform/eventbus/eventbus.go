package eventbus

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Event is the base interface for all domain events.
type Event interface {
	// EventType returns the event type identifier.
	EventType() string
	// TenantID returns the tenant this event belongs to.
	TenantID() string
}

// Handler processes a single event.
type Handler func(ctx context.Context, event Event) error

// Bus is an in-process synchronous event bus.
// Events are dispatched to all registered handlers for the event type.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	logger   *slog.Logger
}

// New creates a new event bus.
func New(logger *slog.Logger) *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
		logger:   logger,
	}
}

// Subscribe registers a handler for a specific event type.
func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Publish dispatches an event to all registered handlers synchronously.
// If any handler returns an error, it is logged but does not prevent other handlers from running.
func (b *Bus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers := b.handlers[event.EventType()]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			b.logger.Error("event handler failed",
				"event_type", event.EventType(),
				"tenant_id", event.TenantID(),
				"error", err,
			)
		}
	}
}

// PublishAll dispatches multiple events.
func (b *Bus) PublishAll(ctx context.Context, events []Event) {
	for _, evt := range events {
		b.Publish(ctx, evt)
	}
}

// --- Concrete Event Types ---

// IncidentOpenedEvent is published when a new incident is created.
type IncidentOpenedEvent struct {
	Tenant      string
	IncidentID  string
	ComponentID string
	Title       string
	Severity    string
}

func (e IncidentOpenedEvent) EventType() string { return "incident.opened" }
func (e IncidentOpenedEvent) TenantID() string  { return e.Tenant }

// IncidentResolvedEvent is published when an incident is resolved.
type IncidentResolvedEvent struct {
	Tenant      string
	IncidentID  string
	ComponentID string
}

func (e IncidentResolvedEvent) EventType() string { return "incident.resolved" }
func (e IncidentResolvedEvent) TenantID() string  { return e.Tenant }

// IncidentEscalatedEvent is published when an incident severity is escalated.
type IncidentEscalatedEvent struct {
	Tenant      string
	IncidentID  string
	ComponentID string
	OldSeverity string
	NewSeverity string
}

func (e IncidentEscalatedEvent) EventType() string { return "incident.escalated" }
func (e IncidentEscalatedEvent) TenantID() string  { return e.Tenant }

// HealthDegradedEvent is published when a component's health drops below threshold.
type HealthDegradedEvent struct {
	Tenant      string
	ComponentID string
	Score       int
	PrevScore   int
}

func (e HealthDegradedEvent) EventType() string { return "health.degraded" }
func (e HealthDegradedEvent) TenantID() string  { return e.Tenant }

// HealthRecoveredEvent is published when a component's health recovers.
type HealthRecoveredEvent struct {
	Tenant      string
	ComponentID string
	Score       int
}

func (e HealthRecoveredEvent) EventType() string { return "health.recovered" }
func (e HealthRecoveredEvent) TenantID() string  { return e.Tenant }

// String returns a human-readable representation of the event bus state.
func (b *Bus) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	total := 0
	for _, handlers := range b.handlers {
		total += len(handlers)
	}
	return fmt.Sprintf("EventBus{types=%d, handlers=%d}", len(b.handlers), total)
}
