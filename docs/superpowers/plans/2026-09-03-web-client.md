# План: web-клиент

Реализует [`2026-09-03-web-client-design.md`](../specs/2026-09-03-web-client-design.md).

## Задачи

| # | Файл | Что делает | Тесты |
|---|---|---|---|
| 1 | `web/` (scaffold) | `npm create vite@latest -- --template react-ts`, базовые зависимости + Vitest + React Testing Library | — |
| 2 | `web/vite.config.ts` (правка) | Прокси `/api` → `http://localhost:8080` в dev | — |
| 3 | `web/src/api/types.ts` | Типы из спеки (`Boss`, `IntentClassification`, `BattleResponse` и т.д.) | — (чистые типы) |
| 4 | `web/src/api/client.ts` | `listBosses`/`classifyTactic`/`runBattle`/`getBattle`, типизированная ошибка `ApiError{error, message}` | `client.test.ts`: мок `fetch`, happy path + не-2xx → `ApiError` с полями |
| 5 | `web/src/api/client.test.ts` | Тесты к п.4 | — |
| 6 | `web/src/screens/BossSelect.tsx` | Экран выбора босса: loading/error/success | `BossSelect.test.tsx`: 3 состояния |
| 7 | `web/src/screens/TacticInput.tsx` | Ввод тактики на 4 героя + отображение раскрытых фаз босса | `TacticInput.test.tsx` |
| 8 | `web/src/screens/IntentReview.tsx` | Показ распознанных правил на герой, явная пометка `low_fallback_used` | `IntentReview.test.tsx`: включая тест на пометку |
| 9 | `web/src/screens/BattleResultScreen.tsx` | Пошаговый `BattleLog` + итог | `BattleResultScreen.test.tsx` |
| 10 | `web/src/App.tsx` | Стейт-машина экранов (`boss-select → tactic-input → intent-review → battle-result`), переходы назад | `App.test.tsx`: полный happy path с замоканным `api/client` |
| 11 | `web/src/main.tsx`, `web/index.html` | Точка входа Vite (в основном из шаблона) | — |

## Порядок выполнения

Scaffold и прокси (1–2) — основа. Типы и клиент (3–5) — API-слой, ничего выше по стеку не работает без него. Экраны (6–9) зависят от типов клиента, но не друг от друга — можно писать в любом порядке между собой; идут в порядке потока пользователя для удобства ручной проверки по ходу. `App.tsx` (10) — последний, собирает всё в стейт-машину.

## Валидация

```bash
cd web && npm run build   # TypeScript компилируется — единственная защита от рассинхронизации типов с backend, раз кодогенерации нет
cd web && npm test        # Vitest, все компонентные тесты
```

Ручной прогон (после п.10, если в среде есть браузер/инструмент для него):

```bash
# backend
cd backend && DATABASE_URL=... go run ./cmd/api

# frontend, отдельный терминал
cd web && npm run dev
# открыть http://localhost:5173, пройти полный цикл руками
```

Результат — третья запись в `docs/superpowers/verification/`, с явной пометкой, если браузерная проверка недоступна в среде выполнения.

## Трассировка требование → задача → тест

| Требование (из спеки) | Задача | Тест |
|---|---|---|
| Прокси избавляет от CORS в dev | vite.config.ts | ручная проверка (сетевой вызов проходит без CORS-ошибки) |
| `moderation_rejected` показывается под конкретным полем ввода | TacticInput.tsx | TacticInput.test.tsx |
| `low_fallback_used` — явный игровой сигнал, не скрытая деталь | IntentReview.tsx | IntentReview.test.tsx |
| Полный поток работает от выбора босса до результата | App.tsx | App.test.tsx: happy path |
| Ошибка API даёт типизированный код, не только текст | client.ts | client.test.ts |

## Статус

Не начато — план написан до кода.
