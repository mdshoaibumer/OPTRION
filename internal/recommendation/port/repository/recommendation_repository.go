package repository

import (
	"context"
	"internal/recommendation/domain/recommendation"

	"github.com/google/uuid"
)

type RecommendationRepository interface {
	Create(ctx context.Context, r *recommendation.Recommendation) error
	FindByID(ctx context.Context, id uuid.UUID) (*recommendation.Recommendation, error)
	ListByIncident(ctx context.Context, incidentID uuid.UUID) ([]*recommendation.Recommendation, error)
}
