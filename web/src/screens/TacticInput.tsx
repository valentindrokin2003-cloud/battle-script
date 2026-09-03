import { useState } from 'react'
import type { Boss, TacticTexts } from '../api/types'
import { ROSTER } from '../roster'

interface Props {
  boss: Boss
  onSubmit: (texts: TacticTexts) => void
  onBack: () => void
}

export function TacticInput({ boss, onSubmit, onBack }: Props) {
  const [texts, setTexts] = useState<TacticTexts>({ tank: '', archer: '', mage: '', healer: '' })

  return (
    <div>
      <h1>{boss.display_name}</h1>
      <section>
        <h2>Фазы босса</h2>
        <ul>
          {boss.phases.map((phase) => (
            <li key={phase.phase_id}>
              <strong>{phase.phase_id}</strong> (порог входа {phase.hp_threshold_enter}): {phase.ability_pattern.join(', ')}
              {phase.provocation ? ` — «${phase.provocation}»` : null}
            </li>
          ))}
        </ul>
      </section>

      <section>
        <h2>Тактика команды</h2>
        {ROSTER.map((hero) => (
          <div key={hero.id}>
            <label htmlFor={`tactic-${hero.id}`}>{hero.label}</label>
            <textarea
              id={`tactic-${hero.id}`}
              value={texts[hero.heroClass]}
              onChange={(e) => setTexts((prev) => ({ ...prev, [hero.heroClass]: e.target.value }))}
            />
          </div>
        ))}
      </section>

      <button type="button" onClick={onBack}>
        Назад
      </button>
      <button type="button" onClick={() => onSubmit(texts)}>
        Дальше
      </button>
    </div>
  )
}
