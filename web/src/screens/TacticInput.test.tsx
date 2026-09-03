import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { TacticInput } from './TacticInput'
import type { Boss } from '../api/types'

const boss: Boss = {
  boss_id: 'frost_warden',
  display_name: 'Ледяной страж',
  phases: [
    { phase_id: 'exposed', hp_threshold_enter: 1.0, ability_pattern: ['cleave_all', 'single_target_hit'], provocation: 'Подходите!' },
    { phase_id: 'shielded', hp_threshold_enter: 0.6, ability_pattern: ['shield_up'], provocation: 'Щит вас не пропустит!' },
  ],
}

describe('TacticInput', () => {
  it('reveals every phase of the chosen boss (fairness rule)', () => {
    render(<TacticInput boss={boss} onSubmit={() => {}} onBack={() => {}} />)
    expect(screen.getByText(/exposed/i)).toBeInTheDocument()
    expect(screen.getByText(/shielded/i)).toBeInTheDocument()
    expect(screen.getByText(/подходите/i)).toBeInTheDocument()
  })

  it('renders one text input per roster hero', () => {
    render(<TacticInput boss={boss} onSubmit={() => {}} onBack={() => {}} />)
    expect(screen.getByLabelText(/танк/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/лучник/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/маг/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/целитель/i)).toBeInTheDocument()
  })

  it('submits the entered text for every hero', async () => {
    const onSubmit = vi.fn()
    render(<TacticInput boss={boss} onSubmit={onSubmit} onBack={() => {}} />)

    await userEvent.type(screen.getByLabelText(/танк/i), 'провоцируй босса')
    await userEvent.type(screen.getByLabelText(/маг/i), 'атакуй в фазе щита')
    await userEvent.click(screen.getByRole('button', { name: /дальше/i }))

    expect(onSubmit).toHaveBeenCalledWith({
      tank: 'провоцируй босса',
      archer: '',
      mage: 'атакуй в фазе щита',
      healer: '',
    })
  })

  it('calls onBack when the back button is clicked', async () => {
    const onBack = vi.fn()
    render(<TacticInput boss={boss} onSubmit={() => {}} onBack={onBack} />)
    await userEvent.click(screen.getByRole('button', { name: /назад/i }))
    expect(onBack).toHaveBeenCalled()
  })
})
