package service

import "testing"

func TestBoss_ValidatePhases_Valid(t *testing.T) {
	if err := frostWarden().ValidatePhases(); err != nil {
		t.Fatalf("expected valid phases, got error: %v", err)
	}
}

func TestBoss_ValidatePhases_Rejections(t *testing.T) {
	cases := []struct {
		name   string
		phases []BossPhase
	}{
		{
			name:   "no phases",
			phases: nil,
		},
		{
			name: "does not start at 1.0",
			phases: []BossPhase{
				{PhaseID: "a", HPThresholdEnter: 0.9},
			},
		},
		{
			name: "gap not strictly descending",
			phases: []BossPhase{
				{PhaseID: "a", HPThresholdEnter: 1.0},
				{PhaseID: "b", HPThresholdEnter: 1.0},
			},
		},
		{
			name: "duplicate phase_id",
			phases: []BossPhase{
				{PhaseID: "a", HPThresholdEnter: 1.0},
				{PhaseID: "a", HPThresholdEnter: 0.5},
			},
		},
		{
			name: "threshold out of range",
			phases: []BossPhase{
				{PhaseID: "a", HPThresholdEnter: 1.0},
				{PhaseID: "b", HPThresholdEnter: -0.1},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := Boss{BossID: "test", Phases: tc.phases}
			if err := b.ValidatePhases(); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestBoss_PhaseForHPFraction(t *testing.T) {
	b := frostWarden()
	cases := []struct {
		fraction  float64
		wantPhase string
	}{
		{1.0, "exposed"},
		{0.8, "exposed"},
		{0.6, "shielded"},
		{0.4, "shielded"},
		{0.25, "enraged"},
		{0.0, "enraged"},
	}
	for _, tc := range cases {
		phase, err := b.PhaseForHPFraction(tc.fraction)
		if err != nil {
			t.Fatalf("PhaseForHPFraction(%v) error: %v", tc.fraction, err)
		}
		if phase.PhaseID != tc.wantPhase {
			t.Errorf("PhaseForHPFraction(%v) = %q, want %q", tc.fraction, phase.PhaseID, tc.wantPhase)
		}
	}
}
