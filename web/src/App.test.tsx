import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { App } from './App'
import * as client from './api/client'
import type { Boss, BattleResponse, IntentClassification } from './api/types'

const boss: Boss = {
  boss_id: 'frost_warden',
  display_name: 'Ледяной страж',
  phases: [{ phase_id: 'exposed', hp_threshold_enter: 1.0, ability_pattern: ['cleave_all'], provocation: '' }],
}

function intentFor(): IntentClassification {
  return {
    hero_class: 'tank',
    schema_version: '2026-09-03.1',
    rules: [{ priority: 0, condition: { type: 'always' }, action: { type: 'basic_attack', target: 'lowest_hp_enemy' } }],
    fallback_action: { type: 'basic_attack', target: 'lowest_hp_enemy' },
    confidence: 'high',
  }
}

const battleResponse: BattleResponse = {
  id: 'battle-1',
  boss_id: 'frost_warden',
  turns: [{ turn_number: 1, events: [] }],
  result: { outcome: 'victory', turns_taken: 1, boss_id: 'frost_warden' },
}

describe('App', () => {
  it('walks the full boss-select -> tactic-input -> intent-review -> battle-result flow', async () => {
    vi.spyOn(client, 'listBosses').mockResolvedValue([boss])
    vi.spyOn(client, 'classifyTactic').mockImplementation(() => Promise.resolve(intentFor()))
    vi.spyOn(client, 'runBattle').mockResolvedValue(battleResponse)

    render(<App />)

    const bossButton = await screen.findByRole('button', { name: /ледяной страж/i })
    await userEvent.click(bossButton)

    await screen.findByLabelText(/танк/i)
    await userEvent.click(screen.getByRole('button', { name: /дальше/i }))

    const startButton = await screen.findByRole('button', { name: /начать бой/i })
    await userEvent.click(startButton)

    expect(await screen.findByText(/victory/i)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /сыграть ещё раз/i }))
    expect(await screen.findByRole('button', { name: /ледяной страж/i })).toBeInTheDocument()
  })
})
