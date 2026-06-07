package rest

// --- Request DTOs ---

// CreateTenantRequest is the request body for creating a tenant.
type CreateTenantRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Plan string `json:"plan"`
}

// Validate checks the request fields.
func (r CreateTenantRequest) Validate() []string {
	var errs []string
	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	if len(r.Name) > 255 {
		errs = append(errs, "name must be at most 255 characters")
	}
	if r.Slug == "" {
		errs = append(errs, "slug is required")
	}
	if r.Plan == "" {
		errs = append(errs, "plan is required")
	}
	return errs
}

// CreateProductRequest is the request body for creating a product.
type CreateProductRequest struct {
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// Validate checks the request fields.
func (r CreateProductRequest) Validate() []string {
	var errs []string
	if r.TenantID == "" {
		errs = append(errs, "tenant_id is required")
	}
	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	if len(r.Name) > 255 {
		errs = append(errs, "name must be at most 255 characters")
	}
	if r.Slug == "" {
		errs = append(errs, "slug is required")
	}
	return errs
}

// CreateEnvironmentRequest is the request body for creating an environment.
type CreateEnvironmentRequest struct {
	TenantID  string `json:"tenant_id"`
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Tier      string `json:"tier"`
}

// Validate checks the request fields.
func (r *CreateEnvironmentRequest) Validate() []string {
	var errs []string
	if r.TenantID == "" {
		errs = append(errs, "tenant_id is required")
	}
	if r.ProductID == "" {
		errs = append(errs, "product_id is required")
	}
	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	if len(r.Name) > 255 {
		errs = append(errs, "name must be at most 255 characters")
	}
	if r.Slug == "" {
		errs = append(errs, "slug is required")
	}
	if r.Tier == "" {
		errs = append(errs, "tier is required")
	}
	return errs
}

// RegisterComponentRequest is the request body for registering a component.
type RegisterComponentRequest struct {
	TenantID      string `json:"tenant_id"`
	ProductID     string `json:"product_id"`
	EnvironmentID string `json:"environment_id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Kind          string `json:"kind"`
	EndpointURL   string `json:"endpoint_url"`
}

// Validate checks the request fields.
func (r *RegisterComponentRequest) Validate() []string {
	var errs []string
	if r.TenantID == "" {
		errs = append(errs, "tenant_id is required")
	}
	if r.ProductID == "" {
		errs = append(errs, "product_id is required")
	}
	if r.EnvironmentID == "" {
		errs = append(errs, "environment_id is required")
	}
	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	if len(r.Name) > 255 {
		errs = append(errs, "name must be at most 255 characters")
	}
	if r.Slug == "" {
		errs = append(errs, "slug is required")
	}
	if r.Kind == "" {
		errs = append(errs, "kind is required")
	}
	return errs
}
