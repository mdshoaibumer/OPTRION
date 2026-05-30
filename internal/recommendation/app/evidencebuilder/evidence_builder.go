package evidencebuilder

import (
	"github.com/google/uuid"
)

// EvidenceBuilder constructs evidence for recommendations.
type EvidenceBuilder interface {
	Build(incidentID uuid.UUID, recommendationID uuid.UUID) ([]string, error)
}
