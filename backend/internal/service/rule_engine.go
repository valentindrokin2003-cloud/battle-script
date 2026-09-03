package service

import "sort"

// SelectAction evaluates program in priority order and returns the
// action of the first rule whose condition holds against ctx. If no
// rule matches (including an empty program, or every matching rule's
// target failing to resolve), it returns fallback. This is the entire
// per-hero, per-turn decision function: given the same program,
// fallback, and ctx, it always returns the same Action — the
// determinism guarantee the HLD requires for a future ranked mode.
func SelectAction(program []Rule, fallback Action, ctx BattleContext) Action {
	sorted := make([]Rule, len(program))
	copy(sorted, program)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Priority < sorted[j].Priority })

	for _, r := range sorted {
		if evaluateCondition(r.Condition, ctx) {
			return r.Action
		}
	}
	return fallback
}

func evaluateCondition(c Condition, ctx BattleContext) bool {
	switch c.Type {
	case ConditionAlways:
		return true

	case ConditionSelfResourceBelow:
		return ctx.Self.Resource(c.Resource).Fraction() < c.Threshold

	case ConditionAllyResourceBelow:
		u, ok := resolveAllyTarget(c.Target, ctx)
		if !ok {
			return false
		}
		return u.Resource(c.Resource).Fraction() < c.Threshold

	case ConditionBossResourceBelow:
		return ctx.Boss.Resource(c.Resource).Fraction() < c.Threshold

	case ConditionBossPhaseIs:
		return ctx.Boss.CurrentPhaseID == c.PhaseID

	case ConditionBossTargeting:
		id, ok := resolveTargetID(c.Target, ctx)
		if !ok {
			return false
		}
		return ctx.Boss.TargetingUnitID == id

	case ConditionAllyStatusIs:
		u, ok := resolveAllyTarget(c.Target, ctx)
		if !ok {
			return false
		}
		return u.Statuses[c.Status]

	default:
		return false
	}
}

// resolveTargetID resolves selector to a stable identifier: a unit ID
// for an ally selector, or the sentinel "boss" for any enemy selector.
// Phase 0 bosses have no minions (see the boss script spec's open
// questions), so every enemy selector resolves to the boss itself.
func resolveTargetID(selector TargetSelector, ctx BattleContext) (string, bool) {
	switch selector {
	case TargetBoss, TargetLowestHPEnemy, TargetHighestHPEnemy:
		return "boss", true
	case TargetSelf:
		return ctx.Self.ID, true
	case TargetLowestHPAlly:
		u, ok := lowestHPAlly(ctx.Allies)
		if !ok {
			return "", false
		}
		return u.ID, true
	}
	if class, isRole := ParseRoleSelector(selector); isRole {
		u, ok := allyByClass(ctx.Allies, class)
		if !ok {
			return "", false
		}
		return u.ID, true
	}
	return "", false
}

// resolveAllyTarget resolves selector to a living ally's full state.
// Enemy selectors (boss, lowest/highest_hp_enemy) never resolve here —
// conditions that need an ally's resources or status must target one.
func resolveAllyTarget(selector TargetSelector, ctx BattleContext) (UnitState, bool) {
	switch selector {
	case TargetSelf:
		return ctx.Self, true
	case TargetLowestHPAlly:
		return lowestHPAlly(ctx.Allies)
	}
	if class, isRole := ParseRoleSelector(selector); isRole {
		return allyByClass(ctx.Allies, class)
	}
	return UnitState{}, false
}

func lowestHPAlly(allies []UnitState) (UnitState, bool) {
	var best UnitState
	found := false
	for _, a := range allies {
		if !found || a.Resource(ResourceHP).Fraction() < best.Resource(ResourceHP).Fraction() {
			best = a
			found = true
		}
	}
	return best, found
}

func allyByClass(allies []UnitState, class HeroClass) (UnitState, bool) {
	for _, a := range allies {
		if a.HeroClass == class {
			return a, true
		}
	}
	return UnitState{}, false
}
