import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { IntentReview } from './IntentReview'
import * as client from '../api/client'
import type { Boss, IntentClassification, TacticTexts } from '../api/types'
import type { HeroClass } from '../api/types'

const boss: Boss = { boss_id: 'frost_warden', display_name: 'Ледяной страж', phases: [] }
const texts: TacticTexts = { tank: 'провоцируй', archer: 'атакуй', mage: 'лёд в фазе щита', healer: 'лечи слабого' }

function intentFor(heroClass: HeroClass, confidence: 'high' | 'low_fallback_used' = 'high'): IntentClassification {
  return {
    hero_class: heroClass,
    schema_version: '2026-09-03.1',
    rules: [{ priority: 0, condition: { type: 'always' }, action: { type: 'basic_attack', target: 'lowest_hp_enemy' } }],
    fallback_action: { type: 'basic_attack', target: 'lowest_hp_enemy' },
    confidence,
  }
}

describe('IntentReview', () => {
  it('classifies every hero and shows the recognized rules', async () => {
    vi.spyOn(client, 'classifyTactic').mockImplementation((req) => Promise.resolve(intentFor(req.hero_class)))
    render(<IntentReview boss={boss} texts={texts} onConfirm={() => {}} onBack={() => {}} />)

    await waitFor(() => expect(screen.getAllByText(/basic_attack/i)).toHaveLength(4))
  })

  it('flags low_fallback_used classifications explicitly', async () => {
    vi.spyOn(client, 'classifyTactic').mockImplementation((req) =>
      Promise.resolve(intentFor(req.hero_class, req.hero_class === 'mage' ? 'low_fallback_used' : 'high')),
    )
    render(<IntentReview boss={boss} texts={texts} onConfirm={() => {}} onBack={() => {}} />)

    expect(await screen.findByText(/не поняли однозначно/i)).toBeInTheDocument()
  })

  it('shows a moderation error for the specific hero that was rejected', async () => {
    vi.spyOn(client, 'classifyTactic').mockImplementation((req) => {
      if (req.hero_class === 'archer') {
        return Promise.reject(new client.ApiError('moderation_rejected', 'напиши тактику словами'))
      }
      return Promise.resolve(intentFor(req.hero_class))
    })
    render(<IntentReview boss={boss} texts={texts} onConfirm={() => {}} onBack={() => {}} />)

    expect(await screen.findByText(/напиши тактику словами/i)).toBeInTheDocument()
  })

  it('confirms with all classifications once every hero is classified', async () => {
    vi.spyOn(client, 'classifyTactic').mockImplementation((req) => Promise.resolve(intentFor(req.hero_class)))
    const onConfirm = vi.fn()
    render(<IntentReview boss={boss} texts={texts} onConfirm={onConfirm} onBack={() => {}} />)

    const button = await screen.findByRole('button', { name: /начать бой/i })
    await userEvent.click(button)

    expect(onConfirm).toHaveBeenCalledWith({
      tank: intentFor('tank'),
      archer: intentFor('archer'),
      mage: intentFor('mage'),
      healer: intentFor('healer'),
    })
  })
})
