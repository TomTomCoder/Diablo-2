package d2hero

import (
	"testing"
	"time"
)

func TestHeroStatsStateApplyDamage(t *testing.T) {
	stats := &HeroStatsState{Health: 10, MaxHealth: 10}

	if died := stats.ApplyDamage(4); died {
		t.Error("hero with 10 HP should not die from 4 damage")
	}

	if stats.Health != 6 {
		t.Errorf("expected HP 6 after 4 damage, got %d", stats.Health)
	}

	if died := stats.ApplyDamage(100); !died {
		t.Error("hero should die when damage exceeds remaining HP")
	}

	if stats.Health != 0 {
		t.Errorf("expected HP to floor at 0, got %d", stats.Health)
	}

	if died := stats.ApplyDamage(1); !died {
		t.Error("applying damage to an already-dead hero should still report died=true")
	}
}

func TestHeroStatsStateHeal(t *testing.T) {
	stats := &HeroStatsState{Health: 4, MaxHealth: 10}

	stats.Heal(3)

	if stats.Health != 7 {
		t.Errorf("expected HP 7 after healing 3, got %d", stats.Health)
	}

	stats.Heal(100)

	if stats.Health != 10 {
		t.Errorf("expected Heal to cap at MaxHealth (10), got %d", stats.Health)
	}
}

func TestHeroStatsStateMagicImmunity(t *testing.T) {
	stats := &HeroStatsState{}
	now := time.Now()

	if stats.IsMagicImmune(now) {
		t.Fatal("a fresh hero should not start magic immune")
	}

	stats.ApplyMagicImmunity(now.Add(8 * time.Second))

	if !stats.IsMagicImmune(now) {
		t.Error("expected the hero to be magic immune immediately after ApplyMagicImmunity")
	}

	if stats.IsMagicImmune(now.Add(9 * time.Second)) {
		t.Error("expected the immunity to have expired after its duration elapsed")
	}
}

func TestGrantExperienceNoLevelUp(t *testing.T) {
	stats := &HeroStatsState{Level: 1, Experience: 0, NextLevelExp: 100}

	stats.GrantExperience(30)

	if stats.Experience != 30 || stats.Level != 1 {
		t.Errorf("expected Experience=30 Level=1, got Experience=%d Level=%d", stats.Experience, stats.Level)
	}
}

func TestGrantExperienceSingleLevelUp(t *testing.T) {
	stats := &HeroStatsState{Level: 1, Experience: 90, NextLevelExp: 100, SkillPoints: 0, StatsPoints: 0}

	stats.GrantExperience(20)

	if stats.Level != 2 {
		t.Fatalf("expected Level=2, got %d", stats.Level)
	}

	if stats.Experience != 10 {
		t.Errorf("expected leftover Experience=10, got %d", stats.Experience)
	}

	if stats.SkillPoints != skillPointsPerLevel || stats.StatsPoints != statsPointsPerLevel {
		t.Errorf("expected SkillPoints=%d StatsPoints=%d, got SkillPoints=%d StatsPoints=%d",
			skillPointsPerLevel, statsPointsPerLevel, stats.SkillPoints, stats.StatsPoints)
	}

	if stats.NextLevelExp != experienceForLevel(2) {
		t.Errorf("expected NextLevelExp=%d, got %d", experienceForLevel(2), stats.NextLevelExp)
	}
}

func TestGrantExperienceMultipleLevelUps(t *testing.T) {
	stats := &HeroStatsState{Level: 1, Experience: 0, NextLevelExp: experienceForLevel(1)}

	// enough to clear level 1 (100) and level 2 (200), landing partway into level 3.
	stats.GrantExperience(350)

	if stats.Level != 3 {
		t.Fatalf("expected Level=3, got %d", stats.Level)
	}

	if stats.SkillPoints != 2*skillPointsPerLevel || stats.StatsPoints != 2*statsPointsPerLevel {
		t.Errorf("expected points for 2 level-ups, got SkillPoints=%d StatsPoints=%d", stats.SkillPoints, stats.StatsPoints)
	}
}
