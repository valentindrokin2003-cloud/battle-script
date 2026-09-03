package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

// NewRouter wires every handler into one Gin engine.
func NewRouter(catalog map[string]service.Boss, moderator service.Moderator, classifier service.IntentClassifier, repo service.BattleRepository, db *sql.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", ReadyzHandler{DB: db}.Check)

	bosses := BossHandler{Catalog: catalog}
	r.GET("/api/v1/bosses", bosses.List)
	r.GET("/api/v1/bosses/:boss_id", bosses.Get)

	tactics := TacticsHandler{Moderator: moderator, Classifier: classifier, Catalog: catalog}
	r.POST("/api/v1/tactics/classify", tactics.Classify)

	battles := BattlesHandler{Catalog: catalog, Repository: repo}
	r.POST("/api/v1/battles", battles.Run)
	r.GET("/api/v1/battles/:id", battles.Get)

	return r
}
