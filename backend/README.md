# Backend

Go module for Battle Script's Phase 0 backend. Implements the design from:

- [`../docs/superpowers/specs/2026-09-03-battle-script-hld-design.md`](../docs/superpowers/specs/2026-09-03-battle-script-hld-design.md)
- [`../docs/superpowers/specs/2026-09-03-hero-class-action-schema-design.md`](../docs/superpowers/specs/2026-09-03-hero-class-action-schema-design.md)
- [`../docs/superpowers/specs/2026-09-03-boss-script-design.md`](../docs/superpowers/specs/2026-09-03-boss-script-design.md)
- [`../docs/superpowers/specs/2026-09-03-battle-turn-resolution-design.md`](../docs/superpowers/specs/2026-09-03-battle-turn-resolution-design.md)

## What's implemented

`internal/service` — pure domain logic, no HTTP/SQL/LLM SDK dependency:

- Closed tactic vocabulary (`vocabulary.go`): hero classes, resources, target selectors, condition types, per-class ability sets, class-level default fallback actions.
- `IntentClassification` type and server-side validation against that vocabulary (`intent.go`) — the LLM provider is never trusted blindly; every field is re-checked here.
- `Boss`/`BossPhase` content model, phase-ordering validation, and HP-fraction-to-phase resolution (`boss.go`).
- The deterministic rule engine (`rule_engine.go`): `SelectAction` evaluates a hero's `TacticProgram` against live battle state and returns one action, falling back to the class default when nothing matches or a target selector can't resolve.
- The full turn loop (`battle_engine.go`, `battle_state.go`): `RunBattle` simulates a complete `BattleSession` turn by turn — heroes act via `SelectAction`, damage/heal/shield-absorption resolves, the boss acts via its per-boss targeting rule and phase ability pattern — and returns a `BattleLog` with the final `BattleResult` (`victory`/`defeat`/`aborted`).
- Phase 0 seed content (`phase0_content.go`): base hero resources and the 3 starter bosses (`frost_warden`, `shadow_hunter`, `stone_giant`) with illustrative-but-runnable numbers.

Every piece above has unit tests. `battle_engine_test.go` includes the property test promised by the boss script spec (phase-gated `frost_bolt` never loses to the naive always-cast version), scenario tests for all three bosses' teaching mechanics, a turn-limit abort test, and a full-battle determinism regression test.

## What's not implemented yet

- HTTP layer (`cmd/api`, `internal/handler`) — no routes exist yet.
- Persistence (`internal/repository`, `db/migrations`) — nothing is durable yet; everything above operates on in-memory Go values, one `RunBattle` call at a time.
- `ModerationModule` and the `IntentClassifier` LLM adapter port — the validator exists, but nothing calls an LLM yet.
- `ClassroomCohort` / `PlayerSession` auth.
- Mana costs, cooldowns, and boss status effects (`cleanse` has no debuff content to remove yet) — explicit non-goals of the current iteration, see the turn resolution spec.

## Running checks

```bash
make check   # fmt-check, vet, golangci-lint, go test -race
```

Requires Go 1.27+ and `golangci-lint` (both installed via Homebrew in this environment).
