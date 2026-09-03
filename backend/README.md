# Backend

Go module for Battle Script's Phase 0 backend. Implements the design from:

- [`../docs/superpowers/specs/2026-09-03-battle-script-hld-design.md`](../docs/superpowers/specs/2026-09-03-battle-script-hld-design.md)
- [`../docs/superpowers/specs/2026-09-03-hero-class-action-schema-design.md`](../docs/superpowers/specs/2026-09-03-hero-class-action-schema-design.md)
- [`../docs/superpowers/specs/2026-09-03-boss-script-design.md`](../docs/superpowers/specs/2026-09-03-boss-script-design.md)

## What's implemented

`internal/service` — pure domain logic, no HTTP/SQL/LLM SDK dependency:

- Closed tactic vocabulary (`vocabulary.go`): hero classes, resources, target selectors, condition types, per-class ability sets, class-level default fallback actions.
- `IntentClassification` type and server-side validation against that vocabulary (`intent.go`) — the LLM provider is never trusted blindly; every field is re-checked here.
- `Boss`/`BossPhase` content model, phase-ordering validation, and HP-fraction-to-phase resolution (`boss.go`).
- The deterministic rule engine (`rule_engine.go`): `SelectAction` evaluates a hero's `TacticProgram` against live battle state and returns one action, falling back to the class default when nothing matches or a target selector can't resolve.

Every piece above has unit tests, including the three boss-teaching scenarios from the boss script spec and a determinism regression test.

## What's not implemented yet

- HTTP layer (`cmd/api`, `internal/handler`) — no routes exist yet.
- Persistence (`internal/repository`, `db/migrations`) — nothing is durable yet; everything above operates on in-memory Go values.
- `ModerationModule` and the `IntentClassifier` LLM adapter port — the validator exists, but nothing calls an LLM yet.
- Turn-by-turn `BattleLog` accumulation and damage/HP resolution — `SelectAction` picks *which* action fires; applying its numeric effect to battle state is separate, unbuilt work.
- `ClassroomCohort` / `PlayerSession` auth.

## Running checks

```bash
make check   # fmt-check, vet, golangci-lint, go test -race
```

Requires Go 1.27+ and `golangci-lint` (both installed via Homebrew in this environment).
