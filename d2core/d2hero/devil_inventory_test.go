package d2hero

import "testing"

func TestInitStorageCreatesEmptySlots(t *testing.T) {
	hero := &HeroState{}
	hero.InitStorage()

	if len(hero.Inventory) != inventoryCapacity {
		t.Fatalf("expected %d inventory slots, got %d", inventoryCapacity, len(hero.Inventory))
	}

	if len(hero.Stash) != stashCapacity {
		t.Fatalf("expected %d stash slots, got %d", stashCapacity, len(hero.Stash))
	}

	for i, slot := range hero.Inventory {
		if slot != "" {
			t.Errorf("expected inventory slot %d to start empty, got %q", i, slot)
		}
	}

	for i, slot := range hero.Stash {
		if slot != "" {
			t.Errorf("expected stash slot %d to start empty, got %q", i, slot)
		}
	}
}

func TestAddToInventoryFillsFirstEmptySlot(t *testing.T) {
	hero := &HeroState{}
	hero.InitStorage()

	slot, err := hero.AddToInventory(ItemPendentifArcane)
	if err != nil {
		t.Fatalf("expected AddToInventory to succeed, got %v", err)
	}

	if slot != 0 || hero.Inventory[0] != ItemPendentifArcane {
		t.Errorf("expected the item in slot 0, got slot %d (%q)", slot, hero.Inventory[0])
	}

	slot, err = hero.AddToInventory(ItemAnneauDuDebut)
	if err != nil {
		t.Fatalf("expected a second AddToInventory to succeed, got %v", err)
	}

	if slot != 1 || hero.Inventory[1] != ItemAnneauDuDebut {
		t.Errorf("expected the second item in slot 1, got slot %d (%q)", slot, hero.Inventory[1])
	}
}

func TestAddToInventoryFailsWhenFull(t *testing.T) {
	hero := &HeroState{Inventory: make([]string, 1)}

	if _, err := hero.AddToInventory(ItemPendentifArcane); err != nil {
		t.Fatalf("expected the first AddToInventory to succeed, got %v", err)
	}

	if _, err := hero.AddToInventory(ItemAnneauDuDebut); err == nil {
		t.Error("expected AddToInventory to fail once the inventory is full")
	}
}

func TestRemoveFromInventoryClearsSlotAndReturnsItem(t *testing.T) {
	hero := &HeroState{Inventory: make([]string, 2)}

	slot, err := hero.AddToInventory(ItemPendentifArcane)
	if err != nil {
		t.Fatalf("expected AddToInventory to succeed, got %v", err)
	}

	removed, err := hero.RemoveFromInventory(slot)
	if err != nil {
		t.Fatalf("expected RemoveFromInventory to succeed, got %v", err)
	}

	if removed != ItemPendentifArcane {
		t.Errorf("expected the removed item code %q, got %q", ItemPendentifArcane, removed)
	}

	if hero.Inventory[slot] != "" {
		t.Errorf("expected the slot cleared after removal, got %q", hero.Inventory[slot])
	}
}

func TestRemoveFromInventoryFailsOnEmptySlot(t *testing.T) {
	hero := &HeroState{Inventory: make([]string, 2)}

	if _, err := hero.RemoveFromInventory(0); err == nil {
		t.Error("expected RemoveFromInventory to fail on an empty slot")
	}
}

func TestRemoveFromInventoryFailsOutOfRange(t *testing.T) {
	hero := &HeroState{Inventory: make([]string, 2)}

	if _, err := hero.RemoveFromInventory(5); err == nil {
		t.Error("expected RemoveFromInventory to fail for an out-of-range slot")
	}

	if _, err := hero.RemoveFromInventory(-1); err == nil {
		t.Error("expected RemoveFromInventory to fail for a negative slot index")
	}
}

// TestStashIsIndependentOfInventory is a regression test for the shared
// addToSlots/removeFromSlots helpers: they must operate on whichever slice
// is passed in, not accidentally share state between Inventory and Stash.
func TestStashIsIndependentOfInventory(t *testing.T) {
	hero := &HeroState{}
	hero.InitStorage()

	if _, err := hero.AddToInventory(ItemPendentifArcane); err != nil {
		t.Fatalf("expected AddToInventory to succeed, got %v", err)
	}

	if _, err := hero.AddToStash(ItemAnneauDuDebut); err != nil {
		t.Fatalf("expected AddToStash to succeed, got %v", err)
	}

	if hero.Inventory[0] != ItemPendentifArcane {
		t.Errorf("expected the inventory item untouched by the stash add, got %q", hero.Inventory[0])
	}

	if hero.Stash[0] != ItemAnneauDuDebut {
		t.Errorf("expected the stash item in its own slot 0, got %q", hero.Stash[0])
	}

	if _, err := hero.RemoveFromStash(0); err != nil {
		t.Fatalf("expected RemoveFromStash to succeed, got %v", err)
	}

	if hero.Inventory[0] != ItemPendentifArcane {
		t.Errorf("expected the inventory item untouched by the stash removal, got %q", hero.Inventory[0])
	}
}

func TestRemoveFromStashFailsOnEmptySlot(t *testing.T) {
	hero := &HeroState{}
	hero.InitStorage()

	if _, err := hero.RemoveFromStash(0); err == nil {
		t.Error("expected RemoveFromStash to fail on an empty slot")
	}
}
