# Архитектура Battle Script — текущее состояние

Живой документ: описывает, как система устроена **сейчас**, а не как решили в конкретную дату. За обоснованием решений и полным контекстом — см. датированные спеки в [`../superpowers/specs/`](../superpowers/specs/); при расхождении между этим файлом и спекой доверять текущему коду и этому файлу, спеки могут отставать.

## Итог в двух словах

Battle Script — веб-автобаттлер, где ребёнок пишет тактику текстом, LLM переводит её в закрытый набор действий, а бой разыгрывает детерминированный движок. Реализовано и протестировано: доменное ядро боевой системы, HTTP API поверх него и персистентность боёв в PostgreSQL. Не реализовано: сам вызов настоящего LLM, фронтенд, авторизация/`PlayerSession`.

## Компоненты и их статус

| Компонент | Статус | Где в коде |
|---|---|---|
| Закрытый словарь тактик (классы, ресурсы, селекторы, условия, способности) | ✅ реализован | `backend/internal/service/vocabulary.go` |
| Валидация `IntentClassification` (LLM — не граница доверия) | ✅ реализован | `backend/internal/service/intent.go` |
| Модель `Boss`/`BossPhase`, резолюция фазы по HP | ✅ реализован | `backend/internal/service/boss.go` |
| Детерминированный выбор действия героя (`SelectAction`) | ✅ реализован | `backend/internal/service/rule_engine.go` |
| Полный цикл боя (урон/лечение/щит/таргетинг босса/`BattleLog`) | ✅ реализован | `backend/internal/service/battle_engine.go`, `battle_state.go` |
| Сид-контент Фазы 0 (3 класса героев, 3 босса, иллюстративные числа) | ✅ реализован | `backend/internal/service/phase0_content.go` |
| `ModerationModule` | ✅ реализован (Phase-0 минимум, не прошёл юридическое ревью) | `backend/internal/service/moderation.go` |
| Порт `IntentClassifier` + retry/fallback-оркестрация | ✅ реализован | `backend/internal/service/classifier.go` |
| `IntentClassifier` — реальный LLM-адаптер | ❌ не реализован (есть только dev-заглушка `LocalHeuristicClassifier`, нет доступа к API-ключу провайдера) | `backend/internal/service/local_heuristic_classifier.go` |
| HTTP API (7 эндпоинтов: healthz, readyz, bosses×2, classify, battles×2) | ✅ реализован (без auth) | `backend/internal/handler/`, `backend/cmd/api/main.go` |
| Персистентность боёв (`internal/repository`, `internal/migrate`, `cmd/migrate`, `db/migrations`) | ✅ реализован (только `battle_sessions`, анонимно — без привязки к игроку) | `backend/internal/repository/`, `backend/internal/migrate/` |
| `ClassroomCohort`/`PlayerSession` авторизация | ❌ не реализована — персистентность есть, но бои анонимны | — |
| Web-клиент | ❌ не реализован | — |

## Архитектура (целевая, из HLD-спеки)

Backend — модульный монолит, один процесс (`backend-api`) для Фазы 0, без отдельного worker-процесса (осознанное решение — см. [ADR-001](../architecture-decisions.md#adr-001)). Слои: `delivery → application/domain → infrastructure`. Все три реализованы: `delivery` (`internal/handler`), `domain` (`internal/service`), `infrastructure` (`internal/repository`, `internal/migrate` — Postgres).

```text
Web App (не реализован)
  -> Backend API: Gin, cmd/api/main.go, internal/handler (реализован) [ЕСТЬ]
      -> ModerationModule (реализован: BasicModerator) [ЕСТЬ]
      -> IntentClassifier port + retry/fallback-оркестрация (реализованы) [ЕСТЬ]
          -> LocalHeuristicClassifier (dev-заглушка, реализована) [ЕСТЬ]
          -> реальный LLM-адаптер (не реализован — нет API-ключа)
      -> BattleEngine (реализован, покрыт тестами) [ЕСТЬ]
      -> BattleRepository port -> PostgresBattleRepository (реализованы) [ЕСТЬ]
          -> PostgreSQL 16 (battle_sessions, анонимно) [ЕСТЬ]
```

Полный путь «сырой текст → действие → сохранённый бой» уже собран и протестирован end-to-end на трёх уровнях: Go-вызовы (`TestFullPipeline_TextToAction`), реальный HTTP поверх `httptest.Server` (`TestHTTPPipeline_ClassifyThenBattle`, включая `POST /battles` → `id` → `GET /battles/{id}`), и вручную через `curl` на живом `go run ./cmd/api` с реальным Postgres. Не хватает только настоящего LLM за портом и авторизации — бои по-прежнему анонимны.

## Доменная модель (реализованная часть)

Полный список сущностей — в [HLD-спеке](../superpowers/specs/2026-09-03-battle-script-hld-design.md#доменная-модель). Реализованы в коде на сегодня:

- `HeroClass`, `ResourceType`, `TargetSelector`, `ConditionType`, `ActionType` — закрытые словари (`vocabulary.go`).
- `Condition`, `Action`, `Rule`, `IntentClassification` — структура тактики (`intent.go`).
- `Boss`, `BossPhase` — контент-модель босса (`boss.go`).
- `HeroDef`, `heroLiveState`, `battleState` (неэкспортируемые, внутреннее состояние симуляции) — `battle_state.go`.
- `BattleEvent`, `BattleTurn`, `BattleResult`, `BattleLog` — журнал и итог боя (`battle_engine.go`).
- `BattleRecord`, `HeroRosterEntry`, `BattleRepository` (порт) — персистентность боя (`battle_repository.go`); `PostgresBattleRepository` — Postgres-адаптер (`internal/repository`).

Не реализованы (есть только в спеке): `ClassroomCohort`, `PlayerSession`, `PromptSubmission`, `ModerationEvent`, `Achievement` — появятся вместе с авторизацией.

## Известные дизайн-решения, вытекшие из реализации (не были в исходных спеках)

Зафиксировано как ADR, не только как код-комментарий — см. [architecture-decisions.md](../architecture-decisions.md):

- Щит поглощает урон раньше HP ([ADR-005](../architecture-decisions.md#adr-005)) — обнаружено при разработке через упавший тест, не было явно в HLD.
- Правило таргетинга `stone_giant` (целится в `role:tank`, отступление уводит удар) — решено в спеке разрешения хода, не в исходной спеке боссов.
- Щит не восстанавливается ⇒ `retreat` по порогу щита фактически необратим до конца боя ([ADR-005, дополнение](../architecture-decisions.md#adr-005)) — вскрыто интеграционным HTTP-тестом, не исправлено, зафиксировано как открытый вопрос для плейтеста.
- `BattleLog`/`BattleResult`/`BattleTurn`/`BattleEvent` изначально не имели JSON-тегов (уходили в API как PascalCase) — вскрыто ручной проверкой `curl`, не юнит-тестами; исправлено, см. [verification-лог](../superpowers/verification/2026-09-03-http-api.md).

## Что читать дальше

- Полное обоснование каждого решения — датированные спеки в [`docs/superpowers/specs/`](../superpowers/specs/), в хронологическом порядке.
- Что именно реализовано и как это проверить — [`backend/README.md`](../../backend/README.md).
- Дорогие/трудно обратимые решения без разворачивания в спеку — [`docs/architecture-decisions.md`](../architecture-decisions.md).
