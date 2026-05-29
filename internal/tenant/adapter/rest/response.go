package rest

import "time"

// --- Response DTOs ---

// TenantResponse is the API response for a tenant.
type TenantResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Plan      string                 `json:"plan"`
	Status    string                 `json:"status"`
	Settings  map[string]interface{} `json:"settings"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ProductResponse is the API response for a product.
type ProductResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EnvironmentResponse is the API response for an environment.
type EnvironmentResponse struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	ProductID string    `json:"product_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Tier      string    `json:"tier"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ComponentResponse is the API response for a component.
type ComponentResponse struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id"`
	ProductID     string                 `json:"product_id"`
	EnvironmentID string                 `json:"environment_id"`
	Name          string                 `json:"name"`
	Slug          string                 `json:"slug"`
	Kind          string                 `json:"kind"`
	EndpointURL   string                 `json:"endpoint_url"`
	Status        string                 `json:"status"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ListResponse is a generic list response with pagination.
type ListResponse[T any] struct {
	Data   []T `json:"data"`
	Count  int `json:"count"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}
