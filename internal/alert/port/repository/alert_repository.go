package repository

import (
	"context"

	"github.com/optrion/optrion/internal/alert/domain/alert"
)

type AlertRepository interface {
	Create(ctx context.Context, a *alert.Alert) error
	Update(ctx context.Context, a *alert.Alert) error
	FindByID(ctx context.Context, id string) (*alert.Alert, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*alert.Alert, error)
}
