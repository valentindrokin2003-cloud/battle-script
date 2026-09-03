// Command api runs the Battle Script Phase 0 HTTP API. Stateless: no
// database, no auth — see docs/superpowers/specs/2026-09-03-http-api-design.md.
package main

import (
	"log"
	"os"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/handler"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// LocalHeuristicClassifier is a dev/test stand-in, not an LLM — see
	// its doc comment. No real adapter exists yet in this environment.
	router := handler.NewRouter(service.Phase0Bosses(), service.BasicModerator{}, service.LocalHeuristicClassifier{})

	log.Printf("Battle Script API listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
