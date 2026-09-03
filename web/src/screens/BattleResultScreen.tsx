import { useEffect, useState } from 'react'
import { ApiError, runBattle } from '../api/client'
import type { Boss, BattleResponse, IntentByHero } from '../api/types'
import { ROSTER } from '../roster'

interface Props {
  boss: Boss
  intents: IntentByHero
  onPlayAgain: () => void
}

type State =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'done'; battle: BattleResponse }

export function BattleResultScreen({ boss, intents, onPlayAgain }: Props) {
  const [state, setState] = useState<State>({ status: 'loading' })

  useEffect(() => {
    let cancelled = false
    runBattle({
      boss_id: boss.boss_id,
      heroes: ROSTER.map((hero) => ({ id: hero.id, hero_class: hero.heroClass, intent: intents[hero.heroClass] })),
    })
      .then((battle) => {
        if (!cancelled) setState({ status: 'done', battle })
      })
      .catch((err: unknown) => {
        const message = err instanceof ApiError ? err.message : err instanceof Error ? err.message : String(err)
        if (!cancelled) setState({ status: 'error', message })
      })
    return () => {
      cancelled = true
    }
    // boss/intents are fixed for the lifetime of this screen.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (state.status === 'loading') return <p>Бой идёт…</p>
  if (state.status === 'error') return <p role="alert">{state.message}</p>

  const { battle } = state
  return (
    <div>
      <h1>Итог: {battle.result.outcome}</h1>
      <p>Ходов: {battle.result.turns_taken}</p>
      {battle.turns.map((turn) => (
        <section key={turn.turn_number}>
          <h2>Ход {turn.turn_number}</h2>
          <ul>
            {turn.events.map((event, i) => (
              <li key={i}>
                {event.actor} → {event.action_type}
                {event.target ? ` (${event.target})` : ''}: {event.amount}
              </li>
            ))}
          </ul>
        </section>
      ))}
      <button type="button" onClick={onPlayAgain}>
        Сыграть ещё раз
      </button>
    </div>
  )
}
