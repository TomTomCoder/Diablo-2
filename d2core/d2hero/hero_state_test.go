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
