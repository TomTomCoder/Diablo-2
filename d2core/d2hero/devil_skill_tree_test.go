package d2hero

import "testing"

// unknownSkillID is any ID not present in DevilSkills.
const unknownSkillID = -1

func TestCanLearnSkill(t *testing.T) {
	if !CanLearnSkill(SkillTraitDeFeu, 1) {
		t.Error("a level 1 hero should be able to learn Trait de feu (RequiredLevel 1)")
	}

	if CanLearnSkill(SkillTraitDeFeu, 0) {
		t.Error("a level 0 hero should not meet Trait de feu's RequiredLevel")
	}

	if CanLearnSkill(unknownSkillID, 99) {
		t.Error("an unknown skill ID should never be learnable, regardless of level")
	}
}

func TestSkillBaseSortDamageFallback(t *testing.T) {
	const fallback = 4

	if got := SkillBaseSortDamage(SkillTraitDeFeu, fallback); got != DevilSkills[SkillTraitDeFeu].BaseSortDamage {
		t.Errorf("expected Trait de feu's own base_sort, got %d", got)
	}

	if got := SkillBaseSortDamage(unknownSkillID, fallback); got != fallback {
		t.Errorf("expected fallback %d for an unknown skill, got %d", fallback, got)
	}
}

func TestSkillEclatDeGlaceIsDistinctFromTraitDeFeu(t *testing.T) {
	if SkillEclatDeGlace == SkillTraitDeFeu {
		t.Fatal("SkillEclatDeGlace must not collide with SkillTraitDeFeu")
	}

	fire, ice := DevilSkills[SkillTraitDeFeu], DevilSkills[SkillEclatDeGlace]

	if fire.Tree != ice.Tree {
		t.Error("expected both starting Élémentalisme skills to share a tree")
	}

	if fire.BaseSortDamage == ice.BaseSortDamage {
		t.Error("expected distinct base_sort values -- Éclat de glace trades damage for its slow effect")
	}
}

func TestSkillManaCostFallback(t *testing.T) {
	const fallback = 2

	if got := SkillManaCost(SkillTraitDeFeu, fallback); got != DevilSkills[SkillTraitDeFeu].ManaCost {
		t.Errorf("expected Trait de feu's own mana cost, got %d", got)
	}

	if got := SkillManaCost(unknownSkillID, fallback); got != fallback {
		t.Errorf("expected fallback %d for an unknown skill, got %d", fallback, got)
	}
}
