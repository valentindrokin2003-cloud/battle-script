// Package migrate applies SQL migration files against a Postgres
// database. It is a deliberately minimal, hand-rolled runner — one
// table doesn't justify pulling in a full migration framework
// dependency, per the persistence spec's "Не-цели".
package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

const ensureSchemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version    int PRIMARY KEY,
	applied_at timestamptz NOT NULL DEFAULT now()
);`

// Run applies every "NNNN_*.sql" file in migrations, in filename order,
// that isn't yet recorded in schema_migrations, each inside its own
// transaction. It returns the filenames actually applied this call — an
// empty slice on a call against an already-migrated database is the
// idempotency guarantee the persistence spec requires.
func Run(ctx context.Context, db *sql.DB, migrations fs.FS) ([]string, error) {
	if _, err := db.ExecContext(ctx, ensureSchemaMigrationsTable); err != nil {
		return nil, fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	files, err := sortedMigrationFiles(migrations)
	if err != nil {
		return nil, err
	}

	var applied []string
	for _, name := range files {
		version, err := versionFromFilename(name)
		if err != nil {
			return applied, fmt.Errorf("migration %s: %w", name, err)
		}

		var alreadyApplied bool
		err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&alreadyApplied)
		if err != nil {
			return applied, fmt.Errorf("check migration %s: %w", name, err)
		}
		if alreadyApplied {
			continue
		}

		if err := applyMigration(ctx, db, migrations, name, version); err != nil {
			return applied, err
		}
		applied = append(applied, name)
	}
	return applied, nil
}

func sortedMigrationFiles(migrations fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(migrations, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func applyMigration(ctx context.Context, db *sql.DB, migrations fs.FS, name string, version int) error {
	sqlBytes, err := fs.ReadFile(migrations, name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for %s: %w", name, err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}
	return nil
}

// versionFromFilename extracts the leading integer from "NNNN_name.sql".
func versionFromFilename(name string) (int, error) {
	prefix, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0, fmt.Errorf("filename %q must start with NNNN_", name)
	}
	v, err := strconv.Atoi(prefix)
	if err != nil {
		return 0, fmt.Errorf("filename %q prefix is not numeric: %w", name, err)
	}
	return v, nil
}
