package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

// BattleHeroInput is one hero's entry in a POST /api/v1/battles request.
// Intent is expected to have come from POST /api/v1/tactics/classify,
// but this stateless endpoint does not verify that provenance — see the
// spec's open questions. It is re-validated regardless.
type BattleHeroInput struct {
	ID        string                       `json:"id"`
	HeroClass string                       `json:"hero_class"`
	Intent    service.IntentClassification `json:"intent"`
}

// BattleRequest is the POST /api/v1/battles body.
type BattleRequest struct {
	BossID string            `json:"boss_id"`
	Heroes []BattleHeroInput `json:"heroes"`
}

// BattlesHandler runs a full BattleSession from a client-supplied
// roster and returns the resulting BattleLog.
type BattlesHandler struct {
	Catalog map[string]service.Boss
}

func (h BattlesHandler) Run(c *gin.Context) {
	var req BattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	boss, ok := h.Catalog[req.BossID]
	if !ok {
		writeError(c, http.StatusNotFound, "unknown_boss", "no such boss_id")
		return
	}

	heroDefs := make([]service.HeroDef, 0, len(req.Heroes))
	for _, hi := range req.Heroes {
		class := service.HeroClass(hi.HeroClass)
		if !class.Valid() {
			writeError(c, http.StatusNotFound, "unknown_hero_class", "no such hero_class")
			return
		}
		if err := service.ValidateIntentClassification(hi.Intent, boss); err != nil {
			writeError(c, http.StatusUnprocessableEntity, "invalid_intent", err.Error())
			return
		}
		heroDefs = append(heroDefs, service.HeroDef{
			ID:        hi.ID,
			HeroClass: class,
			Resources: service.HeroBaseResources(class),
			Program:   hi.Intent.Rules,
			Fallback:  hi.Intent.FallbackAction,
		})
	}

	log := service.RunBattle(boss, heroDefs, service.DefaultMaxTurns)
	c.JSON(http.StatusOK, log)
}
