package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/optrion/optrion/internal/shared/id"
	tenantdomain "github.com/optrion/optrion/internal/tenant/domain"
)

// RegistrationToken is used to authenticate registration requests.
type RegistrationToken string

// NewRegistrationToken generates a new secure registration token.
func NewRegistrationToken() RegistrationToken {
	return RegistrationToken(id.New())
}

// String returns the token as a string.
func (rt RegistrationToken) String() string {
	return string(rt)
}

// RegistrationRequest holds the data for a bulk registration.
type RegistrationRequest struct {
	Tenant      TenantRegistration
	Product     ProductRegistration
	Environment EnvironmentRegistration
	Components  []ComponentRegistration
}

// TenantRegistration represents tenant registration data.
type TenantRegistration struct {
	Name string
	Slug string
	Plan string // free, starter, professional, enterprise
}

// Validate checks if tenant registration data is valid.
func (tr TenantRegistration) Validate() error {
	if strings.TrimSpace(tr.Name) == "" {
		return fmt.Errorf("tenant name is required")
	}
	if strings.TrimSpace(tr.Slug) == "" {
		return fmt.Errorf("tenant slug is required")
	}
	if strings.TrimSpace(tr.Plan) == "" {
		return fmt.Errorf("tenant plan is required")
	}

	plan := tenantdomain.Plan(tr.Plan)
	if !plan.IsValid() {
		return fmt.Errorf("invalid plan: %s", tr.Plan)
	}

	return nil
}

// ProductRegistration represents product registration data.
type ProductRegistration struct {
	Name        string
	Slug        string
	Description string
}

// Validate checks if product registration data is valid.
func (pr ProductRegistration) Validate() error {
	if strings.TrimSpace(pr.Name) == "" {
		return fmt.Errorf("product name is required")
	}
	if strings.TrimSpace(pr.Slug) == "" {
		return fmt.Errorf("product slug is required")
	}
	return nil
}

// EnvironmentRegistration represents environment registration data.
type EnvironmentRegistration struct {
	Name string
	Tier string // development, staging, production
}

// Validate checks if environment registration data is valid.
func (er EnvironmentRegistration) Validate() error {
	if strings.TrimSpace(er.Name) == "" {
		return fmt.Errorf("environment name is required")
	}
	if strings.TrimSpace(er.Tier) == "" {
		return fmt.Errorf("environment tier is required")
	}

	tier := tenantdomain.Tier(er.Tier)
	if !tier.IsValid() {
		return fmt.Errorf("invalid tier: %s", er.Tier)
	}

	return nil
}

// ComponentRegistration represents component registration data.
type ComponentRegistration struct {
	Name        string
	Kind        string // database, cache, api, web, queue, storage, service, external
	Description string
	Endpoint    string // optional: connection string or URL
	Port        int    // optional: port number
}

// Validate checks if component registration data is valid.
func (cr ComponentRegistration) Validate() error {
	if strings.TrimSpace(cr.Name) == "" {
		return fmt.Errorf("component name is required")
	}
	if strings.TrimSpace(cr.Kind) == "" {
		return fmt.Errorf("component kind is required")
	}

	kind := tenantdomain.ComponentKind(cr.Kind)
	if !kind.IsValid() {
		return fmt.Errorf("invalid component kind: %s", cr.Kind)
	}

	return nil
}

// RegistrationResponse contains the result of a registration.
type RegistrationResponse struct {
	TenantID      string
	ProductID     string
	EnvironmentID string
	ComponentIDs  []string
	APIKey        string
	Endpoint      string
	Message       string
}

// RegistrationAudit represents an audit trail for registrations.
type RegistrationAudit struct {
	ID               string
	TenantID         string
	RegistrationType string // "bulk", "auto-discovery", "manual"
	Status           string // "pending", "success", "failed"
	Request          interface{}
	Response         interface{}
	Error            string
	CreatedAt        time.Time
}

// NewRegistrationAudit creates a new audit record.
func NewRegistrationAudit(tenantID, registrationType string, request interface{}) *RegistrationAudit {
	return &RegistrationAudit{
		ID:               id.New(),
		TenantID:         tenantID,
		RegistrationType: registrationType,
		Status:           "pending",
		Request:          request,
		CreatedAt:        time.Now().UTC(),
	}
}

// MarkSuccess marks the registration as successful.
func (ra *RegistrationAudit) MarkSuccess(response interface{}) {
	ra.Status = "success"
	ra.Response = response
}

// MarkFailed marks the registration as failed with an error.
func (ra *RegistrationAudit) MarkFailed(err error) {
	ra.Status = "failed"
	ra.Error = err.Error()
}
