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

// NewTraitDeFeuSkill returns a HeroSkill for "Trait de feu" with just
// enough real data to render and cast without touching Diablo 2's own
// skill data at all: Charclass "" resolves to the game's built-in generic
// skill icon sheet (used for e.g. the default "Attack" skill), since no
// Devil-specific skill icon exists yet (ROADMAP.md Phase 6).
//
// Known gap: HeroSkill.UnmarshalJSON only restores Shallow (the ID) on
// load, and whatever re-resolves that ID into SkillRecord/
// SkillDescriptionRecord after loading a save only knows Diablo 2's own
// skills.txt today. A freshly created character works for the current
// session; surviving a save/load round-trip needs that resolution path
// extended (ROADMAP.md Phase 1, "Sauvegarde").
func NewTraitDeFeuSkill() *HeroSkill {
	return &HeroSkill{
		SkillRecord: &d2records.SkillRecord{
			ID:        SkillTraitDeFeu,
			Charclass: "",
		},
		SkillDescriptionRecord: &d2records.SkillDescriptionRecord{
			IconCel: 0,
		},
		SkillPoints: 1,
		Shallow:     &shallowHeroSkill{SkillID: SkillTraitDeFeu, SkillPoints: 1},
	}
}
