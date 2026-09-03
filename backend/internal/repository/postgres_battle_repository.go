// Package repository holds Postgres implementations of the ports
// defined in internal/service. No business logic here — only mapping
// Go values to rows and back.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/service"
)

// PostgresBattleRepository implements service.BattleRepository.
type PostgresBattleRepository struct {
	db *sql.DB
}

func NewPostgresBattleRepository(db *sql.DB) *PostgresBattleRepository {
	return &PostgresBattleRepository{db: db}
}

func (r *PostgresBattleRepository) Save(ctx context.Context, record service.BattleRecord) (string, error) {
	rosterJSON, err := json.Marshal(record.HeroRoster)
	if err != nil {
		return "", fmt.Errorf("marshal hero_roster: %w", err)
	}
	logJSON, err := json.Marshal(record.Log)
	if err != nil {
		return "", fmt.Errorf("marshal battle_log: %w", err)
	}

	var id string
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO battle_sessions (boss_id, outcome, hero_roster, battle_log)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		record.BossID, string(record.Log.Result.Outcome), rosterJSON, logJSON,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert battle_session: %w", err)
	}
	return id, nil
}

func (r *PostgresBattleRepository) Get(ctx context.Context, id string) (service.BattleRecord, error) {
	var (
		record     service.BattleRecord
		rosterJSON []byte
		logJSON    []byte
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT id, boss_id, hero_roster, battle_log, created_at
		FROM battle_sessions
		WHERE id = $1`, id,
	).Scan(&record.ID, &record.BossID, &rosterJSON, &logJSON, &record.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return service.BattleRecord{}, service.ErrBattleNotFound
	}
	if err != nil {
		return service.BattleRecord{}, fmt.Errorf("select battle_session: %w", err)
	}

	if err := json.Unmarshal(rosterJSON, &record.HeroRoster); err != nil {
		return service.BattleRecord{}, fmt.Errorf("unmarshal hero_roster: %w", err)
	}
	if err := json.Unmarshal(logJSON, &record.Log); err != nil {
		return service.BattleRecord{}, fmt.Errorf("unmarshal battle_log: %w", err)
	}
	return record, nil
}
