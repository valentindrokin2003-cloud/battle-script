# План: персистентность боёв

Реализует [`2026-09-03-battle-persistence-design.md`](../specs/2026-09-03-battle-persistence-design.md).

## Задачи

| # | Файл | Что делает | TDD |
|---|---|---|---|
| 1 | `backend/db/migrations/0001_create_battle_sessions.sql` | DDL таблицы `battle_sessions` | — (SQL-файл, не Go) |
| 2 | `backend/internal/migrate/migrate.go` | Раннер: читает `db/migrations/*.sql`, применяет непринятые в транзакции, пишет в `schema_migrations` | Red: тест на реальной `battle_script_test` — применить, проверить `schema_migrations`, применить второй раз (идемпотентность) |
| 3 | `backend/internal/migrate/migrate_test.go` | Тесты к п.2 | — |
| 4 | `backend/cmd/migrate/main.go` | Тонкая обёртка CLI над `internal/migrate` | Ручная проверка (`go run ./cmd/migrate`), `main` юнитами не тестируется по конвенции |
| 5 | `backend/internal/service/battle_repository.go` | `BattleRecord`, `HeroRosterEntry`, `BattleRepository` порт, `ErrBattleNotFound` | — (только типы/интерфейс, нет логики для теста) |
| 6 | `backend/internal/repository/postgres_battle_repository.go` | `PostgresBattleRepository` — реализация порта на `database/sql`+`pgx/v5/stdlib` | Red: `Save`→`Get` round-trip тест на `battle_script_test`, пишется первым и падает на несуществующем типе |
| 7 | `backend/internal/repository/postgres_battle_repository_test.go` | Тесты к п.6, включая `Get` несуществующего id → `ErrBattleNotFound` | — |
| 8 | `backend/internal/handler/battles.go` (правка) | `BattlesHandler` получает `Repository service.BattleRepository`, `Run` сохраняет результат, новый `Get` для `GET /battles/{id}` | Red: тест на новое поведение (id в ответе, 502 на ошибку сохранения через мок-репозиторий) пишется до правки |
| 9 | `backend/internal/handler/battles_test.go` (правка) | Тесты к п.8 + мок `BattleRepository` для теста 502-ветки | — |
| 10 | `backend/internal/handler/readyz.go` | `GET /readyz` — пингует БД через `*sql.DB` | Red → green |
| 11 | `backend/internal/handler/readyz_test.go` | Тесты к п.10: живая тестовая БД → 200, закрытое соединение → 503 | — |
| 12 | `backend/internal/handler/router.go` (правка) | Регистрирует `GET /battles/:id`, `GET /readyz`, пробрасывает `*sql.DB`/`BattleRepository` | Косвенно через интеграционный тест п.13 |
| 13 | `backend/internal/handler/pipeline_integration_test.go` (правка/добавление) | HTTP-путь: `POST /battles` → `id` → `GET /battles/{id}` возвращает то же самое, на реальной `battle_script_test` | Red → green |
| 14 | `backend/cmd/api/main.go` (правка) | Читает `DATABASE_URL`, открывает пул, применяет `internal/migrate` не запускает сам (только `cmd/migrate` мигрирует) — просто использует БД | Ручная проверка |

## Порядок выполнения

Миграции (1–4) — основа, ничего не работает без применённой схемы. Порт (5) — чистые типы, идёт до адаптера. Адаптер (6–7) зависит от применённой миграции в `battle_script_test` (запускается перед тестами репозитория). Handler (8–11) зависит от порта, но не от конкретного адаптера — тестируется и мок-репозиторием (502-ветка), и реальным (round-trip). Router и интеграционный тест (12–13) — последние. `main.go` (14) — сборочный, последний.

## Валидация

```bash
# один раз перед прогоном тестов, использующих БД
export TEST_DATABASE_URL="postgres://localhost:5432/battle_script_test?sslmode=disable"
cd backend && go run ./cmd/migrate -database "$TEST_DATABASE_URL"

cd backend && make check   # fmt-check, vet, golangci-lint, go test ./... -race
```

Ручной прогон после п.14:

```bash
export DATABASE_URL="postgres://localhost:5432/battle_script_dev?sslmode=disable"
cd backend && go run ./cmd/migrate -database "$DATABASE_URL"
go run ./cmd/api &
curl localhost:8080/readyz
curl -X POST localhost:8080/api/v1/battles -d '{...}'   # получить id
curl localhost:8080/api/v1/battles/<id>                 # тот же результат
```

Результат — вторая запись в `docs/superpowers/verification/`.

## Трассировка требование → задача → тест

| Требование (из спеки) | Задача | Тест |
|---|---|---|
| Миграции идемпотентны | internal/migrate | migrate_test.go: второй прогон — no-op |
| `Get` несуществующего боя → `ErrBattleNotFound` | postgres_battle_repository.go | postgres_battle_repository_test.go |
| Ошибка сохранения после честно разыгранного боя → `502`, явный ответ | handler/battles.go | battles_test.go: мок-репозиторий с ошибкой `Save` |
| `/readyz` реально пингует БД, не заглушка | handler/readyz.go | readyz_test.go: закрытое соединение → 503 |
| `POST` → `id` → `GET` даёт тот же результат через реальный HTTP | pipeline_integration_test.go | `TestHTTPPipeline_PersistedBattleRoundTrip` |

## Статус

Не начато — план написан до кода.
