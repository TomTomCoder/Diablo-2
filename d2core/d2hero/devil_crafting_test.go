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
