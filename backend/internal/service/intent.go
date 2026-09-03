package service

import (
	"errors"
	"fmt"
)

// MaxRulesPerHero is the Phase 0 cap on TacticProgram length, per the
// action schema spec's "Не-цели" (kept small for both LLM classification
// reliability and for a child to reason about their own tactic).
const MaxRulesPerHero = 3

// Condition is one member of the closed condition vocabulary. Fields
// unused by a given Type are left zero; ValidateIntentClassification
// checks that only the fields relevant to Type are meaningfully set.
type Condition struct {
	Type      ConditionType  `json:"type"`
	Target    TargetSelector `json:"target,omitempty"`
	Resource  ResourceType   `json:"resource,omitempty"`
	Threshold float64        `json:"threshold,omitempty"`
	PhaseID   string         `json:"phase_id,omitempty"`
	Status    string         `json:"status,omitempty"`
}

// Action is a single ability invocation: an ActionType from the acting
// hero's closed ability set, plus a target selector.
type Action struct {
	Type   ActionType     `json:"type"`
	Target TargetSelector `json:"target,omitempty"`
}

// Rule is one "condition -> action" entry in a TacticProgram, evaluated
// in Priority order (0 highest).
type Rule struct {
	Priority  int       `json:"priority"`
	Condition Condition `json:"condition"`
	Action    Action    `json:"action"`
}

// Confidence flags whether the LLM adapter classified the prompt
// directly (High) or whether the retry/fallback path from the HLD's
// error handling section had to run (LowFallbackUsed).
type Confidence string

const (
	ConfidenceHigh            Confidence = "high"
	ConfidenceLowFallbackUsed Confidence = "low_fallback_used"
)

// IntentClassification is the structured result IntentClassifier must
// produce from a child's free-text PromptSubmission, per
// docs/superpowers/specs/2026-09-03-hero-class-action-schema-design.md.
type IntentClassification struct {
	HeroClass                HeroClass  `json:"hero_class"`
	SchemaVersion            string     `json:"schema_version"`
	Rules                    []Rule     `json:"rules"`
	FallbackAction           Action     `json:"fallback_action"`
	SourcePromptSubmissionID string     `json:"source_prompt_submission_id"`
	Confidence               Confidence `json:"confidence"`
}

// ValidateIntentClassification checks ic against the closed vocabulary
// and structural rules from the action schema spec. The LLM provider is
// never a trust boundary: every field is re-checked here regardless of
// what the classification prompt already constrained it to.
//
// boss supplies the set of valid phase_id values for boss_phase_is
// conditions; pass a zero Boss (no phases) when validating without a
// specific boss in scope, in which case any boss_phase_is condition is
// rejected.
func ValidateIntentClassification(ic IntentClassification, boss Boss) error {
	if !ic.HeroClass.Valid() {
		return fmt.Errorf("unknown hero_class %q", ic.HeroClass)
	}
	if ic.SchemaVersion == "" {
		return errors.New("schema_version is required")
	}
	if len(ic.Rules) > MaxRulesPerHero {
		return fmt.Errorf("too many rules: got %d, max %d", len(ic.Rules), MaxRulesPerHero)
	}
	for i, r := range ic.Rules {
		if r.Priority != i {
			return fmt.Errorf("rule %d: priority must equal index, got %d", i, r.Priority)
		}
		if err := validateCondition(r.Condition, boss); err != nil {
			return fmt.Errorf("rule %d condition: %w", i, err)
		}
		if err := validateAction(ic.HeroClass, r.Action); err != nil {
			return fmt.Errorf("rule %d action: %w", i, err)
		}
	}
	if err := validateAction(ic.HeroClass, ic.FallbackAction); err != nil {
		return fmt.Errorf("fallback_action: %w", err)
	}
	return nil
}

func validateAction(class HeroClass, a Action) error {
	if !ClassHasAbility(class, a.Type) {
		return fmt.Errorf("action %q is not in %s's ability set", a.Type, class)
	}
	if a.Target != "" && !ValidSelector(a.Target) {
		return fmt.Errorf("action target %q is not a valid selector", a.Target)
	}
	return nil
}

func validateCondition(c Condition, boss Boss) error {
	switch c.Type {
	case ConditionAlways:
		return nil
	case ConditionSelfResourceBelow, ConditionBossResourceBelow:
		return validateResourceThreshold(c)
	case ConditionAllyResourceBelow:
		if err := validateResourceThreshold(c); err != nil {
			return err
		}
		return requireValidTarget(c.Target)
	case ConditionBossPhaseIs:
		if c.PhaseID == "" {
			return errors.New("boss_phase_is requires phase_id")
		}
		if !boss.HasPhase(c.PhaseID) {
			return fmt.Errorf("unknown phase_id %q for boss %q", c.PhaseID, boss.BossID)
		}
		return nil
	case ConditionBossTargeting:
		return requireValidTarget(c.Target)
	case ConditionAllyStatusIs:
		if c.Status == "" {
			return errors.New("ally_status_is requires status")
		}
		return requireValidTarget(c.Target)
	default:
		return fmt.Errorf("unknown condition type %q", c.Type)
	}
}

func validateResourceThreshold(c Condition) error {
	if !c.Resource.Valid() {
		return fmt.Errorf("unknown resource %q", c.Resource)
	}
	if c.Threshold < 0.0 || c.Threshold > 1.0 {
		return fmt.Errorf("threshold %v out of range [0.0, 1.0]", c.Threshold)
	}
	return nil
}

func requireValidTarget(t TargetSelector) error {
	if t == "" {
		return errors.New("target is required")
	}
	if !ValidSelector(t) {
		return fmt.Errorf("target %q is not a valid selector", t)
	}
	return nil
}
