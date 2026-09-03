CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE battle_sessions (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    boss_id     text NOT NULL,
    outcome     text NOT NULL,
    hero_roster jsonb NOT NULL,
    battle_log  jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);
