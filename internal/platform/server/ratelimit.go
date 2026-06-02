package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// RateLimiter defines the interface for rate limiting backends.
type RateLimiter interface {
	// Allow checks if the given key is allowed to proceed.
	// Returns true if allowed, false if rate limited.
	Allow(ctx context.Context, key string) (bool, error)
}

// InMemoryRateLimiter is a simple token bucket rate limiter for single-instance deployments.
type InMemoryRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rate    int           // requests per second
	burst   int           // max burst
	cleanup time.Duration // how often to clean expired buckets
}

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

// NewInMemoryRateLimiter creates a rate limiter with the given requests-per-second limit.
func NewInMemoryRateLimiter(rps int) *InMemoryRateLimiter {
	if rps <= 0 {
		rps = 100
	}
	rl := &InMemoryRateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    rps,
		burst:   rps * 2, // Allow burst of 2x rate
		cleanup: 5 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *InMemoryRateLimiter) Allow(_ context.Context, key string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.buckets[key]
	if !exists {
		rl.buckets[key] = &tokenBucket{
			tokens:     float64(rl.burst) - 1,
			lastRefill: now,
		}
		return true, nil
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens += elapsed * float64(rl.rate)
	if bucket.tokens > float64(rl.burst) {
		bucket.tokens = float64(rl.burst)
	}
	bucket.lastRefill = now

	if bucket.tokens < 1 {
		return false, nil
	}

	bucket.tokens--
	return true, nil
}

func (rl *InMemoryRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			if now.Sub(bucket.lastRefill) > rl.cleanup {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit is middleware that applies per-tenant rate limiting.
// Falls back to per-IP limiting for unauthenticated requests.
func RateLimit(limiter RateLimiter, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use tenant ID if authenticated, otherwise use client IP
			key := TenantIDFromContext(r.Context())
			if key == "" {
				key = ClientIP(r)
			}

			allowed, err := limiter.Allow(r.Context(), key)
			if err != nil {
				logger.Error("rate limiter error", "error", err, "key", key)
				// Fail closed — deny request when rate limiter is unavailable
				w.Header().Set("Retry-After", "5")
				WriteError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", "1")
				WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
