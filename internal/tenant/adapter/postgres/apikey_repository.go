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
		`SELECT id, tenant_id, key_hash, status, expires_at, grace_expires_at
		 FROM api_keys WHERE key_hash = $1`, keyHash,
	)

	var record server.APIKeyRecord
	err := row.Scan(&record.ID, &record.TenantID, &record.KeyHash, &record.Status, &record.ExpiresAt, &record.GraceExpiresAt)
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

// RotateAPIKey creates a new API key and puts the old key into a grace period.
// During the grace period, both old and new keys are valid.
// After grace period expires, the old key is automatically rejected.
// Returns the new raw key (shown once) and the new key ID.
func (r *APIKeyRepository) RotateAPIKey(ctx context.Context, oldKeyID, tenantID, name string, scopes []string, gracePeriod time.Duration) (newRawKey string, newKeyID string, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback is no-op after commit

	now := time.Now().UTC()
	graceExpiresAt := now.Add(gracePeriod)

	// Verify old key exists and belongs to tenant
	var oldKeyStatus string
	err = tx.QueryRow(ctx,
		`SELECT status FROM api_keys WHERE id = $1 AND tenant_id = $2`,
		oldKeyID, tenantID,
	).Scan(&oldKeyStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", "", fmt.Errorf("api key not found")
		}
		return "", "", fmt.Errorf("querying old key: %w", err)
	}
	if oldKeyStatus != "active" {
		return "", "", fmt.Errorf("cannot rotate a non-active key (status: %s)", oldKeyStatus)
	}

	// Mark old key with grace period expiry
	_, err = tx.Exec(ctx,
		`UPDATE api_keys SET rotated_at = $1, grace_expires_at = $2, updated_at = $1
		 WHERE id = $3 AND tenant_id = $4`,
		now, graceExpiresAt, oldKeyID, tenantID,
	)
	if err != nil {
		return "", "", fmt.Errorf("updating old key for rotation: %w", err)
	}

	// Generate the new key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", "", fmt.Errorf("generating random key: %w", err)
	}
	newRawKey = "opk_" + hex.EncodeToString(keyBytes)
	keyHash := server.HashAPIKey(newRawKey)
	keyPrefix := newRawKey[:12]

	newKeyID = id.New()

	_, err = tx.Exec(ctx,
		`INSERT INTO api_keys (id, tenant_id, name, key_hash, key_prefix, scopes, status, rotated_from, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'active', $7, $8, $8)`,
		newKeyID, tenantID, name, keyHash, keyPrefix, scopes, oldKeyID, now,
	)
	if err != nil {
		return "", "", fmt.Errorf("inserting rotated key: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("committing rotation: %w", err)
	}

	return newRawKey, newKeyID, nil
}

// RevokeExpiredGracePeriodKeys revokes keys whose grace period has expired.
// Should be called periodically by a background worker.
func (r *APIKeyRepository) RevokeExpiredGracePeriodKeys(ctx context.Context) (int64, error) {
	result, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET status = 'revoked', updated_at = $1
		 WHERE grace_expires_at IS NOT NULL AND grace_expires_at < $1 AND status = 'active'`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, fmt.Errorf("revoking expired grace period keys: %w", err)
	}
	return result.RowsAffected(), nil
}
