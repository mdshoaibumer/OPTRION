package alertrule_test

import (
	"testing"
	"time"

	"github.com/optrion/optrion/internal/alert/domain/alertrule"
)

func TestAlertRule_Structure(t *testing.T) {
	now := time.Now()
	rule := alertrule.AlertRule{
		ID:                 "r-1",
		TenantID:           "t-1",
		Name:               "High CPU Alert",
		Description:        "Trigger alert when CPU > 90%",
		Severity:           "critical",
		Enabled:            true,
		Conditions:         []alertrule.RuleCondition{{Key: "cpu_usage", Operator: "gt", Value: "90"}},
		Channels:           []string{"ch-1", "ch-2"},
		EscalationPolicyID: "ep-1",
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          "admin",
	}

	if rule.ID != "r-1" {
		t.Errorf("expected ID r-1, got %s", rule.ID)
	}
	if rule.Name != "High CPU Alert" {
		t.Errorf("expected name, got %s", rule.Name)
	}
	if !rule.Enabled {
		t.Error("expected rule to be enabled")
	}
	if len(rule.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(rule.Conditions))
	}
	if rule.Conditions[0].Operator != "gt" {
		t.Errorf("expected operator gt, got %s", rule.Conditions[0].Operator)
	}
	if len(rule.Channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(rule.Channels))
	}
}

func TestRuleCondition_Fields(t *testing.T) {
	conditions := []alertrule.RuleCondition{
		{Key: "cpu_usage", Operator: "gt", Value: "90"},
		{Key: "status", Operator: "eq", Value: "open"},
		{Key: "severity", Operator: "contains", Value: "critical"},
	}

	tests := []struct {
		name    string
		cond    alertrule.RuleCondition
		wantKey string
		wantOp  string
		wantVal string
	}{
		{"greater than", conditions[0], "cpu_usage", "gt", "90"},
		{"equals", conditions[1], "status", "eq", "open"},
		{"contains", conditions[2], "severity", "contains", "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cond.Key != tt.wantKey {
				t.Errorf("expected key %s, got %s", tt.wantKey, tt.cond.Key)
			}
			if tt.cond.Operator != tt.wantOp {
				t.Errorf("expected operator %s, got %s", tt.wantOp, tt.cond.Operator)
			}
			if tt.cond.Value != tt.wantVal {
				t.Errorf("expected value %s, got %s", tt.wantVal, tt.cond.Value)
			}
		})
	}
}
