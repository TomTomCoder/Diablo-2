package d2hero

import "testing"

func newLearnableHeroState(level, skillPoints int) *HeroState {
	return &HeroState{
		Stats:  &HeroStatsState{Level: level, SkillPoints: skillPoints},
		Skills: make(map[int]*HeroSkill),
	}
}

func TestLearnSkillSucceeds(t *testing.T) {
	hero := newLearnableHeroState(1, 1)

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("expected LearnSkill to succeed, got %v", err)
	}

	if _, known := hero.Skills[SkillTraitDeFeu]; !known {
		t.Error("expected Trait de feu to be added to Skills")
	}

	if hero.Stats.SkillPoints != 0 {
		t.Errorf("expected SkillPoints to drop to 0, got %d", hero.Stats.SkillPoints)
	}
}

func TestLearnSkillFailsWithoutPoints(t *testing.T) {
	hero := newLearnableHeroState(1, 0)

	if err := hero.LearnSkill(SkillTraitDeFeu); err == nil {
		t.Error("expected an error with no skill points available")
	}
}

func TestLearnSkillFailsBelowRequiredLevel(t *testing.T) {
	hero := newLearnableHeroState(1, 5)

	if err := hero.LearnSkill(SkillEclairEnChaine); err == nil {
		t.Error("expected an error learning a tier-6 skill at level 1")
	}
}

func TestLearnSkillFailsIfAlreadyKnown(t *testing.T) {
	hero := newLearnableHeroState(1, 2)

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("first LearnSkill should succeed, got %v", err)
	}

	if err := hero.LearnSkill(SkillTraitDeFeu); err == nil {
		t.Error("expected an error learning an already-known skill")
	}
}

func TestLearnSkillFailsForUnknownSkill(t *testing.T) {
	hero := newLearnableHeroState(1, 1)

	if err := hero.LearnSkill(unknownSkillID); err == nil {
		t.Error("expected an error for an unknown skill ID")
	}
}

func TestInvestSkillPointAddsToAnAlreadyLearnedSkill(t *testing.T) {
	hero := newLearnableHeroState(1, 2)

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("LearnSkill should succeed, got %v", err)
	}

	if err := hero.InvestSkillPoint(SkillTraitDeFeu); err != nil {
		t.Fatalf("expected InvestSkillPoint to succeed, got %v", err)
	}

	if got := hero.Skills[SkillTraitDeFeu].SkillPoints; got != 2 {
		t.Errorf("expected 2 points invested in Trait de feu, got %d", got)
	}

	if hero.Stats.SkillPoints != 0 {
		t.Errorf("expected SkillPoints to drop to 0, got %d", hero.Stats.SkillPoints)
	}
}

func TestInvestSkillPointFailsIfNotYetLearned(t *testing.T) {
	hero := newLearnableHeroState(1, 1)

	if err := hero.InvestSkillPoint(SkillTraitDeFeu); err == nil {
		t.Error("expected an error investing in a skill that isn't learned yet")
	}
}

func TestInvestSkillPointFailsWithoutPoints(t *testing.T) {
	hero := newLearnableHeroState(1, 1)

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("LearnSkill should succeed, got %v", err)
	}

	if err := hero.InvestSkillPoint(SkillTraitDeFeu); err == nil {
		t.Error("expected an error with no skill points available")
	}
}

func TestRespecSkillsRefundsPointsAndClearsSkills(t *testing.T) {
	hero := newLearnableHeroState(6, 2)
	hero.LeftSkill = SkillTraitDeFeu
	hero.RightSkill = SkillEclatDeGlace

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("LearnSkill(Trait de feu) should succeed, got %v", err)
	}

	if err := hero.LearnSkill(SkillEclatDeGlace); err != nil {
		t.Fatalf("LearnSkill(Éclat de glace) should succeed, got %v", err)
	}

	if hero.Stats.SkillPoints != 0 {
		t.Fatalf("expected 0 SkillPoints before respec, got %d", hero.Stats.SkillPoints)
	}

	refunded := hero.RespecSkills()

	if refunded != 2 {
		t.Errorf("expected 2 points refunded, got %d", refunded)
	}

	if hero.Stats.SkillPoints != 2 {
		t.Errorf("expected SkillPoints restored to 2, got %d", hero.Stats.SkillPoints)
	}

	if len(hero.Skills) != 0 {
		t.Errorf("expected Skills cleared, got %d entries", len(hero.Skills))
	}

	if hero.LeftSkill != 0 || hero.RightSkill != 0 {
		t.Errorf("expected LeftSkill/RightSkill reset to 0, got %d/%d", hero.LeftSkill, hero.RightSkill)
	}

	// learning again should work -- the points were genuinely returned,
	// not just reported.
	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Errorf("expected to be able to relearn a skill after respec, got %v", err)
	}
}

