package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/optrion/optrion/internal/platform/config"
)

// Redis wraps a redis.Client with lifecycle management.
type Redis struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedis creates and validates a Redis connection.
func NewRedis(ctx context.Context, cfg config.RedisConfig, logger *slog.Logger) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
	})

	// Verify connectivity
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	logger.Info("redis connected",
		"addr", cfg.Addr(),
		"db", cfg.DB,
		"pool_size", cfg.PoolSize,
	)

	return &Redis{
		client: client,
		logger: logger,
	}, nil
}

// Client returns the underlying redis client for use by adapters.
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Health checks Redis connectivity.
func (r *Redis) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}

// Close gracefully shuts down the Redis connection.
func (r *Redis) Close() error {
	r.logger.Info("closing redis connection")
	return r.client.Close()
}
