package repository

import (
	"context"
	"github.com/optrion/optrion/internal/alert/domain/escalationpolicy"

	"github.com/google/uuid"
)

type EscalationPolicyRepository interface {
	Create(ctx context.Context, p *escalationpolicy.EscalationPolicy) error
	Update(ctx context.Context, p *escalationpolicy.EscalationPolicy) error
	FindByID(ctx context.Context, id uuid.UUID) (*escalationpolicy.EscalationPolicy, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*escalationpolicy.EscalationPolicy, error)
}
