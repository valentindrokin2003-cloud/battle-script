# Web

React + TypeScript + Vite SPA for Battle Script's Phase 0. Implements:

- [`../docs/superpowers/specs/2026-09-03-web-client-design.md`](../docs/superpowers/specs/2026-09-03-web-client-design.md)

## What's implemented

- `src/api/` — typed client (`client.ts`) and wire types (`types.ts`) matching the backend's JSON contracts by hand (no codegen yet — see the spec's open questions).
- `src/screens/` — `BossSelect`, `TacticInput`, `IntentReview`, `BattleResultScreen`, each with loading/error/success states.
- `src/App.tsx` — the screen state machine: `boss-select → tactic-input → intent-review → battle-result`, with back-navigation to the previous two screens and a "play again" loop from the result screen.
- `src/roster.ts` — the fixed Phase 0 roster (tank/archer/mage/healer), no hero selection.
- Dev-only Vite proxy (`vite.config.ts`): `/api/*` → `http://localhost:8080`, avoids CORS without touching the backend.

23 component tests (Vitest + React Testing Library) cover every screen's states and the full happy-path flow end to end against a mocked API client. `npm run build` is the only guard against the hand-maintained types drifting from the Go backend.

## What's not verified yet

No browser tool was available in the environment this was built in — the network path (Vite proxy → real backend) was verified with `curl`, but no interactive click-through of the actual UI has happened. See [`docs/superpowers/verification/2026-09-03-web-client.md`](../docs/superpowers/verification/2026-09-03-web-client.md). Do a manual pass in a real browser before showing this to anyone.

## Running

```bash
npm install
npm run dev     # http://localhost:5173, proxies /api to :8080
npm test        # vitest
npm run build   # tsc + vite build
```

Requires the backend running separately (`cd ../backend && DATABASE_URL=... go run ./cmd/api`) for the dev proxy to have something to talk to.
