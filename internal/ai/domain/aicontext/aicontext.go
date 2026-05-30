package aicontext

import (
	"time"

	"github.com/google/uuid"
)

// AIContext is a snapshot of the context sent to the AI provider.
type AIContext struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	IncidentID uuid.UUID
	Snapshot   []byte // JSON or compressed
	CreatedAt  time.Time
}
