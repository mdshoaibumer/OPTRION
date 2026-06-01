package repository

import (
	"context"
	"github.com/optrion/optrion/internal/ai/domain/rootcausereport"

	"github.com/google/uuid"
)

type RootCauseReportRepository interface {
	Create(ctx context.Context, r *rootcausereport.RootCauseReport) error
	FindByID(ctx context.Context, id uuid.UUID) (*rootcausereport.RootCauseReport, error)
	ListByIncident(ctx context.Context, incidentID uuid.UUID) ([]*rootcausereport.RootCauseReport, error)
}
