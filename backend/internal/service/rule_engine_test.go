package service

import "testing"

// TestSelectAction_FrostWardenPhaseGating covers the boss script spec's
// first teaching scenario: a mage's frost_bolt should only fire while
// the boss is in its "shielded" phase, otherwise fall back to a basic
// attack. This tests rule selection, not damage output — BattleEngine's
// damage resolution is a separate, not-yet-built concern.
func TestSelectAction_FrostWardenPhaseGating(t *testing.T) {
	program := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionBossPhaseIs, PhaseID: "shielded"}, Action: Action{Type: ActionFrostBolt, Target: TargetLowestHPEnemy}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	fallback := DefaultFallback(HeroClassMage)
	mage := UnitState{ID: "mage-1", HeroClass: HeroClassMage, Resources: map[ResourceType]ResourceValue{ResourceHP: {100, 100}}}

	for _, tc := range []struct {
		phase    string
		wantType ActionType
	}{
		{"exposed", ActionBasicAttack},
		{"shielded", ActionFrostBolt},
		{"enraged", ActionBasicAttack},
	} {
		ctx := BattleContext{
			Self:   mage,
			Allies: []UnitState{mage},
			Boss:   BossState{CurrentPhaseID: tc.phase},
		}
		got := SelectAction(program, fallback, ctx)
		if got.Type != tc.wantType {
			t.Errorf("phase %q: SelectAction = %v, want %v", tc.phase, got.Type, tc.wantType)
		}
	}
}

// TestSelectAction_ShadowHunterTauntsForHealer covers the boss script
// spec's second teaching scenario: a tank should taunt when the boss is
// about to target the healer.
func TestSelectAction_ShadowHunterTauntsForHealer(t *testing.T) {
	program := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionBossTargeting, Target: "role:healer"}, Action: Action{Type: ActionTaunt, Target: TargetSelf}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	fallback := DefaultFallback(HeroClassTank)
	tank := UnitState{ID: "tank-1", HeroClass: HeroClassTank}
	healer := UnitState{ID: "healer-1", HeroClass: HeroClassHealer}
	roster := []UnitState{tank, healer}

	targetingHealer := BattleContext{Self: tank, Allies: roster, Boss: BossState{TargetingUnitID: "healer-1"}}
	if got := SelectAction(program, fallback, targetingHealer); got.Type != ActionTaunt {
		t.Errorf("boss targeting healer: SelectAction = %v, want taunt", got.Type)
	}

	targetingTank := BattleContext{Self: tank, Allies: roster, Boss: BossState{TargetingUnitID: "tank-1"}}
	if got := SelectAction(program, fallback, targetingTank); got.Type != ActionBasicAttack {
		t.Errorf("boss targeting tank: SelectAction = %v, want basic_attack", got.Type)
	}
}

// TestSelectAction_UnresolvedSelectorSkipsToNextRule covers the spec's
// "Словарь селекторов" rule: a role:healer condition with no healer in
// the roster must not error — it is simply inapplicable, so evaluation
// falls through to the next rule.
func TestSelectAction_UnresolvedSelectorSkipsToNextRule(t *testing.T) {
	program := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionBossTargeting, Target: "role:healer"}, Action: Action{Type: ActionTaunt, Target: TargetSelf}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	tank := UnitState{ID: "tank-1", HeroClass: HeroClassTank}
	ctx := BattleContext{Self: tank, Allies: []UnitState{tank}, Boss: BossState{TargetingUnitID: "tank-1"}}

	got := SelectAction(program, DefaultFallback(HeroClassTank), ctx)
	if got.Type != ActionBasicAttack {
		t.Errorf("SelectAction with no healer in roster = %v, want basic_attack (rule 0 inapplicable)", got.Type)
	}
}

// TestSelectAction_StoneGiantRetreatThreshold covers the boss script
// spec's third teaching scenario: a tank should retreat once its shield
// drops below 30%.
func TestSelectAction_StoneGiantRetreatThreshold(t *testing.T) {
	program := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionSelfResourceBelow, Resource: ResourceShield, Threshold: 0.3}, Action: Action{Type: ActionRetreat, Target: TargetSelf}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	fallback := DefaultFallback(HeroClassTank)

	for _, tc := range []struct {
		name       string
		shieldCur  float64
		wantAction ActionType
	}{
		{"shield healthy", 50, ActionBasicAttack},
		{"shield exactly at threshold", 30, ActionBasicAttack},
		{"shield below threshold", 20, ActionRetreat},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tank := UnitState{ID: "tank-1", HeroClass: HeroClassTank, Resources: map[ResourceType]ResourceValue{
				ResourceShield: {Current: tc.shieldCur, Max: 100},
			}}
			ctx := BattleContext{Self: tank, Allies: []UnitState{tank}}
			got := SelectAction(program, fallback, ctx)
			if got.Type != tc.wantAction {
				t.Errorf("shield=%v: SelectAction = %v, want %v", tc.shieldCur, got.Type, tc.wantAction)
			}
		})
	}
}

// TestSelectAction_EmptyProgramUsesFallback covers the HLD's error
// handling: a fully-failed classification (empty rules) must still
// produce a defined action via the class-level fallback.
func TestSelectAction_EmptyProgramUsesFallback(t *testing.T) {
	fallback := DefaultFallback(HeroClassHealer)
	ctx := BattleContext{Self: UnitState{ID: "h-1", HeroClass: HeroClassHealer}}
	got := SelectAction(nil, fallback, ctx)
	if got != fallback {
		t.Errorf("SelectAction(nil program) = %v, want fallback %v", got, fallback)
	}
}

// TestSelectAction_Deterministic guards the HLD's core honesty
// requirement: identical program, fallback, and context must always
// produce the identical action.
func TestSelectAction_Deterministic(t *testing.T) {
	program := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionBossPhaseIs, PhaseID: "shielded"}, Action: Action{Type: ActionFrostBolt, Target: TargetLowestHPEnemy}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	fallback := DefaultFallback(HeroClassMage)
	ctx := BattleContext{
		Self:   UnitState{ID: "mage-1", HeroClass: HeroClassMage},
		Allies: []UnitState{{ID: "mage-1", HeroClass: HeroClassMage}},
		Boss:   BossState{CurrentPhaseID: "shielded"},
	}
	first := SelectAction(program, fallback, ctx)
	for i := 0; i < 100; i++ {
		if got := SelectAction(program, fallback, ctx); got != first {
			t.Fatalf("iteration %d: SelectAction = %v, want %v (nondeterministic)", i, got, first)
		}
	}
}
