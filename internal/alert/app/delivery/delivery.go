package delivery

import (
	"context"

	"github.com/google/uuid"
)

// DeliveryService handles alert delivery to channels.
type DeliveryService interface {
	Deliver(ctx context.Context, alertID uuid.UUID, channelID uuid.UUID) error
}
