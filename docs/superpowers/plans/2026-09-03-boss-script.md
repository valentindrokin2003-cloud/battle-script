# План: скрипт боссов

Реализует [`2026-09-03-boss-script-design.md`](../specs/2026-09-03-boss-script-design.md).

## Задачи

| # | Файл | Что делает | TDD |
|---|---|---|---|
| 1 | `backend/internal/service/boss.go` | Типы `Boss`/`BossPhase`, `Boss.HasPhase` (используется валидатором из предыдущего плана), `Boss.ValidatePhases` (контроль сид-данных: монотонные пороги от 1.0), `Boss.PhaseForHPFraction` (резолюция текущей фазы по HP) | Табличные тесты на валидные/невалидные наборы фаз + таблица `fraction -> ожидаемая фаза` |
| 2 | `backend/internal/service/boss_test.go` | Тесты к п.1, включая `frostWarden()` — тестовый хелпер с полным JSON-примером боссa из спеки | — |

## Зависимость от предыдущего плана

`Boss.HasPhase` используется в `intent.go`'s `validateCondition` для проверки `phase_id` в условии `boss_phase_is` — этот план физически не мог идти раньше [хero-class-action-schema плана](./2026-09-03-hero-class-action-schema.md), но логически они независимы по содержанию (словарь действий не знает о структуре `Boss`, только импортирует её).

## Валидация

```bash
cd backend && gofmt -l . && go vet ./... && golangci-lint run ./... && go test ./internal/service/... -run TestBoss -v
```

## Трассировка требование → задача → тест

| Требование (из спеки) | Задача | Тест |
|---|---|---|
| Фазы начинаются строго с порога 1.0, без пропусков | `Boss.ValidatePhases` | `TestBoss_ValidatePhases_Rejections` (все подслучаи) |
| Условие не может ссылаться на нераскрытую фазу («правило честности») | `Boss.HasPhase` (вызывается из `intent.go`) | `TestValidateIntentClassification_Rejections/unknown_phase_id` (в предыдущем плане) |
| Резолюция активной фазы по HP | `Boss.PhaseForHPFraction` | `TestBoss_PhaseForHPFraction` (границы 1.0/0.8/0.6/0.4/0.25/0.0) |

## Статус

Выполнено, закоммичено в `3ecb143`. План написан задним числом.
