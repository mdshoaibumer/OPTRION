package app

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	alertpg "github.com/optrion/optrion/internal/alert/adapter/repository"
	"github.com/optrion/optrion/internal/alert/adapter/telegram"
	alertevent "github.com/optrion/optrion/internal/alert/app/event"
	alertservice "github.com/optrion/optrion/internal/alert/app/service"
	healthpg "github.com/optrion/optrion/internal/health/adapter/postgres"
	healthrest "github.com/optrion/optrion/internal/health/adapter/rest"
	"github.com/optrion/optrion/internal/health/anomaly"
	healthapp "github.com/optrion/optrion/internal/health/app"
	"github.com/optrion/optrion/internal/health/collector"
	"github.com/optrion/optrion/internal/health/port"
	"github.com/optrion/optrion/internal/health/scheduler"
	"github.com/optrion/optrion/internal/health/scoring"
	incidentpg "github.com/optrion/optrion/internal/incident/adapter/postgres"
	incidentrest "github.com/optrion/optrion/internal/incident/adapter/rest"
	incidentapp "github.com/optrion/optrion/internal/incident/app"
	"github.com/optrion/optrion/internal/platform/cache"
	"github.com/optrion/optrion/internal/platform/config"
	"github.com/optrion/optrion/internal/platform/database"
	"github.com/optrion/optrion/internal/platform/eventbus"
	"github.com/optrion/optrion/internal/platform/logger"
	"github.com/optrion/optrion/internal/platform/server"
	tenantpg "github.com/optrion/optrion/internal/tenant/adapter/postgres"
	tenantrest "github.com/optrion/optrion/internal/tenant/adapter/rest"
	tenantapp "github.com/optrion/optrion/internal/tenant/app"
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

	// Services
	TenantService   *tenantapp.TenantService
	HealthService   *healthapp.HealthService
	IncidentService *incidentapp.IncidentService
	AlertEngine     alertservice.AlertEngine

	// Infrastructure
	EventBus  *eventbus.Bus
	Scheduler *scheduler.Scheduler

	// Internal components for lifecycle management
	Router *server.Router
	health *server.HealthHandler
}

// Migrations is set by main.go to provide embedded migration files.
var Migrations embed.FS

