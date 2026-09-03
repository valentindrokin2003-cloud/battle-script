package service

import "testing"

func TestParseRoleSelector(t *testing.T) {
	cases := []struct {
		name      string
		selector  TargetSelector
		wantClass HeroClass
		wantOK    bool
	}{
		{"valid healer role", "role:healer", HeroClassHealer, true},
		{"valid tank role", "role:tank", HeroClassTank, true},
		{"unknown class", "role:paladin", "", false},
		{"not a role selector", "self", "", false},
		{"empty", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			class, ok := ParseRoleSelector(tc.selector)
			if ok != tc.wantOK || class != tc.wantClass {
				t.Errorf("ParseRoleSelector(%q) = (%q, %v), want (%q, %v)", tc.selector, class, ok, tc.wantClass, tc.wantOK)
			}
		})
	}
}

func TestValidSelector(t *testing.T) {
	valid := []TargetSelector{TargetSelf, TargetLowestHPEnemy, TargetHighestHPEnemy, TargetLowestHPAlly, TargetBoss, "role:mage"}
	for _, s := range valid {
		if !ValidSelector(s) {
			t.Errorf("ValidSelector(%q) = false, want true", s)
		}
	}
	invalid := []TargetSelector{"", "random_enemy", "role:paladin", "lowest_hp"}
	for _, s := range invalid {
		if ValidSelector(s) {
			t.Errorf("ValidSelector(%q) = true, want false", s)
		}
	}
}

func TestClassHasAbility(t *testing.T) {
	if !ClassHasAbility(HeroClassMage, ActionFrostBolt) {
		t.Error("mage should have frost_bolt")
	}
	if ClassHasAbility(HeroClassMage, ActionTaunt) {
		t.Error("mage should not have taunt")
	}
	if !ClassHasAbility(HeroClassTank, ActionTaunt) {
		t.Error("tank should have taunt")
	}
}

func TestDefaultFallback(t *testing.T) {
	if got := DefaultFallback(HeroClassHealer); got.Type != ActionHeal {
		t.Errorf("healer default fallback = %v, want heal", got.Type)
	}
	if got := DefaultFallback(HeroClassMage); got.Type != ActionBasicAttack {
		t.Errorf("mage default fallback = %v, want basic_attack", got.Type)
	}
}
