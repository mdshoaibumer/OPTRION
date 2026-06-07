package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/optrion/optrion/internal/tenant/domain"
	"github.com/optrion/optrion/internal/tenant/port"
)

// TenantService orchestrates tenant-related use cases.
type TenantService struct {
	tenants      port.TenantRepository
	products     port.ProductRepository
	environments port.EnvironmentRepository
	components   port.ComponentRepository
	audit        port.AuditRepository
	uow          port.UnitOfWork
	logger       *slog.Logger
}

// NewTenantService creates a new TenantService with all dependencies.
func NewTenantService(
	tenants port.TenantRepository,
	products port.ProductRepository,
	environments port.EnvironmentRepository,
	components port.ComponentRepository,
	audit port.AuditRepository,
	uow port.UnitOfWork,
	logger *slog.Logger,
) *TenantService {
	return &TenantService{
		tenants:      tenants,
		products:     products,
		environments: environments,
		components:   components,
		audit:        audit,
		uow:          uow,
		logger:       logger,
	}
}

// --- Commands ---

// CreateTenantCmd holds the data needed to create a tenant.
type CreateTenantCmd struct {
	Name string
	Slug string
	Plan string
}

// CreateTenant registers a new tenant in the platform.
func (s *TenantService) CreateTenant(ctx context.Context, cmd CreateTenantCmd) (*domain.Tenant, error) {
	slug, err := domain.NewSlug(cmd.Slug)
	if err != nil {
		return nil, fmt.Errorf("invalid slug: %w", err)
	}

	plan := domain.Plan(cmd.Plan)
	if !plan.IsValid() {
		return nil, fmt.Errorf("invalid plan: %s", cmd.Plan)
	}

	// Check slug uniqueness
	exists, err := s.tenants.ExistsBySlug(ctx, slug.String())
	if err != nil {
		return nil, fmt.Errorf("checking slug uniqueness: %w", err)
	}
	if exists {
		return nil, domain.ErrTenantSlugTaken{Slug: slug.String()}
	}

	tenant, err := domain.NewTenant(cmd.Name, slug, plan)
	if err != nil {
		return nil, fmt.Errorf("creating tenant: %w", err)
	}

	// Begin transaction
	txCtx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer s.uow.Rollback(txCtx) //nolint:errcheck

	if err := s.tenants.Create(txCtx, tenant); err != nil {
		return nil, err
	}

	// Audit
	audit := domain.NewAuditEvent(tenant.ID, "system", "tenant.created", "tenant", tenant.ID, map[string]interface{}{
		"name": tenant.Name,
		"slug": tenant.Slug.String(),
		"plan": string(tenant.Plan),
	})
	if err := s.audit.Create(txCtx, audit); err != nil {
		s.logger.WarnContext(ctx, "failed to create audit event", "error", err)
	}

	if err := s.uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	s.logger.InfoContext(ctx, "tenant created",
		"tenant_id", tenant.ID,
		"name", tenant.Name,
		"slug", tenant.Slug.String(),
	)

	return tenant, nil
}

// CreateProductCmd holds the data needed to create a product.
type CreateProductCmd struct {
	TenantID    string
	Name        string
	Slug        string
	Description string
}

// CreateProduct registers a new product under a tenant.
func (s *TenantService) CreateProduct(ctx context.Context, cmd CreateProductCmd) (*domain.Product, error) {
	// Verify tenant exists and is active
	tenant, err := s.tenants.GetByID(ctx, cmd.TenantID)
	if err != nil {
		return nil, err
	}
	if !tenant.IsActive() {
		return nil, domain.ErrTenantInactive{ID: tenant.ID, Status: tenant.Status}
	}

	slug, err := domain.NewSlug(cmd.Slug)
	if err != nil {
		return nil, fmt.Errorf("invalid slug: %w", err)
	}

	// Check slug uniqueness within tenant
	exists, err := s.products.ExistsBySlug(ctx, cmd.TenantID, slug.String())
	if err != nil {
		return nil, fmt.Errorf("checking slug uniqueness: %w", err)
	}
	if exists {
		return nil, domain.ErrProductSlugTaken{TenantID: cmd.TenantID, Slug: slug.String()}
	}

	product, err := domain.NewProduct(cmd.TenantID, cmd.Name, slug, cmd.Description)
	if err != nil {
		return nil, fmt.Errorf("creating product: %w", err)
	}

	txCtx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer s.uow.Rollback(txCtx) //nolint:errcheck

	if err := s.products.Create(txCtx, product); err != nil {
		return nil, err
	}

	audit := domain.NewAuditEvent(cmd.TenantID, "system", "product.created", "product", product.ID, map[string]interface{}{
		"name": product.Name,
		"slug": product.Slug.String(),
	})
	if err := s.audit.Create(txCtx, audit); err != nil {
		s.logger.WarnContext(ctx, "failed to create audit event", "error", err)
	}

	if err := s.uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	s.logger.InfoContext(ctx, "product created",
		"tenant_id", cmd.TenantID,
		"product_id", product.ID,
		"name", product.Name,
	)

	return product, nil
}

