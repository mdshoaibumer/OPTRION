package service

import (
	"context"

	"github.com/google/uuid"
)

// EscalationEngine handles escalation logic for alerts.
type EscalationEngine interface {
	StartEscalation(ctx context.Context, alertID uuid.UUID) error
}

type EscalationEngineImpl struct {
	// dependencies: repositories, delivery, policies
}

func (ee *EscalationEngineImpl) StartEscalation(ctx context.Context, alertID uuid.UUID) error {
	// TODO: Implement escalation logic: timers, reminders, escalation steps
	return nil
}
