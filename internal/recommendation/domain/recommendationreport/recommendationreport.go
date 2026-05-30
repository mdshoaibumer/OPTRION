package recommendationreport

import (
	"time"

	"github.com/google/uuid"
)

// RecommendationReport is a set of recommendations for an incident.
type RecommendationReport struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	IncidentID      uuid.UUID
	Recommendations []uuid.UUID
	CreatedAt       time.Time
}
