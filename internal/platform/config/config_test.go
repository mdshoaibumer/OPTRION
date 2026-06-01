package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear relevant env vars to test defaults
	envVars := []string{
		"APP_NAME", "APP_VERSION", "APP_ENV",
		"HTTP_HOST", "HTTP_PORT",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSL_MODE",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD",
		"LOG_LEVEL", "LOG_FORMAT",
	}
	for _, key := range envVars {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// App defaults
	if cfg.App.Name != "optrion" {
		t.Errorf("expected app name 'optrion', got %q", cfg.App.Name)
	}
	if cfg.App.Environment != EnvLocal {
		t.Errorf("expected environment 'local', got %q", cfg.App.Environment)
	}

	// HTTP defaults
	if cfg.HTTP.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.HTTP.Port)
	}
	if cfg.HTTP.ReadTimeout != 10*time.Second {
		t.Errorf("expected read timeout 10s, got %v", cfg.HTTP.ReadTimeout)
	}

	// Database defaults
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected db host 'localhost', got %q", cfg.Database.Host)
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("expected max open conns 25, got %d", cfg.Database.MaxOpenConns)
	}

	// Redis defaults
	if cfg.Redis.Host != "localhost" {
		t.Errorf("expected redis host 'localhost', got %q", cfg.Redis.Host)
	}
	if cfg.Redis.PoolSize != 10 {
		t.Errorf("expected redis pool size 10, got %d", cfg.Redis.PoolSize)
	}

	// Log defaults
	if cfg.Log.Level != "info" {
		t.Errorf("expected log level 'info', got %q", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("expected log format 'json', got %q", cfg.Log.Format)
	}
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "admin")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "optrion_dev")
	t.Setenv("REDIS_HOST", "redis.internal")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.App.Environment != EnvDevelopment {
		t.Errorf("expected environment 'development', got %q", cfg.App.Environment)
	}
	if cfg.HTTP.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.HTTP.Port)
	}
	if cfg.Database.Host != "db.internal" {
		t.Errorf("expected db host 'db.internal', got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("expected db port 5433, got %d", cfg.Database.Port)
	}
}

func TestLoad_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name: "invalid environment",
			envVars: map[string]string{
				"APP_ENV": "staging_invalid",
			},
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"HTTP_PORT": "99999",
			},
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"LOG_LEVEL": "verbose",
			},
		},
		{
			name: "invalid log format",
			envVars: map[string]string{
				"LOG_FORMAT": "yaml",
			},
		},
		{
			name: "production without ssl",
			envVars: map[string]string{
				"APP_ENV":     "production",
				"DB_SSL_MODE": "disable",
				"DB_PASSWORD": "secret",
				"REDIS_HOST":  "localhost",
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_NAME":     "db",
				"LOG_LEVEL":   "info",
				"LOG_FORMAT":  "json",
				"HTTP_PORT":   "8080",
			},
		},
		{
			name: "production without password",
			envVars: map[string]string{
				"APP_ENV":     "production",
				"DB_SSL_MODE": "require",
				"DB_PASSWORD": "",
				"REDIS_HOST":  "localhost",
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_NAME":     "db",
				"LOG_LEVEL":   "info",
				"LOG_FORMAT":  "json",
				"HTTP_PORT":   "8080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset environment
			os.Clearenv()
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			_, err := Load()
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "optrion",
		Password: "secret",
		Name:     "optrion",
		SSLMode:  "disable",
	}

	expected := "postgres://optrion:secret@localhost:5432/optrion?sslmode=disable"
	if got := cfg.DSN(); got != expected {
		t.Errorf("expected DSN %q, got %q", expected, got)
	}
}

func TestRedisConfig_Addr(t *testing.T) {
	cfg := RedisConfig{
		Host: "redis.internal",
		Port: 6380,
	}

	expected := "redis.internal:6380"
	if got := cfg.Addr(); got != expected {
		t.Errorf("expected addr %q, got %q", expected, got)
	}
}

func TestLoad_ProductionCORSWildcardRejected(t *testing.T) {
	os.Clearenv()
	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_SSL_MODE", "require")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("REDIS_HOST", "localhost")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_NAME", "db")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com,*")

	_, err := Load()
	if err == nil {
		t.Error("expected validation error for CORS wildcard in production, got nil")
	}
}

func TestLoad_AIConfigDefaults(t *testing.T) {
	os.Clearenv()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AI.Provider != "gemini" {
		t.Errorf("expected default AI provider 'gemini', got %q", cfg.AI.Provider)
	}
	if cfg.AI.Model != "gemini-2.0-flash" {
		t.Errorf("expected default AI model 'gemini-2.0-flash', got %q", cfg.AI.Model)
	}
	if cfg.AI.MaxTokens != 4096 {
		t.Errorf("expected default max tokens 4096, got %d", cfg.AI.MaxTokens)
	}
	if cfg.AI.Enabled {
		t.Error("expected AI disabled by default")
	}
}
