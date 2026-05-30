package repository

import (
	"context"
	"internal/recommendation/domain/recommendationevidence"

	"github.com/google/uuid"
)

type RecommendationEvidenceRepository interface {
	Create(ctx context.Context, e *recommendationevidence.RecommendationEvidence) error
	FindByID(ctx context.Context, id uuid.UUID) (*recommendationevidence.RecommendationEvidence, error)
	ListByRecommendation(ctx context.Context, recommendationID uuid.UUID) ([]*recommendationevidence.RecommendationEvidence, error)
}