// CreateEnvironmentCmd holds the data needed to create an environment.
type CreateEnvironmentCmd struct {
	TenantID  string
	ProductID string
	Name      string
	Slug      string
	Tier      string
}

// CreateEnvironment registers a new environment under a product.
func (s *TenantService) CreateEnvironment(ctx context.Context, cmd CreateEnvironmentCmd) (*domain.Environment, error) {
	// Verify tenant is active
	tenant, err := s.tenants.GetByID(ctx, cmd.TenantID)
	if err != nil {
		return nil, err
	}
	if !tenant.IsActive() {
		return nil, domain.ErrTenantInactive{ID: tenant.ID, Status: tenant.Status}
	}

	// Verify product exists and belongs to tenant
	product, err := s.products.GetByID(ctx, cmd.ProductID)
	if err != nil {
		return nil, err
	}
	if product.TenantID != cmd.TenantID {
		return nil, domain.ErrProductNotFound{ID: cmd.ProductID}
	}

	slug, err := domain.NewSlug(cmd.Slug)
	if err != nil {
		return nil, fmt.Errorf("invalid slug: %w", err)
	}

	tier := domain.Tier(cmd.Tier)
	if !tier.IsValid() {
		return nil, fmt.Errorf("invalid tier: %s", cmd.Tier)
	}

	// Check slug uniqueness within product
	exists, err := s.environments.ExistsBySlug(ctx, cmd.ProductID, slug.String())
	if err != nil {
		return nil, fmt.Errorf("checking slug uniqueness: %w", err)
	}
	if exists {
		return nil, domain.ErrEnvironmentSlugTaken{ProductID: cmd.ProductID, Slug: slug.String()}
	}

	env, err := domain.NewEnvironment(cmd.TenantID, cmd.ProductID, cmd.Name, slug, tier)
	if err != nil {
		return nil, fmt.Errorf("creating environment: %w", err)
	}

	txCtx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer s.uow.Rollback(txCtx) //nolint:errcheck

	if err := s.environments.Create(txCtx, env); err != nil {
		return nil, err
	}

	audit := domain.NewAuditEvent(cmd.TenantID, "system", "environment.created", "environment", env.ID, map[string]interface{}{
		"name":       env.Name,
		"slug":       env.Slug.String(),
		"tier":       string(env.Tier),
		"product_id": cmd.ProductID,
	})
	if err := s.audit.Create(txCtx, audit); err != nil {
		s.logger.WarnContext(ctx, "failed to create audit event", "error", err)
	}

	if err := s.uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	s.logger.InfoContext(ctx, "environment created",
		"tenant_id", cmd.TenantID,
		"product_id", cmd.ProductID,
		"environment_id", env.ID,
		"name", env.Name,
	)

	return env, nil
}

// RegisterComponentCmd holds the data needed to register a component.
type RegisterComponentCmd struct {
	TenantID      string
	ProductID     string
	EnvironmentID string
	Name          string
	Slug          string
	Kind          string
	EndpointURL   string
}

