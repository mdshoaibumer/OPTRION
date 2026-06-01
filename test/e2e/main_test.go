package e2e

import (
	"log"
	"os"
	"testing"

	"github.com/optrion/optrion/test/testutil"
)

func TestMain(m *testing.M) {
	root, err := testutil.FindRepoRoot()
	if err != nil {
		log.Fatalf("failed to locate repo root: %v", err)
	}

	testutil.LoadDotEnv()

	if err := testutil.EnsureDockerDependencies(root); err != nil {
		log.Fatalf("failed to ensure docker dependencies: %v", err)
	}

	code := m.Run()

	if err := testutil.StopDockerDependencies(root); err != nil {
		log.Printf("warning: failed to stop docker dependencies: %v", err)
	}

	os.Exit(code)
}
