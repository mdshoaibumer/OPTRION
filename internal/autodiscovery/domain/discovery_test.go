package domain_test

import (
	"testing"
	"time"

	"github.com/optrion/optrion/internal/autodiscovery/domain"
)

func TestDefaultDiscoveryConfig(t *testing.T) {
	cfg := domain.DefaultDiscoveryConfig()

	if cfg.Strategy != domain.StrategyEnvVars {
		t.Errorf("expected strategy env-vars, got %s", cfg.Strategy)
	}
	if !cfg.DetectPostgreSQL {
		t.Error("expected DetectPostgreSQL to be true")
	}
	if !cfg.DetectRedis {
		t.Error("expected DetectRedis to be true")
	}
	if !cfg.DetectHTTPServices {
		t.Error("expected DetectHTTPServices to be true")
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", cfg.Timeout)
	}
	if cfg.Retries != 3 {
		t.Errorf("expected 3 retries, got %d", cfg.Retries)
	}
}

func TestPostgreSQLDiscovery_ToComponent(t *testing.T) {
	pg := &domain.PostgreSQLDiscovery{
		Host:     "localhost",
		Port:     5432,
		Database: "optrion",
		User:     "admin",
		Version:  "16.1",
		Healthy:  true,
	}

	comp := pg.ToComponent()

	if comp.Name != "PostgreSQL" {
		t.Errorf("expected name PostgreSQL, got %s", comp.Name)
	}
	if comp.Kind != "database" {
		t.Errorf("expected kind database, got %s", comp.Kind)
	}
	if comp.Endpoint != "localhost" {
		t.Errorf("expected endpoint localhost, got %s", comp.Endpoint)
	}
	if comp.Port != 5432 {
		t.Errorf("expected port 5432, got %d", comp.Port)
	}
	if !comp.Healthy {
		t.Error("expected healthy to be true")
	}
	if comp.Protocol != "postgresql" {
		t.Errorf("expected protocol postgresql, got %s", comp.Protocol)
	}
	if comp.Version != "16.1" {
		t.Errorf("expected version 16.1, got %s", comp.Version)
	}
	if comp.Metadata["database"] != "optrion" {
		t.Errorf("expected metadata database optrion, got %v", comp.Metadata["database"])
	}
}

func TestRedisDiscovery_ToComponent(t *testing.T) {
	rd := &domain.RedisDiscovery{
		Host:    "redis.local",
		Port:    6379,
		Version: "7.2.4",
		Healthy: true,
	}

	comp := rd.ToComponent()

	if comp.Name != "Redis" {
		t.Errorf("expected name Redis, got %s", comp.Name)
	}
	if comp.Kind != "cache" {
		t.Errorf("expected kind cache, got %s", comp.Kind)
	}
	if comp.Port != 6379 {
		t.Errorf("expected port 6379, got %d", comp.Port)
	}
	if comp.Protocol != "redis" {
		t.Errorf("expected protocol redis, got %s", comp.Protocol)
	}
}

func TestHTTPServiceDiscovery_ToComponent(t *testing.T) {
	svc := &domain.HTTPServiceDiscovery{
		Name:            "user-service",
		Endpoint:        "http://user-svc:8080",
		Port:            8080,
		HealthCheckPath: "/healthz",
		Version:         "1.2.3",
		Healthy:         true,
		ResponseTime:    150 * time.Millisecond,
	}

	comp := svc.ToComponent()

	if comp.Name != "user-service" {
		t.Errorf("expected name user-service, got %s", comp.Name)
	}
	if comp.Kind != "api" {
		t.Errorf("expected kind api, got %s", comp.Kind)
	}
	if comp.Port != 8080 {
		t.Errorf("expected port 8080, got %d", comp.Port)
	}
	if comp.Protocol != "http" {
		t.Errorf("expected protocol http, got %s", comp.Protocol)
	}
	if comp.Metadata["response_time_ms"] != int64(150) {
		t.Errorf("expected 150ms response time, got %v", comp.Metadata["response_time_ms"])
	}
}

func TestDiscoveryResult_Empty(t *testing.T) {
	result := &domain.DiscoveryResult{
		Components: make([]*domain.DiscoveredComponent, 0),
		Timestamp:  time.Now(),
	}

	if len(result.Components) != 0 {
		t.Errorf("expected 0 components, got %d", len(result.Components))
	}
}

func TestDiscoveryStrategies(t *testing.T) {
	strategies := []domain.DiscoveryStrategy{
		domain.StrategyEnvVars,
		domain.StrategyDefaults,
		domain.StrategyDNS,
		domain.StrategyKubernetes,
	}

	expected := []string{"env-vars", "defaults", "dns", "kubernetes"}
	for i, s := range strategies {
		if string(s) != expected[i] {
			t.Errorf("expected strategy %s, got %s", expected[i], s)
		}
	}
}
