package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/optrion/optrion/internal/registration/domain"
	"github.com/optrion/optrion/internal/registration/port"
	tenantapp "github.com/optrion/optrion/internal/tenant/app"
)

// RegistrationService orchestrates the registration workflow.
type RegistrationService struct {
	tenantService   *tenantapp.TenantService
	apiKeyGenerator port.APIKeyGenerator
	auditRepository port.RegistrationRepository
	logger          *slog.Logger
	platformURL     string
}

// NewRegistrationService creates a new RegistrationService.
func NewRegistrationService(
	tenantService *tenantapp.TenantService,
	apiKeyGenerator port.APIKeyGenerator,
	auditRepository port.RegistrationRepository,
	logger *slog.Logger,
	platformURL string,
) *RegistrationService {
	return &RegistrationService{
		tenantService:   tenantService,
		apiKeyGenerator: apiKeyGenerator,
		auditRepository: auditRepository,
		logger:          logger,
		platformURL:     platformURL,
	}
}

// Register performs a bulk registration of tenant, product, environment, and components.
// This is the main entry point for the plug-and-play registration workflow.
func (rs *RegistrationService) Register(ctx context.Context, req domain.RegistrationRequest) (*domain.RegistrationResponse, error) {
	// Validate all input data
	if err := req.Tenant.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tenant registration: %w", err)
	}
	if err := req.Product.Validate(); err != nil {
		return nil, fmt.Errorf("invalid product registration: %w", err)
	}
	if err := req.Environment.Validate(); err != nil {
		return nil, fmt.Errorf("invalid environment registration: %w", err)
	}

	for i, comp := range req.Components {
		if err := comp.Validate(); err != nil {
			return nil, fmt.Errorf("invalid component %d registration: %w", i, err)
		}
	}

	// Create audit record
	audit := domain.NewRegistrationAudit("", "bulk", req)

	// Create tenant
	tenantCmd := tenantapp.CreateTenantCmd{
		Name: req.Tenant.Name,
		Slug: req.Tenant.Slug,
		Plan: req.Tenant.Plan,
	}

	tenant, err := rs.tenantService.CreateTenant(ctx, tenantCmd)
	if err != nil {
		audit.MarkFailed(err)
		if err := rs.auditRepository.CreateAudit(ctx, audit); err != nil {
			rs.logger.WarnContext(ctx, "failed to create audit record", "error", err)
		}
		return nil, fmt.Errorf("creating tenant: %w", err)
	}

	audit.TenantID = tenant.ID

	// Create product
	productCmd := tenantapp.CreateProductCmd{
		TenantID:    tenant.ID,
		Name:        req.Product.Name,
		Slug:        req.Product.Slug,
		Description: req.Product.Description,
	}

	product, err := rs.tenantService.CreateProduct(ctx, productCmd)
	if err != nil {
		audit.MarkFailed(err)
		if err := rs.auditRepository.CreateAudit(ctx, audit); err != nil {
			rs.logger.WarnContext(ctx, "failed to create audit record", "error", err)
		}
		return nil, fmt.Errorf("creating product: %w", err)
	}

	// Create environment
	envCmd := tenantapp.CreateEnvironmentCmd{
		TenantID:  tenant.ID,
		ProductID: product.ID,
		Name:      req.Environment.Name,
		Slug:      slugFromName(req.Environment.Name),
		Tier:      req.Environment.Tier,
	}

	env, err := rs.tenantService.CreateEnvironment(ctx, envCmd)
	if err != nil {
		audit.MarkFailed(err)
		if err := rs.auditRepository.CreateAudit(ctx, audit); err != nil {
			rs.logger.WarnContext(ctx, "failed to create audit record", "error", err)
		}
		return nil, fmt.Errorf("creating environment: %w", err)
	}

	// Create components
	componentIDs := make([]string, 0, len(req.Components))
	for _, compReq := range req.Components {
		compCmd := tenantapp.CreateComponentCmd{
			EnvironmentID: env.ID,
			Name:          compReq.Name,
			Kind:          compReq.Kind,
			Description:   compReq.Description,
			Endpoint:      compReq.Endpoint,
			Port:          compReq.Port,
		}

		comp, err := rs.tenantService.CreateComponent(ctx, &compCmd)
		if err != nil {
			audit.MarkFailed(err)
			if err := rs.auditRepository.CreateAudit(ctx, audit); err != nil {
				rs.logger.WarnContext(ctx, "failed to create audit record", "error", err)
			}
			return nil, fmt.Errorf("creating component %s: %w", compReq.Name, err)
		}

		componentIDs = append(componentIDs, comp.ID)
	}

	// Generate API key
	apiKey, err := rs.apiKeyGenerator.Generate(tenant.ID)
	if err != nil {
		audit.MarkFailed(err)
		if err := rs.auditRepository.CreateAudit(ctx, audit); err != nil {
			rs.logger.WarnContext(ctx, "failed to create audit record", "error", err)
		}
		return nil, fmt.Errorf("generating API key: %w", err)
	}

	// Build response
	response := &domain.RegistrationResponse{
		TenantID:      tenant.ID,
		ProductID:     product.ID,
		EnvironmentID: env.ID,
		ComponentIDs:  componentIDs,
		APIKey:        apiKey,
		Endpoint:      rs.platformURL,
		Message:       fmt.Sprintf("Successfully registered %s with %d components", req.Product.Name, len(componentIDs)),
	}

	// Mark audit as successful
	audit.MarkSuccess(response)
	if err := rs.auditRepository.CreateAudit(ctx, audit); err != nil {
		rs.logger.WarnContext(ctx, "failed to create audit record", "error", err)
	}

	rs.logger.InfoContext(ctx, "bulk registration completed",
		"tenant_id", tenant.ID,
		"product_id", product.ID,
		"component_count", len(componentIDs),
	)

	return response, nil
}

// ValidateAPIKey checks if an API key is valid and belongs to a tenant.
func (rs *RegistrationService) ValidateAPIKey(ctx context.Context, apiKey string) (string, error) {
	if len(apiKey) == 0 {
		return "", fmt.Errorf("API key is required")
	}

	// API key validation is handled by the platform auth middleware (SHA-256 hash lookup).
	// This method is used during registration validation to check key format.
	if len(apiKey) < 32 {
		return "", fmt.Errorf("invalid API key format")
	}

	// The actual key validation and tenant resolution is handled by the
	// server.APIKeyValidator interface in the auth middleware.
	// This function validates format only; full validation happens at the HTTP layer.
	return "", nil
}

func slugFromName(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
}
