package server

import (
	"log/slog"
	"net/http"
)

// RouterConfig holds configuration for the HTTP router.
type RouterConfig struct {
	CORSOrigins    []string
	RateLimitRPS   int
	AuthEnabled    bool
	AuthValidator  APIKeyValidator
	TrustedProxies []string // IPs allowed to set X-Forwarded-For
}

// Router sets up all HTTP routes for the application.
type Router struct {
	mux         *http.ServeMux
	logger      *slog.Logger
	config      RouterConfig
	rateLimiter RateLimiter
}

// NewRouter creates a new router with all middleware applied.
func NewRouter(logger *slog.Logger, cfg RouterConfig) *Router {
	var limiter RateLimiter
	if cfg.RateLimitRPS > 0 {
		limiter = NewInMemoryRateLimiter(cfg.RateLimitRPS)
	}

	// Configure trusted proxies for IP extraction security
	if len(cfg.TrustedProxies) > 0 {
		SetTrustedProxies(cfg.TrustedProxies)
	}

	return &Router{
		mux:         http.NewServeMux(),
		logger:      logger,
		config:      cfg,
		rateLimiter: limiter,
	}
}

// Handle registers a handler for the given pattern.
func (rt *Router) Handle(pattern string, handler http.Handler) {
	rt.mux.Handle(pattern, handler)
}

// HandleFunc registers a handler function for the given pattern.
func (rt *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	rt.mux.HandleFunc(pattern, handler)
}

// Mux returns the underlying ServeMux for external route registration.
func (rt *Router) Mux() *http.ServeMux {
	return rt.mux
}

// Handler returns the composed HTTP handler with the full middleware chain.
func (rt *Router) Handler() http.Handler {
	middlewares := []Middleware{
		SecurityHeaders(),
		CORS(rt.config.CORSOrigins...),
		Recovery(rt.logger),
		CorrelationID(),
		RequestID(),
		Logging(rt.logger),
		ContentType("application/json"),
	}

	// Add rate limiting if configured
	if rt.rateLimiter != nil {
		middlewares = append(middlewares, RateLimit(rt.rateLimiter, rt.logger))
	}

	return Chain(middlewares...)(rt.mux)
}

// AuthenticatedHandler wraps a handler with API key authentication and tenant isolation.
// Use this for all tenant-scoped API endpoints.
func (rt *Router) AuthenticatedHandler(handler http.Handler) http.Handler {
	if !rt.config.AuthEnabled || rt.config.AuthValidator == nil {
		return handler
	}
	return Chain(
		APIKeyAuth(rt.config.AuthValidator, rt.logger),
		TenantIsolation(rt.logger),
	)(handler)
}