// RegisterComponent registers a new component within an environment.
func (s *TenantService) RegisterComponent(ctx context.Context, cmd RegisterComponentCmd) (*domain.Component, error) {
	// Verify tenant is active
	tenant, err := s.tenants.GetByID(ctx, cmd.TenantID)
	if err != nil {
		return nil, err
	}
	if !tenant.IsActive() {
		return nil, domain.ErrTenantInactive{ID: tenant.ID, Status: tenant.Status}
	}

	// Verify environment exists and belongs to product/tenant
	env, err := s.environments.GetByID(ctx, cmd.EnvironmentID)
	if err != nil {
		return nil, err
	}
	if env.TenantID != cmd.TenantID || env.ProductID != cmd.ProductID {
		return nil, domain.ErrEnvironmentNotFound{ID: cmd.EnvironmentID}
	}

	slug, err := domain.NewSlug(cmd.Slug)
	if err != nil {
		return nil, fmt.Errorf("invalid slug: %w", err)
	}

	kind := domain.ComponentKind(cmd.Kind)
	if !kind.IsValid() {
		return nil, fmt.Errorf("invalid component kind: %s", cmd.Kind)
	}

	// Check slug uniqueness within environment
	exists, err := s.components.ExistsBySlug(ctx, cmd.EnvironmentID, slug.String())
	if err != nil {
		return nil, fmt.Errorf("checking slug uniqueness: %w", err)
	}
	if exists {
		return nil, domain.ErrComponentSlugTaken{EnvironmentID: cmd.EnvironmentID, Slug: slug.String()}
	}

	comp, err := domain.NewComponent(cmd.TenantID, cmd.ProductID, cmd.EnvironmentID, cmd.Name, slug, kind, cmd.EndpointURL)
	if err != nil {
		return nil, fmt.Errorf("creating component: %w", err)
	}

	txCtx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer s.uow.Rollback(txCtx) //nolint:errcheck

	if err := s.components.Create(txCtx, comp); err != nil {
		return nil, err
	}

	audit := domain.NewAuditEvent(cmd.TenantID, "system", "component.registered", "component", comp.ID, map[string]interface{}{
		"name":           comp.Name,
		"slug":           comp.Slug.String(),
		"kind":           string(comp.Kind),
		"environment_id": cmd.EnvironmentID,
	})
	if err := s.audit.Create(txCtx, audit); err != nil {
		s.logger.WarnContext(ctx, "failed to create audit event", "error", err)
	}

	if err := s.uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	s.logger.InfoContext(ctx, "component registered",
		"tenant_id", cmd.TenantID,
		"environment_id", cmd.EnvironmentID,
		"component_id", comp.ID,
		"name", comp.Name,
		"kind", string(comp.Kind),
	)

	return comp, nil
}

// --- Queries ---

// GetTenant retrieves a tenant by ID.
func (s *TenantService) GetTenant(ctx context.Context, id string) (*domain.Tenant, error) {
	return s.tenants.GetByID(ctx, id)
}

// ListTenants retrieves tenants matching the filter.
func (s *TenantService) ListTenants(ctx context.Context, filter port.TenantFilter) ([]*domain.Tenant, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.tenants.List(ctx, filter)
}

// ListProducts retrieves products for a tenant.
func (s *TenantService) ListProducts(ctx context.Context, tenantID string, filter port.ProductFilter) ([]*domain.Product, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.products.ListByTenant(ctx, tenantID, filter)
}

// ListEnvironments retrieves environments for a product.
func (s *TenantService) ListEnvironments(ctx context.Context, productID string, filter port.EnvironmentFilter) ([]*domain.Environment, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.environments.ListByProduct(ctx, productID, filter)
}

// ListComponents retrieves components for an environment.
func (s *TenantService) ListComponents(ctx context.Context, environmentID string, filter port.ComponentFilter) ([]*domain.Component, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.components.ListByEnvironment(ctx, environmentID, filter)
}

// CreateComponentCmd holds data for creating a component (used by registration workflow).
type CreateComponentCmd struct {
	TenantID      string
	ProductID     string
	EnvironmentID string
	Name          string
	Kind          string
	Description   string
	Endpoint      string
	Port          int
}

// CreateComponent creates a component using a simplified command (delegates to RegisterComponent).
func (s *TenantService) CreateComponent(ctx context.Context, cmd CreateComponentCmd) (*domain.Component, error) {
	// If TenantID/ProductID not provided, look them up from environment
	tenantID := cmd.TenantID
	productID := cmd.ProductID
	if tenantID == "" || productID == "" {
		env, err := s.environments.GetByID(ctx, cmd.EnvironmentID)
		if err != nil {
			return nil, fmt.Errorf("looking up environment: %w", err)
		}
		if tenantID == "" {
			tenantID = env.TenantID
		}
		if productID == "" {
			productID = env.ProductID
		}
	}

	// Generate slug from name
	slugStr := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(cmd.Name), " ", "-"))

	return s.RegisterComponent(ctx, RegisterComponentCmd{
		TenantID:      tenantID,
		ProductID:     productID,
		EnvironmentID: cmd.EnvironmentID,
		Name:          cmd.Name,
		Slug:          slugStr,
		Kind:          cmd.Kind,
		EndpointURL:   cmd.Endpoint,
	})
}

// ListAuditEvents retrieves paginated audit events for a tenant.
func (s *TenantService) ListAuditEvents(ctx context.Context, tenantID string, filter port.AuditFilter) ([]*domain.AuditEvent, int, error) {
	return s.audit.ListByTenant(ctx, tenantID, filter)
}
