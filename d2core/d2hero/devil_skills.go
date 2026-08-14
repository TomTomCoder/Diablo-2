package d2hero

import "github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"

// SkillTraitDeFeu is Devil's own skill ID for "Trait de feu" (Élémentalisme,
// devil_game_design_reference.md §7) -- the first skill given real data
// instead of relying on Diablo 2's own skills.txt.
//
// ponytail: deliberately numbered well above any Diablo 2 skills.txt ID
// range (D2's own skill IDs are loaded at runtime from the player's MPQ
// files and top out in the low hundreds) so a real skill cast from the
// original game's data can never collide with a Devil-specific ID here.
// Devil's own skill data model (30 skills/3 trees, ROADMAP.md Phase 2)
// will replace this with a proper data table.
const SkillTraitDeFeu = 1000

// NewDevilHeroSkill returns a HeroSkill for any skill in DevilSkills, with
// just enough real data to render and cast without touching Diablo 2's own
// skill data at all: Charclass "" resolves to the game's built-in generic
// skill icon sheet (used for e.g. the default "Attack" skill), since no
// Devil-specific skill icons exist yet (ROADMAP.md Phase 6). nil if
// skillID isn't in DevilSkills.
//
// Correction (août 2026): this doc comment used to flag a "known gap" here
// -- HeroSkill.UnmarshalJSON only restoring Shallow (the ID) on load, and
// nothing re-resolving a Devil skill ID (as opposed to a D2 skills.txt one)
// back into a real SkillRecord afterwards. That's since been fixed:
// HeroStateFactory.LoadHeroState calls this very function for any persisted
// Shallow.SkillID found in DevilSkills, restoring a real skill rather than
// leaving it nil -- see that function's own "Bug fix" comment, and
// ROADMAP.md's "Round-trip complet vérifié" entry for the regression test.
func NewDevilHeroSkill(skillID int) *HeroSkill {
	def, ok := DevilSkills[skillID]
	if !ok {
		return nil
	}

	// Correction (août 2026): SkillPage/SkillColumn/SkillRow used to be
	// left at their zero value here -- every single Devil skill landed at
	// the same (page 0, column 0, row 0), which is a real problem, not a
	// cosmetic one: d2game/d2player/skilltree.go's setTab() only shows an
	// icon whose SkillPage matches the open tab (1/2/3), so with every
	// skill stuck at SkillPage 0 the skill tree panel would render every
	// tab completely empty. SkillGridPosition (this package) derives a
	// real, distinct, deterministic position from Tree/RequiredLevel/ID
	// instead. Leftskill/Passive/ListRow were similarly left at their zero
	// value (false/false/0): skill_select_panel.go reads Leftskill to
	// decide what the *left*-click popup shows, so with it always false no
	// Devil skill was ever offered there -- only the right-click popup
	// worked. Devil's design draws no left/right distinction between
	// skills, so every non-passive skill is eligible for both; Passive
	// mirrors DevilSkillDef.Passive so the panel correctly excludes the 5
	// always-on skills that are never meant to be equipped/cast at all.
	page, column, row := SkillGridPosition(skillID)

	return &HeroSkill{
		SkillRecord: &d2records.SkillRecord{
			ID:        skillID,
			Charclass: "",
			Leftskill: true,
			Passive:   def.Passive,
		},
		SkillDescriptionRecord: &d2records.SkillDescriptionRecord{
			IconCel:     0,
			SkillPage:   page,
			SkillColumn: column,
			SkillRow:    row,
			ListRow:     page, // group the equip popup by tree, same as the skill tree tabs
		},
		SkillPoints: 1,
		Shallow:     &shallowHeroSkill{SkillID: skillID, SkillPoints: 1},
	}
}

// maitriseElementaireDamagePercentPerPoint is the placeholder magnitude for
// Maîtrise élémentaire's "+% dégâts élémentaires par point"
// (devil_game_design_reference.md §7) -- the design gives no number, same
// practice as amplificationDamagePercent/eveilDuNexusHealPercent in
// game_server.go.
const maitriseElementaireDamagePercentPerPoint = 5

