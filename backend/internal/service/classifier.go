package service

import "context"

// CurrentSchemaVersion is the IntentClassification schema version this
// backend validates against. See the action schema spec's
// schema_version field.
const CurrentSchemaVersion = "2026-09-03.1"

// ClassificationContext is everything IntentClassifier needs to turn a
// child's moderated free text into a TacticProgram: which hero class is
// acting, and the boss content already revealed to the player (so
// conditions like boss_phase_is are grounded in what the child actually
// saw before writing their tactic).
type ClassificationContext struct {
	HeroClass  HeroClass
	Boss       Boss
	PromptText string // text that has already passed Moderator.Check
}

// IntentClassifier is the port from the HLD/ADR-003: the LLM provider
// lives behind this interface, never called directly from domain code.
type IntentClassifier interface {
	Classify(ctx context.Context, input ClassificationContext) (IntentClassification, error)
}

// ClassifyWithFallback implements the HLD's error handling contract:
// one retry on error or an invalid result, then the class's
// DefaultFallback with Confidence low_fallback_used. This lives outside
// any specific IntentClassifier implementation so the retry/fallback
// behavior is identical whether the adapter behind the port is the dev
// stand-in or a real LLM.
func ClassifyWithFallback(ctx context.Context, classifier IntentClassifier, input ClassificationContext) IntentClassification {
	for attempt := 0; attempt < 2; attempt++ {
		result, err := classifier.Classify(ctx, input)
		if err == nil && ValidateIntentClassification(result, input.Boss) == nil {
			return result
		}
	}
	return IntentClassification{
		HeroClass:      input.HeroClass,
		SchemaVersion:  CurrentSchemaVersion,
		Rules:          nil,
		FallbackAction: DefaultFallback(input.HeroClass),
		Confidence:     ConfidenceLowFallbackUsed,
	}
}
