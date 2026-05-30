package rootcausereport

import (
	"time"

	"github.com/google/uuid"
)

// RootCauseReport is the AI's structured output for an incident.
type RootCauseReport struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	IncidentID         uuid.UUID
	LikelyCause        string
	AffectedComponents []string
	Confidence         float64
	InvestigationHints []string
	RawOutput          []byte // Original JSON from provider
	CreatedAt          time.Time
}
