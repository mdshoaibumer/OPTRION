package recommendationevidence

import (
	"time"

	"github.com/google/uuid"
)

// RecommendationEvidence is supporting evidence for a recommendation.
type RecommendationEvidence struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	IncidentID       uuid.UUID
	RecommendationID uuid.UUID
	Description      string
	CreatedAt        time.Time
}
