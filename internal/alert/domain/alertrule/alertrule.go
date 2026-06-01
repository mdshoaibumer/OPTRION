package alertrule

import (
	"time"
)

// AlertRule defines the conditions for generating alerts.
type AlertRule struct {
	ID                 string
	TenantID           string
	Name               string
	Description        string
	Severity           string // Info, Warning, Minor, Major, Critical
	Enabled            bool
	Conditions         []RuleCondition
	Channels           []string // AlertChannel IDs
	EscalationPolicyID string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CreatedBy          string
	UpdatedBy          string
}

type RuleCondition struct {
	Key      string
	Operator string
	Value    string
}
