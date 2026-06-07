package app_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/optrion/optrion/internal/tenant/app"
	"github.com/optrion/optrion/internal/tenant/domain"
	"github.com/optrion/optrion/internal/tenant/port"
)

// --- Mock Repositories ---

type mockTenantRepo struct {
	tenants       map[string]*domain.Tenant
	slugs         map[string]bool
	createErr     error
	getByIDErr    error
	existsSlugErr error
}

func newMockTenantRepo() *mockTenantRepo {
	return &mockTenantRepo{
		tenants: make(map[string]*domain.Tenant),
		slugs:   make(map[string]bool),
	}
}

func (m *mockTenantRepo) Create(_ context.Context, t *domain.Tenant) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.tenants[t.ID] = t
	m.slugs[t.Slug.String()] = true
	return nil
}

func (m *mockTenantRepo) GetByID(_ context.Context, id string) (*domain.Tenant, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	t, ok := m.tenants[id]
	if !ok {
		return nil, domain.ErrTenantNotFound{ID: id}
	}
	return t, nil
}

func (m *mockTenantRepo) GetBySlug(_ context.Context, slug string) (*domain.Tenant, error) {
	for _, t := range m.tenants {
		if t.Slug.String() == slug {
			return t, nil
		}
	}
	return nil, domain.ErrTenantNotFound{ID: slug}
}

func (m *mockTenantRepo) List(_ context.Context, _ port.TenantFilter) ([]*domain.Tenant, error) {
	result := make([]*domain.Tenant, 0, len(m.tenants))
	for _, t := range m.tenants {
		result = append(result, t)
	}
	return result, nil
}

func (m *mockTenantRepo) Update(_ context.Context, t *domain.Tenant) error {
	m.tenants[t.ID] = t
	return nil
}

func (m *mockTenantRepo) ExistsBySlug(_ context.Context, slug string) (bool, error) {
	if m.existsSlugErr != nil {
		return false, m.existsSlugErr
	}
	return m.slugs[slug], nil
}

type mockProductRepo struct {
	products map[string]*domain.Product
	slugs    map[string]bool
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{
		products: make(map[string]*domain.Product),
		slugs:    make(map[string]bool),
	}
}

func (m *mockProductRepo) Create(_ context.Context, p *domain.Product) error {
	m.products[p.ID] = p
	m.slugs[p.TenantID+":"+p.Slug.String()] = true
	return nil
}

func (m *mockProductRepo) GetByID(_ context.Context, id string) (*domain.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound{ID: id}
	}
	return p, nil
}

