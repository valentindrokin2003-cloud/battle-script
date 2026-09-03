package service

// This file holds Phase 0 seed content: illustrative numbers from
// docs/superpowers/specs/2026-09-03-battle-turn-resolution-design.md.
// Not final balance — needs a playtest pass before Phase 0 ships.

// HeroBaseResources returns the starting Max resources for class, per
// the spec's "Базовые ресурсы героев" table.
func HeroBaseResources(class HeroClass) map[ResourceType]ResourceValue {
	switch class {
	case HeroClassTank:
		return map[ResourceType]ResourceValue{
			ResourceHP:     {Current: 120, Max: 120},
			ResourceShield: {Current: 60, Max: 60},
		}
	case HeroClassArcher:
		return map[ResourceType]ResourceValue{
			ResourceHP:   {Current: 80, Max: 80},
			ResourceMana: {Current: 40, Max: 40},
		}
	case HeroClassMage:
		return map[ResourceType]ResourceValue{
			ResourceHP:   {Current: 70, Max: 70},
			ResourceMana: {Current: 50, Max: 50},
		}
	case HeroClassHealer:
		return map[ResourceType]ResourceValue{
			ResourceHP:   {Current: 75, Max: 75},
			ResourceMana: {Current: 50, Max: 50},
		}
	default:
		return map[ResourceType]ResourceValue{}
	}
}

// FrostWardenBoss is the Phase 0 seed content for the frost_warden boss,
// per the boss script spec.
func FrostWardenBoss() Boss {
	return Boss{
		BossID:      "frost_warden",
		DisplayName: "Ледяной страж",
		MaxHP:       1000,
		Phases: []BossPhase{
			{PhaseID: "exposed", HPThresholdEnter: 1.0, AbilityPattern: []string{"cleave_all", "single_target_hit"}, Provocation: "«Подходите, если смелости хватит!»"},
			{PhaseID: "shielded", HPThresholdEnter: 0.6, AbilityPattern: []string{"shield_up", "single_target_hit"}, Provocation: "«Мой щит вас не пропустит!»"},
			{PhaseID: "enraged", HPThresholdEnter: 0.25, AbilityPattern: []string{"cleave_all", "cleave_all", "single_target_hit"}, Provocation: "«Хватит! Никто не уйдёт!»"},
		},
	}
}

// ShadowHunterBoss is the Phase 0 seed content for the shadow_hunter boss.
func ShadowHunterBoss() Boss {
	return Boss{
		BossID:      "shadow_hunter",
		DisplayName: "Охотник теней",
		MaxHP:       800,
		Phases: []BossPhase{
			{PhaseID: "stalking", HPThresholdEnter: 1.0, AbilityPattern: []string{"single_target_hit"}},
			{PhaseID: "hunting", HPThresholdEnter: 0.5, AbilityPattern: []string{"single_target_hit", "cleave_all"}},
		},
	}
}

// StoneGiantBoss is the Phase 0 seed content for the stone_giant boss.
func StoneGiantBoss() Boss {
	return Boss{
		BossID:      "stone_giant",
		DisplayName: "Каменный великан",
		MaxHP:       1200,
		Phases: []BossPhase{
			{PhaseID: "steady", HPThresholdEnter: 1.0, AbilityPattern: []string{"single_target_hit"}},
			{PhaseID: "crumbling", HPThresholdEnter: 0.4, AbilityPattern: []string{"single_target_hit", "single_target_hit", "cleave_all"}},
		},
	}
}

// heroBaseDamage is the "Таблица эффектов способностей" damage column
// from the turn resolution spec. frost_bolt is intentionally absent —
// its phase-dependent amount is computed in heroDamageAmount, not looked
// up here. heal is also absent — see healAmount.
var heroBaseDamage = map[HeroClass]map[ActionType]float64{
	HeroClassTank: {
		ActionBasicAttack: 9,
	},
	HeroClassArcher: {
		ActionBasicAttack:  10,
		ActionPiercingShot: 8,
		ActionAimedShot:    16,
	},
	HeroClassMage: {
		ActionBasicAttack: 9,
		ActionFireball:    11,
	},
	HeroClassHealer: {
		ActionBasicAttack: 5,
		ActionHeal:        12,
	},
}

type bossDamage struct {
	Single float64
	Cleave float64
}

// bossPhaseDamage is the "Урон single_target_hit/cleave_all по боссам и
// фазам" table from the turn resolution spec.
var bossPhaseDamage = map[string]map[string]bossDamage{
	"frost_warden": {
		"exposed":  {Single: 12, Cleave: 6},
		"shielded": {Single: 12, Cleave: 0},
		"enraged":  {Single: 16, Cleave: 8},
	},
	"shadow_hunter": {
		"stalking": {Single: 10, Cleave: 0},
		"hunting":  {Single: 14, Cleave: 6},
	},
	"stone_giant": {
		"steady":    {Single: 10, Cleave: 0},
		"crumbling": {Single: 22, Cleave: 8},
	},
}
