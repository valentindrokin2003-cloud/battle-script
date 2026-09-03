# План: разрешение боевого хода

Реализует [`2026-09-03-battle-turn-resolution-design.md`](../specs/2026-09-03-battle-turn-resolution-design.md).

## Задачи

| # | Файл | Что делает | TDD |
|---|---|---|---|
| 1 | `backend/internal/service/phase0_content.go` | Иллюстративные сид-данные: `HeroBaseResources`, 3 функции-конструктора боссов (`FrostWardenBoss`/`ShadowHunterBoss`/`StoneGiantBoss`), таблицы урона `heroBaseDamage`/`bossPhaseDamage` | Данные, не логика — покрываются косвенно через тесты п.4 |
| 2 | `backend/internal/service/battle_state.go` | `HeroDef`/`heroLiveState`/`battleState` — мутируемое состояние одного `BattleSession` | Косвенно, через тесты п.4 |
| 3 | `backend/internal/service/battle_engine.go` | `RunBattle` — цикл ходов, разрешение урона/лечения/щита, таргетинг босса по правилам из спеки, `BattleLog`/`BattleResult` | См. ниже — этот шаг единственный в проекте пока прошёл через настоящий red-green цикл |
| 4 | `backend/internal/service/battle_engine_test.go` | Сценарные + property-тесты, тест лимита ходов, тест детерминизма | — |

## TDD-заметка (единственный настоящий red-green в проекте на сейчас)

`TestRunBattle_StoneGiantRetreatSavesTank` был написан по спеке **до** того, как я заметил, что `dealDamageToHero` не проводит урон через щит — тест упал (`318 == 318`, отступление не давало эффекта), что и вскрыло пропуск. Правка (щит поглощает урон до HP) внесена в `battle_engine.go` и отражена в самой спеке (раздел «Базовые ресурсы героев»), а не обойдена подгонкой теста. Остальные файлы этого проекта писались в связке код+тест одновременно, без отдельной red-фазы — отклонение от `superpowers:test-driven-development`, фиксирую явно, а не скрываю.

## Валидация

```bash
cd backend && make check   # gofmt-check, go vet, golangci-lint, go test ./... -race
```

## Трассировка требование → задача → тест

| Требование (из спеки) | Задача | Тест |
|---|---|---|
| Герои действуют раньше босса каждый ход | battle_engine.go: `RunBattle` порядок цикла | Косвенно все `TestRunBattle_*` (босс не мог бы получать урон до атаки героя иначе) |
| `frost_bolt` выгоден только в фазе `shielded` | battle_engine.go: `heroDamageAmount` | `TestRunBattle_FrostWardenPhaseGatingWins` — property-тест, обещанный спекой боссов |
| `stone_giant` целится в `role:tank`, отступление уводит от удара | battle_engine.go: `resolveBossTarget` | `TestRunBattle_StoneGiantRetreatSavesTank` |
| `shadow_hunter` целится в целителя по чётным ходам фазы `hunting`; `taunt` перекрывает это | battle_engine.go: `resolveBossTarget`, `tauntOverrideUnitID` | `TestRunBattle_ShadowHunterTauntProtectsHealer` |
| Щит поглощает урон раньше HP | battle_engine.go: `dealDamageToHero` | `TestRunBattle_StoneGiantRetreatSavesTank` (обнаружил пропуск, см. TDD-заметку) |
| Аварийное завершение по лимиту ходов, не бесконечный цикл | battle_engine.go: `RunBattle` цикл с `maxTurns` | `TestRunBattle_AbortsAtTurnLimit` |
| Детерминизм всего боя, не только одного хода | battle_engine.go: `RunBattle` (чистая функция от `boss`+`heroDefs`) | `TestRunBattle_Deterministic` |

## Статус

Выполнено, закоммичено в `6ace6be`. План написан задним числом; со следующей спеки план пишется до кода.
