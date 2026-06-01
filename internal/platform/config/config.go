package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment represents the application environment.
type Environment string

const (
	EnvLocal       Environment = "local"
	EnvDevelopment Environment = "development"
	EnvProduction  Environment = "production"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Log      LogConfig
	Auth     AuthConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name        string
	Version     string
	Environment Environment
}

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	CORSOrigins     []string
	RateLimitRPS    int
}

// AuthConfig holds API key authentication settings.
type AuthConfig struct {
	Enabled bool
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DSN returns the PostgreSQL connection string.
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
}

// Addr returns the Redis address string.
func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string
	Format string // "json" or "text"
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        envOrDefault("APP_NAME", "optrion"),
			Version:     envOrDefault("APP_VERSION", "0.1.0"),
			Environment: Environment(envOrDefault("APP_ENV", "local")),
		},
		HTTP: HTTPConfig{
			Host:            envOrDefault("HTTP_HOST", "0.0.0.0"),
			Port:            envOrDefaultInt("HTTP_PORT", 8080),
			ReadTimeout:     envOrDefaultDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    envOrDefaultDuration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     envOrDefaultDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envOrDefaultDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
			CORSOrigins:     envOrDefaultSlice("CORS_ALLOWED_ORIGINS", []string{}),
			RateLimitRPS:    envOrDefaultInt("RATE_LIMIT_RPS", 100),
		},
		Database: DatabaseConfig{
			Host:            envOrDefault("DB_HOST", "localhost"),
			Port:            envOrDefaultInt("DB_PORT", 5432),
			User:            envOrDefault("DB_USER", "optrion"),
			Password:        os.Getenv("DB_PASSWORD"),
			Name:            envOrDefault("DB_NAME", "optrion"),
			SSLMode:         envOrDefault("DB_SSL_MODE", "disable"),
			MaxOpenConns:    envOrDefaultInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    envOrDefaultInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: envOrDefaultDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: envOrDefaultDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:         envOrDefault("REDIS_HOST", "localhost"),
			Port:         envOrDefaultInt("REDIS_PORT", 6379),
			Password:     envOrDefault("REDIS_PASSWORD", ""),
			DB:           envOrDefaultInt("REDIS_DB", 0),
			MaxRetries:   envOrDefaultInt("REDIS_MAX_RETRIES", 3),
			DialTimeout:  envOrDefaultDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  envOrDefaultDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: envOrDefaultDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolSize:     envOrDefaultInt("REDIS_POOL_SIZE", 10),
		},
		Log: LogConfig{
			Level:  envOrDefault("LOG_LEVEL", "info"),
			Format: envOrDefault("LOG_FORMAT", "json"),
		},
		Auth: AuthConfig{
			Enabled: envOrDefaultBool("AUTH_ENABLED", true),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	var errs []string

	// Validate environment
	switch c.App.Environment {
	case EnvLocal, EnvDevelopment, EnvProduction:
		// valid
	default:
		errs = append(errs, fmt.Sprintf("invalid APP_ENV: %q (must be local, development, or production)", c.App.Environment))
	}

	// Validate HTTP
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		errs = append(errs, fmt.Sprintf("invalid HTTP_PORT: %d (must be 1-65535)", c.HTTP.Port))
	}

	// Validate Database
	if c.Database.Host == "" {
		errs = append(errs, "DB_HOST is required")
	}
	if c.Database.User == "" {
		errs = append(errs, "DB_USER is required")
	}
	if c.Database.Name == "" {
		errs = append(errs, "DB_NAME is required")
	}
	if c.Database.MaxOpenConns < 1 {
		errs = append(errs, "DB_MAX_OPEN_CONNS must be >= 1")
	}
	if c.Database.MaxIdleConns < 0 {
		errs = append(errs, "DB_MAX_IDLE_CONNS must be >= 0")
	}

	// Validate Redis
	if c.Redis.Host == "" {
		errs = append(errs, "REDIS_HOST is required")
	}
	if c.Redis.PoolSize < 1 {
		errs = append(errs, "REDIS_POOL_SIZE must be >= 1")
	}

	// Validate Log
	switch strings.ToLower(c.Log.Level) {
	case "debug", "info", "warn", "error":
		// valid
	default:
		errs = append(errs, fmt.Sprintf("invalid LOG_LEVEL: %q (must be debug, info, warn, or error)", c.Log.Level))
	}

	switch strings.ToLower(c.Log.Format) {
	case "json", "text":
		// valid
	default:
		errs = append(errs, fmt.Sprintf("invalid LOG_FORMAT: %q (must be json or text)", c.Log.Format))
	}

	// Production-specific validations
	if c.App.Environment == EnvProduction {
		if c.Database.SSLMode == "disable" {
			errs = append(errs, "DB_SSL_MODE must not be 'disable' in production")
		}
		if c.Database.Password == "" {
			errs = append(errs, "DB_PASSWORD is required in production")
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func envOrDefaultInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return i
}

func envOrDefaultDuration(key string, defaultValue time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultValue
	}
	return d
}

func envOrDefaultSlice(key string, defaultValue []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func envOrDefaultBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultValue
	}
	return b
}
