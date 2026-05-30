package repository

import (
	"context"
	"internal/recommendation/domain/recommendationreport"

	"github.com/google/uuid"
)

type RecommendationReportRepository interface {
	Create(ctx context.Context, r *recommendationreport.RecommendationReport) error
	FindByID(ctx context.Context, id uuid.UUID) (*recommendationreport.RecommendationReport, error)
	ListByIncident(ctx context.Context, incidentID uuid.UUID) ([]*recommendationreport.RecommendationReport, error)
}