// MaitriseElementaireDamagePercent returns the total elemental-damage bonus
// percent for a caster with the given number of points invested in
// Maîtrise élémentaire (0 for none invested).
func MaitriseElementaireDamagePercent(points int) int {
	return points * maitriseElementaireDamagePercentPerPoint
}

// resonanceMagiqueDamagePercentPerPoint is the placeholder magnitude for
// Résonance magique's "chaque sort lancé augmente les dégâts du suivant
// (+% temporaire)" (devil_game_design_reference.md §7) -- the design gives
// no number, same practice as maitriseElementaireDamagePercentPerPoint.
const resonanceMagiqueDamagePercentPerPoint = 5

// ResonanceMagiqueDamagePercent returns the damage bonus percent Résonance
// magique grants the next eligible cast, for a caster with the given
// number of points invested (0 for none invested).
func ResonanceMagiqueDamagePercent(points int) int {
	return points * resonanceMagiqueDamagePercentPerPoint
}

// regenerationAccellereePercentPerPoint is the placeholder magnitude for
// Régénération accélérée's "Augmente la vitesse de régénération du mana"
// (devil_game_design_reference.md §7) -- the design gives no number, same
// practice as maitriseElementaireDamagePercentPerPoint.
const regenerationAccellereePercentPerPoint = 10

// RegenerationAccellereePercent returns the mana-regen-speed bonus percent
// for a caster with the given number of points invested in Régénération
// accélérée (0 for none invested).
func RegenerationAccellereePercent(points int) int {
	return points * regenerationAccellereePercentPerPoint
}

// NewTraitDeFeuSkill returns a HeroSkill for "Trait de feu" specifically --
// kept as a thin wrapper since character creation (select_hero_class.go)
// hardcodes everyone's starting skill to it.
func NewTraitDeFeuSkill() *HeroSkill {
	return NewDevilHeroSkill(SkillTraitDeFeu)
}

// TraitDeFeuSynergyTargets are the skills Trait de feu's own invested points
// boost (devil_game_design_reference.md §7 "Règles des synergies": "chaque
// point dans Trait de feu augmente les dégâts de Boule de feu et Météore").
//
// nolint:gochecknoglobals // a read-only lookup table, not mutable shared
// state -- flagged now that golangci-lint actually runs (août 2026).
var TraitDeFeuSynergyTargets = map[int]bool{
	SkillBouleDeFeu: true,
	SkillMeteore:    true,
}

// traitDeFeuSynergyPercentPerPoint is the placeholder magnitude for the
// synergy above -- the design names the rule but gives no number, same
// practice as maitriseElementaireDamagePercentPerPoint.
const traitDeFeuSynergyPercentPerPoint = 3

// TraitDeFeuSynergyDamagePercent returns the damage bonus percent Trait de
// feu's own invested points grant to Boule de feu/Météore, for a caster with
// the given number of points invested in Trait de feu (0 for none invested).
func TraitDeFeuSynergyDamagePercent(points int) int {
	return points * traitDeFeuSynergyPercentPerPoint
}

// bouclierDeManaSynergyPercentPerPoint is the placeholder magnitude for
// Bouclier de mana's own synergy (§7: "chaque point dans Bouclier de mana
// augmente l'absorption de Armure de glace") -- same practice as
// traitDeFeuSynergyPercentPerPoint.
const bouclierDeManaSynergyPercentPerPoint = 2

// BouclierDeManaSynergyReductionPercent returns the extra damage-reduction
// percent Bouclier de mana's own invested points grant to Armure de glace's
// own flat reduction, for a caster with the given number of points invested
// in Bouclier de mana (0 for none invested).
func BouclierDeManaSynergyReductionPercent(points int) int {
	return points * bouclierDeManaSynergyPercentPerPoint
}
