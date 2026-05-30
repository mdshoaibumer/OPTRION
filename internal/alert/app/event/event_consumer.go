package event

import (
	"context"
)

// IncidentEventType represents the type of incident event.
type IncidentEventType string

const (
	IncidentOpened       IncidentEventType = "IncidentOpened"
	IncidentAcknowledged IncidentEventType = "IncidentAcknowledged"
	IncidentResolved     IncidentEventType = "IncidentResolved"
	IncidentClosed       IncidentEventType = "IncidentClosed"
)

// IncidentEvent represents an event from the incident engine.
type IncidentEvent struct {
	ID         string
	TenantID   string
	IncidentID string
	Type       IncidentEventType
	Payload    map[string]interface{}
}

// EventConsumer consumes incident events and triggers alert generation.
type EventConsumer interface {
	Consume(ctx context.Context, event IncidentEvent) error
}
