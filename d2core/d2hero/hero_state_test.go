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

func TestRespecSkillsWithNoSkillsIsNoop(t *testing.T) {
	hero := newLearnableHeroState(1, 3)

	if refunded := hero.RespecSkills(); refunded != 0 {
		t.Errorf("expected 0 points refunded with no skills learned, got %d", refunded)
	}

	if hero.Stats.SkillPoints != 3 {
		t.Errorf("expected SkillPoints unchanged at 3, got %d", hero.Stats.SkillPoints)
	}
}
