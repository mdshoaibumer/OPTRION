package port

import (
	"context"

	"github.com/optrion/optrion/internal/registration/domain"
)

// RegistrationRepository defines the storage interface for registration data.
type RegistrationRepository interface {
	CreateAudit(ctx context.Context, audit *domain.RegistrationAudit) error
	GetAuditByID(ctx context.Context, id string) (*domain.RegistrationAudit, error)
	ListAuditsByTenant(ctx context.Context, tenantID string) ([]*domain.RegistrationAudit, error)
}

// APIKeyGenerator defines the interface for generating API keys.
type APIKeyGenerator interface {
	Generate(tenantID string) (string, error)
}
