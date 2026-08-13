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

// SkillEclairEnChaine is Devil's own skill ID for "Éclair en chaîne"
// (Élémentalisme, devil_game_design_reference.md §7): "Foudre qui rebondit
// sur 3 cibles". Its multi-target hit is resolved by resolveChainHit in
// game_server.go -- see skillEclairEnChaineTargets/eclairEnChaineChainRadius.
const SkillEclairEnChaine = SkillEclatDeGlace + 1

// SkillNovaDeGivre is Devil's own skill ID for "Nova de givre"
// (Élémentalisme, devil_game_design_reference.md §7): "Explosion de froid en
// zone autour du Mage". Unlike the other Élémentalisme spells, it's centered
// on the caster rather than a targeted/nearest NPC -- see resolveNovaHit in
// game_server.go.
const SkillNovaDeGivre = SkillEclairEnChaine + 1

// SkillBouleDeFeu is Devil's own skill ID for "Boule de feu" (Élémentalisme,
// devil_game_design_reference.md §7): "Projectile AoE, dégâts feu élevés".
// Like Nova de givre it's an AoE hit (see resolveAoeHit in game_server.go),
// but centered on the cast's targeted position rather than the caster.
const SkillBouleDeFeu = SkillNovaDeGivre + 1

// SkillTempeteStatique is Devil's own skill ID for "Tempête statique"
// (Élémentalisme, devil_game_design_reference.md §7): "Invoque une zone
// d'éclair persistante".
//
// ponytail: resolved as a single instant AoE hit at the targeted position
// (same resolveAoeHit path as Boule de feu) rather than an actual
// persistent, re-ticking zone -- no zone/duration entity exists yet.
// Upgrade path: a per-cast expiry tracked the same way lastCastAt/
// lastMonsterAttackAt are, re-running resolveAoeHit on that position every
// aiTickInterval until it lapses.
const SkillTempeteStatique = SkillBouleDeFeu + 1

// SkillChampStatique is Devil's own skill ID for "Champ statique" (Arcane,
// devil_game_design_reference.md §7): "Réduit la vie de toutes les entités à
// l'écran d'un % fixe". Devil's first Arcane skill -- percent-of-current-HP
// damage rather than base_sort scaling, so BaseSortDamage is unused (0) for
// it. See resolveChampStatiqueHit in game_server.go.
const SkillChampStatique = SkillTempeteStatique + 1

// SkillTelekinesie is Devil's own skill ID for "Télékinésie" (Arcane,
// devil_game_design_reference.md §7): "Repousse les entités, interaction
// avec les objets à distance". Only the knockback half is modeled --
// see resolveTelekinesieHit in game_server.go and d2mapentity.NPC.Knockback.
// The "interaction avec les objets à distance" half needs the item/ground-
// interaction system, which doesn't exist yet (ROADMAP.md Phase 5).
const SkillTelekinesie = SkillChampStatique + 1

// SkillRalentissement is Devil's own skill ID for "Ralentissement" (Arcane,
// devil_game_design_reference.md §7): "Zone qui réduit la vitesse des
// entités de 50%". Deals no damage (BaseSortDamage 0) -- its whole value is
// the slow, applied to every killable NPC in the targeted area via the same
// d2mapentity.NPC.ApplySlow used by Éclat de glace. See
// resolveRalentissementHit in game_server.go.
const SkillRalentissement = SkillTelekinesie + 1

// DevilSkills is the registry of Devil's own skill data, keyed by ID.
//
// ponytail: a handful of entries instead of the design's full 30 -- this
// phase is about proving the data model (tier gating, synergies, a status
// effect, multi-target/AoE hits) works, not filling it in. Add the rest of
// Élémentalisme next (ROADMAP.md Phase 2), then Arcane/Ésotérisme once
// their prerequisite mechanics (group control, ally summons) exist.
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
	SkillEclairEnChaine: {
		ID:             SkillEclairEnChaine,
		Name:           "Éclair en chaîne",
		Tree:           TreeElementalisme,
		RequiredLevel:  6, // tier 2 of Élémentalisme, per the design's level table
		BaseSortDamage: 5,
		ManaCost:       6,
	},
	SkillNovaDeGivre: {
		ID:             SkillNovaDeGivre,
		Name:           "Nova de givre",
		Tree:           TreeElementalisme,
		RequiredLevel:  6, // tier 2, alongside Éclair en chaîne
		BaseSortDamage: 4, // same cold-family trade-off as Éclat de glace: lower damage, an AoE footprint instead
		ManaCost:       8, // costliest tier-2 spell -- it hits every killable NPC in range, not just one
	},
	SkillBouleDeFeu: {
		ID:             SkillBouleDeFeu,
		Name:           "Boule de feu",
		Tree:           TreeElementalisme,
		RequiredLevel:  12, // tier 3, per the design's level table
		BaseSortDamage: 10, // "dégâts feu élevés" -- highest base_sort of any Élémentalisme spell so far
		ManaCost:       10,
	},
	SkillTempeteStatique: {
		ID:             SkillTempeteStatique,
		Name:           "Tempête statique",
		Tree:           TreeElementalisme,
		RequiredLevel:  12, // tier 3, alongside Boule de feu
		BaseSortDamage: 6,  // lightning-family: between Éclair en chaîne and Boule de feu
		ManaCost:       9,
	},
	SkillChampStatique: {
		ID:            SkillChampStatique,
		Name:          "Champ statique",
		Tree:          TreeArcane,
		RequiredLevel: 1,
		ManaCost:      12, // costliest tier-1 spell -- it hits every killable NPC on the map, not just one
	},
	SkillTelekinesie: {
		ID:            SkillTelekinesie,
		Name:          "Télékinésie",
		Tree:          TreeArcane,
		RequiredLevel: 1,
		ManaCost:      5, // cheap utility/control spell -- no direct damage
	},
	SkillRalentissement: {
		ID:            SkillRalentissement,
		Name:          "Ralentissement",
		Tree:          TreeArcane,
		RequiredLevel: 6, // tier 2, per the design's level table
		ManaCost:      7,
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
