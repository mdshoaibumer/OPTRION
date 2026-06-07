package port

import (
	"context"

	"github.com/optrion/optrion/internal/tenant/domain"
)

// TenantRepository defines the persistence contract for tenants.
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	List(ctx context.Context, filter TenantFilter) ([]*domain.Tenant, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
}

// TenantFilter defines filtering options for listing tenants.
type TenantFilter struct {
	ID     *string // If set, restrict to this specific tenant ID (for tenant isolation)
	Status *domain.Status
	Limit  int
	Offset int
}

// ProductRepository defines the persistence contract for products.
type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	ListByTenant(ctx context.Context, tenantID string, filter ProductFilter) ([]*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	ExistsBySlug(ctx context.Context, tenantID, slug string) (bool, error)
}

// ProductFilter defines filtering options for listing products.
type ProductFilter struct {
	Status *domain.Status
	Limit  int
	Offset int
}

// EnvironmentRepository defines the persistence contract for environments.
type EnvironmentRepository interface {
	Create(ctx context.Context, env *domain.Environment) error
	GetByID(ctx context.Context, id string) (*domain.Environment, error)
	ListByProduct(ctx context.Context, productID string, filter EnvironmentFilter) ([]*domain.Environment, error)
	Update(ctx context.Context, env *domain.Environment) error
	ExistsBySlug(ctx context.Context, productID, slug string) (bool, error)
}

// EnvironmentFilter defines filtering options for listing environments.
type EnvironmentFilter struct {
	Status *domain.Status
	Tier   *domain.Tier
	Limit  int
	Offset int
}

// ComponentRepository defines the persistence contract for components.
type ComponentRepository interface {
	Create(ctx context.Context, comp *domain.Component) error
	GetByID(ctx context.Context, id string) (*domain.Component, error)
	ListByEnvironment(ctx context.Context, environmentID string, filter ComponentFilter) ([]*domain.Component, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.Component, error)
	Update(ctx context.Context, comp *domain.Component) error
	ExistsBySlug(ctx context.Context, environmentID, slug string) (bool, error)
}

// ComponentFilter defines filtering options for listing components.
type ComponentFilter struct {
	Status *domain.Status
	Kind   *domain.ComponentKind
	Limit  int
	Offset int
}

// AuditRepository defines the persistence contract for audit events.
type AuditRepository interface {
	Create(ctx context.Context, event *domain.AuditEvent) error
	ListByEntity(ctx context.Context, tenantID, entityType, entityID string, limit int) ([]*domain.AuditEvent, error)
	ListByTenant(ctx context.Context, tenantID string, filter AuditFilter) ([]*domain.AuditEvent, int, error)
}

// AuditFilter defines filtering options for listing audit events.
type AuditFilter struct {
	Action     *string
	EntityType *string
	Limit      int
	Offset     int
}

// UnitOfWork provides transaction support across repositories.
type UnitOfWork interface {
	// Begin starts a new transaction and returns a context carrying it.
	Begin(ctx context.Context) (context.Context, error)
	// Commit commits the transaction in the context.
	Commit(ctx context.Context) error
	// Rollback rolls back the transaction in the context.
	Rollback(ctx context.Context) error
}

// HealthCheckRepository defines the persistence contract for health check results.
type HealthCheckRepository interface {
	GetLatestByComponent(ctx context.Context, componentID string) (*domain.HealthCheckResult, error)
	ListByTenant(ctx context.Context, tenantID string, limit int) ([]*domain.HealthCheckResult, error)
}
