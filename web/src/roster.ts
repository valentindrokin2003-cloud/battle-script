import type { HeroClass } from './api/types'

// Fixed Phase 0 roster — no hero selection, per the web client spec's
// "Не-цели".
export interface RosterHero {
  id: string
  heroClass: HeroClass
  label: string
}

export const ROSTER: RosterHero[] = [
  { id: 'tank-1', heroClass: 'tank', label: 'Танк' },
  { id: 'archer-1', heroClass: 'archer', label: 'Лучник' },
  { id: 'mage-1', heroClass: 'mage', label: 'Маг' },
  { id: 'healer-1', heroClass: 'healer', label: 'Целитель' },
]
