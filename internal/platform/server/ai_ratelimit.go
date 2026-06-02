package server

import (
	"net/http"
	"sync"
	"time"
)

// AIRateLimiter provides per-tenant rate limiting for expensive AI operations.
// Prevents abuse of AI analysis endpoints which incur external API costs.
type AIRateLimiter struct {
	mu        sync.Mutex
	limits    map[string]*tenantAILimit
	maxPerMin int
}

type tenantAILimit struct {
	count       int
	windowStart time.Time
}

// NewAIRateLimiter creates a rate limiter for AI endpoints.
// maxPerMin is the maximum number of AI requests per tenant per minute.
func NewAIRateLimiter(maxPerMin int) *AIRateLimiter {
	if maxPerMin <= 0 {
		maxPerMin = 10
	}
	rl := &AIRateLimiter{
		limits:    make(map[string]*tenantAILimit),
		maxPerMin: maxPerMin,
	}
	go rl.cleanupLoop()
	return rl
}

// Allow checks if the tenant is within their rate limit.
func (rl *AIRateLimiter) Allow(tenantID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.limits[tenantID]
	if !exists {
		rl.limits[tenantID] = &tenantAILimit{
			count:       1,
			windowStart: now,
		}
		return true
	}

	// Reset window if minute has passed
	if now.Sub(limit.windowStart) > time.Minute {
		limit.count = 1
		limit.windowStart = now
		return true
	}

	if limit.count >= rl.maxPerMin {
		return false
	}

	limit.count++
	return true
}

func (rl *AIRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, limit := range rl.limits {
			if now.Sub(limit.windowStart) > 5*time.Minute {
				delete(rl.limits, key)
			}
		}
		rl.mu.Unlock()
	}
}

// AIRateLimit middleware enforces per-tenant rate limiting on AI endpoints.
func AIRateLimit(limiter *AIRateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := TenantIDFromContext(r.Context())
			if tenantID == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !limiter.Allow(tenantID) {
				w.Header().Set("Retry-After", "60")
				WriteError(w, http.StatusTooManyRequests, "AI analysis rate limit exceeded (max 10 requests per minute)")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
