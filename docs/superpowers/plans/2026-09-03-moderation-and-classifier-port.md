# План: модерация и порт `IntentClassifier`

Реализует [`2026-09-03-moderation-and-classifier-port-design.md`](../specs/2026-09-03-moderation-and-classifier-port-design.md).

## Задачи

| # | Файл | Что делает | TDD |
|---|---|---|---|
| 1 | `backend/internal/service/moderation.go` | `ModerationResult`, `Moderator` интерфейс, `BasicModerator` (длина, профанити, PII-паттерны) | Red: таблица кейсов пишется первой и падает (тип ещё не существует), затем реализация до зелёного |
| 2 | `backend/internal/service/moderation_test.go` | Тесты к п.1 | — |
| 3 | `backend/internal/service/classifier.go` | `ClassificationContext`, `IntentClassifier` интерфейс, `ClassifyWithFallback` (retry-оркестрация, независимая от конкретного адаптера) | Red: тест с мок-классификатором, считающим вызовы, пишется первым |
| 4 | `backend/internal/service/classifier_test.go` | Тесты к п.3, включая мок, падающий N раз подряд | — |
| 5 | `backend/internal/service/local_heuristic_classifier.go` | `LocalHeuristicClassifier` — детерминированная dev-заглушка `IntentClassifier`, явно помеченная как не-production в комментарии пакета | Red: по одному кейсу «текст → ожидаемый TacticProgram» на класс пишется до реализации таблицы шаблонов |
| 6 | `backend/internal/service/local_heuristic_classifier_test.go` | Тесты к п.5 + интеграционный тест полного пути (текст → модерация → классификация → валидация → `SelectAction`) | — |

## Порядок выполнения

Модерация (п.1–2) не зависит ни от чего нового — реализуется первой. Порт и оркестрация (п.3–4) зависят только от уже существующих `IntentClassification`/`ValidateIntentClassification`/`DefaultFallback` (предыдущие планы), не от модерации. `LocalHeuristicClassifier` (п.5–6) зависит от порта (п.3), поэтому идёт последней и её тесты первыми используют весь конвейер целиком.

На этот раз TDD не декларативный, а реальный: для каждого файла тест пишется и запускается **до** реализации (red), затем добавляется код до зелёного — в отличие от предыдущего плана, где это сделал только один тест по факту.

## Валидация

```bash
cd backend && make check   # gofmt-check, go vet, golangci-lint, go test ./... -race
```

## Трассировка требование → задача → тест

| Требование (из спеки) | Задача | Тест |
|---|---|---|
| Модерация — обязательный гейт до LLM, не постобработка | classifier.go: оркестрация вызывает `Moderator.Check` до `IntentClassifier.Classify` | интеграционный тест п.6 (отклонённый текст не доходит до классификатора) |
| Один retry, затем `DefaultFallback` с `low_fallback_used` | classifier.go: `ClassifyWithFallback` | classifier_test.go: мок, падающий 1/2/3+ раз подряд |
| Заглушка explicitly не выдаётся за LLM | local_heuristic_classifier.go: doc-комментарий пакета/типа | ревью кода, не автоматизируемо тестом — фиксируется здесь как ручная проверка перед коммитом |
| Узнанный результат заглушки всегда проходит `ValidateIntentClassification` | local_heuristic_classifier.go | local_heuristic_classifier_test.go: каждый positive-кейс дополнительно валидируется |
| PII/профанити/пустой/слишком длинный текст отклоняются | moderation.go: `BasicModerator.Check` | moderation_test.go: таблица кейсов |

## Статус

Не начато — план написан до кода, реализация следующим шагом в этой же сессии.
