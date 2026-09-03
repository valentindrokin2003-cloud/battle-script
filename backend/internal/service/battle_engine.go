package service

// BattleEvent is one logged action within a BattleTurn, per
// docs/superpowers/specs/2026-09-03-battle-turn-resolution-design.md.
type BattleEvent struct {
	Actor         string // "hero:<id>" or "boss"
	ActionType    string
	Target        string // hero id, "boss", "all", or "" for no target
	Amount        float64
	TargetHPAfter float64
}

// BattleTurn is one full round: every living hero acts in roster order,
// then the boss acts once.
type BattleTurn struct {
	TurnNumber int
	Events     []BattleEvent
}

type BattleOutcome string

const (
	OutcomeVictory BattleOutcome = "victory"
	OutcomeDefeat  BattleOutcome = "defeat"
	OutcomeAborted BattleOutcome = "aborted"
)

// BattleResult is the outcome of a fully-run BattleSession.
type BattleResult struct {
	Outcome    BattleOutcome
	TurnsTaken int
	BossID     string
}

// BattleLog is the complete, ordered record of a BattleSession.
type BattleLog struct {
	BossID string
	Turns  []BattleTurn
	Result BattleResult
}

// DefaultMaxTurns is the Phase 0 defensive turn cap from the turn
// resolution spec — a guard against a stuck configuration, not a
// balance number.
const DefaultMaxTurns = 30

// RunBattle deterministically simulates boss against heroDefs and
// returns the full BattleLog. Same boss, same heroDefs (including their
// TacticProgram), same result, always — this is the HLD's core honesty
// guarantee.
func RunBattle(boss Boss, heroDefs []HeroDef, maxTurns int) BattleLog {
	if maxTurns <= 0 {
		maxTurns = DefaultMaxTurns
	}
	s := newBattleState(boss, heroDefs)
	log := BattleLog{BossID: boss.BossID}

	for round := 1; round <= maxTurns; round++ {
		s.roundNumber = round
		s.beginRound()

		turn := BattleTurn{TurnNumber: round}

		for _, h := range s.heroes {
			if !h.alive() {
				continue
			}
			ctx := s.contextFor(h)
			action := SelectAction(h.Def.Program, h.Def.Fallback, ctx)
			turn.Events = append(turn.Events, s.applyHeroAction(h, action))
			if s.BossHP.Current <= 0 {
				break
			}
		}

		if s.BossHP.Current <= 0 {
			log.Turns = append(log.Turns, turn)
			log.Result = BattleResult{Outcome: OutcomeVictory, TurnsTaken: round, BossID: boss.BossID}
			return log
		}
		if s.allHeroesDead() {
			log.Turns = append(log.Turns, turn)
			log.Result = BattleResult{Outcome: OutcomeDefeat, TurnsTaken: round, BossID: boss.BossID}
			return log
		}

		turn.Events = append(turn.Events, s.applyBossAction())
		log.Turns = append(log.Turns, turn)

		if s.BossHP.Current <= 0 {
			log.Result = BattleResult{Outcome: OutcomeVictory, TurnsTaken: round, BossID: boss.BossID}
			return log
		}
		if s.allHeroesDead() {
			log.Result = BattleResult{Outcome: OutcomeDefeat, TurnsTaken: round, BossID: boss.BossID}
			return log
		}

		s.previousRoundPhaseID = s.bossPhaseID
	}

	log.Result = BattleResult{Outcome: OutcomeAborted, TurnsTaken: maxTurns, BossID: boss.BossID}
	return log
}

// beginRound resolves the boss's phase and announced target for this
// round from state as of the start of the round, before any hero acts.
// Heroes decide their actions against this announced target; only taunt
// (applied during the hero phase) can redirect the boss's actual target
// away from it later in the same round.
func (s *battleState) beginRound() {
	phase, err := s.Boss.PhaseForHPFraction(s.BossHP.Fraction())
	if err != nil {
		phase = s.currentPhase()
	}
	s.bossPhaseID = phase.PhaseID
	s.phaseTransitionThisRound = s.previousRoundPhaseID != "" && s.previousRoundPhaseID != s.bossPhaseID
	s.tauntOverrideUnitID = ""
	s.announcedTargetID = s.resolveBossTarget()
}

