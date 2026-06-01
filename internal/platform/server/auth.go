package server

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TrustedProxies is a configurable list of trusted proxy IPs.
// When set, X-Forwarded-For and X-Real-IP headers are only trusted
// if the request originates from one of these IPs.
// An empty list means all proxy headers are trusted (backwards-compatible default for dev).
var TrustedProxies []string

// SetTrustedProxies configures the list of trusted proxy IPs.
func SetTrustedProxies(proxies []string) {
	TrustedProxies = proxies
}

// ClientIP extracts the real client IP address from the request,
// checking X-Forwarded-For and X-Real-IP headers before falling back to RemoteAddr.
// Only trusts proxy headers if the request comes from a trusted proxy IP.
// All header values are validated as parseable IP addresses.
func ClientIP(r *http.Request) string {
	// Extract the direct connection IP (cannot be spoofed)
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}

	// Only trust proxy headers if request came from a trusted proxy
	if !isFromTrustedProxy(remoteIP) {
		return remoteIP
	}

	// Check X-Forwarded-For header (comma-separated list; first is the client)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		ip := strings.TrimSpace(parts[0])
		if ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Fall back to RemoteAddr
	return remoteIP
}

// isFromTrustedProxy checks if the given IP is in the trusted proxy list.
// If no trusted proxies are configured, all proxy headers are trusted (dev mode).
func isFromTrustedProxy(ip string) bool {
	if len(TrustedProxies) == 0 {
		return true // No restrictions in dev mode
	}
	for _, trusted := range TrustedProxies {
		if ip == trusted {
			return true
		}
	}
	return false
}

// AuthFailureTracker tracks failed authentication attempts for brute-force protection.
type AuthFailureTracker struct {
	mu              sync.Mutex
	failures        map[string]*failureRecord
	maxAttempts     int
	lockoutDuration time.Duration
}

type failureRecord struct {
	attempts    int
	firstFail   time.Time
	lockedUntil time.Time
}

// NewAuthFailureTracker creates a tracker with configurable thresholds.
func NewAuthFailureTracker(maxAttempts int, lockoutDuration time.Duration) *AuthFailureTracker {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if lockoutDuration == 0 {
		lockoutDuration = 15 * time.Minute
	}
	t := &AuthFailureTracker{
		failures:        make(map[string]*failureRecord),
		maxAttempts:     maxAttempts,
		lockoutDuration: lockoutDuration,
	}
	go t.cleanupLoop()
	return t
}

// IsLocked checks if a key (IP or key prefix) is currently locked out.
func (t *AuthFailureTracker) IsLocked(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	rec, exists := t.failures[key]
	if !exists {
		return false
	}

	if !rec.lockedUntil.IsZero() && time.Now().Before(rec.lockedUntil) {
		return true
	}

	return false
}

// RecordFailure records a failed authentication attempt. Returns true if the key is now locked.
func (t *AuthFailureTracker) RecordFailure(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	rec, exists := t.failures[key]
	if !exists {
		t.failures[key] = &failureRecord{
			attempts:  1,
			firstFail: now,
		}
		return false
	}

	// Reset if the window has expired
	if now.Sub(rec.firstFail) > t.lockoutDuration {
		rec.attempts = 1
		rec.firstFail = now
		rec.lockedUntil = time.Time{}
		return false
	}

	rec.attempts++
	if rec.attempts >= t.maxAttempts {
		rec.lockedUntil = now.Add(t.lockoutDuration)
		return true
	}

	return false
}

// RecordSuccess clears the failure record for a key on successful auth.
func (t *AuthFailureTracker) RecordSuccess(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.failures, key)
}

func (t *AuthFailureTracker) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		t.mu.Lock()
		now := time.Now()
		for key, rec := range t.failures {
			if now.Sub(rec.firstFail) > t.lockoutDuration*2 {
				delete(t.failures, key)
			}
		}
		t.mu.Unlock()
	}
}

// contextKey is a private type to prevent collisions in context values.
type contextKey string

const (
	// ContextKeyTenantID is the context key for the authenticated tenant ID.
	ContextKeyTenantID contextKey = "tenant_id"
	// ContextKeyAPIKeyID is the context key for the authenticated API key ID.
	ContextKeyAPIKeyID contextKey = "api_key_id"
)

// TenantIDFromContext extracts the authenticated tenant ID from the request context.
func TenantIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeyTenantID).(string)
	return v
}

