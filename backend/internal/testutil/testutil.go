// Package testutil holds small helpers shared by tests that need a real
// Postgres connection (migrate, repository, handler packages). Per the
// persistence spec, these tests run for real against TEST_DATABASE_URL
// rather than being mocked out, and skip cleanly when it isn't set.
package testutil

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenTestDB opens a connection to TEST_DATABASE_URL, skipping the test
// if the variable isn't set. The caller closes the returned *sql.DB.
func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping test that requires a real Postgres connection")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Fatalf("ping test database: %v", err)
	}
	return db
}
