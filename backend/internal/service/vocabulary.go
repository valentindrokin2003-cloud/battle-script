// Package service holds Battle Script domain logic: the closed tactic
// vocabulary, intent validation, boss phase resolution, and the
// deterministic rule engine. It has no dependency on HTTP, SQL, or any
// LLM SDK — see docs/superpowers/specs/2026-09-03-battle-script-hld-design.md.
package service

import "strings"

// HeroClass is one of the closed set of Phase 0 hero classes.
type HeroClass string

const (
	HeroClassTank   HeroClass = "tank"
	HeroClassArcher HeroClass = "archer"
	HeroClassMage   HeroClass = "mage"
	HeroClassHealer HeroClass = "healer"
)

func (c HeroClass) Valid() bool {
	switch c {
	case HeroClassTank, HeroClassArcher, HeroClassMage, HeroClassHealer:
		return true
	default:
		return false
	}
}

// ResourceType is a resource a threshold condition can check.
type ResourceType string

const (
	ResourceHP     ResourceType = "hp"
	ResourceShield ResourceType = "shield"
	ResourceMana   ResourceType = "mana"
)

func (r ResourceType) Valid() bool {
	switch r {
	case ResourceHP, ResourceShield, ResourceMana:
		return true
	default:
		return false
	}
}

// TargetSelector is a member of the closed selector vocabulary from
// docs/superpowers/specs/2026-09-03-hero-class-action-schema-design.md.
// A "role:<hero_class>" selector is dynamic and validated separately
// via ParseRoleSelector.
type TargetSelector string

const (
	TargetSelf           TargetSelector = "self"
	TargetLowestHPEnemy  TargetSelector = "lowest_hp_enemy"
	TargetHighestHPEnemy TargetSelector = "highest_hp_enemy"
	TargetLowestHPAlly   TargetSelector = "lowest_hp_ally"
	TargetBoss           TargetSelector = "boss"
)

const roleSelectorPrefix = "role:"

// ParseRoleSelector reports whether s is a "role:<hero_class>" selector
// and, if so, returns the referenced class.
func ParseRoleSelector(s TargetSelector) (HeroClass, bool) {
	raw := string(s)
	if !strings.HasPrefix(raw, roleSelectorPrefix) {
		return "", false
	}
	class := HeroClass(strings.TrimPrefix(raw, roleSelectorPrefix))
	if !class.Valid() {
		return "", false
	}
	return class, true
}

// ValidSelector reports whether s is a member of the closed selector
// vocabulary, including the dynamic "role:<hero_class>" form.
func ValidSelector(s TargetSelector) bool {
	switch s {
	case TargetSelf, TargetLowestHPEnemy, TargetHighestHPEnemy, TargetLowestHPAlly, TargetBoss:
		return true
	}
	_, ok := ParseRoleSelector(s)
	return ok
}

// ConditionType is a member of the closed condition vocabulary.
type ConditionType string

const (
	ConditionAlways            ConditionType = "always"
	ConditionSelfResourceBelow ConditionType = "self_resource_below"
	ConditionAllyResourceBelow ConditionType = "ally_resource_below"
	ConditionBossResourceBelow ConditionType = "boss_resource_below"
	ConditionBossPhaseIs       ConditionType = "boss_phase_is"
	ConditionBossTargeting     ConditionType = "boss_targeting"
	ConditionAllyStatusIs      ConditionType = "ally_status_is"
)

// ActionType is a member of a hero class's closed ability set.
type ActionType string

const (
	ActionBasicAttack  ActionType = "basic_attack"
	ActionTaunt        ActionType = "taunt"
	ActionRetreat      ActionType = "retreat"
	ActionPiercingShot ActionType = "piercing_shot"
	ActionAimedShot    ActionType = "aimed_shot"
	ActionFrostBolt    ActionType = "frost_bolt"
	ActionFireball     ActionType = "fireball"
	ActionHeal         ActionType = "heal"
	ActionCleanse      ActionType = "cleanse"
)

// classAbilities is the closed per-class ability set from the action
// schema spec's "Классы героев" section.
var classAbilities = map[HeroClass]map[ActionType]bool{
	HeroClassTank: {
		ActionBasicAttack: true,
		ActionTaunt:       true,
		ActionRetreat:     true,
	},
	HeroClassArcher: {
		ActionBasicAttack:  true,
		ActionPiercingShot: true,
		ActionAimedShot:    true,
	},
	HeroClassMage: {
		ActionBasicAttack: true,
		ActionFrostBolt:   true,
		ActionFireball:    true,
	},
	HeroClassHealer: {
		ActionBasicAttack: true,
		ActionHeal:        true,
		ActionCleanse:     true,
	},
}

// ClassHasAbility reports whether action is in class's closed ability set.
func ClassHasAbility(class HeroClass, action ActionType) bool {
	return classAbilities[class][action]
}

// classDefaultFallback is the class-level default used when parsing or
// classification fails entirely (see HLD "Обработка ошибок"). Healer
// defaults to healing its neediest ally rather than attacking, because a
// fully-failed classification should still keep the class doing its core
// job; every other class defaults to a plain attack on the boss.
var classDefaultFallback = map[HeroClass]Action{
	HeroClassTank:   {Type: ActionBasicAttack, Target: TargetLowestHPEnemy},
	HeroClassArcher: {Type: ActionBasicAttack, Target: TargetLowestHPEnemy},
	HeroClassMage:   {Type: ActionBasicAttack, Target: TargetLowestHPEnemy},
	HeroClassHealer: {Type: ActionHeal, Target: TargetLowestHPAlly},
}

// DefaultFallback returns the class-level default fallback action.
func DefaultFallback(class HeroClass) Action {
	return classDefaultFallback[class]
}
