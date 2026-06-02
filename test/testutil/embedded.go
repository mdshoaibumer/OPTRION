package testutil

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/alicebob/miniredis/v2"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

// EmbeddedPostgres wraps an embedded PostgreSQL instance for testing without Docker.
type EmbeddedPostgres struct {
	db   *embeddedpostgres.EmbeddedPostgres
	Port uint32
}

// StartEmbeddedPostgres starts a local embedded PostgreSQL instance.
// No installation required — the library downloads a pre-built binary on first run.
func StartEmbeddedPostgres() (*EmbeddedPostgres, error) {
	port, err := freePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find free port: %w", err)
	}

	cacheDir := filepath.Join(os.TempDir(), "optrion-embedded-pg")

	db := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(uint32(port)).
			Username("optrion").
			Password("optrion_test").
			Database("optrion").
			RuntimePath(filepath.Join(cacheDir, fmt.Sprintf("runtime-%d", port))).
			DataPath(filepath.Join(cacheDir, fmt.Sprintf("data-%d", port))).
			BinariesPath(filepath.Join(cacheDir, "bin")).
			StartTimeout(90 * time.Second).
			Version(embeddedpostgres.V16),
	)

	if err := db.Start(); err != nil {
		return nil, fmt.Errorf("failed to start embedded postgres: %w", err)
	}

	return &EmbeddedPostgres{db: db, Port: uint32(port)}, nil
}

// Stop shuts down the embedded PostgreSQL instance.
func (ep *EmbeddedPostgres) Stop() error {
	if ep.db != nil {
		return ep.db.Stop()
	}
	return nil
}

// SetEnv configures environment variables so config.Load() picks up this instance.
func (ep *EmbeddedPostgres) SetEnv() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", fmt.Sprintf("%d", ep.Port))
	os.Setenv("DB_USER", "optrion")
	os.Setenv("DB_PASSWORD", "optrion_test")
	os.Setenv("DB_NAME", "optrion")
	os.Setenv("DB_SSL_MODE", "disable")
}

// EmbeddedPostgresSupported checks if the current OS/arch is supported.
func EmbeddedPostgresSupported() bool {
	// embedded-postgres supports: linux, darwin, windows on amd64 and arm64
	switch runtime.GOOS {
	case "linux", "darwin", "windows":
		switch runtime.GOARCH {
		case "amd64", "arm64":
			return true
		}
	}
	return false
}

// MiniRedis wraps a miniredis instance for testing without a real Redis server.
type MiniRedis struct {
	server *miniredis.Miniredis
}

// StartMiniRedis starts an in-memory Redis-compatible server.
func StartMiniRedis() (*MiniRedis, error) {
	s, err := miniredis.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to start miniredis: %w", err)
	}
	return &MiniRedis{server: s}, nil
}

// Stop shuts down the miniredis server.
func (mr *MiniRedis) Stop() {
	if mr.server != nil {
		mr.server.Close()
	}
}

// SetEnv configures environment variables so config.Load() picks up miniredis.
func (mr *MiniRedis) SetEnv() {
	os.Setenv("REDIS_HOST", mr.server.Host())
	os.Setenv("REDIS_PORT", mr.server.Port())
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("REDIS_DB", "0")
}

// freePort finds an available TCP port.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port, nil
}
