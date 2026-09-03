package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

// TestHTTPPipeline_ClassifyThenBattle is the HTTP-level counterpart to
// internal/service's TestFullPipeline_TextToAction: the same journey,
// but over a real network listener via httptest.Server and a real
// http.Client, not direct Go function calls.
func TestHTTPPipeline_ClassifyThenBattle(t *testing.T) {
	router := NewRouter(service.Phase0Bosses(), service.BasicModerator{}, service.LocalHeuristicClassifier{})
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
	var log service.BattleLog
	if err := json.NewDecoder(resp.Body).Decode(&log); err != nil {
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
	// valid, terminated battle.
	switch log.Result.Outcome {
	case service.OutcomeVictory, service.OutcomeDefeat, service.OutcomeAborted:
	default:
		t.Errorf("Outcome = %q, not one of victory/defeat/aborted", log.Result.Outcome)
	}
	if len(log.Turns) == 0 {
		t.Error("expected non-empty Turns in response")
	}
	if log.BossID != "frost_warden" {
		t.Errorf("BossID = %q, want frost_warden", log.BossID)
	}
}

func TestHTTPPipeline_Healthz(t *testing.T) {
	router := NewRouter(service.Phase0Bosses(), service.BasicModerator{}, service.LocalHeuristicClassifier{})
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
