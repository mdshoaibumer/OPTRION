package service

import (
	"context"
	"log/slog"

	"github.com/optrion/optrion/internal/alert/port/repository"
)

// EscalationEngine handles escalation logic for alerts.
type EscalationEngine interface {
	StartEscalation(ctx context.Context, alertID string) error
}

type EscalationEngineImpl struct {
	alerts   repository.AlertRepository
	channels repository.AlertChannelRepository
	policies repository.EscalationPolicyRepository
	sender   ChannelSender
	logger   *slog.Logger
}

func NewEscalationEngine(
	alerts repository.AlertRepository,
	channels repository.AlertChannelRepository,
	policies repository.EscalationPolicyRepository,
	sender ChannelSender,
	logger *slog.Logger,
) *EscalationEngineImpl {
	return &EscalationEngineImpl{
		alerts:   alerts,
		channels: channels,
		policies: policies,
		sender:   sender,
		logger:   logger,
	}
}

func (ee *EscalationEngineImpl) StartEscalation(ctx context.Context, alertID string) error {
	alertInstance, err := ee.alerts.FindByID(ctx, alertID)
	if err != nil {
		return err
	}
	if alertInstance == nil {
		return nil
	}

	ee.logger.Info("escalation started",
		"alert_id", alertID,
		"severity", alertInstance.Severity,
	)

	// In a full implementation, this would:
	// 1. Load the escalation policy for the alert's rule
	// 2. Start a timer for each escalation step
	// 3. If the alert is not acknowledged within the delay, escalate to the next step
	// For now, we log the escalation start
	return nil
}
