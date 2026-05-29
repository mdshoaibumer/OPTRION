package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/optrion/optrion/internal/platform/cache"
	"github.com/optrion/optrion/internal/platform/config"
	"github.com/optrion/optrion/internal/platform/database"
	"github.com/optrion/optrion/internal/platform/logger"
	"github.com/optrion/optrion/internal/platform/server"
)

// Container holds all application dependencies wired together.
// This is the composition root — the only place where concrete implementations
// are selected and composed. No global state. No service locator.
type Container struct {
	Config   *config.Config
	Logger   *slog.Logger
	Database *database.PostgreSQL
	Redis    *cache.Redis
	Server   *server.Server

	// Internal components for lifecycle management
	router *server.Router
	health *server.HealthHandler
}

// NewContainer builds the application container with the full dependency graph.
// Order: Config → Logger → Database → Redis → Router → Server
func NewContainer(ctx context.Context) (*Container, error) {
	c := &Container{}

	// 1. Configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	c.Config = cfg

	// 2. Logger
	c.Logger = logger.New(cfg.Log)
	c.Logger.Info("configuration loaded",
		"environment", cfg.App.Environment,
		"version", cfg.App.Version,
	)

	// 3. Database
	db, err := database.NewPostgreSQL(ctx, cfg.Database, c.Logger.With("component", "database"))
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	c.Database = db

	// 4. Redis
	redis, err := cache.NewRedis(ctx, cfg.Redis, c.Logger.With("component", "redis"))
	if err != nil {
		c.Database.Close()
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}
	c.Redis = redis

	// 5. Health handler
	c.health = server.NewHealthHandler(db, redis, cfg.App.Version, c.Logger.With("component", "health"))

	// 6. Router
	c.router = server.NewRouter(c.Logger.With("component", "http"))
	c.registerRoutes()

	// 7. HTTP Server
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	c.Server = server.NewServer(
		addr,
		c.router.Handler(),
		c.Logger.With("component", "server"),
		cfg.HTTP.ReadTimeout,
		cfg.HTTP.WriteTimeout,
		cfg.HTTP.IdleTimeout,
	)

	return c, nil
}

// registerRoutes sets up all HTTP routes.
func (c *Container) registerRoutes() {
	// Health endpoints (no auth required)
	c.router.HandleFunc("GET /healthz", c.health.Liveness())
	c.router.HandleFunc("GET /readyz", c.health.Readiness())

	// API version info
	c.router.HandleFunc("GET /api/v1/info", c.infoHandler())
}

// infoHandler returns application information.
func (c *Container) infoHandler() http.HandlerFunc {
	type infoResponse struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Environment string `json:"environment"`
	}

	resp := infoResponse{
		Name:        c.Config.App.Name,
		Version:     c.Config.App.Version,
		Environment: string(c.Config.App.Environment),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		server.WriteJSON(w, http.StatusOK, resp)
	}
}

// Shutdown gracefully shuts down all components in reverse order.
func (c *Container) Shutdown(ctx context.Context) {
	c.Logger.Info("shutting down application")

	// Shutdown HTTP server first (stop accepting new requests)
	if c.Server != nil {
		if err := c.Server.Shutdown(ctx); err != nil {
			c.Logger.Error("http server shutdown error", "error", err)
		}
	}

	// Close Redis
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			c.Logger.Error("redis close error", "error", err)
		}
	}

	// Close Database (last, so in-flight requests can complete)
	if c.Database != nil {
		c.Database.Close()
	}

	c.Logger.Info("application shutdown complete")
}
