package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

// NewRouter wires every handler into one Gin engine. No /readyz yet —
// there are no external dependencies to check until persistence exists,
// see the HTTP API spec's "Не-цели".
func NewRouter(catalog map[string]service.Boss, moderator service.Moderator, classifier service.IntentClassifier) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	bosses := BossHandler{Catalog: catalog}
	r.GET("/api/v1/bosses", bosses.List)
	r.GET("/api/v1/bosses/:boss_id", bosses.Get)

	tactics := TacticsHandler{Moderator: moderator, Classifier: classifier, Catalog: catalog}
	r.POST("/api/v1/tactics/classify", tactics.Classify)

	battles := BattlesHandler{Catalog: catalog}
	r.POST("/api/v1/battles", battles.Run)

	return r
}
