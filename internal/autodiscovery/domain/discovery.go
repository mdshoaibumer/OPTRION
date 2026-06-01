package domain

import (
	"time"
)

// DiscoveryResult represents the result of auto-discovery.
type DiscoveryResult struct {
	Components []*DiscoveredComponent
	Timestamp  time.Time
}

// DiscoveredComponent represents a discovered component.
type DiscoveredComponent struct {
	Name        string
	Kind        string
	Endpoint    string
	Port        int
	Healthy     bool
	Protocol    string
	Version     string
	Metadata    map[string]interface{}
	DiscoveredAt time.Time
}

// PostgreSQLDiscovery represents discovered PostgreSQL instance.
type PostgreSQLDiscovery struct {
	Host     string
	Port     int
	Database string
	User     string
	Version  string
	Healthy  bool
}

// ToComponent converts PostgreSQL discovery to a component.
func (pg *PostgreSQLDiscovery) ToComponent() *DiscoveredComponent {
	return &DiscoveredComponent{
		Name:     "PostgreSQL",
		Kind:     "database",
		Endpoint: pg.Host,
		Port:     pg.Port,
		Healthy:  pg.Healthy,
		Protocol: "postgresql",
		Version:  pg.Version,
		Metadata: map[string]interface{}{
			"database": pg.Database,
			"user":     pg.User,
		},
		DiscoveredAt: time.Now().UTC(),
	}
}

// RedisDiscovery represents discovered Redis instance.
type RedisDiscovery struct {
	Host    string
	Port    int
	Version string
	Healthy bool
}

// ToComponent converts Redis discovery to a component.
func (r *RedisDiscovery) ToComponent() *DiscoveredComponent {
	return &DiscoveredComponent{
		Name:     "Redis",
		Kind:     "cache",
		Endpoint: r.Host,
		Port:     r.Port,
		Healthy:  r.Healthy,
		Protocol: "redis",
		Version:  r.Version,
		Metadata: map[string]interface{}{
			"connection_type": "standalone",
		},
		DiscoveredAt: time.Now().UTC(),
	}
}

// HTTPServiceDiscovery represents discovered HTTP service.
type HTTPServiceDiscovery struct {
	Name            string
	Endpoint        string
	Port            int
	HealthCheckPath string
	Version         string
	Healthy         bool
	ResponseTime    time.Duration
}

// ToComponent converts HTTP service discovery to a component.
func (h *HTTPServiceDiscovery) ToComponent() *DiscoveredComponent {
	return &DiscoveredComponent{
		Name:     h.Name,
		Kind:     "api",
		Endpoint: h.Endpoint,
		Port:     h.Port,
		Healthy:  h.Healthy,
		Protocol: "http",
		Version:  h.Version,
		Metadata: map[string]interface{}{
			"health_check_path": h.HealthCheckPath,
			"response_time_ms":  h.ResponseTime.Milliseconds(),
		},
		DiscoveredAt: time.Now().UTC(),
	}
}

// DiscoveryStrategy defines how components are discovered.
type DiscoveryStrategy string

const (
	StrategyEnvVars  DiscoveryStrategy = "env-vars"
	StrategyDefaults DiscoveryStrategy = "defaults"
	StrategyDNS      DiscoveryStrategy = "dns"
	StrategyKubernetes DiscoveryStrategy = "kubernetes"
)

// DiscoveryConfig holds configuration for auto-discovery.
type DiscoveryConfig struct {
	Strategy           DiscoveryStrategy
	DetectPostgreSQL   bool
	DetectRedis        bool
	DetectHTTPServices bool
	Timeout            time.Duration
	Retries            int
}

// DefaultDiscoveryConfig returns sensible defaults for discovery.
func DefaultDiscoveryConfig() DiscoveryConfig {
	return DiscoveryConfig{
		Strategy:           StrategyEnvVars,
		DetectPostgreSQL:   true,
		DetectRedis:        true,
		DetectHTTPServices: true,
		Timeout:            5 * time.Second,
		Retries:            3,
	}
}
