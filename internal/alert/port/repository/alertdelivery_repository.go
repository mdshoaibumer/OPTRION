package repository

import (
	"context"
	"github.com/optrion/optrion/internal/alert/domain/alertdelivery"

	"github.com/google/uuid"
)

type AlertDeliveryRepository interface {
	Create(ctx context.Context, d *alertdelivery.AlertDelivery) error
	Update(ctx context.Context, d *alertdelivery.AlertDelivery) error
	FindByID(ctx context.Context, id uuid.UUID) (*alertdelivery.AlertDelivery, error)
	ListByAlert(ctx context.Context, alertID uuid.UUID) ([]*alertdelivery.AlertDelivery, error)
}
