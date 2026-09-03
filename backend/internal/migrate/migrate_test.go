package migrate_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/valentindrokin2003-cloud/battle-script/backend/db/migrations"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/migrate"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/testutil"
)

func mustExec(t *testing.T, db *sql.DB, query string) {
	t.Helper()
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func TestRun_AppliesThenIsIdempotent(t *testing.T) {
	db := testutil.OpenTestDB(t)
	defer func() { _ = db.Close() }()

	// Deterministic starting point regardless of what earlier test runs
	// left behind in the shared test database.
	mustExec(t, db, `DROP TABLE IF EXISTS battle_sessions`)
	mustExec(t, db, `DROP TABLE IF EXISTS schema_migrations`)

	applied, err := migrate.Run(context.Background(), db, migrations.FS)
	if err != nil {
		t.Fatalf("first Run: %v", err)
	}
	want := []string{"0001_create_battle_sessions.sql"}
	if len(applied) != len(want) || applied[0] != want[0] {
		t.Fatalf("applied = %v, want %v", applied, want)
	}

	var tableExists bool
	err = db.QueryRow(`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'battle_sessions')`).Scan(&tableExists)
	if err != nil {
		t.Fatalf("check battle_sessions exists: %v", err)
	}
	if !tableExists {
		t.Fatal("battle_sessions table was not created")
	}

	appliedAgain, err := migrate.Run(context.Background(), db, migrations.FS)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if len(appliedAgain) != 0 {
		t.Fatalf("second Run applied = %v, want none (idempotency)", appliedAgain)
	}
}
