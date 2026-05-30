package repository

import (
	"context"
	"internal/alert/domain/notificationtemplate"

	"github.com/google/uuid"
)

type NotificationTemplateRepository interface {
	Create(ctx context.Context, t *notificationtemplate.NotificationTemplate) error
	Update(ctx context.Context, t *notificationtemplate.NotificationTemplate) error
	FindByID(ctx context.Context, id uuid.UUID) (*notificationtemplate.NotificationTemplate, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*notificationtemplate.NotificationTemplate, error)
}
