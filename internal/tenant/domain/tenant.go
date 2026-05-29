package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/optrion/optrion/internal/shared/id"
)

// Slug is a URL-safe identifier for human-readable URLs.
type Slug string

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// NewSlug validates and creates a Slug from a string.
func NewSlug(s string) (Slug, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) < 3 {
		return "", fmt.Errorf("slug must be at least 3 characters")
	}
	if len(s) > 100 {
		return "", fmt.Errorf("slug must be at most 100 characters")
	}
	if !slugRegex.MatchString(s) {
		return "", fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens")
	}
	return Slug(s), nil
}

// String returns the slug as a string.
func (s Slug) String() string { return string(s) }

// Plan represents a tenant's subscription plan.
type Plan string

const (
	PlanFree         Plan = "free"
	PlanStarter      Plan = "starter"
	PlanProfessional Plan = "professional"
	PlanEnterprise   Plan = "enterprise"
)

// IsValid checks if the plan is a recognized value.
func (p Plan) IsValid() bool {
	switch p {
	case PlanFree, PlanStarter, PlanProfessional, PlanEnterprise:
		return true
	}
	return false
}

// Status represents an entity's lifecycle status.
type Status string

const (
	StatusActive      Status = "active"
	StatusSuspended   Status = "suspended"
	StatusDeactivated Status = "deactivated"
	StatusArchived    Status = "archived"
)

// Tier represents an environment tier.
type Tier string

const (
	TierDevelopment Tier = "development"
	TierStaging     Tier = "staging"
	TierProduction  Tier = "production"
)

// IsValid checks if the tier is a recognized value.
func (t Tier) IsValid() bool {
	switch t {
	case TierDevelopment, TierStaging, TierProduction:
		return true
	}
	return false
}

// ComponentKind represents the type of a monitored component.
type ComponentKind string

const (
	KindDatabase ComponentKind = "database"
	KindCache    ComponentKind = "cache"
	KindAPI      ComponentKind = "api"
	KindWeb      ComponentKind = "web"
	KindQueue    ComponentKind = "queue"
	KindStorage  ComponentKind = "storage"
	KindService  ComponentKind = "service"
	KindExternal ComponentKind = "external"
)

// IsValid checks if the component kind is recognized.
func (k ComponentKind) IsValid() bool {
	switch k {
	case KindDatabase, KindCache, KindAPI, KindWeb, KindQueue, KindStorage, KindService, KindExternal:
		return true
	}
	return false
}

