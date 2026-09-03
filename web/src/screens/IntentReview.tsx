import { useEffect, useState } from 'react'
import { ApiError, classifyTactic } from '../api/client'
import type { Boss, HeroClass, IntentByHero, IntentClassification, TacticTexts } from '../api/types'
import { ROSTER } from '../roster'

interface Props {
  boss: Boss
  texts: TacticTexts
  onConfirm: (intents: IntentByHero) => void
  onBack: () => void
}

type HeroState =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'done'; intent: IntentClassification }

function describeRule(rule: IntentClassification['rules'][number]) {
  return `${rule.condition.type} -> ${rule.action.type}`
}

export function IntentReview({ boss, texts, onConfirm, onBack }: Props) {
  const [states, setStates] = useState<Record<HeroClass, HeroState>>({
    tank: { status: 'loading' },
    archer: { status: 'loading' },
    mage: { status: 'loading' },
    healer: { status: 'loading' },
  })

  useEffect(() => {
    let cancelled = false
    for (const hero of ROSTER) {
      classifyTactic({ hero_class: hero.heroClass, boss_id: boss.boss_id, prompt_text: texts[hero.heroClass] })
        .then((intent) => {
          if (!cancelled) setStates((prev) => ({ ...prev, [hero.heroClass]: { status: 'done', intent } }))
        })
        .catch((err: unknown) => {
          const message = err instanceof ApiError ? err.message : err instanceof Error ? err.message : String(err)
          if (!cancelled) setStates((prev) => ({ ...prev, [hero.heroClass]: { status: 'error', message } }))
        })
    }
    return () => {
      cancelled = true
    }
    // texts/boss are fixed for the lifetime of this screen (a new one is
    // mounted per attempt from App's state machine).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const allDone = ROSTER.every((hero) => states[hero.heroClass].status === 'done')

  const confirm = () => {
    const intents = {} as IntentByHero
    for (const hero of ROSTER) {
      const state = states[hero.heroClass]
      if (state.status === 'done') intents[hero.heroClass] = state.intent
    }
    onConfirm(intents)
  }

  return (
    <div>
      <h1>Распознанные намерения</h1>
      {ROSTER.map((hero) => {
        const state = states[hero.heroClass]
        return (
          <section key={hero.id}>
            <h2>{hero.label}</h2>
            {state.status === 'loading' && <p>Распознаём тактику…</p>}
            {state.status === 'error' && <p role="alert">{state.message}</p>}
            {state.status === 'done' && (
              <>
                <ul>
                  {state.intent.rules.map((rule, i) => (
                    <li key={i}>{describeRule(rule)}</li>
                  ))}
                </ul>
                {state.intent.confidence === 'low_fallback_used' && (
                  <p role="status">Не поняли однозначно — используем запасное действие: {state.intent.fallback_action.type}</p>
                )}
              </>
            )}
          </section>
        )
      })}
      <button type="button" onClick={onBack}>
        Назад
      </button>
      <button type="button" disabled={!allDone} onClick={confirm}>
        Начать бой
      </button>
    </div>
  )
}
