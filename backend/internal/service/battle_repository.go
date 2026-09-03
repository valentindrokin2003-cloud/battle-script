package service

import (
	"context"
	"errors"
	"time"
)

// ErrBattleNotFound is returned by BattleRepository.Get when no battle
// with the given id exists.
var ErrBattleNotFound = errors.New("battle not found")

// HeroRosterEntry is the minimal per-hero information kept alongside a
// stored battle, for display without parsing the full BattleLog.
type HeroRosterEntry struct {
	ID        string    `json:"id"`
	HeroClass HeroClass `json:"hero_class"`
}

// BattleRecord is a persisted BattleSession.
type BattleRecord struct {
	ID         string
	BossID     string
	HeroRoster []HeroRosterEntry
	Log        BattleLog
	CreatedAt  time.Time
}

// BattleRepository is the port persistence lives behind — domain and
// HTTP handler code depend on this interface, never on a specific
// database. See ADR-003 for the same pattern applied to IntentClassifier.
type BattleRepository interface {
	// Save persists record and returns its generated id.
	Save(ctx context.Context, record BattleRecord) (id string, err error)
	// Get returns ErrBattleNotFound if id doesn't exist.
	Get(ctx context.Context, id string) (BattleRecord, error)
}
