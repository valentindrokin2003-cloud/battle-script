# Персистентность боёв — дизайн

## Краткое описание

Первый шаг персистентности поверх уже реализованного и стейтлес HTTP API: `POST /api/v1/battles` начинает сохранять результат в PostgreSQL и возвращать `id`, добавляется `GET /api/v1/battles/{id}` для повторного чтения. Осознанно **не** вводится `PlayerSession`/`ClassroomCohort` — бои остаются анонимными записями в этом шаге; это прямое продолжение решения из HTTP API спеки («стейтлес API — временное упрощение, не архитектура на вырост»), теперь снимается только его персистентная часть, не авторизационная. `GET /readyz` добавляется впервые — раньше не было внешних зависимостей, которые имело бы смысл проверять, теперь есть Postgres.

## Цели

- Дать порт `BattleRepository` в domain-слое (`internal/service`) — постгрес не должен быть виден domain-коду напрямую, только через интерфейс, как и `IntentClassifier`/`Moderator`.
- Дать Postgres-адаптер этого порта в `internal/repository`.
- Дать минимальный SQL-раннер миграций (`cmd/migrate`), без внешней библиотеки — на одну таблицу полноценный migration-фреймворк избыточен.
- `POST /api/v1/battles` сохраняет `BattleLog` целиком как JSONB (не нормализует по ходам/событиям — нет текущей потребности запрашивать отдельные ходы, YAGNI) и возвращает сгенерированный `id`.
- `GET /api/v1/battles/{id}` — чтение сохранённого боя.
- `GET /readyz` — проверяет реальное соединение с БД (`ping`), в отличие от `/healthz`, который остаётся чистой liveness-проверкой без внешних вызовов.

## Не-цели

- Не вводить `PlayerSession`/`ClassroomCohort`/авторизацию — бои анонимны, поле "чей это бой" не хранится. Явно отдельная будущая работа.
- Не нормализовывать `BattleTurn`/`BattleEvent` в отдельные таблицы — нет запроса, требующего доступ к одному ходу отдельно от всего боя.
- Не вводить ORM — `database/sql` + `pgx` напрямую, простые SQL-запросы, как и предполагает HLD (`repository`-слой — тонкая обвязка, не фреймворк).
- Не подключать Docker/`docker-compose` — в этом окружении их нет; PostgreSQL 16 установлен через Homebrew и запущен как локальный сервис. Docker-based `docker-compose.yml` для командной разработки — открытый вопрос, не блокирует эту спеку.
- Не хранить `IntentClassification`/`PromptSubmission`/`ModerationEvent` отдельно — это персистентность самого боя, не всего пути от текста до намерения; следующий естественный шаг, не этот.

## Схема

```sql
CREATE TABLE battle_sessions (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    boss_id     text NOT NULL,
    outcome     text NOT NULL,       -- денормализовано из battle_log->result->outcome для будущей фильтрации
    hero_roster jsonb NOT NULL,      -- [{id, hero_class}], для отображения без разбора всего battle_log
    battle_log  jsonb NOT NULL,      -- весь BattleLog целиком
    created_at  timestamptz NOT NULL DEFAULT now()
);
```

`outcome` вынесен отдельной колонкой намеренно — единственное поле, для которого уже сейчас предсказуема потребность в фильтрации/индексе (например, «показать только победы»), всё остальное содержимое `BattleLog` остаётся только в JSONB.

## Порт и адаптер

```go
// internal/service
type BattleRecord struct {
    ID         string
    BossID     string
    HeroRoster []HeroRosterEntry // {ID, HeroClass}
    Log        BattleLog
    CreatedAt  time.Time
}

type BattleRepository interface {
    Save(ctx context.Context, record BattleRecord) (id string, err error)
    Get(ctx context.Context, id string) (BattleRecord, error)
}

var ErrBattleNotFound = errors.New("battle not found")
```

`internal/repository.PostgresBattleRepository` реализует этот интерфейс через `database/sql` (драйвер `pgx/v5/stdlib`), без бизнес-логики — только маппинг Go-структур в строки таблицы и обратно.

## Миграции

`cmd/migrate` — отдельный бинарник (как в Sectr), не автозапуск при старте API:

- `db/migrations/0001_create_battle_sessions.sql` — единственная миграция этого шага.
- Раннер читает файлы `NNNN_*.sql` по порядку, применяет ещё не применённые внутри транзакции, отмечает в служебной таблице `schema_migrations (version int primary key, applied_at timestamptz)`.
- Идемпотентен: повторный запуск на уже мигрированной БД — no-op.

## Изменения HTTP API

| Метод | Путь | Изменение |
|---|---|---|
| `POST` | `/api/v1/battles` | Тело запроса не меняется. После `RunBattle` сохраняет `BattleRecord` через `BattleRepository.Save`, возвращает `{id, boss_id, result, turns}` — то же тело, что раньше, плюс `id`. |
| `GET` | `/api/v1/battles/{id}` | Новый. `200` с тем же форматом ответа, что и `POST /battles`, `404 {"error":"battle_not_found"}`, если `id` не существует. |
| `GET` | `/readyz` | Новый. Пингует БД; `200 {"status":"ready"}` при успехе, `503 {"status":"not_ready","reason":"..."}` при ошибке соединения. |

Ошибка сохранения после успешного `RunBattle` (Postgres недоступен и т.п.) — `502`, `{"error":"persistence_failed", ...}`; бой уже был честно разыгран, но результат не сохранился — тело ответа явно говорит, что `id` отсутствует, а не молчит об этом.

## Обработка ошибок

- Недоступность БД при старте `cmd/api` — сервер всё равно стартует (fail open на уровне процесса), но `/readyz` сразу отражает `not_ready`; сами `/battles`-запросы будут падать в `persistence_failed`, пока БД не появится — так оркестратор (k8s/health-check) видит проблему через `/readyz`, не через 5xx на пользовательском трафике.
- Непримененные миграции при старте `cmd/api` не проверяются автоматически в этом шаге (нет схемы "запрети старт, если миграции не применены") — открытый вопрос ниже.

## Тестирование и валидация

- `cmd/migrate` тестируется на реальном `battle_script_test` (запущен локально через Homebrew, без Docker): применить с нуля, применить второй раз (идемпотентность), проверить существование таблицы.
- `PostgresBattleRepository`: `Save` → `Get` round-trip на реальной БД, `Get` несуществующего `id` возвращает `ErrBattleNotFound`.
- HTTP-слой: `httptest` с реальным репозиторием на `battle_script_test` (не мок) — `POST /battles` → `id` в ответе → `GET /battles/{id}` возвращает то же самое; `GET /readyz` — `200` при поднятой тестовой БД.
- Тесты, требующие БД, помечены явно и пропускаются (`t.Skip`), если переменная окружения `TEST_DATABASE_URL` не задана — чтобы `go test ./...` не падал в окружении без Postgres, но в этом окружении Postgres есть и тесты гоняются по-настоящему, не мокаются.

## Открытые вопросы

- Docker/`docker-compose.yml` для командной разработки (у вас, вероятно, Docker есть, в отличие от этого окружения) — стоит добавить отдельным шагом, чтобы не полагаться на Homebrew-специфичный локальный Postgres при разработке на другой машине.
- Автоматическая проверка «миграции применены» при старте `cmd/api` — сейчас полагается на ручную дисциплину (`cmd/migrate` перед `cmd/api`), как и в Sectr; ужесточать или нет — вопрос операционного удобства, не архитектуры.
- Ретеншн — сколько хранить анонимные `battle_sessions` без привязки к игроку — не решено, не блокирует эту спеку.
