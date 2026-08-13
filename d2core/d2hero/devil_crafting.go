package d2hero

import "errors"

// DevilCraftingRecipe is Devil's own Cube de Nexus recipe -- consumes
// InputQuantity of InputItemCode plus InputGoldCost gold, produces
// OutputItemCode (devil_game_design_reference.md §8: "Le Cube de Nexus
// permet de transformer et améliorer les objets").
//
// ponytail: InputQuantity/InputItemCode name a single item type, not a list
// -- Devil doesn't have a multi-ingredient recipe yet, and its whole item
// set today is one weapon (DevilItems), so there's nothing to combine.
type DevilCraftingRecipe struct {
	ID             string
	Name           string
	InputItemCode  string
	InputQuantity  int
	InputGoldCost  int
	OutputItemCode string
}

// RecipeUpgradeBatonApprenti is Devil's own recipe ID for upgrading a
// Bâton de l'Apprenti into a Bâton de l'Initié -- the Cube de Nexus'
// first real recipe. The design gates its introduction behind a
// mandatory Région II quest, which doesn't exist yet (ROADMAP.md); this
// recipe is craftable unconditionally for now.
const RecipeUpgradeBatonApprenti = "dvl_recipe_upgrade_baton_apprenti"

// DevilCraftingRecipes is the registry of Devil's own Cube de Nexus
// recipes, keyed by ID.
//
// ponytail: one recipe instead of a full crafting system -- proves the
// engine (validate an equipped input item + gold, consume both, produce an
// output) works end to end with real Devil items, not just a placeholder.
// The design's own "Codex progressif" (recipes revealed as discovered) is a
// UI-layer concern, not modeled here.
var DevilCraftingRecipes = map[string]*DevilCraftingRecipe{
	RecipeUpgradeBatonApprenti: {
		ID:             RecipeUpgradeBatonApprenti,
		Name:           "Bâton de l'Apprenti → Bâton de l'Initié",
		InputItemCode:  ItemBatonApprenti,
		InputQuantity:  1,
		InputGoldCost:  50,
		OutputItemCode: ItemBatonInitie,
	},
}

// Craft attempts recipeID against h: the hero's own RightHand weapon must
// match the recipe's input item code, and Gold must cover its cost. On
// success, Gold is debited, the weapon's ItemCode becomes the recipe's
// output -- resolveAttackDamage's existing ItemFireDamagePercent lookup
// picks up the new item's bonuses automatically, no further wiring needed
// -- and recipeID is marked discovered in h's Codex (DiscoverRecipe),
// devil_game_design_reference.md §8's "chaque recette découverte (par
// quête ou utilisation)".
//
// ponytail: only checks/mutates the equipped weapon slot, not a real
// inventory/stash -- Devil doesn't have an inventory grid model yet
// (ROADMAP.md Phase 5), so "does the player have this item" can only mean
// "is it equipped" for now. InputQuantity is unused for the same reason
// (nothing to stack N of outside a slot).
func (h *HeroState) Craft(recipeID string) error {
	recipe, ok := DevilCraftingRecipes[recipeID]
	if !ok {
		return errors.New("unknown recipe")
	}

	if h.Equipment.RightHand == nil || h.Equipment.RightHand.ItemCode != recipe.InputItemCode {
		return errors.New("required input item not equipped")
	}

	if h.Gold < recipe.InputGoldCost {
		return errors.New("not enough gold")
	}

	h.Gold -= recipe.InputGoldCost
	h.Equipment.RightHand.ItemCode = recipe.OutputItemCode
	h.DiscoverRecipe(recipeID)

	return nil
}
