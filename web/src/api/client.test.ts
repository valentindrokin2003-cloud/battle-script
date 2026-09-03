import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, classifyTactic, getBattle, listBosses, runBattle } from './client'
import type { Boss, BattleResponse, IntentClassification } from './types'

function mockFetchOnce(status: number, body: unknown) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: status >= 200 && status < 300,
      status,
      json: () => Promise.resolve(body),
    }),
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('listBosses', () => {
  it('returns parsed bosses on success', async () => {
    const bosses: Boss[] = [{ boss_id: 'frost_warden', display_name: 'Ледяной страж', phases: [] }]
    mockFetchOnce(200, bosses)

    const got = await listBosses()

    expect(got).toEqual(bosses)
    expect(fetch).toHaveBeenCalledWith('/api/v1/bosses', undefined)
  })

  it('throws ApiError with backend fields on non-2xx', async () => {
    mockFetchOnce(500, { error: 'internal', message: 'boom' })

    await expect(listBosses()).rejects.toMatchObject(new ApiError('internal', 'boom'))
  })
})

describe('classifyTactic', () => {
  it('posts the request and returns the classification', async () => {
    const intent: IntentClassification = {
      hero_class: 'mage',
      schema_version: '2026-09-03.1',
      rules: [],
      fallback_action: { type: 'basic_attack', target: 'lowest_hp_enemy' },
      confidence: 'high',
    }
    mockFetchOnce(200, intent)

    const got = await classifyTactic({ hero_class: 'mage', boss_id: 'frost_warden', prompt_text: 'text' })

    expect(got).toEqual(intent)
  })

  it('surfaces moderation_rejected as an ApiError', async () => {
    mockFetchOnce(422, { error: 'moderation_rejected', message: 'напиши тактику словами' })

    await expect(classifyTactic({ hero_class: 'mage', boss_id: 'frost_warden', prompt_text: '' })).rejects.toMatchObject(
      new ApiError('moderation_rejected', 'напиши тактику словами'),
    )
  })
})

describe('runBattle', () => {
  it('returns the battle response', async () => {
    const response: BattleResponse = {
      id: 'abc',
      boss_id: 'frost_warden',
      turns: [],
      result: { outcome: 'victory', turns_taken: 3, boss_id: 'frost_warden' },
    }
    mockFetchOnce(200, response)

    const got = await runBattle({ boss_id: 'frost_warden', heroes: [] })

    expect(got).toEqual(response)
  })
})

describe('getBattle', () => {
  it('fetches by id', async () => {
    const response: BattleResponse = {
      id: 'abc',
      boss_id: 'frost_warden',
      turns: [],
      result: { outcome: 'victory', turns_taken: 3, boss_id: 'frost_warden' },
    }
    mockFetchOnce(200, response)

    const got = await getBattle('abc')

    expect(got).toEqual(response)
    expect(fetch).toHaveBeenCalledWith('/api/v1/battles/abc', undefined)
  })

  it('throws battle_not_found as ApiError', async () => {
    mockFetchOnce(404, { error: 'battle_not_found', message: 'no such battle id' })

    await expect(getBattle('missing')).rejects.toMatchObject(new ApiError('battle_not_found', 'no such battle id'))
  })
})