func (m *mockProductRepo) ListByTenant(_ context.Context, _ string, _ port.ProductFilter) ([]*domain.Product, error) {
	result := make([]*domain.Product, 0, len(m.products))
	for _, p := range m.products {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockProductRepo) Update(_ context.Context, p *domain.Product) error {
	m.products[p.ID] = p
	return nil
}

func (m *mockProductRepo) ExistsBySlug(_ context.Context, tenantID, slug string) (bool, error) {
	return m.slugs[tenantID+":"+slug], nil
}

type mockEnvironmentRepo struct {
	environments map[string]*domain.Environment
	slugs        map[string]bool
}

func newMockEnvironmentRepo() *mockEnvironmentRepo {
	return &mockEnvironmentRepo{
		environments: make(map[string]*domain.Environment),
		slugs:        make(map[string]bool),
	}
}

func (m *mockEnvironmentRepo) Create(_ context.Context, e *domain.Environment) error {
	m.environments[e.ID] = e
	m.slugs[e.ProductID+":"+e.Slug.String()] = true
	return nil
}

func (m *mockEnvironmentRepo) GetByID(_ context.Context, id string) (*domain.Environment, error) {
	e, ok := m.environments[id]
	if !ok {
		return nil, domain.ErrEnvironmentNotFound{ID: id}
	}
	return e, nil
}

func (m *mockEnvironmentRepo) ListByProduct(_ context.Context, _ string, _ port.EnvironmentFilter) ([]*domain.Environment, error) {
	result := make([]*domain.Environment, 0, len(m.environments))
	for _, e := range m.environments {
		result = append(result, e)
	}
	return result, nil
}

func (m *mockEnvironmentRepo) Update(_ context.Context, e *domain.Environment) error {
	m.environments[e.ID] = e
	return nil
}

func (m *mockEnvironmentRepo) ExistsBySlug(_ context.Context, productID, slug string) (bool, error) {
	return m.slugs[productID+":"+slug], nil
}

type mockComponentRepo struct {
	components map[string]*domain.Component
	slugs      map[string]bool
}

func newMockComponentRepo() *mockComponentRepo {
	return &mockComponentRepo{
		components: make(map[string]*domain.Component),
		slugs:      make(map[string]bool),
	}
}

func (m *mockComponentRepo) Create(_ context.Context, c *domain.Component) error {
	m.components[c.ID] = c
	m.slugs[c.EnvironmentID+":"+c.Slug.String()] = true
	return nil
}

func (m *mockComponentRepo) GetByID(_ context.Context, id string) (*domain.Component, error) {
	c, ok := m.components[id]
	if !ok {
		return nil, domain.ErrComponentNotFound{ID: id}
	}
	return c, nil
}

func (m *mockComponentRepo) ListByEnvironment(_ context.Context, _ string, _ port.ComponentFilter) ([]*domain.Component, error) {
	result := make([]*domain.Component, 0, len(m.components))
	for _, c := range m.components {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockComponentRepo) ListByTenant(_ context.Context, _ string) ([]*domain.Component, error) {
	result := make([]*domain.Component, 0, len(m.components))
	for _, c := range m.components {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockComponentRepo) Update(_ context.Context, c *domain.Component) error {
	m.components[c.ID] = c
	return nil
}

func (m *mockComponentRepo) ExistsBySlug(_ context.Context, environmentID, slug string) (bool, error) {
	return m.slugs[environmentID+":"+slug], nil
}

type mockAuditRepo struct {
	events []*domain.AuditEvent
}

func (m *mockAuditRepo) Create(_ context.Context, e *domain.AuditEvent) error {
	m.events = append(m.events, e)
	return nil
}

func (m *mockAuditRepo) ListByEntity(_ context.Context, _, _, _ string, _ int) ([]*domain.AuditEvent, error) {
	return m.events, nil
}

func (m *mockAuditRepo) ListByTenant(_ context.Context, _ string, _ port.AuditFilter) ([]*domain.AuditEvent, int, error) {
	return m.events, len(m.events), nil
}

type mockUnitOfWork struct{}

func (m *mockUnitOfWork) Begin(ctx context.Context) (context.Context, error) { return ctx, nil }
func (m *mockUnitOfWork) Commit(_ context.Context) error                     { return nil }
func (m *mockUnitOfWork) Rollback(_ context.Context) error                   { return nil }

// --- Test Helpers ---

func newTestService() (*app.TenantService, *mockTenantRepo, *mockProductRepo, *mockEnvironmentRepo, *mockComponentRepo) {
	tenantRepo := newMockTenantRepo()
	productRepo := newMockProductRepo()
	envRepo := newMockEnvironmentRepo()
	compRepo := newMockComponentRepo()
	auditRepo := &mockAuditRepo{}
	uow := &mockUnitOfWork{}
	logger := slog.Default()

	svc := app.NewTenantService(tenantRepo, productRepo, envRepo, compRepo, auditRepo, uow, logger)
	return svc, tenantRepo, productRepo, envRepo, compRepo
}

// --- Tests ---

func TestCreateTenant_Success(t *testing.T) {
	svc, tenantRepo, _, _, _ := newTestService()
	ctx := context.Background()

	tenant, err := svc.CreateTenant(ctx, app.CreateTenantCmd{
		Name: "GymFlow",
		Slug: "gym-flow",
		Plan: "free",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tenant.Name != "GymFlow" {
		t.Errorf("got name %q, want %q", tenant.Name, "GymFlow")
	}
	if tenant.Slug.String() != "gym-flow" {
		t.Errorf("got slug %q, want %q", tenant.Slug, "gym-flow")
	}
	if tenant.Plan != domain.PlanFree {
		t.Errorf("got plan %s, want %s", tenant.Plan, domain.PlanFree)
	}
	if len(tenantRepo.tenants) != 1 {
		t.Errorf("expected 1 tenant in repo, got %d", len(tenantRepo.tenants))
	}
}

func TestCreateTenant_DuplicateSlug(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	_, _ = svc.CreateTenant(ctx, app.CreateTenantCmd{
		Name: "GymFlow",
		Slug: "gym-flow",
		Plan: "free",
	})

	_, err := svc.CreateTenant(ctx, app.CreateTenantCmd{
		Name: "GymFlow 2",
		Slug: "gym-flow",
		Plan: "starter",
	})

	if err == nil {
		t.Fatal("expected error for duplicate slug")
	}

	var slugErr domain.ErrTenantSlugTaken
	if !containsError(err, &slugErr) {
		t.Errorf("expected ErrTenantSlugTaken, got %T: %v", err, err)
	}
}

func TestCreateTenant_InvalidSlug(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.CreateTenant(ctx, app.CreateTenantCmd{
		Name: "Test",
		Slug: "ab", // too short
		Plan: "free",
	})

	if err == nil {
		t.Fatal("expected error for invalid slug")
	}
}

func TestCreateTenant_InvalidPlan(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.CreateTenant(ctx, app.CreateTenantCmd{
		Name: "Test",
		Slug: "valid-slug",
		Plan: "ultimate",
	})

	if err == nil {
		t.Fatal("expected error for invalid plan")
	}
}

func TestCreateProduct_Success(t *testing.T) {
	svc, _, productRepo, _, _ := newTestService()
	ctx := context.Background()

	// First create a tenant
	tenant, _ := svc.CreateTenant(ctx, app.CreateTenantCmd{
		Name: "GymFlow",
		Slug: "gym-flow",
		Plan: "free",
	})

	product, err := svc.CreateProduct(ctx, app.CreateProductCmd{
		TenantID:    tenant.ID,
		Name:        "Backend API",
		Slug:        "backend-api",
		Description: "Main backend service",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product.TenantID != tenant.ID {
		t.Errorf("expected tenant_id %s, got %s", tenant.ID, product.TenantID)
	}
	if len(productRepo.products) != 1 {
		t.Errorf("expected 1 product in repo, got %d", len(productRepo.products))
	}
}

func TestCreateProduct_TenantNotFound(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.CreateProduct(ctx, app.CreateProductCmd{
		TenantID:    "019450c0-7e90-7a0f-8c1e-123456789abc",
		Name:        "Backend API",
		Slug:        "backend-api",
		Description: "",
	})

	if err == nil {
		t.Fatal("expected error for missing tenant")
	}
}

func TestCreateProduct_TenantInactive(t *testing.T) {
	svc, tenantRepo, _, _, _ := newTestService()
	ctx := context.Background()

	tenant, _ := svc.CreateTenant(ctx, app.CreateTenantCmd{
		Name: "GymFlow",
		Slug: "gym-flow",
		Plan: "free",
	})

	// Suspend the tenant
	tenant.Status = domain.StatusSuspended
	tenantRepo.tenants[tenant.ID] = tenant

	_, err := svc.CreateProduct(ctx, app.CreateProductCmd{
		TenantID: tenant.ID,
		Name:     "Backend API",
		Slug:     "backend-api",
	})

	if err == nil {
		t.Fatal("expected error for inactive tenant")
	}
}

func TestCreateEnvironment_Success(t *testing.T) {
	svc, _, _, envRepo, _ := newTestService()
	ctx := context.Background()

	tenant, _ := svc.CreateTenant(ctx, app.CreateTenantCmd{
		Name: "GymFlow",
		Slug: "gym-flow",
		Plan: "free",
	})

	product, _ := svc.CreateProduct(ctx, app.CreateProductCmd{
		TenantID: tenant.ID,
		Name:     "Backend",
		Slug:     "backend",
	})

	env, err := svc.CreateEnvironment(ctx, app.CreateEnvironmentCmd{
		TenantID:  tenant.ID,
		ProductID: product.ID,
		Name:      "Production",
		Slug:      "production",
		Tier:      "production",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if env.Tier != domain.TierProduction {
		t.Errorf("expected production tier, got %s", env.Tier)
	}
	if len(envRepo.environments) != 1 {
		t.Errorf("expected 1 environment in repo, got %d", len(envRepo.environments))
	}
}

func TestRegisterComponent_Success(t *testing.T) {
	svc, _, _, _, compRepo := newTestService()
	ctx := context.Background()

	tenant, _ := svc.CreateTenant(ctx, app.CreateTenantCmd{
		Name: "GymFlow",
		Slug: "gym-flow",
		Plan: "free",
	})

	product, _ := svc.CreateProduct(ctx, app.CreateProductCmd{
		TenantID: tenant.ID,
		Name:     "Backend",
		Slug:     "backend",
	})

	env, _ := svc.CreateEnvironment(ctx, app.CreateEnvironmentCmd{
		TenantID:  tenant.ID,
		ProductID: product.ID,
		Name:      "Production",
		Slug:      "production",
		Tier:      "production",
	})

	comp, err := svc.RegisterComponent(ctx, app.RegisterComponentCmd{
		TenantID:      tenant.ID,
		ProductID:     product.ID,
		EnvironmentID: env.ID,
		Name:          "PostgreSQL",
		Slug:          "postgres-main",
		Kind:          "database",
		EndpointURL:   "postgresql://localhost:5432/gymflow",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if comp.Kind != domain.KindDatabase {
		t.Errorf("expected database kind, got %s", comp.Kind)
	}
	if len(compRepo.components) != 1 {
		t.Errorf("expected 1 component in repo, got %d", len(compRepo.components))
	}
}

func TestListTenants_LimitCap(t *testing.T) {
	svc, _, _, _, _ := newTestService()
	ctx := context.Background()

	// Service should cap limit at 100
	_, err := svc.ListTenants(ctx, port.TenantFilter{Limit: 500})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// --- helpers ---

func containsError(err error, target interface{}) bool {
	switch target.(type) {
	case *domain.ErrTenantSlugTaken:
		_, ok := err.(domain.ErrTenantSlugTaken)
		return ok
	}
	return false
}
