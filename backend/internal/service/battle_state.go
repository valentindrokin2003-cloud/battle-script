package service

// HeroDef is one hero's starting configuration for a BattleSession: its
// class, resources, and the TacticProgram produced by IntentClassifier
// (already validated by ValidateIntentClassification before it reaches
// here).
type HeroDef struct {
	ID        string
	HeroClass HeroClass
	Resources map[ResourceType]ResourceValue
	Program   []Rule
	Fallback  Action
}

// NewHeroDef builds a HeroDef with the class's Phase 0 base resources
// (see phase0_content.go).
func NewHeroDef(id string, class HeroClass, program []Rule) HeroDef {
	return HeroDef{
		ID:        id,
		HeroClass: class,
		Resources: HeroBaseResources(class),
		Program:   program,
		Fallback:  DefaultFallback(class),
	}
}

type heroLiveState struct {
	Def             HeroDef
	Resources       map[ResourceType]ResourceValue
	Statuses        map[string]bool
	FrontlineStreak int
}

func newHeroLiveState(def HeroDef) *heroLiveState {
	resources := make(map[ResourceType]ResourceValue, len(def.Resources))
	for k, v := range def.Resources {
		resources[k] = v
	}
	return &heroLiveState{Def: def, Resources: resources, Statuses: map[string]bool{}}
}

func (h *heroLiveState) unitState() UnitState {
	return UnitState{ID: h.Def.ID, HeroClass: h.Def.HeroClass, Resources: h.Resources, Statuses: h.Statuses}
}

func (h *heroLiveState) alive() bool {
	return h.Resources[ResourceHP].Current > 0
}

// battleState is the live, mutable state of one BattleSession. It is
// unexported: callers only interact with it through RunBattle.
type battleState struct {
	Boss   Boss
	BossHP ResourceValue

	bossPhaseID              string
	previousRoundPhaseID     string
	phaseTransitionThisRound bool
	bossActionsTaken         int
	roundNumber              int
	announcedTargetID        string
	tauntOverrideUnitID      string

	heroes []*heroLiveState
}

func newBattleState(boss Boss, heroDefs []HeroDef) *battleState {
	heroes := make([]*heroLiveState, len(heroDefs))
	for i, d := range heroDefs {
		heroes[i] = newHeroLiveState(d)
	}
	return &battleState{
		Boss:   boss,
		BossHP: ResourceValue{Current: boss.MaxHP, Max: boss.MaxHP},
		heroes: heroes,
	}
}

func (s *battleState) allHeroesDead() bool {
	for _, h := range s.heroes {
		if h.alive() {
			return false
		}
	}
	return true
}

func (s *battleState) livingAllies() []UnitState {
	out := make([]UnitState, 0, len(s.heroes))
	for _, h := range s.heroes {
		if h.alive() {
			out = append(out, h.unitState())
		}
	}
	return out
}

func (s *battleState) heroByID(id string) *heroLiveState {
	for _, h := range s.heroes {
		if h.Def.ID == id {
			return h
		}
	}
	return nil
}

func (s *battleState) bossState() BossState {
	return BossState{
		Resources:       map[ResourceType]ResourceValue{ResourceHP: s.BossHP},
		CurrentPhaseID:  s.bossPhaseID,
		TargetingUnitID: s.announcedTargetID,
	}
}

func (s *battleState) contextFor(self *heroLiveState) BattleContext {
	return BattleContext{
		Self:   self.unitState(),
		Allies: s.livingAllies(),
		Boss:   s.bossState(),
	}
}

// currentPhase returns the boss's script phase matching s.bossPhaseID.
func (s *battleState) currentPhase() BossPhase {
	for _, p := range s.Boss.Phases {
		if p.PhaseID == s.bossPhaseID {
			return p
		}
	}
	return BossPhase{}
}
