package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/optrion/optrion/internal/autodiscovery/domain"
	"github.com/redis/go-redis/v9"
)

// DiscoveryService performs automatic discovery of infrastructure components.
type DiscoveryService struct {
	config domain.DiscoveryConfig
	logger *slog.Logger
}

// NewDiscoveryService creates a new DiscoveryService.
func NewDiscoveryService(logger *slog.Logger) *DiscoveryService {
	return &DiscoveryService{
		config: domain.DefaultDiscoveryConfig(),
		logger: logger,
	}
}

// Discover performs auto-discovery of all configured components.
func (ds *DiscoveryService) Discover(ctx context.Context) (*domain.DiscoveryResult, error) {
	result := &domain.DiscoveryResult{
		Components: make([]*domain.DiscoveredComponent, 0),
		Timestamp:  time.Now().UTC(),
	}

	// Discover PostgreSQL
	if ds.config.DetectPostgreSQL {
		pgComp, err := ds.DiscoverPostgreSQL(ctx)
		if err != nil {
			ds.logger.WarnContext(ctx, "PostgreSQL discovery failed", "error", err)
		} else if pgComp != nil {
			result.Components = append(result.Components, pgComp)
		}
	}

	// Discover Redis
	if ds.config.DetectRedis {
		redisComp, err := ds.DiscoverRedis(ctx)
		if err != nil {
			ds.logger.WarnContext(ctx, "Redis discovery failed", "error", err)
		} else if redisComp != nil {
			result.Components = append(result.Components, redisComp)
		}
	}

	// Discover HTTP services
	if ds.config.DetectHTTPServices {
		httpComps, err := ds.DiscoverHTTPServices(ctx)
		if err != nil {
			ds.logger.WarnContext(ctx, "HTTP services discovery failed", "error", err)
		} else {
			result.Components = append(result.Components, httpComps...)
		}
	}

	return result, nil
}

// DiscoverPostgreSQL attempts to discover PostgreSQL instances.
func (ds *DiscoveryService) DiscoverPostgreSQL(ctx context.Context) (*domain.DiscoveredComponent, error) {
	// Check common PostgreSQL connection string env vars
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = os.Getenv("POSTGRES_URL")
	}
	if connStr == "" {
		// Try to construct from individual env vars
		host := os.Getenv("POSTGRES_HOST")
		if host == "" {
			host = os.Getenv("DB_HOST")
		}
		if host == "" {
			host = "localhost"
		}

		port := os.Getenv("POSTGRES_PORT")
		if port == "" {
			port = os.Getenv("DB_PORT")
		}
		if port == "" {
			port = "5432"
		}

		user := os.Getenv("POSTGRES_USER")
		if user == "" {
			user = os.Getenv("DB_USER")
		}
		if user == "" {
			user = "postgres"
		}

		password := os.Getenv("POSTGRES_PASSWORD")
		if password == "" {
			password = os.Getenv("DB_PASSWORD")
		}

		database := os.Getenv("POSTGRES_DB")
		if database == "" {
			database = os.Getenv("DB_NAME")
		}
		if database == "" {
			database = "postgres"
		}

		if password != "" {
			connStr = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, database)
		} else {
			connStr = fmt.Sprintf("postgresql://%s@%s:%s/%s", user, host, port, database)
		}
	}

	// Try to connect
	ctx, cancel := context.WithTimeout(ctx, ds.config.Timeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer conn.Close(ctx)

	// Get version
	var version string
	err = conn.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("failed to get PostgreSQL version: %w", err)
	}

	// Extract host and port
	config := conn.Config()
	host := config.Host
	port := int(config.Port)

	pgDiscovery := &domain.PostgreSQLDiscovery{
		Host:     host,
		Port:     port,
		Database: config.Database,
		User:     config.User,
		Version:  version,
		Healthy:  true,
	}

	ds.logger.InfoContext(ctx, "PostgreSQL discovered",
		"host", host,
		"port", port,
		"version", version,
	)

	return pgDiscovery.ToComponent(), nil
}

// DiscoverRedis attempts to discover Redis instances.
func (ds *DiscoveryService) DiscoverRedis(ctx context.Context) (*domain.DiscoveredComponent, error) {
	// Check common Redis env vars
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		host := os.Getenv("REDIS_HOST")
		if host == "" {
			host = "localhost"
		}

		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}

		redisURL = fmt.Sprintf("redis://%s:%s", host, port)
	}

	ctx, cancel := context.WithTimeout(ctx, ds.config.Timeout)
	defer cancel()

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	defer client.Close()

	// Test connection
	cmd := client.Ping(ctx)
	if cmd.Err() != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", cmd.Err())
	}

	// Get version
	info := client.Info(ctx, "server")
	if info.Err() != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", info.Err())
	}

	// Extract host and port
	parts := strings.Split(opt.Addr, ":")
	var port int
	if len(parts) > 1 {
		p, _ := strconv.Atoi(parts[1])
		port = p
	} else {
		port = 6379
	}

	redisDiscovery := &domain.RedisDiscovery{
		Host:    opt.Addr,
		Port:    port,
		Version: "unknown",
		Healthy: true,
	}

	ds.logger.InfoContext(ctx, "Redis discovered",
		"host", redisDiscovery.Host,
		"port", redisDiscovery.Port,
	)

	return redisDiscovery.ToComponent(), nil
}

// DiscoverHTTPServices attempts to discover HTTP service instances.
func (ds *DiscoveryService) DiscoverHTTPServices(ctx context.Context) ([]*domain.DiscoveredComponent, error) {
	components := make([]*domain.DiscoveredComponent, 0)

	// Check common service env vars
	services := []struct {
		name    string
		envVars []string
		port    int
	}{
		{
			name:    "Backend",
			envVars: []string{"SERVICE_URL", "API_URL", "BACKEND_URL"},
			port:    3000,
		},
		{
			name:    "Frontend",
			envVars: []string{"FRONTEND_URL", "WEB_URL"},
			port:    8080,
		},
		{
			name:    "Worker",
			envVars: []string{"WORKER_URL"},
			port:    9000,
		},
	}

	for _, svc := range services {
		var endpoint string
		for _, envVar := range svc.envVars {
			endpoint = os.Getenv(envVar)
			if endpoint != "" {
				break
			}
		}

		if endpoint == "" {
			// Try default localhost
			endpoint = fmt.Sprintf("http://localhost:%d", svc.port)
		}

		// Test connectivity
		ctx, cancel := context.WithTimeout(ctx, ds.config.Timeout)
		healthCheckPath := "/health"

		httpDiscovery := &domain.HTTPServiceDiscovery{
			Name:            svc.name,
			Endpoint:        endpoint,
			Port:            svc.port,
			HealthCheckPath: healthCheckPath,
			Healthy:         false,
		}

		// Try health check
		client := &http.Client{Timeout: ds.config.Timeout}
		resp, err := client.Get(fmt.Sprintf("%s%s", endpoint, healthCheckPath))
		if err == nil && resp.StatusCode == http.StatusOK {
			httpDiscovery.Healthy = true
			ds.logger.InfoContext(ctx, "HTTP service discovered",
				"name", svc.name,
				"endpoint", endpoint,
				"port", svc.port,
			)
			components = append(components, httpDiscovery.ToComponent())
		}

		cancel()
	}

	return components, nil
}

// PortIsOpen checks if a port is open on localhost.
func (ds *DiscoveryService) PortIsOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), ds.config.Timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
