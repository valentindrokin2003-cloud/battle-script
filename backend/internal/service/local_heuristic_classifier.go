package service

import (
	"context"
	"strings"
)

// LocalHeuristicClassifier is a keyword-matching dev/test stand-in for
// IntentClassifier. It is NOT an LLM and must never be deployed as the
// Phase 0 pilot's actual classifier — it exists only so the rest of the
// pipeline (moderation, validation, the rule engine, and eventually the
// HTTP layer) can be built and tested before a real LLM API key is
// available. See the moderation/classifier port spec's "Не-цели".
type LocalHeuristicClassifier struct{}

func (LocalHeuristicClassifier) Classify(_ context.Context, input ClassificationContext) (IntentClassification, error) {
	lower := strings.ToLower(input.PromptText)

	if rules, ok := matchRules(input.HeroClass, lower); ok {
		return IntentClassification{
			HeroClass:                input.HeroClass,
			SchemaVersion:            CurrentSchemaVersion,
			Rules:                    rules,
			FallbackAction:           DefaultFallback(input.HeroClass),
			SourcePromptSubmissionID: "",
			Confidence:               ConfidenceHigh,
		}, nil
	}

	// Not recognized: report low confidence directly rather than an
	// error, so ClassifyWithFallback's generic retry path isn't the only
	// tested route to a fallback result — see the spec's rationale.
	return IntentClassification{
		HeroClass:      input.HeroClass,
		SchemaVersion:  CurrentSchemaVersion,
		Rules:          nil,
		FallbackAction: DefaultFallback(input.HeroClass),
		Confidence:     ConfidenceLowFallbackUsed,
	}, nil
}

func matchRules(class HeroClass, lower string) ([]Rule, bool) {
	contains := func(subs ...string) bool {
		for _, s := range subs {
			if strings.Contains(lower, s) {
				return true
			}
		}
		return false
	}

	switch class {
	case HeroClassMage:
		if contains("ледян", "лёд", "frost") && contains("щит") {
			return []Rule{
				{Priority: 0, Condition: Condition{Type: ConditionBossPhaseIs, PhaseID: "shielded"}, Action: Action{Type: ActionFrostBolt, Target: TargetLowestHPEnemy}},
				{Priority: 1, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionBasicAttack, Target: TargetLowestHPEnemy}},
			}, true
		}

	case HeroClassTank:
		hasTaunt := contains("провоц", "таунт")
		hasRetreat := contains("отступ") && contains("щит")
		switch {
		case hasTaunt && hasRetreat:
			return []Rule{
				{Priority: 0, Condition: Condition{Type: ConditionBossTargeting, Target: "role:healer"}, Action: Action{Type: ActionTaunt, Target: TargetSelf}},
				{Priority: 1, Condition: Condition{Type: ConditionSelfResourceBelow, Resource: ResourceShield, Threshold: 0.3}, Action: Action{Type: ActionRetreat, Target: TargetSelf}},
			}, true
		case hasTaunt:
			return []Rule{
				{Priority: 0, Condition: Condition{Type: ConditionBossTargeting, Target: "role:healer"}, Action: Action{Type: ActionTaunt, Target: TargetSelf}},
			}, true
		case hasRetreat:
			return []Rule{
				{Priority: 0, Condition: Condition{Type: ConditionSelfResourceBelow, Resource: ResourceShield, Threshold: 0.3}, Action: Action{Type: ActionRetreat, Target: TargetSelf}},
			}, true
		}

	case HeroClassHealer:
		if contains("лечи", "лечить", "исцел") && contains("меньше", "слаб") {
			return []Rule{
				{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionHeal, Target: TargetLowestHPAlly}},
			}, true
		}

	case HeroClassArcher:
		if contains("сильный удар", "прицел", "aimed") {
			return []Rule{
				{Priority: 0, Condition: Condition{Type: ConditionAlways}, Action: Action{Type: ActionAimedShot, Target: TargetLowestHPEnemy}},
			}, true
		}
	}

	return nil, false
}
