package escalationpolicy

import (
	"time"

	"github.com/google/uuid"
)

// EscalationPolicy defines how and when to escalate alerts.
type EscalationPolicy struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	Steps       []EscalationStep
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

type EscalationStep struct {
	DelayMinutes int
	ChannelIDs   []uuid.UUID
	Reminder     bool
}
