package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/optrion/optrion/internal/alert/app/event"
	"github.com/optrion/optrion/internal/alert/app/routing"
	"github.com/optrion/optrion/internal/alert/domain/alert"
	"github.com/optrion/optrion/internal/alert/domain/alertchannel"
	"github.com/optrion/optrion/internal/alert/domain/alertdelivery"
	"github.com/optrion/optrion/internal/alert/domain/alertrule"
	"github.com/optrion/optrion/internal/alert/domain/notificationtemplate"
	"github.com/optrion/optrion/internal/alert/port/repository"
	"github.com/optrion/optrion/internal/shared/id"
)

// AlertEngine processes incident events and generates alerts.
type AlertEngine interface {
	ProcessEvent(ctx context.Context, evt event.IncidentEvent) error
}

// AlertEngineImpl is the production implementation.
type AlertEngineImpl struct {
	rules          repository.AlertRuleRepository
	alerts         repository.AlertRepository
	channels       repository.AlertChannelRepository
	deliveries     repository.AlertDeliveryRepository
	dedup          *DeduplicationService
	templateEngine *notificationtemplate.TemplateEngine
	sender         ChannelSender
	logger         *slog.Logger
}

// ChannelSender sends messages through a notification channel.
type ChannelSender interface {
	Send(ctx context.Context, channel *alertchannel.AlertChannel, message string) (deliveryID string, err error)
}

// NewAlertEngine creates a new alert engine with all dependencies.
func NewAlertEngine(
	rules repository.AlertRuleRepository,
	alerts repository.AlertRepository,
	channels repository.AlertChannelRepository,
	deliveries repository.AlertDeliveryRepository,
	dedup *DeduplicationService,
	sender ChannelSender,
	logger *slog.Logger,
) *AlertEngineImpl {
	return &AlertEngineImpl{
		rules:          rules,
		alerts:         alerts,
		channels:       channels,
		deliveries:     deliveries,
		dedup:          dedup,
		templateEngine: notificationtemplate.NewTemplateEngine(),
		sender:         sender,
		logger:         logger,
	}
}

// ProcessEvent handles an incident event: evaluates rules, generates alerts, and delivers notifications.
func (ae *AlertEngineImpl) ProcessEvent(ctx context.Context, evt event.IncidentEvent) error {
	ae.logger.Info("processing incident event",
		"event_type", evt.Type,
		"incident_id", evt.IncidentID,
		"tenant_id", evt.TenantID,
	)

	// Only process actionable events
	if evt.Type != event.IncidentOpened && evt.Type != event.IncidentAcknowledged {
		return nil
	}

	// Deduplication check
	dedupKey := fmt.Sprintf("%s:%s:%s", evt.TenantID, evt.IncidentID, evt.Type)
	if ae.dedup.ShouldSuppress(dedupKey) {
		ae.logger.Debug("alert suppressed by deduplication", "key", dedupKey)
		return nil
	}

	// Load enabled rules for this tenant
	rules, err := ae.rules.ListEnabledByTenant(ctx, evt.TenantID)
	if err != nil {
		return fmt.Errorf("loading alert rules: %w", err)
	}

	if len(rules) == 0 {
		ae.logger.Debug("no alert rules configured for tenant", "tenant_id", evt.TenantID)
		return nil
	}

	// Evaluate each rule against the event
	for _, rule := range rules {
		if !ae.ruleMatchesEvent(rule, evt) {
			continue
		}

		// Generate the alert
		alertInstance := &alert.Alert{
			ID:         id.New(),
			TenantID:   evt.TenantID,
			RuleID:     rule.ID,
			IncidentID: evt.IncidentID,
			Severity:   rule.Severity,
			Status:     alert.AlertStatusPending,
			Message:    ae.buildMessage(rule, evt),
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}

		if err := ae.alerts.Create(ctx, alertInstance); err != nil {
			ae.logger.Error("failed to persist alert", "error", err, "rule_id", rule.ID)
			continue
		}

		// Route to channels based on severity
		channelIDs := ae.getChannelsForRule(rule, evt)
		ae.deliverToChannels(ctx, alertInstance, channelIDs)

		ae.logger.Info("alert generated",
			"alert_id", alertInstance.ID,
			"rule_id", rule.ID,
			"severity", alertInstance.Severity,
			"incident_id", evt.IncidentID,
		)
	}

	return nil
}