// TestRespecSkillsAlsoRefundsAttributePoints checks the "et d'attributs"
// half of "Respec partiel" (devil_game_design_reference.md §10), completed
// alongside HeroStatsState.SpendAttributePoint/RespecAllAttributePoints --
// previously RespecSkills only ever refunded skills.
func TestRespecSkillsAlsoRefundsAttributePoints(t *testing.T) {
	hero := newLearnableHeroState(1, 0)
	hero.Stats.StatsPoints = 1
	hero.Stats.Vitality = 10

	if err := hero.Stats.SpendAttributePoint(AttributeVitality); err != nil {
		t.Fatalf("test setup: SpendAttributePoint failed: %v", err)
	}

	hero.RespecSkills()

	if hero.Stats.Vitality != 10 {
		t.Errorf("expected Vitality restored to its base of 10, got %d", hero.Stats.Vitality)
	}

	if hero.Stats.StatsPoints != 1 {
		t.Errorf("expected the attribute point refunded to StatsPoints, got %d", hero.Stats.StatsPoints)
	}
}

func TestRespecSingleAttributePointRefundsOnlyThatAttribute(t *testing.T) {
	hero := newLearnableHeroState(1, 0)
	hero.Stats.StatsPoints = 2
	hero.Stats.Strength = 10
	hero.Stats.Dexterity = 10

	if err := hero.Stats.SpendAttributePoint(AttributeStrength); err != nil {
		t.Fatalf("test setup: SpendAttributePoint(Strength) failed: %v", err)
	}

	if err := hero.Stats.SpendAttributePoint(AttributeDexterity); err != nil {
		t.Fatalf("test setup: SpendAttributePoint(Dexterity) failed: %v", err)
	}

	if err := hero.RespecSingleAttributePoint(AttributeStrength); err != nil {
		t.Fatalf("expected RespecSingleAttributePoint to succeed, got %v", err)
	}

	if hero.Stats.Strength != 10 {
		t.Errorf("expected Strength restored to its base of 10, got %d", hero.Stats.Strength)
	}

	if hero.Stats.Dexterity != 11 {
		t.Errorf("expected Dexterity to remain untouched at 11, got %d", hero.Stats.Dexterity)
	}
}

func TestRespecSingleAttributePointFailsIfNothingSpent(t *testing.T) {
	hero := newLearnableHeroState(1, 0)

	if err := hero.RespecSingleAttributePoint(AttributeVitality); err == nil {
		t.Error("expected an error respeccing an attribute with nothing spent on it")
	}
}

func TestRespecSingleSkillRefundsOnlyThatSkill(t *testing.T) {
	hero := newLearnableHeroState(6, 2)
	hero.LeftSkill = SkillTraitDeFeu
	hero.RightSkill = SkillEclatDeGlace

	if err := hero.LearnSkill(SkillTraitDeFeu); err != nil {
		t.Fatalf("LearnSkill(Trait de feu) should succeed, got %v", err)
	}

	if err := hero.LearnSkill(SkillEclatDeGlace); err != nil {
		t.Fatalf("LearnSkill(Éclat de glace) should succeed, got %v", err)
	}

	refunded, err := hero.RespecSingleSkill(SkillTraitDeFeu)
	if err != nil {
		t.Fatalf("expected RespecSingleSkill to succeed, got %v", err)
	}

	if refunded != 1 {
		t.Errorf("expected 1 point refunded, got %d", refunded)
	}

	if hero.Stats.SkillPoints != 1 {
		t.Errorf("expected SkillPoints restored to 1, got %d", hero.Stats.SkillPoints)
	}

	if _, stillKnown := hero.Skills[SkillTraitDeFeu]; stillKnown {
		t.Error("expected Trait de feu to be forgotten")
	}

	if _, stillKnown := hero.Skills[SkillEclatDeGlace]; !stillKnown {
		t.Error("expected Éclat de glace to remain learned, untouched by the single-skill respec")
	}

	if hero.LeftSkill != 0 {
		t.Errorf("expected LeftSkill (which was the forgotten skill) reset to 0, got %d", hero.LeftSkill)
	}

	if hero.RightSkill != SkillEclatDeGlace {
		t.Errorf("expected RightSkill (a different, still-known skill) left untouched, got %d", hero.RightSkill)
	}
}

func TestRespecSingleSkillFailsIfNotLearned(t *testing.T) {
	hero := newLearnableHeroState(1, 0)

	if _, err := hero.RespecSingleSkill(SkillTraitDeFeu); err == nil {
		t.Error("expected an error respeccing a skill that was never learned")
	}
}

func TestRespecSkillsWithNoSkillsIsNoop(t *testing.T) {
	hero := newLearnableHeroState(1, 3)

	if refunded := hero.RespecSkills(); refunded != 0 {
		t.Errorf("expected 0 points refunded with no skills learned, got %d", refunded)
	}

	if hero.Stats.SkillPoints != 3 {
		t.Errorf("expected SkillPoints unchanged at 3, got %d", hero.Stats.SkillPoints)
	}
}
