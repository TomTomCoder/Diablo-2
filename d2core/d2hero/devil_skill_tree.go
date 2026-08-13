package d2hero

// SkillTree identifies which of Devil's three skill trees a skill belongs
// to (devil_game_design_reference.md §7).
type SkillTree int

// Devil's three skill trees.
const (
	TreeElementalisme SkillTree = iota // dégâts directs
	TreeArcane                         // contrôle/amplification
	TreeEsoterisme                     // défense/survie/invocations
)

// DevilSkillDef is Devil's own skill data model. Deliberately separate from
// Diablo 2's skills.txt/SkillRecord schema: different mana-cost shape, and a
// simpler synergy model (one target skill per point invested, vs. D2's
// often multi-target synergies) -- see ROADMAP.md Phase 2 for why reusing
// D2's columns wasn't a good fit.
type DevilSkillDef struct {
	ID             int
	Name           string
	Tree           SkillTree
	RequiredLevel  int // "palier" -- the character level needed to spend a point here
	BaseSortDamage int
	ManaCost       int

	// SynergyTargetID/SynergyPercent: each point invested in this skill adds
	// SynergyPercent% bonus damage to the skill with ID SynergyTargetID.
	// SynergyTargetID == 0 means this skill has no synergy.
	SynergyTargetID int
	SynergyPercent  int
}

// SkillEclatDeGlace is Devil's own skill ID for "Éclat de glace"
// (Élémentalisme, devil_game_design_reference.md §7): "Projectile froid,
// ralentit la cible". Its slow effect lands on d2mapentity.NPC via
// ApplySlow -- see resolveMeleeHit in game_server.go.
//
// ponytail: same namespacing reasoning as SkillTraitDeFeu -- numbered right
// after it, still nowhere near a real D2 skills.txt ID range.
const SkillEclatDeGlace = SkillTraitDeFeu + 1

// DevilSkills is the registry of Devil's own skill data, keyed by ID.
//
// ponytail: two entries instead of the design's full 30 -- this phase is
// about proving the data model (tier gating, synergies, and now a status
// effect) works, not filling it in. Add the rest of Élémentalisme next
// (ROADMAP.md Phase 2), then Arcane/Ésotérisme once their prerequisite
// mechanics (group control, ally summons) exist.
var DevilSkills = map[int]*DevilSkillDef{
	SkillTraitDeFeu: {
		ID:             SkillTraitDeFeu,
		Name:           "Trait de feu",
		Tree:           TreeElementalisme,
		RequiredLevel:  1,
		BaseSortDamage: 6,
		ManaCost:       3,
	},
	SkillEclatDeGlace: {
		ID:             SkillEclatDeGlace,
		Name:           "Éclat de glace",
		Tree:           TreeElementalisme,
		RequiredLevel:  1,
		BaseSortDamage: 4, // lower base damage than Trait de feu -- its value is the slow, not raw damage
		ManaCost:       4,
	},
}

// CanLearnSkill reports whether a hero of the given level may spend a
// point in skillID, per its RequiredLevel ("palier" gating,
// devil_game_design_reference.md §7). False for an unknown skill ID.
func CanLearnSkill(skillID, heroLevel int) bool {
	def, ok := DevilSkills[skillID]
	if !ok {
		return false
	}

	return heroLevel >= def.RequiredLevel
}

// SkillBaseSortDamage returns skillID's base_sort damage, or fallback if
// it's not in DevilSkills.
func SkillBaseSortDamage(skillID, fallback int) int {
	if def, ok := DevilSkills[skillID]; ok {
		return def.BaseSortDamage
	}

	return fallback
}

// SkillManaCost returns skillID's mana cost, or fallback if it's not in
// DevilSkills.
func SkillManaCost(skillID, fallback int) int {
	if def, ok := DevilSkills[skillID]; ok {
		return def.ManaCost
	}

	return fallback
}
