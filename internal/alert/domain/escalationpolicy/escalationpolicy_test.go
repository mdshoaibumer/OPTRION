package escalationpolicy_test

import (
	"testing"
	"time"

	"github.com/optrion/optrion/internal/alert/domain/escalationpolicy"
)

func TestEscalationPolicy_Structure(t *testing.T) {
	policy := escalationpolicy.EscalationPolicy{
		ID:          "ep-1",
		TenantID:    "t-1",
		Name:        "Critical SLA",
		Description: "Escalate critical alerts within 15 minutes",
		Steps: []escalationpolicy.EscalationStep{
			{DelayMinutes: 5, ChannelIDs: []string{"ch-telegram"}, Reminder: false},
			{DelayMinutes: 10, ChannelIDs: []string{"ch-email", "ch-telegram"}, Reminder: true},
			{DelayMinutes: 15, ChannelIDs: []string{"ch-pagerduty"}, Reminder: true},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "admin",
	}

	if policy.ID != "ep-1" {
		t.Errorf("expected ID ep-1, got %s", policy.ID)
	}
	if policy.Name != "Critical SLA" {
		t.Errorf("expected name, got %s", policy.Name)
	}
	if len(policy.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(policy.Steps))
	}
}

func TestEscalationStep_DelayOrdering(t *testing.T) {
	steps := []escalationpolicy.EscalationStep{
		{DelayMinutes: 5, ChannelIDs: []string{"ch-1"}},
		{DelayMinutes: 15, ChannelIDs: []string{"ch-1", "ch-2"}},
		{DelayMinutes: 30, ChannelIDs: []string{"ch-3"}},
	}

	for i := 1; i < len(steps); i++ {
		if steps[i].DelayMinutes <= steps[i-1].DelayMinutes {
			t.Errorf("step %d delay (%d) should be greater than step %d delay (%d)",
				i, steps[i].DelayMinutes, i-1, steps[i-1].DelayMinutes)
		}
	}
}

func TestEscalationStep_ReminderFlag(t *testing.T) {
	step := escalationpolicy.EscalationStep{
		DelayMinutes: 10,
		ChannelIDs:   []string{"ch-1"},
		Reminder:     true,
	}

	if !step.Reminder {
		t.Error("expected reminder to be true")
	}
	if step.DelayMinutes != 10 {
		t.Errorf("expected 10 minutes delay, got %d", step.DelayMinutes)
	}
}

func TestEscalationPolicy_EmptySteps(t *testing.T) {
	policy := escalationpolicy.EscalationPolicy{
		ID:       "ep-empty",
		TenantID: "t-1",
		Name:     "Empty Policy",
		Steps:    nil,
	}

	if len(policy.Steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(policy.Steps))
	}
}
