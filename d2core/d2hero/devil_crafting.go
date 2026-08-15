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
//
// nolint:gochecknoglobals // a read-only registry, not mutable shared state
// -- flagged now that golangci-lint actually runs (août 2026).
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

// consumeFromInventory looks for at least quantity slots in h.Inventory
// holding itemCode; if found, clears all of them and returns the index of
// one of them (for the caller to place a craft's output into, reusing the
// slot rather than needing a separate empty one). Leaves the inventory
// untouched and returns ok: false otherwise -- including for a
// non-positive quantity, which has no sensible "consume" behavior.
func (h *HeroState) consumeFromInventory(itemCode string, quantity int) (outputSlot int, ok bool) {
	if quantity <= 0 {
		return 0, false
	}

	matches := make([]int, 0, quantity)

	for i, code := range h.Inventory {
		if code == itemCode {
			matches = append(matches, i)
			if len(matches) == quantity {
				break
			}
		}
	}

	if len(matches) < quantity {
		return 0, false
	}

	for _, i := range matches {
		h.Inventory[i] = ""
	}

	return matches[0], true
}

// Craft attempts recipeID against h: its input item must be found either
// equipped (RightHand) or in h.Inventory (InputQuantity copies, see
// consumeFromInventory), and Gold must cover its cost. On success, Gold is
// debited, the input is replaced in place by the recipe's output -- the
// equipped weapon's ItemCode is swapped directly (resolveAttackDamage's
// existing ItemFireDamagePercent lookup picks up the new item's bonuses
// automatically, no further wiring needed), or one consumed inventory slot
// becomes the output -- and recipeID is marked discovered in h's Codex
// (DiscoverRecipe), devil_game_design_reference.md §8's "chaque recette
// découverte (par quête ou utilisation)".
//
// The equipped slot is checked first, matching Craft's original
// equipped-only behavior exactly when the item happens to be equipped;
// the inventory is only consulted as a fallback.
func (h *HeroState) Craft(recipeID string) error {
	recipe, ok := DevilCraftingRecipes[recipeID]
	if !ok {
		return errors.New("unknown recipe")
	}

	if h.Gold < recipe.InputGoldCost {
		return errors.New("not enough gold")
	}

	if h.Equipment.RightHand != nil && h.Equipment.RightHand.ItemCode == recipe.InputItemCode {
		h.Gold -= recipe.InputGoldCost
		h.Equipment.RightHand.ItemCode = recipe.OutputItemCode
		h.DiscoverRecipe(recipeID)

		return nil
	}

	if slot, found := h.consumeFromInventory(recipe.InputItemCode, recipe.InputQuantity); found {
		h.Gold -= recipe.InputGoldCost
		h.Inventory[slot] = recipe.OutputItemCode
		h.DiscoverRecipe(recipeID)

		return nil
	}

	return errors.New("required input item not equipped or in inventory")
}
