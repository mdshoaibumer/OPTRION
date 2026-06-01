package repository

import (
	"context"

	"github.com/optrion/optrion/internal/alert/domain/alertrule"
)

type AlertRuleRepository interface {
	Create(ctx context.Context, r *alertrule.AlertRule) error
	Update(ctx context.Context, r *alertrule.AlertRule) error
	FindByID(ctx context.Context, id string) (*alertrule.AlertRule, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*alertrule.AlertRule, error)
	ListEnabledByTenant(ctx context.Context, tenantID string) ([]*alertrule.AlertRule, error)
}
