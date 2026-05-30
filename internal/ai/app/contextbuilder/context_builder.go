package contextbuilder

import (
	"github.com/google/uuid"
)

// ContextBuilder constructs minimal context for AI analysis.
type ContextBuilder interface {
	Build(incidentID uuid.UUID) ([]byte, error)
}
