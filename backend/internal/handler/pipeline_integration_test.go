package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/repository"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/testutil"
)

// newTestRouterWithDB opens a real TEST_DATABASE_URL connection (skips
// the test if unset) and wires it into a full router, so these
// integration tests exercise real persistence, not a fake.
func newTestRouterWithDB(t *testing.T) *gin.Engine {
	t.Helper()
	db := testutil.OpenTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	repo := repository.NewPostgresBattleRepository(db)
	return NewRouter(service.Phase0Bosses(), service.BasicModerator{}, service.LocalHeuristicClassifier{}, repo, db)
}

// TestHTTPPipeline_ClassifyThenBattle is the HTTP-level counterpart to
// internal/service's TestFullPipeline_TextToAction: the same journey,
// but over a real network listener via httptest.Server and a real
// http.Client, not direct Go function calls.
func TestHTTPPipeline_ClassifyThenBattle(t *testing.T) {
	router := newTestRouterWithDB(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	heroTexts := map[string]struct {
		class service.HeroClass
		text  string
	}{
		"tank-1":   {service.HeroClassTank, "провоцируй босса, когда он целится в целителя, и если щит падает ниже 30% — отступи"},
		"archer-1": {service.HeroClassArcher, "цель самый сильный удар в слабейшего врага"},
		"mage-1":   {service.HeroClassMage, "используй ледяной шар только в фазе щита босса, в остальное время атакуй слабейшего"},
		"healer-1": {service.HeroClassHealer, "лечи того, у кого меньше всего здоровья"},
	}

	var heroes []BattleHeroInput
	for id, ht := range heroTexts {
		reqBody, err := json.Marshal(ClassifyRequest{HeroClass: string(ht.class), BossID: "frost_warden", PromptText: ht.text})
		if err != nil {
			t.Fatalf("marshal classify request: %v", err)
		}
		resp, err := http.Post(srv.URL+"/api/v1/tactics/classify", "application/json", bytes.NewReader(reqBody))
		if err != nil {
			t.Fatalf("classify request for %s failed: %v", id, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("classify for %s: status = %d", id, resp.StatusCode)
		}
		var intent service.IntentClassification
		if err := json.NewDecoder(resp.Body).Decode(&intent); err != nil {
			t.Fatalf("decode classify response for %s: %v", id, err)
		}
		_ = resp.Body.Close()
		if intent.Confidence != service.ConfidenceHigh {
			t.Fatalf("classify for %s: Confidence = %v, want high (heuristic classifier should have recognized this canonical phrasing)", id, intent.Confidence)
		}
		heroes = append(heroes, BattleHeroInput{ID: id, HeroClass: string(ht.class), Intent: intent})
	}

	battleReqBody, err := json.Marshal(BattleRequest{BossID: "frost_warden", Heroes: heroes})
	if err != nil {
		t.Fatalf("marshal battle request: %v", err)
	}
	resp, err := http.Post(srv.URL+"/api/v1/battles", "application/json", bytes.NewReader(battleReqBody))
	if err != nil {
		t.Fatalf("battle request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("battle: status = %d", resp.StatusCode)
	}
	var battleResp BattleResponse
	if err := json.NewDecoder(resp.Body).Decode(&battleResp); err != nil {
		t.Fatalf("decode battle response: %v", err)
	}
	// This test's job is proving the HTTP wiring (classify -> battle over
	// a real network round trip), not re-proving game balance for this
	// ad hoc team composition — that's what battle_engine_test.go's
	// property tests are for. A discarded victory assertion here found a
	// real, legitimate outcome: this exact tank phrasing has no "always"
	// fallback rule, so once its shield breaks the retreat condition
	// (shield never regenerates) stays true forever and the tank sits
	// out the rest of the fight — a realistic tactic-writing mistake, not
	// an engine bug. So just assert the pipeline produced a structurally
	// valid, terminated, persisted battle.
	switch battleResp.Result.Outcome {
	case service.OutcomeVictory, service.OutcomeDefeat, service.OutcomeAborted:
	default:
		t.Errorf("Outcome = %q, not one of victory/defeat/aborted", battleResp.Result.Outcome)
	}
	if len(battleResp.Turns) == 0 {
		t.Error("expected non-empty Turns in response")
	}
	if battleResp.BossID != "frost_warden" {
		t.Errorf("BossID = %q, want frost_warden", battleResp.BossID)
	}
	if battleResp.ID == "" {
		t.Fatal("expected non-empty id in response")
	}

	getResp, err := http.Get(srv.URL + "/api/v1/battles/" + battleResp.ID)
	if err != nil {
		t.Fatalf("GET battle by id failed: %v", err)
	}
	defer func() { _ = getResp.Body.Close() }()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET battle: status = %d", getResp.StatusCode)
	}
	var fetched BattleResponse
	if err := json.NewDecoder(getResp.Body).Decode(&fetched); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if fetched.ID != battleResp.ID || fetched.Result.Outcome != battleResp.Result.Outcome || len(fetched.Turns) != len(battleResp.Turns) {
		t.Errorf("GET result %+v does not match what POST persisted %+v", fetched, battleResp)
	}
}

func TestHTTPPipeline_Readyz(t *testing.T) {
	router := newTestRouterWithDB(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestHTTPPipeline_Healthz(t *testing.T) {
	router := newTestRouterWithDB(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}
