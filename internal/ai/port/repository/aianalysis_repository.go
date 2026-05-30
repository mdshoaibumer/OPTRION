package repository

import (
	"context"
	"internal/ai/domain/aianalysis"

	"github.com/google/uuid"
)

type AIAnalysisRepository interface {
	Create(ctx context.Context, a *aianalysis.AIAnalysis) error
	FindByID(ctx context.Context, id uuid.UUID) (*aianalysis.AIAnalysis, error)
	ListByIncident(ctx context.Context, incidentID uuid.UUID) ([]*aianalysis.AIAnalysis, error)
}
