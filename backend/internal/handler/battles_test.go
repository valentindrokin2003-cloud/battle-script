package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

func newBattlesHandler() BattlesHandler {
	return BattlesHandler{Catalog: service.Phase0Bosses()}
}

func doBattle(t *testing.T, r *gin.Engine, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/battles", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func alwaysBasicAttackIntent(class service.HeroClass) service.IntentClassification {
	return service.IntentClassification{
		HeroClass:     class,
		SchemaVersion: service.CurrentSchemaVersion,
		Rules: []service.Rule{
			{Priority: 0, Condition: service.Condition{Type: service.ConditionAlways}, Action: service.DefaultFallback(class)},
		},
		FallbackAction: service.DefaultFallback(class),
		Confidence:     service.ConfidenceHigh,
	}
}

func fullTeamRequest() BattleRequest {
	return BattleRequest{
		BossID: "frost_warden",
		Heroes: []BattleHeroInput{
			{ID: "tank-1", HeroClass: "tank", Intent: alwaysBasicAttackIntent(service.HeroClassTank)},
			{ID: "archer-1", HeroClass: "archer", Intent: alwaysBasicAttackIntent(service.HeroClassArcher)},
			{ID: "mage-1", HeroClass: "mage", Intent: alwaysBasicAttackIntent(service.HeroClassMage)},
			{ID: "healer-1", HeroClass: "healer", Intent: alwaysBasicAttackIntent(service.HeroClassHealer)},
		},
	}
}

func TestBattlesHandler_Run_Victory(t *testing.T) {
	h := newBattlesHandler()
	r := newTestRouter()
	r.POST("/api/v1/battles", h.Run)

	rec := doBattle(t, r, fullTeamRequest())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	var log service.BattleLog
	if err := json.Unmarshal(rec.Body.Bytes(), &log); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if log.Result.Outcome != service.OutcomeVictory {
		t.Errorf("Outcome = %v, want victory", log.Result.Outcome)
	}
	if len(log.Turns) == 0 {
		t.Error("expected non-empty Turns in response")
	}
}

func TestBattlesHandler_Run_InvalidIntent(t *testing.T) {
	h := newBattlesHandler()
	r := newTestRouter()
	r.POST("/api/v1/battles", h.Run)

	req := fullTeamRequest()
	req.Heroes[0].Intent.Rules = []service.Rule{
		{Priority: 0, Condition: service.Condition{Type: service.ConditionAlways}, Action: service.Action{Type: service.ActionHeal, Target: service.TargetSelf}}, // heal is not a tank ability
	}

	rec := doBattle(t, r, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body: %s", rec.Code, rec.Body.String())
	}
	assertErrorBody(t, rec.Body.Bytes(), "invalid_intent")
}

func TestBattlesHandler_Run_UnknownBoss(t *testing.T) {
	h := newBattlesHandler()
	r := newTestRouter()
	r.POST("/api/v1/battles", h.Run)

	req := fullTeamRequest()
	req.BossID = "does_not_exist"

	rec := doBattle(t, r, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body: %s", rec.Code, rec.Body.String())
	}
	assertErrorBody(t, rec.Body.Bytes(), "unknown_boss")
}
