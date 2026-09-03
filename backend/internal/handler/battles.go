package handler

import (
	"errors"
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

// BattleResponse is the POST /api/v1/battles and GET /api/v1/battles/:id
// response body: the persisted id alongside the same BattleLog shape as
// before persistence existed.
type BattleResponse struct {
	ID string `json:"id"`
	service.BattleLog
}

// BattlesHandler runs a full BattleSession from a client-supplied
// roster, persists it, and returns the resulting BattleLog plus its id.
type BattlesHandler struct {
	Catalog    map[string]service.Boss
	Repository service.BattleRepository
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

	roster := make([]service.HeroRosterEntry, len(heroDefs))
	for i, hd := range heroDefs {
		roster[i] = service.HeroRosterEntry{ID: hd.ID, HeroClass: hd.HeroClass}
	}
	id, err := h.Repository.Save(c.Request.Context(), service.BattleRecord{
		BossID:     boss.BossID,
		HeroRoster: roster,
		Log:        log,
	})
	if err != nil {
		writeError(c, http.StatusBadGateway, "persistence_failed", "battle was resolved but could not be saved: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, BattleResponse{ID: id, BattleLog: log})
}

// Get handles GET /api/v1/battles/:id.
func (h BattlesHandler) Get(c *gin.Context) {
	record, err := h.Repository.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrBattleNotFound) {
		writeError(c, http.StatusNotFound, "battle_not_found", "no such battle id")
		return
	}
	if err != nil {
		writeError(c, http.StatusBadGateway, "persistence_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, BattleResponse{ID: record.ID, BattleLog: record.Log})
}
