# Ручная проверка: персистентность боёв

Спека: [`2026-09-03-battle-persistence-design.md`](../specs/2026-09-03-battle-persistence-design.md). План: [`2026-09-03-battle-persistence.md`](../plans/2026-09-03-battle-persistence.md).

## Окружение

PostgreSQL 16 установлен через Homebrew (`brew install postgresql@16`, `brew services start postgresql@16`) — в этом окружении нет Docker. Базы `battle_script_dev` и `battle_script_test` созданы локально.

## Что проверялось

`cmd/migrate` и `cmd/api` на реальном порту с реальным Postgres, не только `httptest`/моки.

## Прогон

```
$ go run ./cmd/migrate   # DATABASE_URL=postgres://localhost:5432/battle_script_dev?sslmode=disable
2026/09/03 applied 0001_create_battle_sessions.sql

$ curl -s localhost:8099/readyz
{"status":"ready"}

$ curl -s -X POST localhost:8099/api/v1/battles -d '{...4 героя...}'
# -> HTTP 200, {"id":"615e8baf-...", "boss_id":"frost_warden", "turns":[...], "result":{"outcome":"victory","turns_taken":24,...}}

$ curl -s localhost:8099/api/v1/battles/615e8baf-7885-4196-905b-6234bd66660e
# -> HTTP 200, тот же result/turns, что вернул POST — реально прочитано из таблицы battle_sessions, не из памяти процесса
```

## Статус

Пройдено. `make check` чист (включая реальные тесты `internal/migrate`, `internal/repository`, `internal/handler` — все на `battle_script_test`, не смоканы). Отдельно юнит-тестами хендлера проверена 502-ветка (`persistence_failed`) через фейковый репозиторий с инжектированной ошибкой — она не проверялась вручную curl'ом, потому что для этого пришлось бы искусственно ронять живую БД посреди прогона, а это уже покрыто на уровне мока честно и без риска сломать текущую dev-базу.
