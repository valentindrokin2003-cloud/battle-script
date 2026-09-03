package service

import "testing"

func frostWarden() Boss {
	return Boss{
		BossID:      "frost_warden",
		DisplayName: "Ледяной страж",
		MaxHP:       1000,
		Phases: []BossPhase{
			{PhaseID: "exposed", HPThresholdEnter: 1.0, AbilityPattern: []string{"cleave_all", "single_target_hit"}},
			{PhaseID: "shielded", HPThresholdEnter: 0.6, AbilityPattern: []string{"shield_up", "single_target_hit"}},
			{PhaseID: "enraged", HPThresholdEnter: 0.25, AbilityPattern: []string{"cleave_all", "cleave_all", "single_target_hit"}},
		},
	}
}

// validMageClassification mirrors the worked example from the action
// schema spec's "Схема IntentClassification (JSON)" section.
func validMageClassification() IntentClassification {
	return IntentClassification{
		HeroClass:     HeroClassMage,
		SchemaVersion: "2026-09-03.1",
		Rules: []Rule{
			{
				Priority:  0,
				Condition: Condition{Type: ConditionBossPhaseIs, PhaseID: "shielded"},
				Action:    Action{Type: ActionFrostBolt, Target: TargetLowestHPEnemy},
			},
			{
				Priority:  1,
				Condition: Condition{Type: ConditionAlways},
				Action:    Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy},
			},
		},
		FallbackAction:           Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy},
		SourcePromptSubmissionID: "11111111-1111-1111-1111-111111111111",
		Confidence:               ConfidenceHigh,
	}
}

func TestValidateIntentClassification_Valid(t *testing.T) {
	if err := ValidateIntentClassification(validMageClassification(), frostWarden()); err != nil {
		t.Fatalf("expected valid classification, got error: %v", err)
	}
}

func TestValidateIntentClassification_Rejections(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*IntentClassification)
	}{
		{
			name:   "unknown hero class",
			mutate: func(ic *IntentClassification) { ic.HeroClass = "paladin" },
		},
		{
			name:   "missing schema version",
			mutate: func(ic *IntentClassification) { ic.SchemaVersion = "" },
		},
		{
			name: "too many rules",
			mutate: func(ic *IntentClassification) {
				ic.Rules = append(ic.Rules,
					Rule{Priority: 2, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
					Rule{Priority: 3, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
				)
			},
		},
		{
			name:   "priority out of order",
			mutate: func(ic *IntentClassification) { ic.Rules[1].Priority = 5 },
		},
		{
			name: "action not in class ability set",
			mutate: func(ic *IntentClassification) {
				ic.Rules[0].Action = Action{Type: ActionTaunt, Target: TargetSelf}
			},
		},
		{
			name: "unknown phase_id",
			mutate: func(ic *IntentClassification) {
				ic.Rules[0].Condition.PhaseID = "does_not_exist"
			},
		},
		{
			name: "threshold out of range",
			mutate: func(ic *IntentClassification) {
				ic.Rules[0] = Rule{
					Priority:  0,
					Condition: Condition{Type: ConditionSelfResourceBelow, Resource: ResourceHP, Threshold: 1.5},
					Action:    Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy},
				}
			},
		},
		{
			name: "invalid target selector",
			mutate: func(ic *IntentClassification) {
				ic.Rules[1].Action = Action{Type: ActionBasicAttack, Target: "random_enemy"}
			},
		},
		{
			name:   "fallback action not in class ability set",
			mutate: func(ic *IntentClassification) { ic.FallbackAction = Action{Type: ActionHeal, Target: TargetSelf} },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ic := validMageClassification()
			tc.mutate(&ic)
			if err := ValidateIntentClassification(ic, frostWarden()); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestValidateIntentClassification_MaxThreeRules(t *testing.T) {
	ic := validMageClassification()
	ic.Rules = append(ic.Rules,
		Rule{Priority: 2, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
	)
	if err := ValidateIntentClassification(ic, frostWarden()); err != nil {
		t.Fatalf("3 rules should be accepted, got error: %v", err)
	}
}
