package service

import (
	"context"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/alert/app/event"
	"github.com/optrion/optrion/internal/alert/domain/alert"
	"github.com/optrion/optrion/internal/alert/domain/alertchannel"
	"github.com/optrion/optrion/internal/alert/domain/alertdelivery"
	"github.com/optrion/optrion/internal/alert/domain/alertrule"
	"github.com/optrion/optrion/internal/platform/config"
	"github.com/optrion/optrion/internal/platform/logger"
)

// --- Mock Repositories ---

type mockAlertRuleRepo struct {
	rules []*alertrule.AlertRule
}

func (m *mockAlertRuleRepo) Create(_ context.Context, r *alertrule.AlertRule) error { return nil }
func (m *mockAlertRuleRepo) Update(_ context.Context, r *alertrule.AlertRule) error { return nil }
func (m *mockAlertRuleRepo) FindByID(_ context.Context, id string) (*alertrule.AlertRule, error) {
	for _, r := range m.rules {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, nil
}
func (m *mockAlertRuleRepo) ListByTenant(_ context.Context, _ string) ([]*alertrule.AlertRule, error) {
	return m.rules, nil
}
func (m *mockAlertRuleRepo) ListEnabledByTenant(_ context.Context, _ string) ([]*alertrule.AlertRule, error) {
	var enabled []*alertrule.AlertRule
	for _, r := range m.rules {
		if r.Enabled {
			enabled = append(enabled, r)
		}
	}
	return enabled, nil
}

type mockAlertRepo struct {
	alerts []*alert.Alert
}

func (m *mockAlertRepo) Create(_ context.Context, a *alert.Alert) error {
	m.alerts = append(m.alerts, a)
	return nil
}
func (m *mockAlertRepo) Update(_ context.Context, a *alert.Alert) error { return nil }
func (m *mockAlertRepo) FindByID(_ context.Context, id string) (*alert.Alert, error) {
	return nil, nil
}
func (m *mockAlertRepo) ListByTenant(_ context.Context, _ string) ([]*alert.Alert, error) {
	return m.alerts, nil
}

type mockChannelRepo struct {
	channels map[string]*alertchannel.AlertChannel
}

func (m *mockChannelRepo) Create(_ context.Context, c *alertchannel.AlertChannel) error { return nil }
func (m *mockChannelRepo) Update(_ context.Context, c *alertchannel.AlertChannel) error { return nil }
func (m *mockChannelRepo) FindByID(_ context.Context, id string) (*alertchannel.AlertChannel, error) {
	ch, ok := m.channels[id]
	if !ok {
		return nil, nil
	}
	return ch, nil
}
func (m *mockChannelRepo) ListByTenant(_ context.Context, _ string) ([]*alertchannel.AlertChannel, error) {
	return nil, nil
}

type mockDeliveryRepo struct {
	deliveries []*alertdelivery.AlertDelivery
}

func (m *mockDeliveryRepo) Create(_ context.Context, d *alertdelivery.AlertDelivery) error {
	m.deliveries = append(m.deliveries, d)
	return nil
}
func (m *mockDeliveryRepo) Update(_ context.Context, d *alertdelivery.AlertDelivery) error {
	return nil
}
func (m *mockDeliveryRepo) FindByID(_ context.Context, id string) (*alertdelivery.AlertDelivery, error) {
	return nil, nil
}
func (m *mockDeliveryRepo) ListByAlert(_ context.Context, _ string) ([]*alertdelivery.AlertDelivery, error) {
	return m.deliveries, nil
}

type mockSender struct {
	sent []string
}

func (m *mockSender) Send(_ context.Context, _ *alertchannel.AlertChannel, msg string) (string, error) {
	m.sent = append(m.sent, msg)
	return "delivery-123", nil
}

// --- Tests ---

func TestAlertEngine_ProcessEvent_NoRules(t *testing.T) {
	engine := newTestEngine(&mockAlertRuleRepo{rules: nil}, nil, nil, nil, nil)
	err := engine.ProcessEvent(context.Background(), event.IncidentEvent{
		Type:       event.IncidentOpened,
		TenantID:   "tenant-1",
		IncidentID: "inc-1",
		Payload:    map[string]interface{}{"severity": "critical"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAlertEngine_ProcessEvent_MatchingRule_GeneratesAlert(t *testing.T) {
	alertRepo := &mockAlertRepo{}
	channelRepo := &mockChannelRepo{
		channels: map[string]*alertchannel.AlertChannel{
			"ch-1": {
				ID:       "ch-1",
				TenantID: "tenant-1",
				Type:     alertchannel.ChannelTypeTelegram,
				Name:     "ops-channel",
				Config:   map[string]string{"bot_token": "test", "chat_id": "123"},
				Enabled:  true,
			},
		},
	}
	deliveryRepo := &mockDeliveryRepo{}
	sender := &mockSender{}
	ruleRepo := &mockAlertRuleRepo{
		rules: []*alertrule.AlertRule{
			{
				ID:         "rule-1",
				TenantID:   "tenant-1",
				Name:       "Critical Alerts",
				Severity:   "critical",
				Enabled:    true,
				Conditions: []alertrule.RuleCondition{{Key: "severity", Operator: "eq", Value: "critical"}},
				Channels:   []string{"ch-1"},
			},
		},
	}

	engine := newTestEngine(ruleRepo, alertRepo, channelRepo, deliveryRepo, sender)
	err := engine.ProcessEvent(context.Background(), event.IncidentEvent{
		ID:         "evt-1",
		Type:       event.IncidentOpened,
		TenantID:   "tenant-1",
		IncidentID: "inc-1",
		Payload:    map[string]interface{}{"severity": "critical", "title": "DB down"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(alertRepo.alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alertRepo.alerts))
	}

	a := alertRepo.alerts[0]
	if a.TenantID != "tenant-1" {
		t.Errorf("expected tenant-1, got %s", a.TenantID)
	}
	if a.RuleID != "rule-1" {
		t.Errorf("expected rule-1, got %s", a.RuleID)
	}
	if a.Severity != "critical" {
		t.Errorf("expected critical severity, got %s", a.Severity)
	}
	if a.Status != alert.AlertStatusPending {
		t.Errorf("expected pending status, got %s", a.Status)
	}

	if len(sender.sent) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(sender.sent))
	}
}

func TestAlertEngine_ProcessEvent_DisabledRule_NoAlert(t *testing.T) {
	alertRepo := &mockAlertRepo{}
	ruleRepo := &mockAlertRuleRepo{
		rules: []*alertrule.AlertRule{
			{
				ID:       "rule-1",
				TenantID: "tenant-1",
				Enabled:  false,
			},
		},
	}

	engine := newTestEngine(ruleRepo, alertRepo, nil, nil, nil)
	err := engine.ProcessEvent(context.Background(), event.IncidentEvent{
		Type:       event.IncidentOpened,
		TenantID:   "tenant-1",
		IncidentID: "inc-1",
		Payload:    map[string]interface{}{"severity": "critical"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(alertRepo.alerts) != 0 {
		t.Fatalf("expected 0 alerts for disabled rule, got %d", len(alertRepo.alerts))
	}
}

func TestAlertEngine_ProcessEvent_NonActionableEvent_Ignored(t *testing.T) {
	alertRepo := &mockAlertRepo{}
	ruleRepo := &mockAlertRuleRepo{
		rules: []*alertrule.AlertRule{
			{
				ID:       "rule-1",
				TenantID: "tenant-1",
				Enabled:  true,
			},
		},
	}

	engine := newTestEngine(ruleRepo, alertRepo, nil, nil, nil)
	// Resolved events are not actionable
	err := engine.ProcessEvent(context.Background(), event.IncidentEvent{
		Type:       event.IncidentResolved,
		TenantID:   "tenant-1",
		IncidentID: "inc-1",
		Payload:    map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(alertRepo.alerts) != 0 {
		t.Fatalf("expected 0 alerts for non-actionable event, got %d", len(alertRepo.alerts))
	}
}

func TestAlertEngine_ProcessEvent_Deduplication(t *testing.T) {
	alertRepo := &mockAlertRepo{}
	channelRepo := &mockChannelRepo{
		channels: map[string]*alertchannel.AlertChannel{
			"ch-1": {ID: "ch-1", Enabled: true, Config: map[string]string{}},
		},
	}
	sender := &mockSender{}
	ruleRepo := &mockAlertRuleRepo{
		rules: []*alertrule.AlertRule{
			{
				ID:         "rule-1",
				TenantID:   "tenant-1",
				Enabled:    true,
				Severity:   "critical",
				Conditions: []alertrule.RuleCondition{},
				Channels:   []string{"ch-1"},
			},
		},
	}

	engine := newTestEngine(ruleRepo, alertRepo, channelRepo, &mockDeliveryRepo{}, sender)

	evt := event.IncidentEvent{
		Type:       event.IncidentOpened,
		TenantID:   "tenant-1",
		IncidentID: "inc-1",
		Payload:    map[string]interface{}{},
	}

	// First call should create alert
	_ = engine.ProcessEvent(context.Background(), evt)
	if len(alertRepo.alerts) != 1 {
		t.Fatalf("expected 1 alert after first call, got %d", len(alertRepo.alerts))
	}

	// Second call should be deduplicated
	_ = engine.ProcessEvent(context.Background(), evt)
	if len(alertRepo.alerts) != 1 {
		t.Fatalf("expected 1 alert after dedup, got %d", len(alertRepo.alerts))
	}
}

func TestAlertEngine_ProcessEvent_ConditionMismatch_NoAlert(t *testing.T) {
	alertRepo := &mockAlertRepo{}
	ruleRepo := &mockAlertRuleRepo{
		rules: []*alertrule.AlertRule{
			{
				ID:         "rule-1",
				TenantID:   "tenant-1",
				Enabled:    true,
				Severity:   "critical",
				Conditions: []alertrule.RuleCondition{{Key: "severity", Operator: "eq", Value: "critical"}},
				Channels:   []string{"ch-1"},
			},
		},
	}

	engine := newTestEngine(ruleRepo, alertRepo, nil, nil, nil)
	err := engine.ProcessEvent(context.Background(), event.IncidentEvent{
		Type:       event.IncidentOpened,
		TenantID:   "tenant-1",
		IncidentID: "inc-1",
		Payload:    map[string]interface{}{"severity": "warning"}, // doesn't match "critical"
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(alertRepo.alerts) != 0 {
		t.Fatalf("expected 0 alerts for mismatched condition, got %d", len(alertRepo.alerts))
	}
}

func TestAlertEngine_ProcessEvent_DisabledChannel_NoDelivery(t *testing.T) {
	alertRepo := &mockAlertRepo{}
	channelRepo := &mockChannelRepo{
		channels: map[string]*alertchannel.AlertChannel{
			"ch-1": {ID: "ch-1", Enabled: false, Config: map[string]string{}},
		},
	}
	sender := &mockSender{}
	ruleRepo := &mockAlertRuleRepo{
		rules: []*alertrule.AlertRule{
			{
				ID:         "rule-1",
				TenantID:   "tenant-1",
				Enabled:    true,
				Severity:   "critical",
				Conditions: []alertrule.RuleCondition{},
				Channels:   []string{"ch-1"},
			},
		},
	}

	engine := newTestEngine(ruleRepo, alertRepo, channelRepo, &mockDeliveryRepo{}, sender)
	_ = engine.ProcessEvent(context.Background(), event.IncidentEvent{
		Type:       event.IncidentOpened,
		TenantID:   "tenant-1",
		IncidentID: "inc-2",
		Payload:    map[string]interface{}{},
	})

	// Alert should be created but no delivery
	if len(alertRepo.alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alertRepo.alerts))
	}
	if len(sender.sent) != 0 {
		t.Fatalf("expected 0 sends for disabled channel, got %d", len(sender.sent))
	}
}

func newTestEngine(
	rules *mockAlertRuleRepo,
	alerts *mockAlertRepo,
	channels *mockChannelRepo,
	deliveries *mockDeliveryRepo,
	sender *mockSender,
) *AlertEngineImpl {
	log := logger.New(config.LogConfig{Level: "error", Format: "text"})

	if alerts == nil {
		alerts = &mockAlertRepo{}
	}
	if channels == nil {
		channels = &mockChannelRepo{channels: map[string]*alertchannel.AlertChannel{}}
	}
	if deliveries == nil {
		deliveries = &mockDeliveryRepo{}
	}
	if sender == nil {
		sender = &mockSender{}
	}

	return NewAlertEngine(
		rules, alerts, channels, deliveries,
		NewDeduplicationService(5*time.Minute),
		sender, log,
	)
}
