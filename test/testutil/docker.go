package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"github.com/optrion/optrion/internal/platform/config"
)

const (
	composeProject  = "optrion_e2e"
	composeFileName = "deploy/docker/docker-compose.e2e.yml"
)

// EnsureDockerDependencies starts a dedicated PostgreSQL+Redis compose environment for E2E tests.
func EnsureDockerDependencies(root string) error {
	LoadDotEnv()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := ensureHostOverrides(); err != nil {
		return err
	}

	if err := startCompose(root, "postgres", "redis"); err != nil {
		if dbReachable(cfg) && redisReachable(cfg) {
			return nil
		}
		return err
	}

	postgresPort, err := discoverComposePublishedPort(root, "postgres", 5432)
	if err != nil {
		return fmt.Errorf("failed to resolve postgres published port: %w", err)
	}
	_ = os.Setenv("DB_PORT", strconv.Itoa(postgresPort))
	cfg.Database.Port = postgresPort

	redisPort, err := discoverComposePublishedPort(root, "redis", 6379)
	if err != nil {
		return fmt.Errorf("failed to resolve redis published port: %w", err)
	}
	_ = os.Setenv("REDIS_PORT", strconv.Itoa(redisPort))
	cfg.Redis.Port = redisPort

	if err := waitForPostgres(cfg, 60*time.Second); err != nil {
		return fmt.Errorf("postgres readiness failed: %w", err)
	}
	if err := waitForRedis(cfg, 60*time.Second); err != nil {
		return fmt.Errorf("redis readiness failed: %w", err)
	}

	return nil
}

// StopDockerDependencies shuts down any Docker Compose services started for tests.
func StopDockerDependencies(root string) error {
	composePath := filepath.Join(root, composeFileName)
	cmd := exec.Command("docker", "compose", "-f", composePath, "-p", composeProject, "down", "-v")
	cmd.Dir = filepath.Dir(composePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func startCompose(root string, services ...string) error {
	composePath := filepath.Join(root, composeFileName)
	if len(services) == 0 {
		services = []string{"postgres", "redis"}
	}
	args := append([]string{"compose", "-f", composePath, "-p", composeProject, "up", "-d"}, services...)
	cmd := exec.Command("docker", args...)
	cmd.Dir = filepath.Dir(composePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func dbReachable(cfg *config.Config) bool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=%s", cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.SSLMode)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return false
	}
	_ = conn.Close(ctx)
	return true
}

func redisReachable(cfg *config.Config) bool {
	rClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := rClient.Ping(ctx).Result()
	_ = rClient.Close()
	return err == nil
}

func waitForPostgres(cfg *config.Config, timeout time.Duration) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=%s", cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.SSLMode)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn, err := pgx.Connect(ctx, dsn)
			if err == nil {
				_ = conn.Close(ctx)
				return nil
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func waitForRedis(cfg *config.Config, timeout time.Duration) error {
	rClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := rClient.Ping(ctx).Result()
			if err == nil {
				return nil
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func discoverComposePublishedPort(root, service string, targetPort int) (int, error) {
	composePath := filepath.Join(root, composeFileName)
	args := []string{"compose", "-f", composePath, "-p", composeProject, "port", service, fmt.Sprintf("%d", targetPort)}
	cmd := exec.Command("docker", args...)
	cmd.Dir = filepath.Dir(composePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	value := strings.TrimSpace(string(output))
	idx := strings.LastIndex(value, ":")
	if idx < 0 {
		return 0, fmt.Errorf("unexpected docker compose port output: %q", value)
	}

	port, err := strconv.Atoi(value[idx+1:])
	if err != nil {
		return 0, fmt.Errorf("invalid published port from docker compose output: %w", err)
	}

	return port, nil
}

func ensureHostOverrides() error {
	if host := os.Getenv("DB_HOST"); host == "postgres" {
		return os.Setenv("DB_HOST", "localhost")
	}
	if host := os.Getenv("REDIS_HOST"); host == "redis" {
		return os.Setenv("REDIS_HOST", "localhost")
	}
	return nil
}
