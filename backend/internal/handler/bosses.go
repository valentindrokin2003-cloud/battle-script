package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

// BossPhaseResponse mirrors service.BossPhase on the wire. It
// deliberately includes AbilityPattern — the boss script spec's
// fairness rule requires everything a tactic condition can reference to
// be visible to the player before they write it.
type BossPhaseResponse struct {
	PhaseID          string   `json:"phase_id"`
	HPThresholdEnter float64  `json:"hp_threshold_enter"`
	AbilityPattern   []string `json:"ability_pattern"`
	Provocation      string   `json:"provocation"`
}

// BossResponse mirrors service.Boss on the wire.
type BossResponse struct {
	BossID      string              `json:"boss_id"`
	DisplayName string              `json:"display_name"`
	Phases      []BossPhaseResponse `json:"phases"`
}

func toBossResponse(b service.Boss) BossResponse {
	phases := make([]BossPhaseResponse, len(b.Phases))
	for i, p := range b.Phases {
		phases[i] = BossPhaseResponse{
			PhaseID:          p.PhaseID,
			HPThresholdEnter: p.HPThresholdEnter,
			AbilityPattern:   p.AbilityPattern,
			Provocation:      p.Provocation,
		}
	}
	return BossResponse{BossID: b.BossID, DisplayName: b.DisplayName, Phases: phases}
}

// BossHandler serves the Phase 0 boss content catalog.
type BossHandler struct {
	Catalog map[string]service.Boss
}

func (h BossHandler) List(c *gin.Context) {
	out := make([]BossResponse, 0, len(h.Catalog))
	for _, b := range h.Catalog {
		out = append(out, toBossResponse(b))
	}
	c.JSON(http.StatusOK, out)
}

func (h BossHandler) Get(c *gin.Context) {
	boss, ok := h.Catalog[c.Param("boss_id")]
	if !ok {
		writeError(c, http.StatusNotFound, "unknown_boss", "no such boss_id")
		return
	}
	c.JSON(http.StatusOK, toBossResponse(boss))
}
