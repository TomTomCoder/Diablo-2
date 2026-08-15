package d2hero

import "testing"

func TestUseGlypheDOubliOnSkillConsumesItemAndRespecsThatSkill(t *testing.T) {
	hero := newLearnableHeroState(6, 2)
	hero.Inventory = []string{ItemGlypheDOubli, ""}

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("test setup: LearnSkill(Trait de feu) failed: %v", err)
	}

	if err := hero.LearnSkill(SkillEclatDeGlace); err != nil {
		t.Fatalf("test setup: LearnSkill(Éclat de glace) failed: %v", err)
	}

	refunded, err := hero.UseGlypheDOubliOnSkill(SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("expected UseGlypheDOubliOnSkill to succeed, got %v", err)
	}

	if refunded != 1 {
		t.Errorf("expected 1 point refunded, got %d", refunded)
	}

	if _, known := hero.Skills[SkillTraitDeFeu]; known {
		t.Error("expected Trait de feu forgotten")
	}

	if _, known := hero.Skills[SkillEclatDeGlace]; !known {
		t.Error("expected Éclat de glace left untouched")
	}

	if hero.Inventory[0] != "" {
		t.Errorf("expected the glyph consumed (slot cleared), got %q", hero.Inventory[0])
	}
}

func TestUseGlypheDOubliOnSkillFailsWithoutTheItem(t *testing.T) {
	hero := newLearnableHeroState(6, 1)
	hero.Inventory = []string{"", ""}

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("test setup: LearnSkill failed: %v", err)
	}

	if _, err := hero.UseGlypheDOubliOnSkill(SkillTraitDeFeu); err == nil {
		t.Fatal("expected UseGlypheDOubliOnSkill to fail without the glyph in inventory")
	}

	if _, known := hero.Skills[SkillTraitDeFeu]; !known {
		t.Error("expected Skills untouched when UseGlypheDOubliOnSkill fails")
	}
}

// TestUseGlypheDOubliOnSkillRestoresItemToItsOriginalSlotOnFailure is a
// regression test for a real bug caught before it shipped: an earlier
// draft restored a failed respec's consumed glyph via append(h.Inventory,
// ...) instead of back into its own slot -- since Inventory is a
// fixed-size slice (40 slots, InitStorage) that
// d2game/d2player/devilInventoryPanel relies on never changing length,
// that would have silently grown it past capacity on every failed use.
func TestUseGlypheDOubliOnSkillRestoresItemToItsOriginalSlotOnFailure(t *testing.T) {
	hero := newLearnableHeroState(6, 0)
	hero.Inventory = []string{"", ItemGlypheDOubli, ""}

	// SkillTraitDeFeu was never learned, so RespecSingleSkill fails.
	if _, err := hero.UseGlypheDOubliOnSkill(SkillTraitDeFeu); err == nil {
		t.Fatal("test setup: expected UseGlypheDOubliOnSkill to fail (skill never learned)")
	}

	if len(hero.Inventory) != 3 {
		t.Fatalf("expected Inventory to keep its original length 3, got %d", len(hero.Inventory))
	}

	if hero.Inventory[1] != ItemGlypheDOubli {
		t.Errorf("expected the glyph restored to its original slot (1), got %q at slot 1 (inventory: %v)",
			hero.Inventory[1], hero.Inventory)
	}
}

func TestUseGlypheDOubliOnAttributeConsumesItemAndRefundsThatAttribute(t *testing.T) {
	hero := newLearnableHeroState(1, 0)
	hero.Stats.StatsPoints = 2
	hero.Stats.Strength = 10
	hero.Stats.Dexterity = 10
	hero.Inventory = []string{ItemGlypheDOubli, ""}

	if err := hero.Stats.SpendAttributePoint(AttributeStrength); err != nil {
		t.Fatalf("test setup: SpendAttributePoint(Strength) failed: %v", err)
	}

	if err := hero.Stats.SpendAttributePoint(AttributeDexterity); err != nil {
		t.Fatalf("test setup: SpendAttributePoint(Dexterity) failed: %v", err)
	}

	if err := hero.UseGlypheDOubliOnAttribute(AttributeStrength); err != nil {
		t.Fatalf("expected UseGlypheDOubliOnAttribute to succeed, got %v", err)
	}

	if hero.Stats.Strength != 10 {
		t.Errorf("expected Strength restored to its base of 10, got %d", hero.Stats.Strength)
	}

	if hero.Stats.Dexterity != 11 {
		t.Errorf("expected Dexterity to remain untouched at 11, got %d", hero.Stats.Dexterity)
	}

	if hero.Inventory[0] != "" {
		t.Errorf("expected the glyph consumed (slot cleared), got %q", hero.Inventory[0])
	}
}

func TestUseGlypheDOubliOnAttributeFailsWithoutTheItem(t *testing.T) {
	hero := newLearnableHeroState(1, 0)
	hero.Stats.StatsPoints = 1
	hero.Inventory = []string{"", ""}

	if err := hero.Stats.SpendAttributePoint(AttributeStrength); err != nil {
		t.Fatalf("test setup: SpendAttributePoint failed: %v", err)
	}

	if err := hero.UseGlypheDOubliOnAttribute(AttributeStrength); err == nil {
		t.Fatal("expected UseGlypheDOubliOnAttribute to fail without the glyph in inventory")
	}

	if hero.Stats.StrengthSpent != 1 {
		t.Errorf("expected the spent point untouched when UseGlypheDOubliOnAttribute fails, got StrengthSpent=%d",
			hero.Stats.StrengthSpent)
	}
}

// TestUseGlypheDOubliOnAttributeRestoresItemToItsOriginalSlotOnFailure
// mirrors UseGlypheDOubliOnSkill's own regression test for the same class
// of bug (restore must go back into the exact slot cleared, never
// appended -- Inventory is a fixed-size 40-slot slice).
func TestUseGlypheDOubliOnAttributeRestoresItemToItsOriginalSlotOnFailure(t *testing.T) {
	hero := newLearnableHeroState(1, 0)
	hero.Inventory = []string{"", ItemGlypheDOubli, ""}

	// AttributeVitality has nothing spent on it, so RespecSingleAttributePoint fails.
	if err := hero.UseGlypheDOubliOnAttribute(AttributeVitality); err == nil {
		t.Fatal("test setup: expected UseGlypheDOubliOnAttribute to fail (nothing spent)")
	}

	if len(hero.Inventory) != 3 {
		t.Fatalf("expected Inventory to keep its original length 3, got %d", len(hero.Inventory))
	}

	if hero.Inventory[1] != ItemGlypheDOubli {
		t.Errorf("expected the glyph restored to its original slot (1), got %q at slot 1 (inventory: %v)",
			hero.Inventory[1], hero.Inventory)
	}
}