func (s *battleState) applyHeroAction(h *heroLiveState, action Action) BattleEvent {
	switch action.Type {
	case ActionBasicAttack, ActionPiercingShot, ActionAimedShot, ActionFrostBolt, ActionFireball:
		amount := s.heroDamageAmount(h, action.Type)
		s.dealDamageToBoss(amount)
		s.advanceFrontlineStreak(h, false)
		return BattleEvent{Actor: "hero:" + h.Def.ID, ActionType: string(action.Type), Target: "boss", Amount: amount, TargetHPAfter: s.BossHP.Current}

	case ActionHeal:
		targetID, amount := s.healAmount(h, action)
		s.advanceFrontlineStreak(h, false)
		if targetID == "" {
			return BattleEvent{Actor: "hero:" + h.Def.ID, ActionType: string(action.Type)}
		}
		s.healHero(targetID, amount)
		return BattleEvent{Actor: "hero:" + h.Def.ID, ActionType: string(action.Type), Target: targetID, Amount: amount, TargetHPAfter: s.heroByID(targetID).Resources[ResourceHP].Current}

	case ActionCleanse:
		targetID, ok := resolveTargetID(action.Target, s.contextFor(h))
		if ok {
			s.cleanseStatuses(targetID)
		}
		s.advanceFrontlineStreak(h, false)
		return BattleEvent{Actor: "hero:" + h.Def.ID, ActionType: string(action.Type), Target: targetID}

	case ActionTaunt:
		s.tauntOverrideUnitID = h.Def.ID
		s.advanceFrontlineStreak(h, false)
		return BattleEvent{Actor: "hero:" + h.Def.ID, ActionType: string(action.Type), Target: h.Def.ID}

	case ActionRetreat:
		h.Statuses["retreated"] = true
		s.advanceFrontlineStreak(h, true)
		return BattleEvent{Actor: "hero:" + h.Def.ID, ActionType: string(action.Type), Target: h.Def.ID}

	default:
		return BattleEvent{Actor: "hero:" + h.Def.ID, ActionType: string(action.Type)}
	}
}

// advanceFrontlineStreak implements the tank passive: FrontlineStreak
// increments on any round the tank acts without retreating, and resets
// the moment it retreats.
func (s *battleState) advanceFrontlineStreak(h *heroLiveState, retreated bool) {
	if h.Def.HeroClass != HeroClassTank {
		return
	}
	if retreated {
		h.FrontlineStreak = 0
		return
	}
	h.Statuses["retreated"] = false
	h.FrontlineStreak++
}

// heroDamageAmount applies the base damage table plus fixed class
// passives from the turn resolution spec.
func (s *battleState) heroDamageAmount(h *heroLiveState, action ActionType) float64 {
	var amount float64
	if action == ActionFrostBolt {
		if s.bossPhaseID == "shielded" {
			amount = 18
		} else {
			amount = 5
		}
	} else {
		amount = heroBaseDamage[h.Def.HeroClass][action]
	}

	switch h.Def.HeroClass {
	case HeroClassTank:
		amount *= 1 + 0.05*float64(h.FrontlineStreak)
	case HeroClassArcher:
		if s.BossHP.Fraction() < 0.3 {
			amount *= 1.10
		}
	case HeroClassMage:
		if s.phaseTransitionThisRound {
			amount *= 1.15
		}
	}
	return amount
}

// healAmount resolves the heal's target and applies the healer's
// triage passive: +20% when the target is the lowest-HP living ally.
func (s *battleState) healAmount(h *heroLiveState, action Action) (targetID string, amount float64) {
	target, ok := resolveAllyTarget(action.Target, s.contextFor(h))
	if !ok {
		return "", 0
	}
	amount = heroBaseDamage[HeroClassHealer][ActionHeal]
	if lowest, found := lowestHPAlly(s.livingAllies()); found && lowest.ID == target.ID {
		amount *= 1.20
	}
	return target.ID, amount
}

func (s *battleState) dealDamageToBoss(amount float64) {
	s.BossHP.Current -= amount
	if s.BossHP.Current < 0 {
		s.BossHP.Current = 0
	}
}

func (s *battleState) healHero(id string, amount float64) {
	h := s.heroByID(id)
	if h == nil {
		return
	}
	hp := h.Resources[ResourceHP]
	hp.Current += amount
	if hp.Current > hp.Max {
		hp.Current = hp.Max
	}
	h.Resources[ResourceHP] = hp
}

func (s *battleState) cleanseStatuses(id string) {
	h := s.heroByID(id)
	if h == nil {
		return
	}
	for status := range h.Statuses {
		if status == "retreated" {
			continue // retreat is a chosen tactic, not a debuff to cleanse away
		}
		delete(h.Statuses, status)
	}
}

