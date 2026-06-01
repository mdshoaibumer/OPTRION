package repository

import (
	"context"

	"github.com/optrion/optrion/internal/alert/domain/alertdelivery"
)

type AlertDeliveryRepository interface {
	Create(ctx context.Context, d *alertdelivery.AlertDelivery) error
	Update(ctx context.Context, d *alertdelivery.AlertDelivery) error
	FindByID(ctx context.Context, id string) (*alertdelivery.AlertDelivery, error)
	ListByAlert(ctx context.Context, alertID string) ([]*alertdelivery.AlertDelivery, error)
}
