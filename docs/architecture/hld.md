# Архитектура Battle Script — текущее состояние

Живой документ: описывает, как система устроена **сейчас**, а не как решили в конкретную дату. За обоснованием решений и полным контекстом — см. датированные спеки в [`../superpowers/specs/`](../superpowers/specs/); при расхождении между этим файлом и спекой доверять текущему коду и этому файлу, спеки могут отставать.

## Итог в двух словах

Battle Script — веб-автобаттлер, где ребёнок пишет тактику текстом, LLM переводит её в закрытый набор действий, а бой разыгрывает детерминированный движок. Реализовано и протестировано: доменное ядро боевой системы (Go, без внешних зависимостей). Не реализовано: HTTP-сервер, БД, сам вызов LLM, фронтенд, авторизация.

## Компоненты и их статус

| Компонент | Статус | Где в коде |
|---|---|---|
| Закрытый словарь тактик (классы, ресурсы, селекторы, условия, способности) | ✅ реализован | `backend/internal/service/vocabulary.go` |
| Валидация `IntentClassification` (LLM — не граница доверия) | ✅ реализован | `backend/internal/service/intent.go` |
| Модель `Boss`/`BossPhase`, резолюция фазы по HP | ✅ реализован | `backend/internal/service/boss.go` |
| Детерминированный выбор действия героя (`SelectAction`) | ✅ реализован | `backend/internal/service/rule_engine.go` |
| Полный цикл боя (урон/лечение/щит/таргетинг босса/`BattleLog`) | ✅ реализован | `backend/internal/service/battle_engine.go`, `battle_state.go` |
| Сид-контент Фазы 0 (3 класса героев, 3 босса, иллюстративные числа) | ✅ реализован | `backend/internal/service/phase0_content.go` |
| `ModerationModule` | ❌ не реализован | — |
| `IntentClassifier` — реальный LLM-адаптер | ❌ не реализован (есть только валидатор ответа) | — |
| HTTP API (`cmd/api`, `internal/handler`) | ❌ не реализован | — |
| Персистентность (`internal/repository`, `db/migrations`, Postgres) | ❌ не реализован | — |
| `ClassroomCohort`/`PlayerSession` авторизация | ❌ не реализована | — |
| Web-клиент | ❌ не реализован | — |

## Архитектура (целевая, из HLD-спеки)

Backend — модульный монолит, один процесс (`backend-api`) для Фазы 0, без отдельного worker-процесса (осознанное решение — см. [ADR-001](../architecture-decisions.md#adr-001)). Слои: `delivery → application/domain → infrastructure`. Сейчас существует только `domain`-слой (`internal/service`); `delivery` и `infrastructure` ещё не начаты.

```text
Web App (не реализован)
  -> Backend API (не реализован)
      -> ModerationModule (не реализован)
      -> IntentClassifier port -> LLM Adapter (не реализованы; есть только
         server-side валидатор ожидаемого формата ответа)
      -> BattleEngine (реализован, покрыт тестами) [ЭТО ЕСТЬ]
      -> PostgreSQL (не реализован)
```

## Доменная модель (реализованная часть)

Полный список сущностей — в [HLD-спеке](../superpowers/specs/2026-09-03-battle-script-hld-design.md#доменная-модель). Реализованы в коде на сегодня:

- `HeroClass`, `ResourceType`, `TargetSelector`, `ConditionType`, `ActionType` — закрытые словари (`vocabulary.go`).
- `Condition`, `Action`, `Rule`, `IntentClassification` — структура тактики (`intent.go`).
- `Boss`, `BossPhase` — контент-модель босса (`boss.go`).
- `HeroDef`, `heroLiveState`, `battleState` (неэкспортируемые, внутреннее состояние симуляции) — `battle_state.go`.
- `BattleEvent`, `BattleTurn`, `BattleResult`, `BattleLog` — журнал и итог боя (`battle_engine.go`).

Не реализованы (есть только в спеке): `ClassroomCohort`, `PlayerSession`, `PromptSubmission`, `ModerationEvent`, `Achievement` — эти сущности появятся вместе с HTTP-слоем/персистентностью.

## Известные дизайн-решения, вытекшие из реализации (не были в исходных спеках)

Зафиксировано как ADR, не только как код-комментарий — см. [architecture-decisions.md](../architecture-decisions.md):

- Щит поглощает урон раньше HP ([ADR-005](../architecture-decisions.md#adr-005)) — обнаружено при разработке через упавший тест, не было явно в HLD.
- Правило таргетинга `stone_giant` (целится в `role:tank`, отступление уводит удар) — решено в спеке разрешения хода, не в исходной спеке боссов.

## Что читать дальше

- Полное обоснование каждого решения — датированные спеки в [`docs/superpowers/specs/`](../superpowers/specs/), в хронологическом порядке.
- Что именно реализовано и как это проверить — [`backend/README.md`](../../backend/README.md).
- Дорогие/трудно обратимые решения без разворачивания в спеку — [`docs/architecture-decisions.md`](../architecture-decisions.md).
