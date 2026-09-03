import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { BossSelect } from './BossSelect'
import * as client from '../api/client'
import { ApiError } from '../api/client'
import type { Boss } from '../api/types'

const bosses: Boss[] = [
  { boss_id: 'frost_warden', display_name: 'Ледяной страж', phases: [] },
  { boss_id: 'stone_giant', display_name: 'Каменный великан', phases: [] },
]

describe('BossSelect', () => {
  it('shows a loading state while fetching', () => {
    vi.spyOn(client, 'listBosses').mockReturnValue(new Promise(() => {}))
    render(<BossSelect onSelect={() => {}} />)
    expect(screen.getByText(/загрузка/i)).toBeInTheDocument()
  })

  it('shows an error state on failure', async () => {
    vi.spyOn(client, 'listBosses').mockRejectedValue(new ApiError('internal', 'что-то пошло не так'))
    render(<BossSelect onSelect={() => {}} />)
    expect(await screen.findByText(/что-то пошло не так/i)).toBeInTheDocument()
  })

  it('lists bosses and calls onSelect when one is chosen', async () => {
    vi.spyOn(client, 'listBosses').mockResolvedValue(bosses)
    const onSelect = vi.fn()
    render(<BossSelect onSelect={onSelect} />)

    const button = await screen.findByRole('button', { name: /ледяной страж/i })
    await userEvent.click(button)

    await waitFor(() => expect(onSelect).toHaveBeenCalledWith(bosses[0]))
  })
})
