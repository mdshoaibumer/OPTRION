package e2e

import (
	"context"
	"testing"

	"github.com/optrion/optrion/test/testutil"
)

func TestE2E_MigrationValidation(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	defer env.Teardown(t)

	ctx := context.Background()
	pool := env.AppContainer.Database.Pool()

	requiredTables := []string{"tenants", "health_metrics", "incidents", "incident_events", "schema_migrations"}
	for _, table := range requiredTables {
		var exists bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, table).Scan(&exists); err != nil {
			t.Fatalf("failed to verify table %s exists: %v", table, err)
		}
		if !exists {
			t.Fatalf("expected migration-created table %s to exist", table)
		}
	}

	var migrationCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&migrationCount); err != nil {
		t.Fatalf("failed to count schema_migrations entries: %v", err)
	}
	if migrationCount < 18 {
		t.Fatalf("expected at least 18 applied migrations, got %d", migrationCount)
	}
}
