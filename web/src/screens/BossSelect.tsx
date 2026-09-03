import { useEffect, useState } from 'react'
import { listBosses } from '../api/client'
import type { Boss } from '../api/types'

interface Props {
  onSelect: (boss: Boss) => void
}

type State =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'success'; bosses: Boss[] }

export function BossSelect({ onSelect }: Props) {
  const [state, setState] = useState<State>({ status: 'loading' })

  useEffect(() => {
    let cancelled = false
    listBosses()
      .then((bosses) => {
        if (!cancelled) setState({ status: 'success', bosses })
      })
      .catch((err: unknown) => {
        if (!cancelled) setState({ status: 'error', message: err instanceof Error ? err.message : String(err) })
      })
    return () => {
      cancelled = true
    }
  }, [])

  if (state.status === 'loading') return <p>Загрузка боссов…</p>
  if (state.status === 'error') return <p role="alert">{state.message}</p>

  return (
    <div>
      <h1>Выбери босса</h1>
      <ul>
        {state.bosses.map((boss) => (
          <li key={boss.boss_id}>
            <button type="button" onClick={() => onSelect(boss)}>
              {boss.display_name}
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
