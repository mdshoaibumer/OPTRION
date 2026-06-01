package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/platform/server"
	"github.com/optrion/optrion/internal/shared/id"
)

// APIKeyRepository implements server.APIKeyValidator using PostgreSQL.
type APIKeyRepository struct {
	pool *pgxpool.Pool
}

// NewAPIKeyRepository creates a new API key repository.
func NewAPIKeyRepository(pool *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{pool: pool}
}

// ValidateKeyHash looks up an API key by its hash.
func (r *APIKeyRepository) ValidateKeyHash(ctx context.Context, keyHash string) (*server.APIKeyRecord, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, key_hash, status, expires_at
		 FROM api_keys WHERE key_hash = $1`, keyHash,
	)

	var record server.APIKeyRecord
	err := row.Scan(&record.ID, &record.TenantID, &record.KeyHash, &record.Status, &record.ExpiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying api key: %w", err)
	}
	return &record, nil
}

// RecordUsage updates the last_used_at timestamp.
func (r *APIKeyRepository) RecordUsage(ctx context.Context, keyID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET last_used_at = $1 WHERE id = $2`,
		time.Now().UTC(), keyID,
	)
	if err != nil {
		return fmt.Errorf("recording api key usage: %w", err)
	}
	return nil
}

// CreateAPIKey generates a new API key for a tenant and stores it.
// Returns the raw key (only shown once) and the key ID.
func (r *APIKeyRepository) CreateAPIKey(ctx context.Context, tenantID, name string, scopes []string, expiresAt *time.Time) (rawKey string, keyID string, err error) {
	// Generate a cryptographically secure random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", "", fmt.Errorf("generating random key: %w", err)
	}
	rawKey = "opk_" + hex.EncodeToString(keyBytes) // opk_ prefix for "optrion key"
	keyHash := server.HashAPIKey(rawKey)
	keyPrefix := rawKey[:12] // "opk_" + first 8 hex chars

	keyID = id.New()

	_, err = r.pool.Exec(ctx,
		`INSERT INTO api_keys (id, tenant_id, name, key_hash, key_prefix, scopes, status, expires_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'active', $7, $8, $8)`,
		keyID, tenantID, name, keyHash, keyPrefix, scopes, expiresAt, time.Now().UTC(),
	)
	if err != nil {
		return "", "", fmt.Errorf("inserting api key: %w", err)
	}

	return rawKey, keyID, nil
}

// RevokeAPIKey marks an API key as revoked.
func (r *APIKeyRepository) RevokeAPIKey(ctx context.Context, keyID, tenantID string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET status = 'revoked', updated_at = $1 WHERE id = $2 AND tenant_id = $3`,
		time.Now().UTC(), keyID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("revoking api key: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("api key not found")
	}
	return nil
}
