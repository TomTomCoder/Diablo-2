package d2hero

// ItemBatonApprenti is Devil's own item code for "Bâton de l'Apprenti"
// (devil_mage_character_design.md §5), the Mage's starting weapon.
//
// ponytail: a string code namespaced with "dvl_" rather than a bare name,
// so it can never collide with a Diablo 2 items.txt code loaded from the
// player's own MPQ files at runtime (same reasoning as SkillTraitDeFeu's
// numbering -- see that constant's comment).
const ItemBatonApprenti = "dvl_baton_apprenti"

// ItemBatonInitie is Devil's own item code for "Bâton de l'Initié", the
// Cube de Nexus upgrade of Bâton de l'Apprenti (devil_game_design_reference.md
// §8: "Le Cube de Nexus permet de transformer et améliorer les objets").
// See devil_crafting.go.
const ItemBatonInitie = "dvl_baton_initie"

// DevilItemDef is Devil's own item data model -- separate from Diablo 2's
// items.txt (Devil's equipment is magic-only: staffs, orbs, robes,
// amulets, rings, grimoires -- not weapons/armor in D2's sense). See
// ROADMAP.md Phase 5.
type DevilItemDef struct {
	Code        string
	Name        string
	EnergyBonus int

	// FireDamagePercent is a damage *modifier* (e.g. "+10% dégâts Feu"),
	// applied on top of a cast's base_sort/Energy damage -- not a source of
	// base damage itself. See resolveAttackDamage's caller in game_server.go.
	FireDamagePercent int
}

// DevilItems is the registry of Devil's own item data, keyed by code.
//
// ponytail: one entry (the starting staff) instead of the design's full
// item set (§5, §8) -- this proves an equipped item can modify combat
// damage at all, it doesn't fill in the item system.
var DevilItems = map[string]*DevilItemDef{
	ItemBatonApprenti: {
		Code:              ItemBatonApprenti,
		Name:              "Bâton de l'Apprenti",
		EnergyBonus:       5,
		FireDamagePercent: 10,
	},
	ItemBatonInitie: {
		Code:              ItemBatonInitie,
		Name:              "Bâton de l'Initié",
		EnergyBonus:       8,  // better than Bâton de l'Apprenti's 5 -- it's an upgrade
		FireDamagePercent: 15, // better than Bâton de l'Apprenti's 10
	},
}

// ItemFireDamagePercent returns itemCode's Fire damage modifier percent
// (0 if it's not in DevilItems or grants no such bonus).
func ItemFireDamagePercent(itemCode string) int {
	if def, ok := DevilItems[itemCode]; ok {
		return def.FireDamagePercent
	}

	return 0
}

// ItemEnergyBonus returns itemCode's Energy bonus (0 if it's not in
// DevilItems or grants no such bonus).
func ItemEnergyBonus(itemCode string) int {
	if def, ok := DevilItems[itemCode]; ok {
		return def.EnergyBonus
	}

	return 0
}
