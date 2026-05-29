package server

import (
	"log/slog"
	"net/http"
)

// Router sets up all HTTP routes for the application.
type Router struct {
	mux    *http.ServeMux
	logger *slog.Logger
}

// NewRouter creates a new router with all middleware applied.
func NewRouter(logger *slog.Logger) *Router {
	return &Router{
		mux:    http.NewServeMux(),
		logger: logger,
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

// Handler returns the composed HTTP handler with middleware chain.
func (rt *Router) Handler() http.Handler {
	middleware := Chain(
		SecurityHeaders(),
		CORS(),
		Recovery(rt.logger),
		CorrelationID(),
		RequestID(),
		Logging(rt.logger),
		ContentType("application/json"),
	)

	return middleware(rt.mux)
}
