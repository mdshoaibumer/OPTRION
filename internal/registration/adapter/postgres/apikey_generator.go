package postgres

import (
	"context"
	"fmt"

	tenantpg "github.com/optrion/optrion/internal/tenant/adapter/postgres"
)

// APIKeyGeneratorAdapter adapts the tenant APIKeyRepository to satisfy port.APIKeyGenerator.
type APIKeyGeneratorAdapter struct {
	repo *tenantpg.APIKeyRepository
}

// NewAPIKeyGeneratorAdapter creates a new adapter.
func NewAPIKeyGeneratorAdapter(repo *tenantpg.APIKeyRepository) *APIKeyGeneratorAdapter {
	return &APIKeyGeneratorAdapter{repo: repo}
}

// Generate creates a new API key for the given tenant and returns the raw key.
func (a *APIKeyGeneratorAdapter) Generate(tenantID string) (string, error) {
	rawKey, _, err := a.repo.CreateAPIKey(context.Background(), tenantID, "default", []string{"read", "write"}, nil)
	if err != nil {
		return "", fmt.Errorf("generating api key: %w", err)
	}
	return rawKey, nil
}
