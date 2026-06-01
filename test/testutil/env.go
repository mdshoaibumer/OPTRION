package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"github.com/optrion/optrion/internal/app"
	"github.com/optrion/optrion/internal/platform/config"
	"github.com/optrion/optrion/migrations"
)

type TestEnv struct {
	AppContainer *app.Container
	Server       *httptest.Server
	Client       *http.Client
	DBName       string
	RedisClient  *redis.Client
}

func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	LoadDotEnv()

	ctx := context.Background()

	// 1. Load config and randomize DB name
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Generate a unique suffix for the database
	randBytes := make([]byte, 4)
	if _, err := rand.Read(randBytes); err != nil {
		t.Fatalf("Failed to generate random database suffix: %v", err)
	}
	dbName := fmt.Sprintf("optrion_e2e_%s", hex.EncodeToString(randBytes))

	// Connect to default 'postgres' database to create our test database
	rootDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/postgres?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.SSLMode,
	)
	conn, err := pgx.Connect(ctx, rootDSN)
	if err != nil {
		t.Fatalf("Failed to connect to root database to create test DB: %v", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		t.Fatalf("Failed to create test database %s: %v", dbName, err)
	}

	// Override config with our new DB name and Redis DB 15
	cfg.Database.Name = dbName
	cfg.Redis.DB = 15

	// Set environment variables so config.Load() picks them up inside app bootstrap
	t.Setenv("DB_NAME", dbName)
	t.Setenv("REDIS_DB", "15")

	// 2. Set up application container
	app.Migrations = migrations.FS
	container, err := app.NewContainer(ctx)
	if err != nil {
		// Clean up the DB since container creation failed
		_, _ = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", dbName))
		t.Fatalf("Failed to create application container: %v", err)
	}

	// Disable health check scheduler triggers to avoid interference
	container.Scheduler.Stop()

	// Start test server using the exported Router handler
	ts := httptest.NewServer(container.Router.Handler())

	// Set up Redis client for direct validation/flush
	rClient := redis.NewClient(&redis.Options{
		Addr:     container.Config.Redis.Addr(),
		Password: container.Config.Redis.Password,
		DB:       15,
	})

	// Flush Redis to ensure isolation
	if err := rClient.FlushDB(ctx).Err(); err != nil {
		t.Logf("Warning: failed to flush Redis: %v", err)
	}

	return &TestEnv{
		AppContainer: container,
		Server:       ts,
		Client:       ts.Client(),
		DBName:       dbName,
		RedisClient:  rClient,
	}
}

func (env *TestEnv) Teardown(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	// 1. Shutdown container and server
	env.Server.Close()
	env.AppContainer.Shutdown(ctx)

	// 2. Drop the test database
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Teardown failed to load config: %v", err)
	}

	rootDSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/postgres?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.SSLMode,
	)
	conn, err := pgx.Connect(ctx, rootDSN)
	if err != nil {
		t.Fatalf("Teardown failed to connect to root database to drop test DB: %v", err)
	}
	defer conn.Close(ctx)

	// Terminate other sessions and force drop
	_, _ = conn.Exec(ctx, fmt.Sprintf(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = '%s'
		AND pid <> pg_backend_pid();`, env.DBName))

	_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", env.DBName))
	if err != nil {
		t.Errorf("Failed to drop database %s: %v", env.DBName, err)
	}

	// 3. Flush and close Redis
	if env.RedisClient != nil {
		_ = env.RedisClient.FlushDB(ctx).Err()
		_ = env.RedisClient.Close()
	}
}

func FindRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", fmt.Errorf("repository root not found")
}

func LoadDotEnv() {
	root, err := FindRepoRoot()
	if err != nil {
		return
	}

	for _, filename := range []string{".env.local", ".env.development"} {
		path := filepath.Join(root, filename)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		_ = loadEnvFile(path)
	}

	if host := os.Getenv("DB_HOST"); host == "postgres" {
		_ = os.Setenv("DB_HOST", "localhost")
	}
	if host := os.Getenv("REDIS_HOST"); host == "redis" {
		_ = os.Setenv("REDIS_HOST", "localhost")
	}
}

func loadEnvFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}

	return nil
}
