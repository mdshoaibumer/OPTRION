package alertrule

import (
	"time"

	"github.com/google/uuid"
)

// AlertRule defines the conditions for generating alerts.
type AlertRule struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	Name               string
	Description        string
	Severity           string // Info, Warning, Minor, Major, Critical
	Enabled            bool
	Conditions         []RuleCondition
	Channels           []uuid.UUID // AlertChannel IDs
	EscalationPolicyID uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CreatedBy          uuid.UUID
	UpdatedBy          uuid.UUID
}

type RuleCondition struct {
	Key      string
	Operator string
	Value    string
}
