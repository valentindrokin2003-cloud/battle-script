package service

// ResourceValue is a unit's current/max pair for one resource. Threshold
// conditions compare against Fraction(), never the raw Current value,
// per the action schema spec's [0.0, 1.0] threshold contract.
type ResourceValue struct {
	Current float64
	Max     float64
}

// Fraction returns Current/Max, or 0 if Max is non-positive.
func (r ResourceValue) Fraction() float64 {
	if r.Max <= 0 {
		return 0
	}
	return r.Current / r.Max
}

// UnitState is one hero's live battle state.
type UnitState struct {
	ID        string
	HeroClass HeroClass
	Resources map[ResourceType]ResourceValue
	Statuses  map[string]bool
}

// Resource returns the named resource, or a zero ResourceValue (Fraction
// 0) if the unit does not track it.
func (u UnitState) Resource(rt ResourceType) ResourceValue {
	return u.Resources[rt]
}

// BossState is the boss's live battle state, distinct from the static
// Boss script content in boss.go.
type BossState struct {
	Resources       map[ResourceType]ResourceValue
	CurrentPhaseID  string
	TargetingUnitID string // "" if the boss has not committed to a single-target action this turn
}

// Resource returns the named resource, or a zero ResourceValue if the
// boss does not track it.
func (b BossState) Resource(rt ResourceType) ResourceValue {
	return b.Resources[rt]
}

// BattleContext is everything the rule engine needs to evaluate one
// hero's TacticProgram on one turn. Allies includes Self, so
// lowest_hp_ally and role:<class> selectors can resolve to the acting
// hero itself.
type BattleContext struct {
	Self   UnitState
	Allies []UnitState
	Boss   BossState
}
