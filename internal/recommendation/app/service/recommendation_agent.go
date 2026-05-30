package service

import (
	"context"

	"github.com/google/uuid"
)

// RecommendationAgent analyzes incidents and produces actionable recommendations.
type RecommendationAgent interface {
	Recommend(ctx context.Context, incidentID uuid.UUID) error
}
