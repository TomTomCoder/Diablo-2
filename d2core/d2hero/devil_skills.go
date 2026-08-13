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
// Known gap: HeroSkill.UnmarshalJSON only restores Shallow (the ID) on
// load, and whatever re-resolves that ID into SkillRecord/
// SkillDescriptionRecord after loading a save only knows Diablo 2's own
// skills.txt today. A freshly created/learned skill works for the current
// session; surviving a save/load round-trip needs that resolution path
// extended (ROADMAP.md Phase 1, "Sauvegarde").
func NewDevilHeroSkill(skillID int) *HeroSkill {
	if _, ok := DevilSkills[skillID]; !ok {
		return nil
	}

	return &HeroSkill{
		SkillRecord: &d2records.SkillRecord{
			ID:        skillID,
			Charclass: "",
		},
		SkillDescriptionRecord: &d2records.SkillDescriptionRecord{
			IconCel: 0,
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
