package e2e

import (
	"log"
	"os"
	"testing"

	"github.com/optrion/optrion/test/testutil"
)

func TestMain(m *testing.M) {
	testutil.LoadDotEnv()

	// Use embedded PostgreSQL + miniredis if SKIP_DOCKER is set (no Docker/install needed)
	if os.Getenv("SKIP_DOCKER") == "true" {
		runWithEmbedded(m)
		return
	}

	// Default: Docker-based E2E
	root, err := testutil.FindRepoRoot()
	if err != nil {
		log.Fatalf("failed to locate repo root: %v", err)
	}

	if err := testutil.EnsureDockerDependencies(root); err != nil {
		log.Fatalf("failed to ensure docker dependencies: %v", err)
	}

	code := m.Run()

	if err := testutil.StopDockerDependencies(root); err != nil {
		log.Printf("warning: failed to stop docker dependencies: %v", err)
	}

	os.Exit(code)
}

func runWithEmbedded(m *testing.M) {
	if !testutil.EmbeddedPostgresSupported() {
		log.Fatalf("embedded postgres is not supported on this OS/arch")
	}

	log.Println("Starting embedded PostgreSQL (no Docker)...")
	pg, err := testutil.StartEmbeddedPostgres()
	if err != nil {
		log.Fatalf("failed to start embedded postgres: %v", err)
	}
	pg.SetEnv()

	log.Println("Starting miniredis (no Docker)...")
	mr, err := testutil.StartMiniRedis()
	if err != nil {
		pg.Stop()
		log.Fatalf("failed to start miniredis: %v", err)
	}
	mr.SetEnv()

	code := m.Run()

	mr.Stop()
	if err := pg.Stop(); err != nil {
		log.Printf("warning: failed to stop embedded postgres: %v", err)
	}

	os.Exit(code)
}
