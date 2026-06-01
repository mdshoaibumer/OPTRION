package delivery

import (
	"context"
)

// DeliveryService handles alert delivery to channels.
type DeliveryService interface {
	Deliver(ctx context.Context, alertID string, channelID string) error
}
