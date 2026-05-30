package aianalysis

import (
	"time"

	"github.com/google/uuid"
)

// AIAnalysis represents an immutable AI analysis record for an incident.
type AIAnalysis struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	IncidentID  uuid.UUID
	ContextID   uuid.UUID
	ReportID    uuid.UUID
	Provider    string
	RequestedAt time.Time
	CompletedAt time.Time
	Status      string // requested, completed, failed
	CreatedAt   time.Time
}
