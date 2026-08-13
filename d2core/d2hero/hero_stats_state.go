package d2hero

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"
)

// HeroStatsState is a serializable state of hero stats.
type HeroStatsState struct {
	Level      int `json:"level"`
	Experience int `json:"experience"`

	Strength  int `json:"strength"`
	Energy    int `json:"energy"`
	Dexterity int `json:"dexterity"`
	Vitality  int `json:"vitality"`
	// there are stats and skills points remaining to add.
	StatsPoints int `json:"statsPoints"`
	SkillPoints int `json:"skillPoints"`

	Health     int     `json:"health"`
	MaxHealth  int     `json:"maxHealth"`
	Mana       int     `json:"mana"`
	MaxMana    int     `json:"maxMana"`
	Stamina    float64 `json:"-"` // only MaxStamina is saved, Stamina gets reset on entering world
	MaxStamina int     `json:"maxStamina"`

	// Elemental resistances (devil_game_design_reference.md §6): Feu, Froid,
	// Foudre, Ombre. Unlike Diablo 2, these can never go negative -- always
	// clamp assignments through CapResistance.
	FireResist      int `json:"fireResist"`
	ColdResist      int `json:"coldResist"`
	LightningResist int `json:"lightningResist"`
	ShadowResist    int `json:"shadowResist"`

	// ManaShieldActive is Devil's "Bouclier de mana" (Ésotérisme §7)
	// defensive spell -- while true, ApplyDamageWithManaShield drains Mana
	// before touching Health. Toggled by casting SkillBouclierDeMana, see
	// GameServer.resolveBouclierDeManaHit.
	ManaShieldActive bool `json:"manaShieldActive"`

	// ArmureDeGlaceActive is Devil's "Armure de glace" (Ésotérisme §7)
	// defensive spell: "Réduit les dégâts reçus et ralentit les attaquants
	// au contact". While true, GameServer.tryMonsterAttack reduces incoming
	// damage and slows the attacking NPC. Toggled by casting
	// SkillArmureDeGlace, see GameServer.resolveArmureDeGlaceHit.
	ArmureDeGlaceActive bool `json:"armureDeGlaceActive"`

	// values which are not saved/loaded(computed)
	NextLevelExp int `json:"-"`
}

// CreateHeroStatsState generates a running state from a hero stats.
func (f *HeroStateFactory) CreateHeroStatsState(heroClass d2enum.Hero, classStats *d2records.CharStatRecord) *HeroStatsState {
	result := HeroStatsState{
		Level:        1,
		Experience:   0,
		NextLevelExp: f.asset.Records.GetExperienceBreakpoint(heroClass, 1),
		Strength:     classStats.InitStr,
		Dexterity:    classStats.InitDex,
		Vitality:     classStats.InitVit,
		Energy:       classStats.InitEne,
		StatsPoints:  0,
		SkillPoints:  0,

		MaxHealth:  classStats.InitVit * classStats.LifePerVit,
		MaxMana:    classStats.InitEne * classStats.ManaPerEne,
		MaxStamina: classStats.InitStamina,
		// https://github.com/OpenDiablo2/OpenDiablo2/issues/814
	}

	result.Mana = result.MaxMana
	result.Health = result.MaxHealth
	result.Stamina = float64(result.MaxStamina)

	return &result
}

// skillPointsPerLevel/statsPointsPerLevel: what each level grants, per
// "À chaque montée de niveau : 5 points d'attributs à répartir librement +
// 1 point de compétence" (devil_game_design_reference.md §5).
const (
	skillPointsPerLevel = 1
	statsPointsPerLevel = 5
)

// experienceForLevel returns the experience needed to advance from level
// to level+1.
//
// ponytail: a flat placeholder curve (100 * level) -- no real Devil
// experience table exists yet. Doesn't reuse Diablo 2's own
// GetExperienceBreakpoint (used for the very first threshold in
// CreateHeroStatsState below): that table only has real values once the
// player's own MPQ files are loaded, so it can't be exercised in tests
// without them, and it's per-D2-class data that doesn't necessarily fit
// Devil's single-Mage design anyway.
func experienceForLevel(level int) int {
	return level * 100
}

// GrantExperience adds amount to Experience, leveling up (possibly more
// than once, if amount clears several thresholds) while there's enough to
// cross NextLevelExp. Each level grants skillPointsPerLevel/
// statsPointsPerLevel and moves NextLevelExp to the next threshold.
func (s *HeroStatsState) GrantExperience(amount int) {
	s.Experience += amount

	for s.Experience >= s.NextLevelExp {
		s.Experience -= s.NextLevelExp
		s.Level++
		s.SkillPoints += skillPointsPerLevel
		s.StatsPoints += statsPointsPerLevel
		s.NextLevelExp = experienceForLevel(s.Level)
	}
}

// resistanceCap is the maximum any elemental resistance can reach; unlike
// Diablo 2, resistances here can never go negative either (see
// CapResistance) -- devil_game_design_reference.md §6/§12.
const resistanceCap = 75

// CapResistance clamps a resistance value to [0, resistanceCap]. Always run
// resistance assignments through this rather than setting the field
// directly, so nothing can push a resistance negative (Devil explicitly
// removes Diablo 2's punitive negative-resistance mechanic) or over the cap.
func CapResistance(value int) int {
	if value < 0 {
		return 0
	}

	if value > resistanceCap {
		return resistanceCap
	}

	return value
}

// MitigateDamage reduces amount by resistPercent% (expected already capped
// via CapResistance), flooring at 0.
func MitigateDamage(amount, resistPercent int) int {
	mitigated := amount - (amount*resistPercent)/100
	if mitigated < 0 {
		return 0
	}

	return mitigated
}

// ApplyDamage reduces Health by amount and reports whether the hero died.
// Mirrors d2mapentity.NPC.ApplyDamage -- floors at 0, and further damage
// to an already-dead hero is a no-op that still reports died=true.
func (s *HeroStatsState) ApplyDamage(amount int) (died bool) {
	if s.Health <= 0 {
		return true
	}

	s.Health -= amount

	if s.Health < 0 {
		s.Health = 0
	}

	return s.Health == 0
}

// ApplyDamageWithManaShield applies amount, draining it from Mana first if
// ManaShieldActive, with any remainder past available Mana falling through
// to Health via ApplyDamage.
//
// ponytail: 1:1 absorption (1 damage drains 1 mana) -- the design names the
// mechanic (§6/§7) but doesn't specify a ratio. No duration or real
// cast/toggle exists yet either; something else sets ManaShieldActive
// directly (ROADMAP.md Phase 2).
func (s *HeroStatsState) ApplyDamageWithManaShield(amount int) (died bool) {
	if !s.ManaShieldActive || s.Mana <= 0 {
		return s.ApplyDamage(amount)
	}

	absorbed := amount
	if absorbed > s.Mana {
		absorbed = s.Mana
	}

	s.Mana -= absorbed

	if remainder := amount - absorbed; remainder > 0 {
		return s.ApplyDamage(remainder)
	}

	return s.Health <= 0
}
