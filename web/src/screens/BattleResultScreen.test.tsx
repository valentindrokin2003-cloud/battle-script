import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { BattleResultScreen } from './BattleResultScreen'
import * as client from '../api/client'
import type { Boss, BattleResponse, IntentByHero } from '../api/types'

const boss: Boss = { boss_id: 'frost_warden', display_name: 'Ледяной страж', phases: [] }

const emptyIntent = {
  hero_class: 'tank' as const,
  schema_version: '2026-09-03.1',
  rules: [],
  fallback_action: { type: 'basic_attack', target: 'lowest_hp_enemy' },
  confidence: 'high' as const,
}
const intents: IntentByHero = { tank: emptyIntent, archer: emptyIntent, mage: emptyIntent, healer: emptyIntent }

const response: BattleResponse = {
  id: 'battle-1',
  boss_id: 'frost_warden',
  turns: [
    {
      turn_number: 1,
      events: [{ actor: 'hero:mage-1', action_type: 'basic_attack', target: 'boss', amount: 9, target_hp_after: 991 }],
    },
  ],
  result: { outcome: 'victory', turns_taken: 1, boss_id: 'frost_warden' },
}

describe('BattleResultScreen', () => {
  it('shows a loading state while the battle runs', () => {
    vi.spyOn(client, 'runBattle').mockReturnValue(new Promise(() => {}))
    render(<BattleResultScreen boss={boss} intents={intents} onPlayAgain={() => {}} />)
    expect(screen.getByText(/бой идёт/i)).toBeInTheDocument()
  })

  it('shows an error state on failure', async () => {
    vi.spyOn(client, 'runBattle').mockRejectedValue(new client.ApiError('invalid_intent', 'что-то не так'))
    render(<BattleResultScreen boss={boss} intents={intents} onPlayAgain={() => {}} />)
    expect(await screen.findByText(/что-то не так/i)).toBeInTheDocument()
  })

  it('renders the turn-by-turn log and the final outcome', async () => {
    vi.spyOn(client, 'runBattle').mockResolvedValue(response)
    render(<BattleResultScreen boss={boss} intents={intents} onPlayAgain={() => {}} />)

    expect(await screen.findByText(/victory/i)).toBeInTheDocument()
    expect(screen.getByText(/basic_attack/i)).toBeInTheDocument()
  })

  it('calls onPlayAgain when the button is clicked', async () => {
    vi.spyOn(client, 'runBattle').mockResolvedValue(response)
    const onPlayAgain = vi.fn()
    render(<BattleResultScreen boss={boss} intents={intents} onPlayAgain={onPlayAgain} />)

    const button = await screen.findByRole('button', { name: /сыграть ещё раз/i })
    await userEvent.click(button)
    expect(onPlayAgain).toHaveBeenCalled()
  })
})
