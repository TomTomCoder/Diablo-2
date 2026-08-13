package d2hero

import (
	"errors"
	"time"

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

	// StrengthSpent/EnergySpent/DexteritySpent/VitalitySpent count how much
	// of the corresponding attribute above came from spending a level-up
	// point via SpendAttributePoint, as opposed to the class's base stats
	// from character creation. Needed so RespecAllAttributePoints/
	// RefundAttributePoint can refund exactly what was spent without also
	// undoing the base values (devil_game_design_reference.md §10 "Respec
	// partiel"/"Glyphe d'oubli").
	StrengthSpent  int `json:"strengthSpent"`
	EnergySpent    int `json:"energySpent"`
	DexteritySpent int `json:"dexteritySpent"`
	VitalitySpent  int `json:"vitalitySpent"`

	Health     int     `json:"health"`
	MaxHealth  int     `json:"maxHealth"`
	Mana       int     `json:"mana"`
	MaxMana    int     `json:"maxMana"`
	Stamina    float64 `json:"-"` // only MaxStamina is saved, Stamina gets reset on entering world
	MaxStamina int     `json:"maxStamina"`

	// LifePerVit/ManaPerEne are Diablo 2's own charstats.txt columns
	// (d2records.CharStatRecord), copied once at creation so
	// SpendAttributePoint/RefundAttributePoint/RespecAllAttributePoints can
	// keep MaxHealth/MaxMana in sync with Vitality/Energy after creation --
	// devil_game_design_reference.md §5: "Vitality: Détermine les points de
	// vie", "Energy: Augmente ... la réserve de mana". Without these,
	// MaxHealth/MaxMana would stay frozen at their character-creation value
	// forever, since CreateHeroStatsState is the only place classStats
	// itself is ever available.
	LifePerVit int `json:"lifePerVit"`
	ManaPerEne int `json:"manaPerEne"`

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

	// MagicImmuneUntil is when the hero's magic immunity (Éveil du Nexus,
	// Ésotérisme's ultimate: "immunité magique pendant 8 secondes") expires.
	// Zero value means not immune. See ApplyMagicImmunity/IsMagicImmune.
	MagicImmuneUntil time.Time `json:"-"`

	// ResonanceMagiqueBonusUntil is when the temporary damage bonus armed by
	// the hero's last cast (Résonance magique, Arcane §7: "chaque sort lancé
	// augmente les dégâts du suivant") expires unused. Zero value means no
	// bonus armed. Unlike the other *Until fields, this one is also cleared
	// the moment it's used -- see ApplyResonanceMagiqueBonus/
	// ConsumeResonanceMagiqueBonus.
	ResonanceMagiqueBonusUntil time.Time `json:"-"`

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

		LifePerVit: classStats.LifePerVit,
		ManaPerEne: classStats.ManaPerEne,
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

// Attribute identifies one of the four attributes a level-up's "5 points
// d'attributs à répartir librement" (devil_game_design_reference.md §6) can
// be spent on via SpendAttributePoint.
type Attribute int

// Devil's four attributes -- Strength, Energy, Dexterity, Vitality.
const (
	AttributeStrength Attribute = iota
	AttributeEnergy
	AttributeDexterity
	AttributeVitality
)

// SpendAttributePoint spends one of s.StatsPoints on attr, increasing it (and
// its matching *Spent counter, for RefundAttributePoint/
// RespecAllAttributePoints) by 1.
func (s *HeroStatsState) SpendAttributePoint(attr Attribute) error {
	if s.StatsPoints <= 0 {
		return errors.New("no attribute points available")
	}

	switch attr {
	case AttributeStrength:
		s.Strength++
		s.StrengthSpent++
	case AttributeEnergy:
		s.Energy++
		s.EnergySpent++
		s.MaxMana += s.ManaPerEne
		s.Mana += s.ManaPerEne
	case AttributeDexterity:
		s.Dexterity++
		s.DexteritySpent++
	case AttributeVitality:
		s.Vitality++
		s.VitalitySpent++
		s.MaxHealth += s.LifePerVit
		s.Health += s.LifePerVit
	default:
		return errors.New("unknown attribute")
	}

	s.StatsPoints--

	return nil
}

// RefundAttributePoint undoes one previously spent point on attr (the
// design's "Glyphe d'oubli" applied to "1 point d'attribut" instead of a
// skill -- devil_game_design_reference.md §10), returning it to
// s.StatsPoints. Errors if attr has no spent points to refund.
func (s *HeroStatsState) RefundAttributePoint(attr Attribute) error {
	switch attr {
	case AttributeStrength:
		if s.StrengthSpent <= 0 {
			return errors.New("no spent points to refund on this attribute")
		}

		s.Strength--
		s.StrengthSpent--
	case AttributeEnergy:
		if s.EnergySpent <= 0 {
			return errors.New("no spent points to refund on this attribute")
		}

		s.Energy--
		s.EnergySpent--
		s.MaxMana -= s.ManaPerEne
		s.Mana = min(s.Mana, s.MaxMana)
	case AttributeDexterity:
		if s.DexteritySpent <= 0 {
			return errors.New("no spent points to refund on this attribute")
		}

		s.Dexterity--
		s.DexteritySpent--
	case AttributeVitality:
		if s.VitalitySpent <= 0 {
			return errors.New("no spent points to refund on this attribute")
		}

		s.Vitality--
		s.VitalitySpent--
		s.MaxHealth -= s.LifePerVit
		s.Health = min(s.Health, s.MaxHealth)
	default:
		return errors.New("unknown attribute")
	}

	s.StatsPoints++

	return nil
}

// RespecAllAttributePoints undoes every point ever spent via
// SpendAttributePoint across all four attributes, refunding them to
// StatsPoints -- the attribute half of the design's "Respec partiel"
// (devil_game_design_reference.md §10: "Tous les points de compétences et
// d'attributs"). Reports how many points were refunded in total.
func (s *HeroStatsState) RespecAllAttributePoints() (pointsRefunded int) {
	pointsRefunded = s.StrengthSpent + s.EnergySpent + s.DexteritySpent + s.VitalitySpent

	s.Strength -= s.StrengthSpent
	s.Energy -= s.EnergySpent
	s.Dexterity -= s.DexteritySpent
	s.Vitality -= s.VitalitySpent

	s.MaxMana -= s.EnergySpent * s.ManaPerEne
	s.Mana = min(s.Mana, s.MaxMana)
	s.MaxHealth -= s.VitalitySpent * s.LifePerVit
	s.Health = min(s.Health, s.MaxHealth)

	s.StrengthSpent, s.EnergySpent, s.DexteritySpent, s.VitalitySpent = 0, 0, 0, 0
	s.StatsPoints += pointsRefunded

	return pointsRefunded
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
// mechanic (§7) but doesn't specify a ratio.
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

// Heal increases Health by amount, capped at MaxHealth. The inverse of
// ApplyDamage.
func (s *HeroStatsState) Heal(amount int) {
	s.Health += amount

	if s.Health > s.MaxHealth {
		s.Health = s.MaxHealth
	}
}

// RestoreMana increases Mana by amount, capped at MaxMana. Mana's
// equivalent of Heal.
func (s *HeroStatsState) RestoreMana(amount int) {
	s.Mana += amount

	if s.Mana > s.MaxMana {
		s.Mana = s.MaxMana
	}
}

// ApplyMagicImmunity marks the hero as immune to damage (see IsMagicImmune)
// until the given time. Devil's "Éveil du Nexus" ultimate (Ésotérisme §7).
func (s *HeroStatsState) ApplyMagicImmunity(until time.Time) {
	s.MagicImmuneUntil = until
}

// IsMagicImmune reports whether the hero's magic immunity is still active
// at now.
func (s *HeroStatsState) IsMagicImmune(now time.Time) bool {
	return now.Before(s.MagicImmuneUntil)
}

// ApplyResonanceMagiqueBonus arms the hero's next-cast damage bonus (see
// ConsumeResonanceMagiqueBonus) until the given time. Devil's "Résonance
// magique" (Arcane §7): every cast re-arms this, regardless of whether it
// was already active.
func (s *HeroStatsState) ApplyResonanceMagiqueBonus(until time.Time) {
	s.ResonanceMagiqueBonusUntil = until
}

// ConsumeResonanceMagiqueBonus reports whether the hero's Résonance
// magique bonus is still active at now, clearing it either way so it can
// only ever benefit one cast.
func (s *HeroStatsState) ConsumeResonanceMagiqueBonus(now time.Time) bool {
	active := now.Before(s.ResonanceMagiqueBonusUntil)
	s.ResonanceMagiqueBonusUntil = time.Time{}

	return active
}
