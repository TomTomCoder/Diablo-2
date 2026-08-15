package d2hero

import "testing"

func TestInitBeltCreatesEmptySlots(t *testing.T) {
	hero := &HeroState{}
	hero.InitBelt(4)

	if len(hero.Belt) != 4 {
		t.Fatalf("expected 4 belt slots, got %d", len(hero.Belt))
	}

	for i, slot := range hero.Belt {
		if slot != "" {
			t.Errorf("expected slot %d to start empty, got %q", i, slot)
		}
	}
}

func TestAddPotionToBeltFillsFirstEmptySlot(t *testing.T) {
	hero := &HeroState{}
	hero.InitBelt(2)

	if _, err := hero.AddPotionToBelt(ItemPotionDeMana); err != nil {
		t.Fatalf("expected AddPotionToBelt to succeed, got %v", err)
	}

	if hero.Belt[0] != ItemPotionDeMana {
		t.Errorf("expected the potion in slot 0, got %q", hero.Belt[0])
	}

	if _, err := hero.AddPotionToBelt(ItemPotionDeMana); err != nil {
		t.Fatalf("expected a second AddPotionToBelt to succeed, got %v", err)
	}

	if hero.Belt[1] != ItemPotionDeMana {
		t.Errorf("expected the second potion in slot 1, got %q", hero.Belt[1])
	}
}

func TestAddPotionToBeltFailsWhenFull(t *testing.T) {
	hero := &HeroState{}
	hero.InitBelt(1)

	if _, err := hero.AddPotionToBelt(ItemPotionDeMana); err != nil {
		t.Fatalf("expected the first AddPotionToBelt to succeed, got %v", err)
	}

	if _, err := hero.AddPotionToBelt(ItemPotionDeMana); err == nil {
		t.Error("expected AddPotionToBelt to fail once the belt is full")
	}
}

func TestUsePotionRestoresManaAndClearsSlot(t *testing.T) {
	hero := &HeroState{Stats: &HeroStatsState{Mana: 0, MaxMana: 100}}
	hero.InitBelt(2)

	if _, err := hero.AddPotionToBelt(ItemPotionDeMana); err != nil {
		t.Fatalf("expected AddPotionToBelt to succeed, got %v", err)
	}

	if err := hero.UsePotion(0); err != nil {
		t.Fatalf("expected UsePotion to succeed, got %v", err)
	}

	if want := ItemManaRestoreAmount(ItemPotionDeMana); hero.Stats.Mana != want {
		t.Errorf("expected Mana restored to %d, got %d", want, hero.Stats.Mana)
	}

	if hero.Belt[0] != "" {
		t.Errorf("expected the slot cleared after use, got %q", hero.Belt[0])
	}
}

func TestUsePotionFailsOnEmptySlot(t *testing.T) {
	hero := &HeroState{Stats: &HeroStatsState{}}
	hero.InitBelt(2)

	if err := hero.UsePotion(0); err == nil {
		t.Error("expected UsePotion to fail on an empty slot")
	}
}

func TestUsePotionFailsOutOfRange(t *testing.T) {
	hero := &HeroState{Stats: &HeroStatsState{}}
	hero.InitBelt(2)

	if err := hero.UsePotion(5); err == nil {
		t.Error("expected UsePotion to fail for an out-of-range slot")
	}

	if err := hero.UsePotion(-1); err == nil {
		t.Error("expected UsePotion to fail for a negative slot index")
	}
}

func TestMoveToBeltMovesThePotionAndClearsTheInventorySlot(t *testing.T) {
	hero := &HeroState{Inventory: []string{ItemPotionDeMana}}
	hero.InitBelt(2)

	itemCode, beltSlot, err := hero.MoveToBelt(0)
	if err != nil {
		t.Fatalf("expected MoveToBelt to succeed, got %v", err)
	}

	if itemCode != ItemPotionDeMana {
		t.Errorf("expected the returned itemCode %q, got %q", ItemPotionDeMana, itemCode)
	}

	if hero.Inventory[0] != "" {
		t.Errorf("expected the inventory slot cleared, got %q", hero.Inventory[0])
	}

	if hero.Belt[beltSlot] != ItemPotionDeMana {
		t.Errorf("expected the potion in belt slot %d, got %q", beltSlot, hero.Belt[beltSlot])
	}
}

// TestMoveToBeltRejectsNonPotionItems is a regression test for the
// design's own restriction ("Ceinture : ... potions de mana et
// parchemins") -- a non-potion item (ItemManaRestoreAmount 0) must be
// rejected and the inventory left untouched, not silently accepted.
func TestMoveToBeltRejectsNonPotionItems(t *testing.T) {
	hero := &HeroState{Inventory: []string{ItemPendentifArcane}}
	hero.InitBelt(2)

	if _, _, err := hero.MoveToBelt(0); err == nil {
		t.Fatal("expected MoveToBelt to reject a non-potion item")
	}

	if hero.Inventory[0] != ItemPendentifArcane {
		t.Errorf("expected the item put back in the inventory, got %q", hero.Inventory[0])
	}
}

// TestMoveToBeltRollsBackWhenBeltIsFull mirrors
// TestMoveToStashRollsBackWhenStashIsFull (devil_inventory_test.go) for
// the same "don't lose the item" property.
func TestMoveToBeltRollsBackWhenBeltIsFull(t *testing.T) {
	hero := &HeroState{Inventory: []string{ItemPotionDeMana}}
	hero.InitBelt(1)
	hero.Belt[0] = ItemPotionDeMana // already full

	if _, _, err := hero.MoveToBelt(0); err == nil {
		t.Fatal("expected MoveToBelt to fail when the belt is full")
	}

	if hero.Inventory[0] != ItemPotionDeMana {
		t.Errorf("expected the item put back in the inventory, got %q", hero.Inventory[0])
	}
}

func TestMoveFromBeltMovesThePotionAndClearsTheBeltSlot(t *testing.T) {
	hero := &HeroState{Inventory: make([]string, 1)}
	hero.InitBelt(1)
	hero.Belt[0] = ItemPotionDeMana

	itemCode, inventorySlot, err := hero.MoveFromBelt(0)
	if err != nil {
		t.Fatalf("expected MoveFromBelt to succeed, got %v", err)
	}

	if itemCode != ItemPotionDeMana {
		t.Errorf("expected the returned itemCode %q, got %q", ItemPotionDeMana, itemCode)
	}

	if hero.Belt[0] != "" {
		t.Errorf("expected the belt slot cleared, got %q", hero.Belt[0])
	}

	if hero.Inventory[inventorySlot] != ItemPotionDeMana {
		t.Errorf("expected the potion in inventory slot %d, got %q", inventorySlot, hero.Inventory[inventorySlot])
	}
}

func TestMoveFromBeltRollsBackWhenInventoryIsFull(t *testing.T) {
	hero := &HeroState{Inventory: []string{ItemPendentifArcane}} // already full, capacity 1
	hero.InitBelt(1)
	hero.Belt[0] = ItemPotionDeMana

	if _, _, err := hero.MoveFromBelt(0); err == nil {
		t.Fatal("expected MoveFromBelt to fail when the inventory is full")
	}

	if hero.Belt[0] != ItemPotionDeMana {
		t.Errorf("expected the potion put back in the belt, got %q", hero.Belt[0])
	}
}

func TestMoveFromBeltFailsOnEmptySlot(t *testing.T) {
	hero := &HeroState{}
	hero.InitBelt(1)

	if _, _, err := hero.MoveFromBelt(0); err == nil {
		t.Error("expected MoveFromBelt to fail on an empty belt slot")
	}
}
