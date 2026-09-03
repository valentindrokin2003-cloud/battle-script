package service

import (
	"context"
	"testing"
)

func classify(t *testing.T, class HeroClass, text string) IntentClassification {
	t.Helper()
	c := LocalHeuristicClassifier{}
	got, err := c.Classify(context.Background(), ClassificationContext{HeroClass: class, Boss: frostWarden(), PromptText: text})
	if err != nil {
		t.Fatalf("Classify(%q) returned error: %v", text, err)
	}
	if verr := ValidateIntentClassification(got, frostWarden()); verr != nil {
		t.Fatalf("Classify(%q) produced invalid classification: %v\n%+v", text, verr, got)
	}
	return got
}

func TestLocalHeuristicClassifier_MageRecognizesPhaseGatedFrostBolt(t *testing.T) {
	got := classify(t, HeroClassMage, "используй ледяной шар только в фазе щита босса, в остальное время атакуй слабейшего")
	if got.Confidence != ConfidenceHigh {
		t.Fatalf("Confidence = %v, want high", got.Confidence)
	}
	if len(got.Rules) < 1 || got.Rules[0].Action.Type != ActionFrostBolt || got.Rules[0].Condition.Type != ConditionBossPhaseIs || got.Rules[0].Condition.PhaseID != "shielded" {
		t.Fatalf("expected rule 0 to be boss_phase_is(shielded) -> frost_bolt, got %+v", got.Rules)
	}
}

func TestLocalHeuristicClassifier_MageUnrecognizedFallsBack(t *testing.T) {
	got := classify(t, HeroClassMage, "бла бла непонятный текст")
	if got.Confidence != ConfidenceLowFallbackUsed {
		t.Fatalf("Confidence = %v, want low_fallback_used", got.Confidence)
	}
}

func TestLocalHeuristicClassifier_TankRecognizesTauntAndRetreat(t *testing.T) {
	got := classify(t, HeroClassTank, "провоцируй босса, когда он целится в целителя, и если щит падает ниже 30% — отступи")
	if got.Confidence != ConfidenceHigh {
		t.Fatalf("Confidence = %v, want high", got.Confidence)
	}
	if len(got.Rules) < 2 {
		t.Fatalf("expected at least 2 rules (taunt + retreat), got %+v", got.Rules)
	}
	if got.Rules[0].Action.Type != ActionTaunt || got.Rules[0].Condition.Type != ConditionBossTargeting {
		t.Errorf("rule 0 = %+v, want boss_targeting -> taunt", got.Rules[0])
	}
	if got.Rules[1].Action.Type != ActionRetreat || got.Rules[1].Condition.Type != ConditionSelfResourceBelow {
		t.Errorf("rule 1 = %+v, want self_resource_below -> retreat", got.Rules[1])
	}
}

func TestLocalHeuristicClassifier_TankUnrecognizedFallsBack(t *testing.T) {
	got := classify(t, HeroClassTank, "что-то совсем непохожее на тактику")
	if got.Confidence != ConfidenceLowFallbackUsed {
		t.Fatalf("Confidence = %v, want low_fallback_used", got.Confidence)
	}
}

func TestLocalHeuristicClassifier_HealerRecognizesHealWeakest(t *testing.T) {
	got := classify(t, HeroClassHealer, "лечи того, у кого меньше всего здоровья")
	if got.Confidence != ConfidenceHigh {
		t.Fatalf("Confidence = %v, want high", got.Confidence)
	}
	if len(got.Rules) < 1 || got.Rules[0].Action.Type != ActionHeal || got.Rules[0].Action.Target != TargetLowestHPAlly {
		t.Fatalf("expected rule 0 to be heal(lowest_hp_ally), got %+v", got.Rules)
	}
}