// NewContainer builds the application container with the full dependency graph.
// Order: Config → Logger → Database → Redis → EventBus → Services → Router → Server
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

	// 5. Run migrations
	migrator := database.NewMigrator(db.Pool(), c.Logger.With("component", "migrator"))
	if err := migrator.MigrateFS(ctx, Migrations, "."); err != nil {
		c.Database.Close()
		_ = redis.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	// 6. Event Bus
	c.EventBus = eventbus.New(c.Logger.With("component", "eventbus"))

	// 7. Wire tenant bounded context
	pool := db.Pool()
	tenantRepo := tenantpg.NewTenantRepository(pool)
	productRepo := tenantpg.NewProductRepository(pool)
	environmentRepo := tenantpg.NewEnvironmentRepository(pool)
	componentRepo := tenantpg.NewComponentRepository(pool)
	auditRepo := tenantpg.NewAuditRepository(pool)
	uow := tenantpg.NewUnitOfWork(pool)

	c.TenantService = tenantapp.NewTenantService(
		tenantRepo, productRepo, environmentRepo, componentRepo,
		auditRepo, uow, c.Logger.With("component", "tenant"),
	)

	// 8. Wire health monitoring bounded context
	metricRepo := healthpg.NewHealthMetricRepository(pool)
	snapshotRepo := healthpg.NewSnapshotRepository(pool)
	scoreRepo := healthpg.NewScoreRepository(pool)
	componentHealthRepo := healthpg.NewComponentHealthRepository(pool)
	anomalyRepo := healthpg.NewAnomalyRepository(pool)
	scoringEngine := scoring.NewEngine()
	detector := anomaly.NewDetector(3.0, 60, 10)

	c.HealthService = healthapp.NewHealthService(
		metricRepo, snapshotRepo, scoreRepo, componentHealthRepo,
		anomalyRepo, scoringEngine, detector, c.Logger.With("component", "health-monitor"),
	)

	// 9. Scheduler (collectors registered via seed/bootstrap)
	c.Scheduler = scheduler.NewScheduler(
		func(sctx context.Context, result *port.CollectorResult) {
			c.HealthService.ProcessCollectorResult(sctx, result)
		},
		c.Logger.With("component", "scheduler"),
	)

	// Register internal collectors for OPTRION's own infrastructure
	serverCollector := collector.NewServerCollector("system", "optrion-server")
	c.Scheduler.Register(serverCollector, 60*time.Second)

	// 10. Wire incident intelligence bounded context
	incidentRepo := incidentpg.NewIncidentRepository(pool)
	eventRepo := incidentpg.NewEventRepository(pool)
	ruleRepo := incidentpg.NewRuleRepository(pool)
	commentRepo := incidentpg.NewCommentRepository(pool)
	timelineRepo := incidentpg.NewTimelineRepository(pool)

	c.IncidentService = incidentapp.NewIncidentService(
		incidentRepo, eventRepo, ruleRepo, commentRepo, timelineRepo,
		c.Logger.With("component", "incident"),
	)

	// 11. Wire alert engine
	telegramSender := telegram.NewTelegramSender(30, 2*time.Second)
	dedup := alertservice.NewDeduplicationService(5 * time.Minute)
	alertRuleRepo := alertpg.NewAlertRulePostgresRepository(pool)
	alertRepo := alertpg.NewAlertPostgresRepository(pool)
	alertChannelRepo := alertpg.NewAlertChannelPostgresRepository(pool)
	alertDeliveryRepo := alertpg.NewAlertDeliveryPostgresRepository(pool)
	c.AlertEngine = alertservice.NewAlertEngine(
		alertRuleRepo, alertRepo, alertChannelRepo, alertDeliveryRepo,
		dedup, telegramSender,
		c.Logger.With("component", "alert-engine"),
	)

	// 12. Subscribe event bus handlers (cross-context communication)
	c.EventBus.Subscribe("incident.opened", func(ctx context.Context, evt eventbus.Event) error {
		incEvt, ok := evt.(eventbus.IncidentOpenedEvent)
		if !ok {
			return nil
		}
		return c.AlertEngine.ProcessEvent(ctx, alertEventFromIncidentOpened(incEvt))
	})

	// 13. API Key authentication
	apiKeyRepo := tenantpg.NewAPIKeyRepository(pool)

	// 14. Health handler
	c.health = server.NewHealthHandler(db, redis, cfg.App.Version, c.Logger.With("component", "health"))

	// 15. Router (with auth, rate limiting, CORS config)
	routerCfg := server.RouterConfig{
		CORSOrigins:   cfg.HTTP.CORSOrigins,
		RateLimitRPS:  cfg.HTTP.RateLimitRPS,
		AuthEnabled:   cfg.Auth.Enabled,
		AuthValidator: apiKeyRepo,
	}
	c.Router = server.NewRouter(c.Logger.With("component", "http"), routerCfg)
	c.registerRoutes()

	// 16. HTTP Server
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	c.Server = server.NewServer(
		addr,
		c.Router.Handler(),
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
	c.Router.HandleFunc("GET /healthz", c.health.Liveness())
	c.Router.HandleFunc("GET /readyz", c.health.Readiness())

	// API version info
	c.Router.HandleFunc("GET /api/v1/info", c.infoHandler())

	// Tenant domain routes
	tenantHandler := tenantrest.NewHandler(c.TenantService, c.Logger.With("component", "tenant-api"))
	tenantHandler.RegisterRoutes(c.Router.Mux())

	// Health monitoring routes
	healthHandler := healthrest.NewHandler(c.HealthService, c.Logger.With("component", "health-api"))
	healthHandler.RegisterRoutes(c.Router.Mux())

	// Incident intelligence routes
	incidentHandler := incidentrest.NewHandler(c.IncidentService, c.Logger.With("component", "incident-api"))
	incidentHandler.RegisterRoutes(c.Router.Mux())
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

	// Stop scheduler first
	if c.Scheduler != nil {
		c.Scheduler.Stop()
	}

	// Shutdown HTTP server (stop accepting new requests)
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

// alertEventFromIncidentOpened converts an event bus event to an alert engine event.
func alertEventFromIncidentOpened(evt eventbus.IncidentOpenedEvent) alertevent.IncidentEvent {
	return alertevent.IncidentEvent{
		ID:         evt.IncidentID,
		TenantID:   evt.Tenant,
		IncidentID: evt.IncidentID,
		Type:       alertevent.IncidentOpened,
		Payload: map[string]interface{}{
			"component_id": evt.ComponentID,
			"title":        evt.Title,
			"severity":     evt.Severity,
		},
	}
}
