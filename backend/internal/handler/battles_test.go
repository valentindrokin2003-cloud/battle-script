package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

func newBattlesHandler(repo service.BattleRepository) BattlesHandler {
	return BattlesHandler{Catalog: service.Phase0Bosses(), Repository: repo}
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
	repo := newFakeBattleRepository()
	h := newBattlesHandler(repo)
	r := newTestRouter()
	r.POST("/api/v1/battles", h.Run)

	rec := doBattle(t, r, fullTeamRequest())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	var resp BattleResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.ID == "" {
		t.Error("expected non-empty id in response")
	}
	if resp.Result.Outcome != service.OutcomeVictory {
		t.Errorf("Outcome = %v, want victory", resp.Result.Outcome)
	}
	if len(resp.Turns) == 0 {
		t.Error("expected non-empty Turns in response")
	}
	if _, err := repo.Get(context.Background(), resp.ID); err != nil {
		t.Errorf("battle was not actually persisted: %v", err)
	}
}

func TestBattlesHandler_Run_PersistenceFailure(t *testing.T) {
	repo := newFakeBattleRepository()
	repo.saveErr = errors.New("connection refused")
	h := newBattlesHandler(repo)
	r := newTestRouter()
	r.POST("/api/v1/battles", h.Run)

	rec := doBattle(t, r, fullTeamRequest())

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502; body: %s", rec.Code, rec.Body.String())
	}
	assertErrorBody(t, rec.Body.Bytes(), "persistence_failed")
}

func TestBattlesHandler_Get_Found(t *testing.T) {
	repo := newFakeBattleRepository()
	h := newBattlesHandler(repo)
	r := newTestRouter()
	r.POST("/api/v1/battles", h.Run)
	r.GET("/api/v1/battles/:id", h.Get)

	postRec := doBattle(t, r, fullTeamRequest())
	var posted BattleResponse
	if err := json.Unmarshal(postRec.Body.Bytes(), &posted); err != nil {
		t.Fatalf("invalid JSON from POST: %v", err)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/battles/"+posted.ID, nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", getRec.Code, getRec.Body.String())
	}
	var got BattleResponse
	if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.ID != posted.ID || got.Result.Outcome != posted.Result.Outcome {
		t.Errorf("GET result %+v does not match POST result %+v", got, posted)
	}
}

func TestBattlesHandler_Get_NotFound(t *testing.T) {
	h := newBattlesHandler(newFakeBattleRepository())
	r := newTestRouter()
	r.GET("/api/v1/battles/:id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/battles/does-not-exist", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body: %s", rec.Code, rec.Body.String())
	}
	assertErrorBody(t, rec.Body.Bytes(), "battle_not_found")
}

func TestBattlesHandler_Run_InvalidIntent(t *testing.T) {
	h := newBattlesHandler(newFakeBattleRepository())
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
	h := newBattlesHandler(newFakeBattleRepository())
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
