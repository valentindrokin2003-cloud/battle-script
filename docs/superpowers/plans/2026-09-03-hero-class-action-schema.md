# План: схема действий классов героев

Реализует [`2026-09-03-hero-class-action-schema-design.md`](../specs/2026-09-03-hero-class-action-schema-design.md).

## Задачи

| # | Файл | Что делает | TDD |
|---|---|---|---|
| 1 | `backend/internal/service/vocabulary.go` | Закрытые словари: `HeroClass`, `ResourceType`, `TargetSelector` (+ `role:<class>` парсинг), `ConditionType`, `ActionType`, per-class ability set, class-level fallback | Табличные тесты на валидные/невалидные значения каждого словаря |
| 2 | `backend/internal/service/vocabulary_test.go` | Тесты к п.1 | — |
| 3 | `backend/internal/service/intent.go` | Типы `Condition`/`Action`/`Rule`/`IntentClassification` + `ValidateIntentClassification` (server-side, не доверяет LLM вслепую) | Один валидный кейс (пример мага из спеки) + таблица отклонений (по одному на каждое правило валидации) |
| 4 | `backend/internal/service/intent_test.go` | Тесты к п.3 | — |
| 5 | `backend/internal/service/battle_context.go` | Типы `ResourceValue`/`UnitState`/`BossState`/`BattleContext` — состояние на момент решения одного героя | Косвенно, через тесты п.7 |
| 6 | `backend/internal/service/rule_engine.go` | `SelectAction` — реализация «Общей структуры: Tactic Program» из спеки: обход правил по приоритету, разрешение селекторов, fallback | — |
| 7 | `backend/internal/service/rule_engine_test.go` | Сценарные тесты `SelectAction` по всем 3 боссам из спеки боссов (см. зависимость ниже) + тест на нерезолвящийся селектор + тест на детерминизм | — |

## Порядок выполнения

Словарь (п.1) реализовывался первым, так как валидатор (п.3) на него ссылается напрямую — обратный порядок сделал бы intent.go некомпилируемым. Тесты `SelectAction` (п.7) используют боевые сценарии трёх стартовых боссов, поэтому фактически написаны после [плана скрипта боссов](./2026-09-03-boss-script.md) — зависимость на уровне тестовых данных, не на уровне типов (`rule_engine.go` не импортирует ничего боссо-специфичного, кроме уже реализованного `BossState.CurrentPhaseID`).

## Валидация

```bash
cd backend && gofmt -l . && go vet ./... && golangci-lint run ./... && go test ./internal/service/... -run 'TestParseRoleSelector|TestValidSelector|TestClassHasAbility|TestDefaultFallback|TestValidateIntentClassification|TestSelectAction' -v
```

## Трассировка требование → задача → тест

| Требование (из спеки) | Задача | Тест |
|---|---|---|
| Максимум 3 правила на героя | intent.go: `MaxRulesPerHero` | `TestValidateIntentClassification_Rejections/too_many_rules`, `TestValidateIntentClassification_MaxThreeRules` |
| LLM — не граница доверия, каждое поле перепроверяется | intent.go: `ValidateIntentClassification` | вся таблица `TestValidateIntentClassification_Rejections` |
| `role:<hero_class>` — динамический селектор | vocabulary.go: `ParseRoleSelector` | `TestParseRoleSelector` |
| Порог `threshold` только в `[0.0, 1.0]` | intent.go: `validateResourceThreshold` | `TestValidateIntentClassification_Rejections/threshold_out_of_range` |
| Правила исполняются по приоритету, первое истинное побеждает | rule_engine.go: `SelectAction` | `TestSelectAction_FrostWardenPhaseGating` |
| Нерезолвящийся селектор = правило неприменимо, не ошибка | rule_engine.go: `resolveAllyTarget`/`resolveTargetID` | `TestSelectAction_UnresolvedSelectorSkipsToNextRule` |
| Пустая программа/полный отказ LLM всё равно даёт действие | rule_engine.go: `SelectAction` fallback-путь | `TestSelectAction_EmptyProgramUsesFallback` |
| Детерминизм: одинаковый вход → одинаковый выход | rule_engine.go: `SelectAction` (чистая функция) | `TestSelectAction_Deterministic` |

## Статус

Выполнено, закоммичено в `935db90`. Этот план написан задним числом — рабочий процесс шёл спека → сразу код, план пропустили; см. обсуждение в сессии 2026-09-03.
