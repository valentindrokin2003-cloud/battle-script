# Web-клиент — дизайн

## Краткое описание

Первый веб-клиент поверх уже реализованного и протестированного HTTP API: минимальный, но полный игровой цикл Фазы 0 — выбор босса → раскрытие его фаз (правило честности) → ввод тактики текстом для 4 героев → показ распознанного намерения для подтверждения (обучающий крючок из HLD) → розыгрыш боя по ходам → результат. React + TypeScript + Vite SPA, без роутинга (один линейный поток экранов, стейт-машина на React state), без Redux/тяжёлых стейт-менеджеров — YAGNI для объёма Фазы 0.

## Цели

- Реализовать полный цикл на фронтенде, зеркально совпадающий с уже задокументированным HTTP-путём: `GET /bosses` → `POST /tactics/classify` ×4 → показ намерений → `POST /battles` → `GET /battles/{id}` (для повторного просмотра) → рендер `BattleLog`.
- Типизированный API-клиент, соответствующий wire-контрактам из HTTP API и persistence спек буквально (одни и те же имена полей).
- В dev-режиме — прокси Vite на backend (`/api/*` → `http://localhost:8080`), чтобы не решать CORS для Фазы 0 (сборка/деплой раздельно — открытый вопрос, не блокирует этот шаг).
- Компонентные тесты (Vitest + React Testing Library) на логику стейт-машины экранов и рендеринг ключевых состояний (loading/error/пусто/успех) — по аналогии с требованием из методологии Sectr «verify loading, error, empty, success states».

## Не-цели

- Не вводить роутинг (`react-router` и т.п.) — один линейный поток, отдельные URL не нужны для Фазы 0.
- Не вводить дизайн-систему/тяжёлый CSS-фреймворк — минимальная читаемая вёрстка, косметика — отдельная работа после того, как цикл целиком работает.
- Не решать деплой/раздельную сборку фронта и бэка — только dev-режим (`vite dev` + прокси).
- Не давать выбор состава героев — фиксированный ростер из 4 классов (танк/лучник/маг/целитель), как и заложено в контенте Фазы 0.
- Не визуализировать `BattleLog` анимацией/canvas — текстовый пошаговый лог из уже готовых полей (`actor`, `action_type`, `target`, `amount`), анимация — отдельная будущая работа.

## Экраны (стейт-машина)

```text
boss-select
  -> (GET /api/v1/bosses, выбор одного босса)
tactic-input
  -> (текстовое поле на каждого из 4 фиксированных героев, показ раскрытых фаз выбранного босса)
intent-review
  -> (POST /api/v1/tactics/classify на каждого героя, показ Rules/Confidence,
      предупреждение "не поняли однозначно" при confidence low_fallback_used)
battle-result
  -> (POST /api/v1/battles с подтверждёнными intent'ами, пошаговый BattleLog, итог)
```

Переход `intent-review -> tactic-input` возможен (переписать тактику), `tactic-input -> boss-select` — тоже (сменить босса). `battle-result` — терминальный экран с кнопкой «сыграть ещё раз» (`-> boss-select`).

## API-клиент

Один модуль `src/api/client.ts`, типы 1:1 с Go DTO (те же имена полей, snake_case через ручную аннотацию интерфейсов — без кодогенерации в Фазе 0, вручную поддерживаемое соответствие, задокументированное как риск в открытых вопросах):

```ts
export interface BossPhase { phase_id: string; hp_threshold_enter: number; ability_pattern: string[]; provocation: string }
export interface Boss { boss_id: string; display_name: string; phases: BossPhase[] }

export interface Condition { type: string; target?: string; resource?: string; threshold?: number; phase_id?: string; status?: string }
export interface Action { type: string; target?: string }
export interface Rule { priority: number; condition: Condition; action: Action }
export interface IntentClassification {
  hero_class: string; schema_version: string; rules: Rule[];
  fallback_action: Action; source_prompt_submission_id?: string; confidence: "high" | "low_fallback_used";
}

export interface BattleEvent { actor: string; action_type: string; target?: string; amount: number; target_hp_after: number }
export interface BattleTurn { turn_number: number; events: BattleEvent[] }
export interface BattleResult { outcome: "victory" | "defeat" | "aborted"; turns_taken: number; boss_id: string }
export interface BattleResponse { id: string; boss_id: string; turns: BattleTurn[]; result: BattleResult }

export function listBosses(): Promise<Boss[]>
export function classifyTactic(req: { hero_class: string; boss_id: string; prompt_text: string }): Promise<IntentClassification>
export function runBattle(req: { boss_id: string; heroes: { id: string; hero_class: string; intent: IntentClassification }[] }): Promise<BattleResponse>
export function getBattle(id: string): Promise<BattleResponse>
```

Каждая функция кидает типизированную ошибку с полями backend'а (`error`, `message`) при не-2xx ответе — компоненты различают `moderation_rejected`/`invalid_intent`/сетевые ошибки по этому полю, не по тексту сообщения.

## Обработка ошибок в UI

- `moderation_rejected` на `classifyTactic` — показывается прямо под полем ввода конкретного героя, ребёнок может переписать текст, не теряя остальных.
- `invalid_intent`/`unknown_boss` на `runBattle` — не должны происходить при штатном использовании (UI сам не даёт добраться до этого состояния), но на случай гонки/бага показывается общая ошибка экрана с возможностью вернуться к `tactic-input`.
- Сетевая ошибка (backend недоступен) — общий error-стейт с кнопкой повтора, на каждом экране, где есть сетевой вызов.
- `confidence: "low_fallback_used"` — не ошибка транспорта, а игровой сигнал; показывается как явная пометка «не поняли однозначно» на экране `intent-review`, ключевая часть обучающего цикла из HLD, не второстепенная деталь UI.

## Тестирование и валидация

- Vitest + React Testing Library: рендер каждого экрана в loading/error/success состояниях с замоканным `api/client`.
- Тест на полный happy path стейт-машины (переходы `boss-select -> tactic-input -> intent-review -> battle-result`) с замоканными ответами API.
- Тест на явный показ пометки `low_fallback_used`.
- `npm run build` — проверка, что TypeScript компилируется без ошибок (это и есть основная защита от рассинхронизации типов с backend, раз кодогенерации нет).
- Ручная проверка (`docs/superpowers/verification/`): `vite dev` с прокси на реальный `go run ./cmd/api`, полный цикл руками через реальный браузер, если инструмент для этого доступен в среде выполнения; если недоступен — прогон через `curl`-эквивалентные проверки сети (devtools недоступны) фиксируется как явное ограничение проверки, а не как «протестировано».

## Открытые вопросы

- Ручное дублирование типов API-клиента с Go DTO — риск рассинхронизации при изменении backend; кодогенерация (например, из OpenAPI) — кандидат на будущую спеку, не блокирует эту.
- Раздельный деплой фронта/бэка, CORS в продакшене — не решено, вне скоупа Фазы 0 (dev-прокси Vite закрывает текущую потребность).
- Показ `IntentClassification` в `intent-review` — сырой JSON правил или перевод в человекочитаемое предложение («Танк будет провоцировать босса...») — вопрос уже поднимался в спеке схемы действий как открытый; для первой версии экрана будет читаемое, но дословное отображение правил (`condition -> action` по-русски), не полный natural-language перевод — компромисс ради скорости первой версии.
