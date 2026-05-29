package domain

import "fmt"

// ErrTenantNotFound indicates a tenant was not found.
type ErrTenantNotFound struct {
	ID string
}

func (e ErrTenantNotFound) Error() string {
	return fmt.Sprintf("tenant not found: %s", e.ID)
}

// ErrTenantSlugTaken indicates the slug is already in use.
type ErrTenantSlugTaken struct {
	Slug string
}

func (e ErrTenantSlugTaken) Error() string {
	return fmt.Sprintf("tenant slug already taken: %s", e.Slug)
}

// ErrProductNotFound indicates a product was not found.
type ErrProductNotFound struct {
	ID string
}

func (e ErrProductNotFound) Error() string {
	return fmt.Sprintf("product not found: %s", e.ID)
}

// ErrProductSlugTaken indicates the product slug is taken within the tenant.
type ErrProductSlugTaken struct {
	TenantID string
	Slug     string
}

func (e ErrProductSlugTaken) Error() string {
	return fmt.Sprintf("product slug already taken in tenant %s: %s", e.TenantID, e.Slug)
}

// ErrEnvironmentNotFound indicates an environment was not found.
type ErrEnvironmentNotFound struct {
	ID string
}

func (e ErrEnvironmentNotFound) Error() string {
	return fmt.Sprintf("environment not found: %s", e.ID)
}

// ErrEnvironmentSlugTaken indicates the environment slug is taken within the product.
type ErrEnvironmentSlugTaken struct {
	ProductID string
	Slug      string
}

func (e ErrEnvironmentSlugTaken) Error() string {
	return fmt.Sprintf("environment slug already taken in product %s: %s", e.ProductID, e.Slug)
}

// ErrComponentNotFound indicates a component was not found.
type ErrComponentNotFound struct {
	ID string
}

func (e ErrComponentNotFound) Error() string {
	return fmt.Sprintf("component not found: %s", e.ID)
}

// ErrComponentSlugTaken indicates the component slug is taken within the environment.
type ErrComponentSlugTaken struct {
	EnvironmentID string
	Slug          string
}

func (e ErrComponentSlugTaken) Error() string {
	return fmt.Sprintf("component slug already taken in environment %s: %s", e.EnvironmentID, e.Slug)
}

// ErrTenantInactive indicates an operation was attempted on an inactive tenant.
type ErrTenantInactive struct {
	ID     string
	Status Status
}

func (e ErrTenantInactive) Error() string {
	return fmt.Sprintf("tenant %s is not active (status: %s)", e.ID, e.Status)
}