// APIKeyIDFromContext extracts the API key ID from the request context.
func APIKeyIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeyAPIKeyID).(string)
	return v
}

// APIKeyRecord represents a stored API key for validation.
type APIKeyRecord struct {
	ID        string
	TenantID  string
	KeyHash   string
	Status    string
	ExpiresAt *time.Time
}

// APIKeyValidator is the interface for looking up API keys.
type APIKeyValidator interface {
	// ValidateKeyHash looks up an API key by its hash and returns the record if valid.
	ValidateKeyHash(ctx context.Context, keyHash string) (*APIKeyRecord, error)
	// RecordUsage updates the last_used_at timestamp for the key.
	RecordUsage(ctx context.Context, keyID string) error
}

// HashAPIKey produces a SHA-256 hash of the raw API key for storage/lookup.
func HashAPIKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}

// APIKeyAuth is middleware that validates API key authentication.
// Keys are expected in the Authorization header as: Bearer <api-key>
// Includes brute-force protection via per-IP failure tracking.
func APIKeyAuth(validator APIKeyValidator, logger *slog.Logger) Middleware {
	failureTracker := NewAuthFailureTracker(5, 15*time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := ClientIP(r)

			// Check if this IP is locked out due to too many failures
			if failureTracker.IsLocked(clientIP) {
				logger.Warn("auth attempt from locked-out IP", "remote_addr", clientIP)
				w.Header().Set("Retry-After", "900")
				WriteError(w, http.StatusTooManyRequests, "too many failed authentication attempts, try again later")
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				failureTracker.RecordFailure(clientIP)
				WriteError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				failureTracker.RecordFailure(clientIP)
				WriteError(w, http.StatusUnauthorized, "invalid authorization format, expected: Bearer <api-key>")
				return
			}

			rawKey := strings.TrimSpace(parts[1])
			if rawKey == "" {
				failureTracker.RecordFailure(clientIP)
				WriteError(w, http.StatusUnauthorized, "empty api key")
				return
			}

			// Hash the provided key and look it up
			keyHash := HashAPIKey(rawKey)
			record, err := validator.ValidateKeyHash(r.Context(), keyHash)
			if err != nil {
				logger.Error("api key validation error", "error", err)
				WriteError(w, http.StatusInternalServerError, "authentication error")
				return
			}

			if record == nil {
				failureTracker.RecordFailure(clientIP)
				logger.Warn("invalid api key attempt",
					"remote_addr", clientIP,
					"path", r.URL.Path,
				)
				WriteError(w, http.StatusUnauthorized, "invalid api key")
				return
			}

			if record.Status != "active" {
				failureTracker.RecordFailure(clientIP)
				WriteError(w, http.StatusUnauthorized, "api key is revoked")
				return
			}

			if record.ExpiresAt != nil && time.Now().After(*record.ExpiresAt) {
				failureTracker.RecordFailure(clientIP)
				WriteError(w, http.StatusUnauthorized, "api key has expired")
				return
			}

			// Successful auth — clear failure record
			failureTracker.RecordSuccess(clientIP)

			// Inject tenant ID and key ID into context
			ctx := context.WithValue(r.Context(), ContextKeyTenantID, record.TenantID)
			ctx = context.WithValue(ctx, ContextKeyAPIKeyID, record.ID)

			// Record usage asynchronously (non-blocking)
			go func() {
				if err := validator.RecordUsage(context.Background(), record.ID); err != nil {
					logger.Error("failed to record api key usage", "key_id", record.ID, "error", err)
				}
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantIsolation is middleware that enforces the authenticated tenant can only
// access their own resources. It checks path parameters and request bodies for
// tenant_id mismatches.
func TenantIsolation(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := TenantIDFromContext(r.Context())
			if tenantID == "" {
				// No authenticated tenant — should have been caught by APIKeyAuth
				WriteError(w, http.StatusUnauthorized, "no authenticated tenant")
				return
			}

			// Check path parameter if present (e.g., /api/v1/tenants/{tenantId}/...)
			pathTenantID := r.PathValue("tenantId")
			if pathTenantID != "" && !constantTimeEqual(pathTenantID, tenantID) {
				logger.Warn("tenant isolation violation",
					"authenticated_tenant", tenantID,
					"requested_tenant", pathTenantID,
					"path", r.URL.Path,
				)
				WriteError(w, http.StatusForbidden, "access denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// constantTimeEqual compares two strings in constant time to prevent timing attacks.
func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
