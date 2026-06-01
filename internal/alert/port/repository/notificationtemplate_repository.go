package repository

import (
	"context"

	"github.com/optrion/optrion/internal/alert/domain/notificationtemplate"
)

type NotificationTemplateRepository interface {
	Create(ctx context.Context, t *notificationtemplate.NotificationTemplate) error
	Update(ctx context.Context, t *notificationtemplate.NotificationTemplate) error
	FindByID(ctx context.Context, id string) (*notificationtemplate.NotificationTemplate, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*notificationtemplate.NotificationTemplate, error)
}
