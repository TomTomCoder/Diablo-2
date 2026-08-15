package d2hero

import "testing"

func TestUseRespecEssenceConsumesItemAndRespecs(t *testing.T) {
	hero := newLearnableHeroState(6, 0)
	hero.Inventory = []string{ItemEssenceDeBossOrdinaire, ""}

	if err := hero.LearnSkill(SkillTraitDeFeu); err == nil {
		t.Fatal("test setup: expected LearnSkill to fail with 0 SkillPoints")
	}

	hero.Stats.SkillPoints = 1

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("test setup: LearnSkill failed: %v", err)
	}

	refunded, err := hero.UseRespecEssence()
	if err != nil {
		t.Fatalf("expected UseRespecEssence to succeed, got %v", err)
	}

	if refunded != 1 {
		t.Errorf("expected 1 point refunded, got %d", refunded)
	}

	if len(hero.Skills) != 0 {
		t.Errorf("expected Skills cleared by the respec, got %d entries", len(hero.Skills))
	}

	if hero.Inventory[0] != "" {
		t.Errorf("expected the essence consumed (slot cleared), got %q", hero.Inventory[0])
	}
}

func TestUseRespecEssenceFailsWithoutTheItem(t *testing.T) {
	hero := newLearnableHeroState(6, 1)
	hero.Inventory = []string{"", ""}

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("test setup: LearnSkill failed: %v", err)
	}

	if _, err := hero.UseRespecEssence(); err == nil {
		t.Fatal("expected UseRespecEssence to fail without the essence in inventory")
	}

	if _, known := hero.Skills[SkillTraitDeFeu]; !known {
		t.Error("expected Skills untouched when UseRespecEssence fails")
	}
}
