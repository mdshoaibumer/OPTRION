package repository

import (
	"context"

	"github.com/optrion/optrion/internal/alert/domain/alertchannel"
)

type AlertChannelRepository interface {
	Create(ctx context.Context, c *alertchannel.AlertChannel) error
	Update(ctx context.Context, c *alertchannel.AlertChannel) error
	FindByID(ctx context.Context, id string) (*alertchannel.AlertChannel, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*alertchannel.AlertChannel, error)
}
