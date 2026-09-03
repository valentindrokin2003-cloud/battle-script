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

Фаза 0 (по дорожной карте бизнес-плана): закрытый прототип для школьного пилота, веб-платформа. Архитектура зафиксирована в трёх спеках в [`docs/superpowers/specs/`](./docs/superpowers/specs/): HLD, схема действий классов героев, скрипт боссов.

Backend: доменное ядро (`internal/service`) реализовано и покрыто тестами — закрытый словарь тактик, валидация `IntentClassification`, модель боссов, детерминированный движок выбора действия. HTTP-слой, персистентность, модерация, LLM-адаптер и накопление лога боя ещё не реализованы — подробности в [`backend/README.md`](./backend/README.md).
