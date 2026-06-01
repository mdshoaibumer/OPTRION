package repository

import (
	"context"

	"github.com/optrion/optrion/internal/alert/domain/escalationpolicy"
)

type EscalationPolicyRepository interface {
	Create(ctx context.Context, p *escalationpolicy.EscalationPolicy) error
	Update(ctx context.Context, p *escalationpolicy.EscalationPolicy) error
	FindByID(ctx context.Context, id string) (*escalationpolicy.EscalationPolicy, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*escalationpolicy.EscalationPolicy, error)
}
