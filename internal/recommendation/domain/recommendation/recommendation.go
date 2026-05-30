package recommendation

import (
	"time"

	"github.com/google/uuid"
)

// Recommendation is an immutable engineering recommendation for an incident.
type Recommendation struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	IncidentID  uuid.UUID
	ReportID    uuid.UUID
	Category    string
	Priority    string
	Title       string
	Description string
	Confidence  float64
	RiskLevel   string
	EvidenceIDs []uuid.UUID
	CreatedAt   time.Time
}
