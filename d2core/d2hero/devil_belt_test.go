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

	if err := hero.AddPotionToBelt(ItemPotionDeMana); err != nil {
		t.Fatalf("expected AddPotionToBelt to succeed, got %v", err)
	}

	if hero.Belt[0] != ItemPotionDeMana {
		t.Errorf("expected the potion in slot 0, got %q", hero.Belt[0])
	}

	if err := hero.AddPotionToBelt(ItemPotionDeMana); err != nil {
		t.Fatalf("expected a second AddPotionToBelt to succeed, got %v", err)
	}

	if hero.Belt[1] != ItemPotionDeMana {
		t.Errorf("expected the second potion in slot 1, got %q", hero.Belt[1])
	}
}

func TestAddPotionToBeltFailsWhenFull(t *testing.T) {
	hero := &HeroState{}
	hero.InitBelt(1)

	if err := hero.AddPotionToBelt(ItemPotionDeMana); err != nil {
		t.Fatalf("expected the first AddPotionToBelt to succeed, got %v", err)
	}

	if err := hero.AddPotionToBelt(ItemPotionDeMana); err == nil {
		t.Error("expected AddPotionToBelt to fail once the belt is full")
	}
}

func TestUsePotionRestoresManaAndClearsSlot(t *testing.T) {
	hero := &HeroState{Stats: &HeroStatsState{Mana: 0, MaxMana: 100}}
	hero.InitBelt(2)

	if err := hero.AddPotionToBelt(ItemPotionDeMana); err != nil {
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
