package escalationpolicy

import (
	"time"
)

// EscalationPolicy defines how and when to escalate alerts.
type EscalationPolicy struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Steps       []EscalationStep
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
	UpdatedBy   string
}

type EscalationStep struct {
	DelayMinutes int
	ChannelIDs   []string
	Reminder     bool
}
