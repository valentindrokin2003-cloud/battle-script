# Battle Script

Автобаттлер, где стратегический слой — это тактика, которую ребёнок пишет простым языком, а LLM переводит её в закрытый набор действий для детерминированного боя.

## Документы

- [`BUSINESS_PLAN.md`](./BUSINESS_PLAN.md) — концепция, рынок, юнит-экономика, риски, дорожная карта.
- [`COMPETITIVE_ANALYSIS.md`](./COMPETITIVE_ANALYSIS.md) — детальный разбор конкурентов.
- [`docs/superpowers/specs/`](./docs/superpowers/specs/) — versioned design-спеки (spec-driven development), по одной на архитектурное решение или фичу.
- [`docs/superpowers/plans/`](./docs/superpowers/plans/) — исполняемые планы реализации, привязанные к спекам.
- [`backend/README.md`](./backend/README.md) — что уже реализовано в backend и что нет.
- [`docs/architecture/hld.md`](./docs/architecture/hld.md) — текущее состояние архитектуры (живой документ, не история решений).
- [`docs/architecture-decisions.md`](./docs/architecture-decisions.md) — ADR-лог: только дорогие/трудно обратимые решения.

## Текущий статус

Фаза 0 (по дорожной карте бизнес-плана): закрытый прототип для школьного пилота, веб-платформа. Архитектура зафиксирована в семи спеках в [`docs/superpowers/specs/`](./docs/superpowers/specs/): HLD, схема действий классов героев, скрипт боссов, разрешение боевого хода, модерация и порт `IntentClassifier`, HTTP API, персистентность боёв, web-клиент.

**Backend**: доменное ядро (`internal/service`, детерминированный боевой движок, модерация, классификация намерений), HTTP API (`internal/handler`, 7 эндпоинтов) и персистентность боёв в PostgreSQL (`internal/repository`, `internal/migrate`) — реализованы и покрыты тестами, часть которых гоняется на реальном локальном Postgres, не на моках.

**Frontend** (`web/`): React/TypeScript SPA — полный игровой цикл (выбор босса → тактика на 4 героев → показ распознанного намерения → бой → результат), 23 компонентных теста, сетевой путь через Vite-прокси подтверждён на реальном backend. Интерактивная браузерная проверка не выполнена — в среде разработки не было браузерного инструмента, см. [`docs/superpowers/verification/`](./docs/superpowers/verification/).

Ещё не реализовано: реальный LLM-адаптер (сейчас — dev-заглушка на ключевых словах, нет доступа к API-ключу провайдера в этой среде), авторизация/`PlayerSession` (бои анонимны). Подробности — [`backend/README.md`](./backend/README.md), [`web/README.md`](./web/README.md) и [`docs/architecture/hld.md`](./docs/architecture/hld.md).
