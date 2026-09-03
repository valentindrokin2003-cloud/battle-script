// Command api runs the Battle Script Phase 0 HTTP API. No auth yet —
// see docs/superpowers/specs/2026-09-03-http-api-design.md. Persists
// battles to Postgres; run cmd/migrate first, this command does not
// apply migrations itself (see the persistence spec's open questions).
package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/handler"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/repository"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := repository.NewPostgresBattleRepository(db)

	// LocalHeuristicClassifier is a dev/test stand-in, not an LLM — see
	// its doc comment. No real adapter exists yet in this environment.
	router := handler.NewRouter(service.Phase0Bosses(), service.BasicModerator{}, service.LocalHeuristicClassifier{}, repo, db)

	log.Printf("Battle Script API listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
