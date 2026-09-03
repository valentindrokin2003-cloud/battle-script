package service

import (
	"errors"
	"fmt"
	"sort"
)

// BossPhase is one phase of a boss script, per
// docs/superpowers/specs/2026-09-03-boss-script-design.md. HPThresholdEnter
// is the fraction of MaxHP (in [0.0, 1.0]) at or below which this phase
// begins.
type BossPhase struct {
	PhaseID          string   `json:"phase_id"`
	HPThresholdEnter float64  `json:"hp_threshold_enter"`
	AbilityPattern   []string `json:"ability_pattern"`
	Provocation      string   `json:"provocation"`
}

// Boss is a fixed, deterministic boss script. It is content, not
// runtime battle state: BattleEngine reads it, never mutates it.
type Boss struct {
	BossID      string      `json:"boss_id"`
	DisplayName string      `json:"display_name"`
	MaxHP       float64     `json:"max_hp"`
	Phases      []BossPhase `json:"phases"`
}

// HasPhase reports whether phaseID is one of boss's phases.
func (b Boss) HasPhase(phaseID string) bool {
	for _, p := range b.Phases {
		if p.PhaseID == phaseID {
			return true
		}
	}
	return false
}

// ValidatePhases checks the "правило честности" structural requirement
// from the boss script spec: phases must be ordered strictly descending
// by HPThresholdEnter, starting at exactly 1.0, with no gaps or
// duplicate thresholds. This is a content/seed-data check, run when a
// Boss is loaded, not during battle simulation.
func (b Boss) ValidatePhases() error {
	if len(b.Phases) == 0 {
		return errors.New("boss must have at least one phase")
	}
	sorted := make([]BossPhase, len(b.Phases))
	copy(sorted, b.Phases)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].HPThresholdEnter > sorted[j].HPThresholdEnter
	})
	if sorted[0].HPThresholdEnter != 1.0 {
		return fmt.Errorf("first phase must enter at hp threshold 1.0, got %v", sorted[0].HPThresholdEnter)
	}
	seen := map[string]bool{}
	for i, p := range sorted {
		if p.HPThresholdEnter < 0.0 || p.HPThresholdEnter > 1.0 {
			return fmt.Errorf("phase %q threshold %v out of range [0.0, 1.0]", p.PhaseID, p.HPThresholdEnter)
		}
		if seen[p.PhaseID] {
			return fmt.Errorf("duplicate phase_id %q", p.PhaseID)
		}
		seen[p.PhaseID] = true
		if i > 0 && p.HPThresholdEnter >= sorted[i-1].HPThresholdEnter {
			return fmt.Errorf("phase thresholds must be strictly descending: %q (%v) is not below previous phase (%v)",
				p.PhaseID, p.HPThresholdEnter, sorted[i-1].HPThresholdEnter)
		}
	}
	return nil
}

// PhaseForHPFraction returns the phase active at the given HP fraction
// (current HP / max HP, in [0.0, 1.0]): the phase with the highest
// HPThresholdEnter that is still >= fraction.
func (b Boss) PhaseForHPFraction(fraction float64) (BossPhase, error) {
	sorted := make([]BossPhase, len(b.Phases))
	copy(sorted, b.Phases)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].HPThresholdEnter > sorted[j].HPThresholdEnter
	})
	active := BossPhase{}
	found := false
	for _, p := range sorted {
		if fraction <= p.HPThresholdEnter {
			active = p
			found = true
		}
	}
	if !found {
		return BossPhase{}, fmt.Errorf("no phase covers hp fraction %v", fraction)
	}
	return active, nil
}
