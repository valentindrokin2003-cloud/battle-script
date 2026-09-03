package repository_test

import (
	"context"
	"testing"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/repository"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/testutil"
)

func sampleBattleRecord() service.BattleRecord {
	return service.BattleRecord{
		BossID: "frost_warden",
		HeroRoster: []service.HeroRosterEntry{
			{ID: "mage-1", HeroClass: service.HeroClassMage},
		},
		Log: service.BattleLog{
			BossID: "frost_warden",
			Turns: []service.BattleTurn{
				{TurnNumber: 1, Events: []service.BattleEvent{
					{Actor: "hero:mage-1", ActionType: "basic_attack", Target: "boss", Amount: 9, TargetHPAfter: 991},
				}},
			},
			Result: service.BattleResult{Outcome: service.OutcomeVictory, TurnsTaken: 1, BossID: "frost_warden"},
		},
	}
}

func TestPostgresBattleRepository_SaveThenGet(t *testing.T) {
	db := testutil.OpenTestDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewPostgresBattleRepository(db)
	record := sampleBattleRecord()

	id, err := repo.Save(context.Background(), record)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if id == "" {
		t.Fatal("Save returned empty id")
	}

	got, err := repo.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != id {
		t.Errorf("ID = %q, want %q", got.ID, id)
	}
	if got.BossID != record.BossID {
		t.Errorf("BossID = %q, want %q", got.BossID, record.BossID)
	}
	if len(got.HeroRoster) != 1 || got.HeroRoster[0].ID != "mage-1" {
		t.Errorf("HeroRoster = %+v, want one entry mage-1", got.HeroRoster)
	}
	if got.Log.Result.Outcome != service.OutcomeVictory {
		t.Errorf("Log.Result.Outcome = %v, want victory", got.Log.Result.Outcome)
	}
	if len(got.Log.Turns) != 1 || len(got.Log.Turns[0].Events) != 1 {
		t.Errorf("Log.Turns round-trip mismatch: %+v", got.Log.Turns)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt was not set")
	}
}

func TestPostgresBattleRepository_GetNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	defer func() { _ = db.Close() }()

	repo := repository.NewPostgresBattleRepository(db)

	_, err := repo.Get(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != service.ErrBattleNotFound {
		t.Errorf("Get error = %v, want ErrBattleNotFound", err)
	}
}
