package service

import "testing"

func fullTeam(mageProgram []Rule, tankProgram []Rule) []HeroDef {
	return []HeroDef{
		NewHeroDef("tank-1", HeroClassTank, tankProgram),
		NewHeroDef("archer-1", HeroClassArcher, []Rule{}),
		NewHeroDef("mage-1", HeroClassMage, mageProgram),
		NewHeroDef("healer-1", HeroClassHealer, []Rule{
			{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionHeal, Target: TargetLowestHPAlly}},
		}),
	}
}

// TestRunBattle_FrostWardenPhaseGatingWins is the property test promised
// by the boss script spec: a mage who only casts frost_bolt while the
// boss is shielded (and otherwise attacks normally) must not lose to a
// mage who always casts frost_bolt regardless of phase.
func TestRunBattle_FrostWardenPhaseGatingWins(t *testing.T) {
	naive := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionFrostBolt, Target: TargetLowestHPEnemy}},
	}
	correct := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionBossPhaseIs, PhaseID: "shielded"}, Action: Action{Type: ActionFrostBolt, Target: TargetLowestHPEnemy}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	tankDefend := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}

	naiveLog := RunBattle(FrostWardenBoss(), fullTeam(naive, tankDefend), DefaultMaxTurns)
	correctLog := RunBattle(FrostWardenBoss(), fullTeam(correct, tankDefend), DefaultMaxTurns)

	if naiveLog.Result.Outcome != OutcomeVictory {
		t.Fatalf("naive mage run: expected victory, got %v", naiveLog.Result.Outcome)
	}
	if correctLog.Result.Outcome != OutcomeVictory {
		t.Fatalf("correct mage run: expected victory, got %v", correctLog.Result.Outcome)
	}
	if correctLog.Result.TurnsTaken > naiveLog.Result.TurnsTaken {
		t.Errorf("phase-gated frost_bolt took %d turns, naive always-frost_bolt took %d turns; expected gated to not be slower",
			correctLog.Result.TurnsTaken, naiveLog.Result.TurnsTaken)
	}
}

// TestRunBattle_StoneGiantRetreatSavesTank covers the boss script spec's
// third teaching scenario: a tank that retreats once its shield drops
// below 30% takes less cumulative damage over the fight than a tank
// that never retreats.
func TestRunBattle_StoneGiantRetreatSavesTank(t *testing.T) {
	noRetreat := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	withRetreat := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionSelfResourceBelow, Resource: ResourceShield, Threshold: 0.3}, Action: Action{Type: ActionRetreat, Target: TargetSelf}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	mageAttack := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}

	withoutLog := RunBattle(StoneGiantBoss(), fullTeam(mageAttack, noRetreat), DefaultMaxTurns)
	withLog := RunBattle(StoneGiantBoss(), fullTeam(mageAttack, withRetreat), DefaultMaxTurns)

	tankDamage := func(log BattleLog) float64 {
		var total float64
		for _, turn := range log.Turns {
			for _, ev := range turn.Events {
				if ev.Actor == "boss" && ev.Target == "tank-1" {
					total += ev.Amount
				}
			}
		}
		return total
	}

	withoutDamage := tankDamage(withoutLog)
	withDamage := tankDamage(withLog)
	if withDamage >= withoutDamage {
		t.Errorf("tank with retreat rule took %v damage, without retreat took %v; expected retreat to reduce damage taken", withDamage, withoutDamage)
	}
}

// TestRunBattle_ShadowHunterTauntProtectsHealer covers the boss script
// spec's second teaching scenario end-to-end: a tank that taunts
// whenever the boss is about to target the healer keeps the healer
// alive longer than a tank that ignores boss targeting.
func TestRunBattle_ShadowHunterTauntProtectsHealer(t *testing.T) {
	ignoreTargeting := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	protectHealer := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionBossTargeting, Target: "role:healer"}, Action: Action{Type: ActionTaunt, Target: TargetSelf}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	mageAttack := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}

	unprotected := RunBattle(ShadowHunterBoss(), fullTeam(mageAttack, ignoreTargeting), DefaultMaxTurns)
	protected := RunBattle(ShadowHunterBoss(), fullTeam(mageAttack, protectHealer), DefaultMaxTurns)

	healerDamage := func(log BattleLog) float64 {
		var total float64
		for _, turn := range log.Turns {
			for _, ev := range turn.Events {
				if ev.Actor == "boss" && ev.Target == "healer-1" {
					total += ev.Amount
				}
			}
		}
		return total
	}

	if got := healerDamage(protected); got != 0 {
		t.Errorf("healer with taunt-protecting tank took %v damage, want 0", got)
	}
	if got := healerDamage(unprotected); got == 0 {
		t.Error("healer without a protecting tank took 0 damage, expected shadow_hunter to hit the healer at least once in hunting phase")
	}
}

func TestRunBattle_AbortsAtTurnLimit(t *testing.T) {
	// A healer-only "team" with heavy self-heal and a boss that always
	// misses (basic_attack does nothing meaningful against a healer
	// spamming heal on itself) is contrived on purpose: nobody can win,
	// so the engine must abort rather than loop forever.
	stalemateTeam := []HeroDef{
		NewHeroDef("healer-1", HeroClassHealer, []Rule{
			{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionHeal, Target: TargetSelf}},
		}),
	}
	log := RunBattle(StoneGiantBoss(), stalemateTeam, 5)
	if log.Result.Outcome != OutcomeAborted {
		t.Fatalf("expected aborted outcome, got %v after %d turns", log.Result.Outcome, log.Result.TurnsTaken)
	}
	if log.Result.TurnsTaken != 5 {
		t.Errorf("expected TurnsTaken = 5 (the maxTurns cap), got %d", log.Result.TurnsTaken)
	}
}

func TestRunBattle_Deterministic(t *testing.T) {
	program := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionBossPhaseIs, PhaseID: "shielded"}, Action: Action{Type: ActionFrostBolt, Target: TargetLowestHPEnemy}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}
	tankProgram := []Rule{
		{Priority: 0, Condition: Condition{Type: ConditionSelfResourceBelow, Resource: ResourceShield, Threshold: 0.3}, Action: Action{Type: ActionRetreat, Target: TargetSelf}},
		{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	}

	first := RunBattle(FrostWardenBoss(), fullTeam(program, tankProgram), DefaultMaxTurns)
	for i := 0; i < 20; i++ {
		got := RunBattle(FrostWardenBoss(), fullTeam(program, tankProgram), DefaultMaxTurns)
		if got.Result != first.Result {
			t.Fatalf("iteration %d: result = %+v, want %+v (nondeterministic)", i, got.Result, first.Result)
		}
		if len(got.Turns) != len(first.Turns) {
			t.Fatalf("iteration %d: turn count = %d, want %d (nondeterministic)", i, len(got.Turns), len(first.Turns))
		}
	}
}
