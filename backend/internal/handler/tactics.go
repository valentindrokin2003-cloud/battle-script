package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

// ClassifyRequest is the POST /api/v1/tactics/classify body.
type ClassifyRequest struct {
	HeroClass  string `json:"hero_class"`
	BossID     string `json:"boss_id"`
	PromptText string `json:"prompt_text"`
}

// TacticsHandler runs a child's free text through moderation, then
// classification, per the HTTP API spec's ordering requirement:
// Moderator.Check always runs before the classifier is ever called.
type TacticsHandler struct {
	Moderator  service.Moderator
	Classifier service.IntentClassifier
	Catalog    map[string]service.Boss
}

func (h TacticsHandler) Classify(c *gin.Context) {
	var req ClassifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	class := service.HeroClass(req.HeroClass)
	if !class.Valid() {
		writeError(c, http.StatusNotFound, "unknown_hero_class", "no such hero_class")
		return
	}
	boss, ok := h.Catalog[req.BossID]
	if !ok {
		writeError(c, http.StatusNotFound, "unknown_boss", "no such boss_id")
		return
	}

	modResult := h.Moderator.Check(req.PromptText)
	if !modResult.Allowed {
		writeError(c, http.StatusUnprocessableEntity, "moderation_rejected", modResult.Reason)
		return
	}

	result := service.ClassifyWithFallback(c.Request.Context(), h.Classifier, service.ClassificationContext{
		HeroClass:  class,
		Boss:       boss,
		PromptText: req.PromptText,
	})
	c.JSON(http.StatusOK, result)
}