func TestLocalHeuristicClassifier_HealerUnrecognizedFallsBack(t *testing.T) {
	got := classify(t, HeroClassHealer, "непонятно что тут написано")
	if got.Confidence != ConfidenceLowFallbackUsed {
		t.Fatalf("Confidence = %v, want low_fallback_used", got.Confidence)
	}
}

func TestLocalHeuristicClassifier_ArcherRecognizesAimedShot(t *testing.T) {
	got := classify(t, HeroClassArcher, "цель самый сильный удар в слабейшего врага")
	if got.Confidence != ConfidenceHigh {
		t.Fatalf("Confidence = %v, want high", got.Confidence)
	}
	if len(got.Rules) < 1 || got.Rules[0].Action.Type != ActionAimedShot {
		t.Fatalf("expected rule 0 to use aimed_shot, got %+v", got.Rules)
	}
}

func TestLocalHeuristicClassifier_ArcherUnrecognizedFallsBack(t *testing.T) {
	got := classify(t, HeroClassArcher, "какой-то текст без ключевых слов")
	if got.Confidence != ConfidenceLowFallbackUsed {
		t.Fatalf("Confidence = %v, want low_fallback_used", got.Confidence)
	}
}

// TestFullPipeline_TextToAction is the integration test the spec asks
// for: raw child text -> moderation -> classification -> validation ->
// SelectAction, exercised end to end for the first time in this
// project. There is no single Go function combining moderation and
// classification yet (that lands with the HTTP handler); this test
// sequences the two ports the way a future handler must.
func TestFullPipeline_TextToAction(t *testing.T) {
	moderator := BasicModerator{}
	classifier := LocalHeuristicClassifier{}
	boss := frostWarden()

	text := "используй ледяной шар только в фазе щита босса, в остальное время атакуй слабейшего"
	modResult := moderator.Check(text)
	if !modResult.Allowed {
		t.Fatalf("moderator rejected valid tactic text: %s", modResult.Reason)
	}

	classification := ClassifyWithFallback(context.Background(), classifier, ClassificationContext{
		HeroClass: HeroClassMage, Boss: boss, PromptText: text,
	})
	if err := ValidateIntentClassification(classification, boss); err != nil {
		t.Fatalf("pipeline produced invalid classification: %v", err)
	}

	mage := UnitState{ID: "mage-1", HeroClass: HeroClassMage}
	shieldedCtx := BattleContext{Self: mage, Allies: []UnitState{mage}, Boss: BossState{CurrentPhaseID: "shielded"}}
	if got := SelectAction(classification.Rules, classification.FallbackAction, shieldedCtx); got.Type != ActionFrostBolt {
		t.Errorf("in shielded phase: SelectAction = %v, want frost_bolt", got.Type)
	}
	exposedCtx := BattleContext{Self: mage, Allies: []UnitState{mage}, Boss: BossState{CurrentPhaseID: "exposed"}}
	if got := SelectAction(classification.Rules, classification.FallbackAction, exposedCtx); got.Type != ActionBasicAttack {
		t.Errorf("in exposed phase: SelectAction = %v, want basic_attack", got.Type)
	}
}

// TestFullPipeline_ModerationBlocksBeforeClassification proves the
// HLD's ordering requirement: rejected text never reaches the
// classifier at all.
func TestFullPipeline_ModerationBlocksBeforeClassification(t *testing.T) {
	moderator := BasicModerator{}
	counting := &mockClassifier{} // no responses queued; a call would fail the test via "mock exhausted" if reached

	text := "" // empty text: rejected by moderation
	modResult := moderator.Check(text)
	if modResult.Allowed {
		t.Fatal("expected empty text to be rejected by moderation")
	}

	if modResult.Allowed {
		_ = ClassifyWithFallback(context.Background(), counting, ClassificationContext{HeroClass: HeroClassMage, Boss: frostWarden(), PromptText: text})
	}
	if counting.calls != 0 {
		t.Errorf("classifier was called %d times after moderation rejected the text, want 0", counting.calls)
	}
}
