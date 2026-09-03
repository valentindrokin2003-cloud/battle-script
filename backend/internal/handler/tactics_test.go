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

func newTacticsHandler() TacticsHandler {
	return TacticsHandler{
		Moderator:  service.BasicModerator{},
		Classifier: service.LocalHeuristicClassifier{},
		Catalog:    service.Phase0Bosses(),
	}
}

func doClassify(t *testing.T, r *gin.Engine, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tactics/classify", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestTacticsHandler_Classify_ValidText(t *testing.T) {
	h := newTacticsHandler()
	r := newTestRouter()
	r.POST("/api/v1/tactics/classify", h.Classify)

	rec := doClassify(t, r, ClassifyRequest{
		HeroClass:  "mage",
		BossID:     "frost_warden",
		PromptText: "используй ледяной шар только в фазе щита босса, в остальное время атакуй слабейшего",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	var got service.IntentClassification
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Confidence != service.ConfidenceHigh {
		t.Errorf("Confidence = %v, want high", got.Confidence)
	}
	if err := service.ValidateIntentClassification(got, service.FrostWardenBoss()); err != nil {
		t.Errorf("response failed validation: %v", err)
	}
}

func TestTacticsHandler_Classify_ModerationRejected(t *testing.T) {
	h := newTacticsHandler()
	r := newTestRouter()
	r.POST("/api/v1/tactics/classify", h.Classify)

	rec := doClassify(t, r, ClassifyRequest{
		HeroClass:  "tank",
		BossID:     "frost_warden",
		PromptText: "", // rejected by BasicModerator: empty text
	})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body: %s", rec.Code, rec.Body.String())
	}
	assertErrorBody(t, rec.Body.Bytes(), "moderation_rejected")
}

func TestTacticsHandler_Classify_UnknownBoss(t *testing.T) {
	h := newTacticsHandler()
	r := newTestRouter()
	r.POST("/api/v1/tactics/classify", h.Classify)

	rec := doClassify(t, r, ClassifyRequest{HeroClass: "tank", BossID: "does_not_exist", PromptText: "atakuj"})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body: %s", rec.Code, rec.Body.String())
	}
	assertErrorBody(t, rec.Body.Bytes(), "unknown_boss")
}

func TestTacticsHandler_Classify_UnknownHeroClass(t *testing.T) {
	h := newTacticsHandler()
	r := newTestRouter()
	r.POST("/api/v1/tactics/classify", h.Classify)

	rec := doClassify(t, r, ClassifyRequest{HeroClass: "paladin", BossID: "frost_warden", PromptText: "atakuj"})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body: %s", rec.Code, rec.Body.String())
	}
	assertErrorBody(t, rec.Body.Bytes(), "unknown_hero_class")
}
