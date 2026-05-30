package repository

import (
	"context"
	"internal/alert/domain/alert"

	"github.com/google/uuid"
)

type AlertRepository interface {
	Create(ctx context.Context, a *alert.Alert) error
	Update(ctx context.Context, a *alert.Alert) error
	FindByID(ctx context.Context, id uuid.UUID) (*alert.Alert, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*alert.Alert, error)
}
