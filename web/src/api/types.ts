// Wire types mirroring the Go backend's JSON contracts exactly (same
// field names, snake_case). Kept in sync by hand — see
// docs/superpowers/specs/2026-09-03-web-client-design.md's open
// questions re: no codegen yet.

export interface BossPhase {
  phase_id: string
  hp_threshold_enter: number
  ability_pattern: string[]
  provocation: string
}

export interface Boss {
  boss_id: string
  display_name: string
  phases: BossPhase[]
}

export interface Condition {
  type: string
  target?: string
  resource?: string
  threshold?: number
  phase_id?: string
  status?: string
}

export interface Action {
  type: string
  target?: string
}

export interface Rule {
  priority: number
  condition: Condition
  action: Action
}

export type Confidence = 'high' | 'low_fallback_used'

export interface IntentClassification {
  hero_class: string
  schema_version: string
  rules: Rule[]
  fallback_action: Action
  source_prompt_submission_id?: string
  confidence: Confidence
}

export interface BattleEvent {
  actor: string
  action_type: string
  target?: string
  amount: number
  target_hp_after: number
}

export interface BattleTurn {
  turn_number: number
  events: BattleEvent[]
}

export type BattleOutcome = 'victory' | 'defeat' | 'aborted'

export interface BattleResult {
  outcome: BattleOutcome
  turns_taken: number
  boss_id: string
}

export interface BattleResponse {
  id: string
  boss_id: string
  turns: BattleTurn[]
  result: BattleResult
}

export const HERO_CLASSES = ['tank', 'archer', 'mage', 'healer'] as const
export type HeroClass = (typeof HERO_CLASSES)[number]

export type TacticTexts = Record<HeroClass, string>
export type IntentByHero = Record<HeroClass, IntentClassification>

export interface ClassifyRequest {
  hero_class: HeroClass
  boss_id: string
  prompt_text: string
}

export interface BattleHeroInput {
  id: string
  hero_class: HeroClass
  intent: IntentClassification
}

export interface BattleRequest {
  boss_id: string
  heroes: BattleHeroInput[]
}
