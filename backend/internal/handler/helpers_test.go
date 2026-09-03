package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter() *gin.Engine {
	return gin.New()
}

func assertErrorBody(t *testing.T, body []byte, wantCode string) {
	t.Helper()
	var e ErrorResponse
	if err := json.Unmarshal(body, &e); err != nil {
		t.Fatalf("invalid error JSON: %v (body: %s)", err, body)
	}
	if e.Error != wantCode {
		t.Errorf("error code = %q, want %q", e.Error, wantCode)
	}
	if e.Message == "" {
		t.Error("error message is empty")
	}
}

// fakeBattleRepository is an in-memory service.BattleRepository for
// handler unit tests that don't need a real database — real Postgres
// round-tripping is covered separately by
// TestHTTPPipeline_PersistedBattleRoundTrip.
type fakeBattleRepository struct {
	records map[string]service.BattleRecord
	nextID  int
	saveErr error
	getErr  error
}

func newFakeBattleRepository() *fakeBattleRepository {
	return &fakeBattleRepository{records: map[string]service.BattleRecord{}}
}

func (f *fakeBattleRepository) Save(_ context.Context, record service.BattleRecord) (string, error) {
	if f.saveErr != nil {
		return "", f.saveErr
	}
	f.nextID++
	id := fmt.Sprintf("fake-%d", f.nextID)
	record.ID = id
	f.records[id] = record
	return id, nil
}

func (f *fakeBattleRepository) Get(_ context.Context, id string) (service.BattleRecord, error) {
	if f.getErr != nil {
		return service.BattleRecord{}, f.getErr
	}
	record, ok := f.records[id]
	if !ok {
		return service.BattleRecord{}, service.ErrBattleNotFound
	}
	return record, nil
}
