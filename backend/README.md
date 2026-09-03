# Backend

Go module for Battle Script's Phase 0 backend. Implements the design from:

- [`../docs/superpowers/specs/2026-09-03-battle-script-hld-design.md`](../docs/superpowers/specs/2026-09-03-battle-script-hld-design.md)
- [`../docs/superpowers/specs/2026-09-03-hero-class-action-schema-design.md`](../docs/superpowers/specs/2026-09-03-hero-class-action-schema-design.md)
- [`../docs/superpowers/specs/2026-09-03-boss-script-design.md`](../docs/superpowers/specs/2026-09-03-boss-script-design.md)
- [`../docs/superpowers/specs/2026-09-03-battle-turn-resolution-design.md`](../docs/superpowers/specs/2026-09-03-battle-turn-resolution-design.md)
- [`../docs/superpowers/specs/2026-09-03-moderation-and-classifier-port-design.md`](../docs/superpowers/specs/2026-09-03-moderation-and-classifier-port-design.md)
- [`../docs/superpowers/specs/2026-09-03-http-api-design.md`](../docs/superpowers/specs/2026-09-03-http-api-design.md)
- [`../docs/superpowers/specs/2026-09-03-battle-persistence-design.md`](../docs/superpowers/specs/2026-09-03-battle-persistence-design.md)

## What's implemented

`internal/service` — pure domain logic, no HTTP/SQL/LLM SDK dependency:

- Closed tactic vocabulary (`vocabulary.go`): hero classes, resources, target selectors, condition types, per-class ability sets, class-level default fallback actions.
- `IntentClassification` type and server-side validation against that vocabulary (`intent.go`) — the LLM provider is never trusted blindly; every field is re-checked here.
- `Boss`/`BossPhase` content model, phase-ordering validation, and HP-fraction-to-phase resolution (`boss.go`).
- The deterministic rule engine (`rule_engine.go`): `SelectAction` evaluates a hero's `TacticProgram` against live battle state and returns one action, falling back to the class default when nothing matches or a target selector can't resolve.
- The full turn loop (`battle_engine.go`, `battle_state.go`): `RunBattle` simulates a complete `BattleSession` turn by turn — heroes act via `SelectAction`, damage/heal/shield-absorption resolves, the boss acts via its per-boss targeting rule and phase ability pattern — and returns a `BattleLog` with the final `BattleResult` (`victory`/`defeat`/`aborted`).
- Phase 0 seed content (`phase0_content.go`): base hero resources and the 3 starter bosses (`frost_warden`, `shadow_hunter`, `stone_giant`) with illustrative-but-runnable numbers.
- `ModerationModule` (`moderation.go`): `BasicModerator` — length cap, profanity list, email/phone-like pattern rejection. Minimal, not reviewed by a content/legal professional yet — see the spec's open questions before this goes anywhere near a real pilot.
- The `IntentClassifier` port and its retry/fallback orchestration (`classifier.go`): `ClassifyWithFallback` implements the HLD's error handling (one retry, then the class default with `low_fallback_used`) independent of which adapter sits behind the port.
- `LocalHeuristicClassifier` (`local_heuristic_classifier.go`): a keyword-matching **dev/test stand-in**, not an LLM. No real LLM adapter exists — this environment has no provider API key. Swapping in a real adapter later requires no domain/HTTP changes, only a new `IntentClassifier` implementation.

Every piece above has unit tests. `battle_engine_test.go` includes the property test promised by the boss script spec (phase-gated `frost_bolt` never loses to the naive always-cast version), scenario tests for all three bosses' teaching mechanics, a turn-limit abort test, and a full-battle determinism regression test. `local_heuristic_classifier_test.go` includes the project's first end-to-end pipeline test: raw text → moderation → classification → validation → `SelectAction`.

`internal/handler` — thin Gin delivery layer:

- `GET /healthz` (pure liveness), `GET /readyz` (pings Postgres), `GET /api/v1/bosses`, `GET /api/v1/bosses/:boss_id`, `POST /api/v1/tactics/classify`, `POST /api/v1/battles`, `GET /api/v1/battles/:id` — see the HTTP API and persistence specs for the full contract.
- `cmd/api/main.go` wires `BasicModerator` + `LocalHeuristicClassifier` (the dev stand-in, still no real LLM) + `PostgresBattleRepository` into the router and runs on `$PORT` (default 8080). Requires `DATABASE_URL`.
- `TestHTTPPipeline_ClassifyThenBattle` exercises the full path over a real `httptest.Server` and `http.Client`, including the persisted round trip (`POST /battles` → `id` → `GET /battles/:id`), against a real Postgres.

`internal/migrate`, `internal/repository`, `cmd/migrate`, `db/migrations` — persistence:

- `db/migrations/0001_create_battle_sessions.sql` — one table, `battle_sessions`, storing each `BattleLog` as JSONB with `boss_id`/`outcome` denormalized for future filtering.
- `internal/migrate`: a minimal hand-rolled runner (no framework dependency for one table) — applies pending `NNNN_*.sql` files in order, tracked in `schema_migrations`, idempotent. `cmd/migrate` is the CLI: `go run ./cmd/migrate` (reads `-database` or `$DATABASE_URL`).
- `internal/repository.PostgresBattleRepository` implements the `service.BattleRepository` port (`Save`/`Get`) via `database/sql` + `pgx/v5/stdlib` — no ORM.
- Still anonymous: battles aren't tied to a player/session — that's `PlayerSession`/`ClassroomCohort` auth, explicitly separate future work.

Tests that need a real database (`internal/migrate`, `internal/repository`, several `internal/handler` tests) run for real against `TEST_DATABASE_URL`, skipping cleanly if it isn't set — not mocked out. PostgreSQL 16 is installed via Homebrew and running locally in this environment (no Docker here).

## What's not implemented yet

- A real LLM adapter behind `IntentClassifier` — blocked on provider API access, not on architecture.
- `ClassroomCohort` / `PlayerSession` auth — no session concept exists yet; every endpoint is open, and persisted battles are anonymous rows.
- Automatic migration-applied checks at `cmd/api` startup — relies on running `cmd/migrate` first, by convention, same as Sectr.
- Mana costs, cooldowns, and boss status effects (`cleanse` has no debuff content to remove yet, and can't target one specific status yet either — see the moderation/classifier spec's open questions) — explicit non-goals of the current iteration.

## Running checks

```bash
make check   # fmt-check, vet, golangci-lint, go test -race
```

Requires Go 1.27+ and `golangci-lint` (both installed via Homebrew). Tests touching persistence additionally need `TEST_DATABASE_URL` (e.g. `postgres://localhost:5432/battle_script_test?sslmode=disable`) pointing at a migrated database; without it they skip rather than fail.
