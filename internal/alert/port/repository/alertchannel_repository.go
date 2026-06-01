package repository

import (
	"context"
	"github.com/optrion/optrion/internal/alert/domain/alertchannel"

	"github.com/google/uuid"
)

type AlertChannelRepository interface {
	Create(ctx context.Context, c *alertchannel.AlertChannel) error
	Update(ctx context.Context, c *alertchannel.AlertChannel) error
	FindByID(ctx context.Context, id uuid.UUID) (*alertchannel.AlertChannel, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*alertchannel.AlertChannel, error)
}
