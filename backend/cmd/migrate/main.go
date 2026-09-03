// Command migrate applies pending SQL migrations to a Postgres
// database. Not run automatically by cmd/api — operator discipline, per
// the persistence spec's open questions.
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/valentindrokin2003-cloud/battle-script/backend/db/migrations"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/migrate"
)

func main() {
	dsn := flag.String("database", os.Getenv("DATABASE_URL"), "Postgres connection string (or set DATABASE_URL)")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("missing -database (and DATABASE_URL is not set)")
	}

	db, err := sql.Open("pgx", *dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	applied, err := migrate.Run(context.Background(), db, migrations.FS)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if len(applied) == 0 {
		log.Println("no pending migrations")
		return
	}
	for _, name := range applied {
		log.Printf("applied %s", name)
	}
}