// Tenant is the aggregate root for multi-tenancy.
// All resources in OPTRION belong to exactly one Tenant.
type Tenant struct {
	ID        string
	Name      string
	Slug      Slug
	Plan      Plan
	Status    Status
	Settings  map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTenant creates a new Tenant with validation.
func NewTenant(name string, slug Slug, plan Plan) (*Tenant, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("tenant name is required")
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("tenant name must be at most 255 characters")
	}
	if !plan.IsValid() {
		return nil, fmt.Errorf("invalid plan: %s", plan)
	}

	now := time.Now().UTC()
	return &Tenant{
		ID:        id.New(),
		Name:      strings.TrimSpace(name),
		Slug:      slug,
		Plan:      plan,
		Status:    StatusActive,
		Settings:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// IsActive returns true if the tenant is in active status.
func (t *Tenant) IsActive() bool {
	return t.Status == StatusActive
}

// Suspend marks the tenant as suspended.
func (t *Tenant) Suspend() error {
	if t.Status != StatusActive {
		return fmt.Errorf("can only suspend active tenants, current status: %s", t.Status)
	}
	t.Status = StatusSuspended
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Activate re-activates a suspended tenant.
func (t *Tenant) Activate() error {
	if t.Status != StatusSuspended {
		return fmt.Errorf("can only activate suspended tenants, current status: %s", t.Status)
	}
	t.Status = StatusActive
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Product represents a logical grouping of services within a tenant.
type Product struct {
	ID          string
	TenantID    string
	Name        string
	Slug        Slug
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewProduct creates a new Product with validation.
func NewProduct(tenantID, name string, slug Slug, description string) (*Product, error) {
	if !id.IsValid(tenantID) {
		return nil, fmt.Errorf("invalid tenant ID: %s", tenantID)
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("product name must be at most 255 characters")
	}

	now := time.Now().UTC()
	return &Product{
		ID:          id.New(),
		TenantID:    tenantID,
		Name:        strings.TrimSpace(name),
		Slug:        slug,
		Description: strings.TrimSpace(description),
		Status:      StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// IsActive returns true if the product is active.
func (p *Product) IsActive() bool {
	return p.Status == StatusActive
}

// Archive marks the product as archived.
func (p *Product) Archive() error {
	if p.Status == StatusArchived {
		return fmt.Errorf("product is already archived")
	}
	p.Status = StatusArchived
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// Environment represents a deployment stage within a product.
type Environment struct {
	ID        string
	TenantID  string
	ProductID string
	Name      string
	Slug      Slug
	Tier      Tier
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewEnvironment creates a new Environment with validation.
func NewEnvironment(tenantID, productID, name string, slug Slug, tier Tier) (*Environment, error) {
	if !id.IsValid(tenantID) {
		return nil, fmt.Errorf("invalid tenant ID: %s", tenantID)
	}
	if !id.IsValid(productID) {
		return nil, fmt.Errorf("invalid product ID: %s", productID)
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("environment name is required")
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("environment name must be at most 255 characters")
	}
	if !tier.IsValid() {
		return nil, fmt.Errorf("invalid tier: %s", tier)
	}

	now := time.Now().UTC()
	return &Environment{
		ID:        id.New(),
		TenantID:  tenantID,
		ProductID: productID,
		Name:      strings.TrimSpace(name),
		Slug:      slug,
		Tier:      tier,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// IsActive returns true if the environment is active.
func (e *Environment) IsActive() bool {
	return e.Status == StatusActive
}

// Component represents a monitored entity within an environment.
type Component struct {
	ID            string
	TenantID      string
	ProductID     string
	EnvironmentID string
	Name          string
	Slug          Slug
	Kind          ComponentKind
	EndpointURL   string
	Status        Status
	Metadata      map[string]interface{}
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewComponent creates a new Component with validation.
func NewComponent(tenantID, productID, environmentID, name string, slug Slug, kind ComponentKind, endpointURL string) (*Component, error) {
	if !id.IsValid(tenantID) {
		return nil, fmt.Errorf("invalid tenant ID: %s", tenantID)
	}
	if !id.IsValid(productID) {
		return nil, fmt.Errorf("invalid product ID: %s", productID)
	}
	if !id.IsValid(environmentID) {
		return nil, fmt.Errorf("invalid environment ID: %s", environmentID)
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("component name is required")
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("component name must be at most 255 characters")
	}
	if !kind.IsValid() {
		return nil, fmt.Errorf("invalid component kind: %s", kind)
	}

	now := time.Now().UTC()
	return &Component{
		ID:            id.New(),
		TenantID:      tenantID,
		ProductID:     productID,
		EnvironmentID: environmentID,
		Name:          strings.TrimSpace(name),
		Slug:          slug,
		Kind:          kind,
		EndpointURL:   strings.TrimSpace(endpointURL),
		Status:        StatusActive,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// IsActive returns true if the component is active.
func (c *Component) IsActive() bool {
	return c.Status == StatusActive
}

// AuditEvent records a state change in the system.
type AuditEvent struct {
	ID         string
	TenantID   string
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	Payload    map[string]interface{}
	OccurredAt time.Time
}

// NewAuditEvent creates a new audit event.
func NewAuditEvent(tenantID, actorID, action, entityType, entityID string, payload map[string]interface{}) *AuditEvent {
	if payload == nil {
		payload = make(map[string]interface{})
	}
	return &AuditEvent{
		ID:         id.New(),
		TenantID:   tenantID,
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Payload:    payload,
		OccurredAt: time.Now().UTC(),
	}
}