// applyBossAction executes the next ability in the boss's current
// phase pattern against s.tauntOverrideUnitID if set, else the round's
// announced target.
func (s *battleState) applyBossAction() BattleEvent {
	phase := s.currentPhase()
	if len(phase.AbilityPattern) == 0 {
		return BattleEvent{Actor: "boss"}
	}
	ability := phase.AbilityPattern[s.bossActionsTaken%len(phase.AbilityPattern)]
	s.bossActionsTaken++

	targetID := s.announcedTargetID
	if s.tauntOverrideUnitID != "" && s.isAliveID(s.tauntOverrideUnitID) {
		targetID = s.tauntOverrideUnitID
	}

	dmg := bossPhaseDamage[s.Boss.BossID][s.bossPhaseID]

	switch ability {
	case "single_target_hit":
		s.dealDamageToHero(targetID, dmg.Single)
		h := s.heroByID(targetID)
		hpAfter := 0.0
		if h != nil {
			hpAfter = h.Resources[ResourceHP].Current
		}
		return BattleEvent{Actor: "boss", ActionType: ability, Target: targetID, Amount: dmg.Single, TargetHPAfter: hpAfter}

	case "cleave_all":
		for _, h := range s.heroes {
			if h.alive() {
				s.dealDamageToHero(h.Def.ID, dmg.Cleave)
			}
		}
		return BattleEvent{Actor: "boss", ActionType: ability, Target: "all", Amount: dmg.Cleave}

	default:
		// e.g. "shield_up": self-buff flavor, no numeric effect in Phase 0
		// (see turn resolution spec's "Не-цели").
		return BattleEvent{Actor: "boss", ActionType: ability}
	}
}

// dealDamageToHero applies amount to the hero's Shield resource first
// (if they have one) and spills any remainder onto HP — this is what
// makes self_resource_below(shield, ...) conditions like the HLD's
// worked tank example ("если щит падает ниже 30% — отступи") meaningful.
func (s *battleState) dealDamageToHero(id string, amount float64) {
	h := s.heroByID(id)
	if h == nil {
		return
	}
	remaining := amount
	if shield, ok := h.Resources[ResourceShield]; ok && shield.Max > 0 {
		absorbed := remaining
		if absorbed > shield.Current {
			absorbed = shield.Current
		}
		shield.Current -= absorbed
		h.Resources[ResourceShield] = shield
		remaining -= absorbed
	}
	if remaining <= 0 {
		return
	}
	hp := h.Resources[ResourceHP]
	hp.Current -= remaining
	if hp.Current < 0 {
		hp.Current = 0
	}
	h.Resources[ResourceHP] = hp
}

func (s *battleState) isAliveID(id string) bool {
	h := s.heroByID(id)
	return h != nil && h.alive()
}

// resolveBossTarget implements the "Таргетинг босса по ходам" section
// of the turn resolution spec: boss-specific overrides first, falling
// through to the shared default rule.
func (s *battleState) resolveBossTarget() string {
	switch s.Boss.BossID {
	case "shadow_hunter":
		if s.bossPhaseID == "hunting" && s.roundNumber%2 == 0 {
			if id, ok := s.aliveAllyIDByClass(HeroClassHealer); ok {
				return id
			}
		}
	case "stone_giant":
		if s.bossPhaseID == "crumbling" {
			if id, ok := s.aliveAllyIDByClass(HeroClassTank); ok && !s.heroByID(id).Statuses["retreated"] {
				return id
			}
		}
	}
	return s.defaultBossTarget()
}

// defaultBossTarget is the shared default: lowest-HP living hero that
// is not retreated, falling back to lowest-HP living hero overall if
// every living hero is retreated.
func (s *battleState) defaultBossTarget() string {
	best := ""
	bestFraction := 2.0 // above any real fraction
	for _, h := range s.heroes {
		if !h.alive() || h.Statuses["retreated"] {
			continue
		}
		f := h.Resources[ResourceHP].Fraction()
		if best == "" || f < bestFraction {
			best, bestFraction = h.Def.ID, f
		}
	}
	if best != "" {
		return best
	}
	for _, h := range s.heroes {
		if !h.alive() {
			continue
		}
		f := h.Resources[ResourceHP].Fraction()
		if best == "" || f < bestFraction {
			best, bestFraction = h.Def.ID, f
		}
	}
	return best
}

func (s *battleState) aliveAllyIDByClass(class HeroClass) (string, bool) {
	for _, h := range s.heroes {
		if h.alive() && h.Def.HeroClass == class {
			return h.Def.ID, true
		}
	}
	return "", false
}
