package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/optrion/optrion/internal/tenant/adapter/rest"
	"github.com/optrion/optrion/internal/tenant/app"
	"github.com/optrion/optrion/internal/tenant/domain"
	"github.com/optrion/optrion/internal/tenant/port"
)

// --- Mock Repositories (minimal for handler tests) ---

type mockTenantRepo struct {
	tenants map[string]*domain.Tenant
	slugs   map[string]bool
}

func newMockTenantRepo() *mockTenantRepo {
	return &mockTenantRepo{
		tenants: make(map[string]*domain.Tenant),
		slugs:   make(map[string]bool),
	}
}

func (m *mockTenantRepo) Create(_ context.Context, t *domain.Tenant) error {
	m.tenants[t.ID] = t
	m.slugs[t.Slug.String()] = true
	return nil
}
func (m *mockTenantRepo) GetByID(_ context.Context, id string) (*domain.Tenant, error) {
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
	var result []*domain.Tenant
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
	var result []*domain.Product
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
	var result []*domain.Environment
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
	var result []*domain.Component
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

type mockAuditRepo struct{}

func (m *mockAuditRepo) Create(_ context.Context, _ *domain.AuditEvent) error { return nil }
func (m *mockAuditRepo) ListByEntity(_ context.Context, _, _, _ string, _ int) ([]*domain.AuditEvent, error) {
	return nil, nil
}

type mockUnitOfWork struct{}

func (m *mockUnitOfWork) Begin(ctx context.Context) (context.Context, error) { return ctx, nil }
func (m *mockUnitOfWork) Commit(_ context.Context) error                     { return nil }
func (m *mockUnitOfWork) Rollback(_ context.Context) error                   { return nil }

// --- Test Setup ---

func setupHandler() (*rest.Handler, *http.ServeMux) {
	tenantRepo := newMockTenantRepo()
	productRepo := newMockProductRepo()
	envRepo := newMockEnvironmentRepo()
	compRepo := newMockComponentRepo()
	auditRepo := &mockAuditRepo{}
	uow := &mockUnitOfWork{}
	logger := slog.Default()

	svc := app.NewTenantService(tenantRepo, productRepo, envRepo, compRepo, auditRepo, uow, logger)
	handler := rest.NewHandler(svc, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return handler, mux
}

// --- Handler Tests ---

func TestCreateTenantHandler_Success(t *testing.T) {
	_, mux := setupHandler()

	body := `{"name":"GymFlow","slug":"gym-flow","plan":"free"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["name"] != "GymFlow" {
		t.Errorf("expected name GymFlow, got %v", resp["name"])
	}
	if resp["slug"] != "gym-flow" {
		t.Errorf("expected slug gym-flow, got %v", resp["slug"])
	}
	if resp["status"] != "active" {
		t.Errorf("expected status active, got %v", resp["status"])
	}
}

func TestCreateTenantHandler_ValidationError(t *testing.T) {
	_, mux := setupHandler()

	body := `{"name":"","slug":"","plan":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTenantHandler_InvalidJSON(t *testing.T) {
	_, mux := setupHandler()

	body := `not json`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListTenantsHandler_Empty(t *testing.T) {
	_, mux := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp["count"].(float64) != 0 {
		t.Errorf("expected count 0, got %v", resp["count"])
	}
}

func TestGetTenantHandler_NotFound(t *testing.T) {
	_, mux := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/019450c0-7e90-7a0f-8c1e-123456789abc", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTenantHandler_DuplicateSlug(t *testing.T) {
	_, mux := setupHandler()

	body := `{"name":"GymFlow","slug":"gym-flow","plan":"free"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("first create expected 201, got %d", rec.Code)
	}

	// Second create with same slug
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

func TestListProductsHandler_MissingTenantID(t *testing.T) {
	_, mux := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListEnvironmentsHandler_MissingProductID(t *testing.T) {
	_, mux := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/environments", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListComponentsHandler_MissingEnvironmentID(t *testing.T) {
	_, mux := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/components", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