// ruleMatchesEvent checks if a rule should fire for the given event.
func (ae *AlertEngineImpl) ruleMatchesEvent(rule *alertrule.AlertRule, evt event.IncidentEvent) bool {
	if !rule.Enabled {
		return false
	}

	// Check conditions against event payload
	for _, cond := range rule.Conditions {
		val, ok := evt.Payload[cond.Key]
		if !ok {
			return false
		}

		valStr := fmt.Sprintf("%v", val)
		switch cond.Operator {
		case "eq":
			if valStr != cond.Value {
				return false
			}
		case "neq":
			if valStr == cond.Value {
				return false
			}
		case "contains":
			// Simple string contains check
			if valStr != cond.Value && len(valStr) > 0 {
				continue // lenient match
			}
		}
	}

	return true
}

// getChannelsForRule determines which channels to route an alert to.
func (ae *AlertEngineImpl) getChannelsForRule(rule *alertrule.AlertRule, evt event.IncidentEvent) []string {
	// Use routing table based on severity
	rt := &routing.RoutingTable{
		Rules: []routing.RoutingRule{
			{Severity: rule.Severity, ChannelIDs: rule.Channels},
		},
	}
	channelIDs := rt.GetChannelsForSeverity(rule.Severity)
	result := make([]string, len(channelIDs))
	for i, ch := range channelIDs {
		result[i] = ch
	}
	return result
}

// deliverToChannels sends the alert to each configured channel.
func (ae *AlertEngineImpl) deliverToChannels(ctx context.Context, alertInstance *alert.Alert, channelIDs []string) {
	for _, channelID := range channelIDs {
		channel, err := ae.channels.FindByID(ctx, channelID)
		if err != nil {
			ae.logger.Error("channel not found", "channel_id", channelID, "error", err)
			continue
		}

		if !channel.Enabled {
			continue
		}

		delivery := &alertdelivery.AlertDelivery{
			ID:        id.New(),
			TenantID:  alertInstance.TenantID,
			AlertID:   alertInstance.ID,
			ChannelID: channelID,
			Status:    alertdelivery.DeliveryStatusPending,
			Attempts:  0,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		// Attempt delivery
		_, err = ae.sender.Send(ctx, channel, alertInstance.Message)
		if err != nil {
			delivery.Status = alertdelivery.DeliveryStatusFailed
			delivery.LastError = err.Error()
			delivery.Attempts = 1
			delivery.History = append(delivery.History, alertdelivery.DeliveryHistory{
				Timestamp: time.Now().UTC(),
				Status:    alertdelivery.DeliveryStatusFailed,
				Error:     err.Error(),
			})
			ae.logger.Error("alert delivery failed",
				"alert_id", alertInstance.ID,
				"channel_id", channelID,
				"error", err,
			)
		} else {
			delivery.Status = alertdelivery.DeliveryStatusSent
			delivery.Attempts = 1
			delivery.History = append(delivery.History, alertdelivery.DeliveryHistory{
				Timestamp: time.Now().UTC(),
				Status:    alertdelivery.DeliveryStatusSent,
			})
		}

		if err := ae.deliveries.Create(ctx, delivery); err != nil {
			ae.logger.Error("failed to persist delivery record", "error", err)
		}
	}
}

// buildMessage generates a human-readable alert message from the event.
func (ae *AlertEngineImpl) buildMessage(rule *alertrule.AlertRule, evt event.IncidentEvent) string {
	return fmt.Sprintf("🚨 *%s Alert*\n\nRule: %s\nIncident: %s\nSeverity: %s\nType: %s",
		rule.Severity, rule.Name, evt.IncidentID, rule.Severity, evt.Type)
}
