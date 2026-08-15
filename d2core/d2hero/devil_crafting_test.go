package d2hero

import (
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2inventory"
)

func heroWithBatonApprenti(gold int) *HeroState {
	return &HeroState{
		Gold: gold,
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: ItemBatonApprenti},
		},
	}
}

func TestCraftUpgradesWeaponAndDeductsGold(t *testing.T) {
	hero := heroWithBatonApprenti(100)

	if err := hero.Craft(RecipeUpgradeBatonApprenti); err != nil {
		t.Fatalf("expected craft to succeed, got error: %v", err)
	}

	if hero.Equipment.RightHand.ItemCode != ItemBatonInitie {
		t.Errorf("expected weapon upgraded to %q, got %q", ItemBatonInitie, hero.Equipment.RightHand.ItemCode)
	}

	if want := 100 - 50; hero.Gold != want {
		t.Errorf("expected Gold %d after crafting, got %d", want, hero.Gold)
	}

	if !hero.HasDiscoveredRecipe(RecipeUpgradeBatonApprenti) {
		t.Error("expected the recipe marked discovered in the Codex after a successful craft")
	}
}

func TestCraftFailsWithoutEnoughGold(t *testing.T) {
	hero := heroWithBatonApprenti(10)

	if err := hero.Craft(RecipeUpgradeBatonApprenti); err == nil {
		t.Fatal("expected craft to fail with insufficient gold")
	}

	if hero.Equipment.RightHand.ItemCode != ItemBatonApprenti {
		t.Error("expected the weapon to remain unchanged after a failed craft")
	}

	if hero.Gold != 10 {
		t.Errorf("expected Gold untouched after a failed craft, got %d", hero.Gold)
	}

	if hero.HasDiscoveredRecipe(RecipeUpgradeBatonApprenti) {
		t.Error("expected a failed craft to not mark the recipe discovered")
	}
}

func TestCraftFailsWithWrongItemEquipped(t *testing.T) {
	hero := &HeroState{
		Gold: 100,
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: "something-else"},
		},
	}

	if err := hero.Craft(RecipeUpgradeBatonApprenti); err == nil {
		t.Fatal("expected craft to fail when the required input item isn't equipped")
	}

	if hero.Gold != 100 {
		t.Errorf("expected Gold untouched after a failed craft, got %d", hero.Gold)
	}
}

func TestCraftFailsWithNoWeaponEquipped(t *testing.T) {
	hero := &HeroState{Gold: 100}

	if err := hero.Craft(RecipeUpgradeBatonApprenti); err == nil {
		t.Fatal("expected craft to fail with no weapon equipped")
	}
}

func TestCraftFailsWithUnknownRecipe(t *testing.T) {
	hero := heroWithBatonApprenti(100)

	if err := hero.Craft("not-a-real-recipe"); err == nil {
		t.Fatal("expected craft to fail for an unknown recipe ID")
	}
}

// TestCraftConsumesFromInventoryWhenNotEquipped covers the fallback path:
// the input item sitting in the inventory rather than equipped.
func TestCraftConsumesFromInventoryWhenNotEquipped(t *testing.T) {
	hero := &HeroState{
		Gold:      100,
		Inventory: []string{"", ItemBatonApprenti, ""},
	}

	if err := hero.Craft(RecipeUpgradeBatonApprenti); err != nil {
		t.Fatalf("expected craft to succeed from inventory, got error: %v", err)
	}

	if hero.Inventory[1] != ItemBatonInitie {
		t.Errorf("expected slot 1 replaced with %q, got %q", ItemBatonInitie, hero.Inventory[1])
	}

	if want := 100 - 50; hero.Gold != want {
		t.Errorf("expected Gold %d after crafting, got %d", want, hero.Gold)
	}

	if !hero.HasDiscoveredRecipe(RecipeUpgradeBatonApprenti) {
		t.Error("expected the recipe marked discovered in the Codex after a successful craft")
	}
}

// TestCraftPrefersEquippedOverInventory proves the equipped slot is
// checked first: with the input item in both places, only the equipped
// one should be consumed.
func TestCraftPrefersEquippedOverInventory(t *testing.T) {
	hero := heroWithBatonApprenti(100)
	hero.Inventory = []string{ItemBatonApprenti}

	if err := hero.Craft(RecipeUpgradeBatonApprenti); err != nil {
		t.Fatalf("expected craft to succeed, got error: %v", err)
	}

	if hero.Equipment.RightHand.ItemCode != ItemBatonInitie {
		t.Errorf("expected the equipped weapon upgraded, got %q", hero.Equipment.RightHand.ItemCode)
	}

	if hero.Inventory[0] != ItemBatonApprenti {
		t.Errorf("expected the inventory copy left untouched, got %q", hero.Inventory[0])
	}
}

func TestCraftFailsWhenInputNeitherEquippedNorInInventory(t *testing.T) {
	hero := &HeroState{Gold: 100, Inventory: []string{"", ""}}

	if err := hero.Craft(RecipeUpgradeBatonApprenti); err == nil {
		t.Fatal("expected craft to fail when the input item is nowhere to be found")
	}

	if hero.Gold != 100 {
		t.Errorf("expected Gold untouched after a failed craft, got %d", hero.Gold)
	}
}
