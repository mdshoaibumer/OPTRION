package repository

import (
	"context"
	"github.com/optrion/optrion/internal/alert/domain/alertrule"

	"github.com/google/uuid"
)

type AlertRuleRepository interface {
	Create(ctx context.Context, r *alertrule.AlertRule) error
	Update(ctx context.Context, r *alertrule.AlertRule) error
	FindByID(ctx context.Context, id uuid.UUID) (*alertrule.AlertRule, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*alertrule.AlertRule, error)
}
