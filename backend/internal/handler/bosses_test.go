package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

func TestBossHandler_List(t *testing.T) {
	h := BossHandler{Catalog: service.Phase0Bosses()}
	r := newTestRouter()
	r.GET("/api/v1/bosses", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bosses", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	var bosses []BossResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &bosses); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(bosses) != 3 {
		t.Fatalf("got %d bosses, want 3", len(bosses))
	}
	for _, b := range bosses {
		if len(b.Phases) == 0 {
			t.Errorf("boss %q has no phases in response", b.BossID)
		}
		for _, p := range b.Phases {
			if len(p.AbilityPattern) == 0 {
				t.Errorf("boss %q phase %q has empty ability_pattern in response (fairness rule requires this to be visible)", b.BossID, p.PhaseID)
			}
		}
	}
}

func TestBossHandler_Get_Found(t *testing.T) {
	h := BossHandler{Catalog: service.Phase0Bosses()}
	r := newTestRouter()
	r.GET("/api/v1/bosses/:boss_id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bosses/frost_warden", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	var boss BossResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &boss); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if boss.BossID != "frost_warden" {
		t.Errorf("boss_id = %q, want frost_warden", boss.BossID)
	}
}

func TestBossHandler_Get_NotFound(t *testing.T) {
	h := BossHandler{Catalog: service.Phase0Bosses()}
	r := newTestRouter()
	r.GET("/api/v1/bosses/:boss_id", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bosses/does_not_exist", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body: %s", rec.Code, rec.Body.String())
	}
	assertErrorBody(t, rec.Body.Bytes(), "unknown_boss")
}
