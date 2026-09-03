import type {
  Boss,
  BattleRequest,
  BattleResponse,
  ClassifyRequest,
  IntentClassification,
} from './types'

// ApiError carries the backend's own {error, message} shape, per the
// web client spec — callers branch on `code`, not on message text.
export class ApiError extends Error {
  code: string

  constructor(code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init)
  const body = await res.json()
  if (!res.ok) {
    throw new ApiError(body.error ?? 'unknown_error', body.message ?? 'request failed')
  }
  return body as T
}

export function listBosses(): Promise<Boss[]> {
  return request<Boss[]>('/api/v1/bosses')
}

export function classifyTactic(req: ClassifyRequest): Promise<IntentClassification> {
  return request<IntentClassification>('/api/v1/tactics/classify', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
}

export function runBattle(req: BattleRequest): Promise<BattleResponse> {
  return request<BattleResponse>('/api/v1/battles', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
}

export function getBattle(id: string): Promise<BattleResponse> {
  return request<BattleResponse>(`/api/v1/battles/${encodeURIComponent(id)}`)
}
