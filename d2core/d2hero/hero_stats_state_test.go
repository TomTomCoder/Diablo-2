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

func TestHeroStatsStateRestoreMana(t *testing.T) {
	stats := &HeroStatsState{Mana: 2, MaxMana: 10}

	stats.RestoreMana(3)

	if stats.Mana != 5 {
		t.Errorf("expected Mana 5 after restoring 3, got %d", stats.Mana)
	}

	stats.RestoreMana(100)

	if stats.Mana != 10 {
		t.Errorf("expected RestoreMana to cap at MaxMana (10), got %d", stats.Mana)
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

func TestSpendAttributePointIncreasesAttributeAndSpendsAPoint(t *testing.T) {
	stats := &HeroStatsState{StatsPoints: 2, Vitality: 10}

	if err := stats.SpendAttributePoint(AttributeVitality); err != nil {
		t.Fatalf("expected SpendAttributePoint to succeed, got %v", err)
	}

	if stats.Vitality != 11 {
		t.Errorf("expected Vitality=11, got %d", stats.Vitality)
	}

	if stats.VitalitySpent != 1 {
		t.Errorf("expected VitalitySpent=1, got %d", stats.VitalitySpent)
	}

	if stats.StatsPoints != 1 {
		t.Errorf("expected StatsPoints to drop to 1, got %d", stats.StatsPoints)
	}
}

// TestSpendAttributePointGrowsMaxHealthAndMana is a regression test:
// MaxHealth/MaxMana were previously only ever computed once, in
// CreateHeroStatsState -- a Vitality/Energy point spent afterward had zero
// effect on them, even though devil_game_design_reference.md §5 states
// Vitality "détermine les points de vie" and Energy increases "la réserve
// de mana".
func TestSpendAttributePointGrowsMaxHealthAndMana(t *testing.T) {
	stats := &HeroStatsState{
		StatsPoints: 2,
		Vitality:    10, Health: 50, MaxHealth: 50, LifePerVit: 4,
		Energy: 10, Mana: 30, MaxMana: 30, ManaPerEne: 3,
	}

	if err := stats.SpendAttributePoint(AttributeVitality); err != nil {
		t.Fatalf("expected SpendAttributePoint(Vitality) to succeed, got %v", err)
	}

	if stats.MaxHealth != 54 || stats.Health != 54 {
		t.Errorf("expected MaxHealth=Health=54, got MaxHealth=%d Health=%d", stats.MaxHealth, stats.Health)
	}

	if err := stats.SpendAttributePoint(AttributeEnergy); err != nil {
		t.Fatalf("expected SpendAttributePoint(Energy) to succeed, got %v", err)
	}

	if stats.MaxMana != 33 || stats.Mana != 33 {
		t.Errorf("expected MaxMana=Mana=33, got MaxMana=%d Mana=%d", stats.MaxMana, stats.Mana)
	}
}

func TestSpendAttributePointFailsWithoutPoints(t *testing.T) {
	stats := &HeroStatsState{StatsPoints: 0}

	if err := stats.SpendAttributePoint(AttributeStrength); err == nil {
		t.Error("expected an error with no attribute points available")
	}
}

func TestSpendAttributePointFailsForUnknownAttribute(t *testing.T) {
	stats := &HeroStatsState{StatsPoints: 1}

	if err := stats.SpendAttributePoint(Attribute(99)); err == nil {
		t.Error("expected an error for an unknown attribute")
	}
}

func TestRefundAttributePointUndoesASpentPoint(t *testing.T) {
	stats := &HeroStatsState{StatsPoints: 1, Energy: 5}

	if err := stats.SpendAttributePoint(AttributeEnergy); err != nil {
		t.Fatalf("test setup: SpendAttributePoint failed: %v", err)
	}

	if err := stats.RefundAttributePoint(AttributeEnergy); err != nil {
		t.Fatalf("expected RefundAttributePoint to succeed, got %v", err)
	}

	if stats.Energy != 5 {
		t.Errorf("expected Energy restored to 5, got %d", stats.Energy)
	}

	if stats.EnergySpent != 0 {
		t.Errorf("expected EnergySpent reset to 0, got %d", stats.EnergySpent)
	}

	if stats.StatsPoints != 1 {
		t.Errorf("expected the point refunded to StatsPoints, got %d", stats.StatsPoints)
	}
}

func TestRefundAttributePointFailsIfNothingSpent(t *testing.T) {
	stats := &HeroStatsState{}

	if err := stats.RefundAttributePoint(AttributeStrength); err == nil {
		t.Error("expected an error refunding an attribute with nothing spent on it")
	}
}

func TestRefundAttributePointShrinksMaxHealthAndClampsCurrent(t *testing.T) {
	stats := &HeroStatsState{StatsPoints: 1, Vitality: 10, Health: 50, MaxHealth: 50, LifePerVit: 4}

	if err := stats.SpendAttributePoint(AttributeVitality); err != nil {
		t.Fatalf("test setup: SpendAttributePoint failed: %v", err)
	}
	// take some damage so current Health sits well below the new MaxHealth.
	stats.Health = 10

	if err := stats.RefundAttributePoint(AttributeVitality); err != nil {
		t.Fatalf("expected RefundAttributePoint to succeed, got %v", err)
	}

	if stats.MaxHealth != 50 {
		t.Errorf("expected MaxHealth restored to 50, got %d", stats.MaxHealth)
	}

	if stats.Health != 10 {
		t.Errorf("expected Health left untouched at 10 (below the new cap), got %d", stats.Health)
	}
}

func TestRefundAttributePointClampsCurrentHealthIfAboveNewMax(t *testing.T) {
	stats := &HeroStatsState{StatsPoints: 1, Vitality: 10, Health: 50, MaxHealth: 50, LifePerVit: 4}

	if err := stats.SpendAttributePoint(AttributeVitality); err != nil {
		t.Fatalf("test setup: SpendAttributePoint failed: %v", err)
	}
	// stats.Health is now 54 (full), matching the grown MaxHealth.

	if err := stats.RefundAttributePoint(AttributeVitality); err != nil {
		t.Fatalf("expected RefundAttributePoint to succeed, got %v", err)
	}

	if stats.Health != 50 {
		t.Errorf("expected Health clamped down to the restored MaxHealth of 50, got %d", stats.Health)
	}
}

func TestRespecAllAttributePointsRefundsEverything(t *testing.T) {
	stats := &HeroStatsState{StatsPoints: 4, Strength: 10, Energy: 10, Dexterity: 10, Vitality: 10}

	for _, attr := range []Attribute{AttributeStrength, AttributeEnergy, AttributeDexterity, AttributeVitality} {
		if err := stats.SpendAttributePoint(attr); err != nil {
			t.Fatalf("test setup: SpendAttributePoint(%v) failed: %v", attr, err)
		}
	}

	if refunded := stats.RespecAllAttributePoints(); refunded != 4 {
		t.Errorf("expected 4 points refunded, got %d", refunded)
	}

	if stats.Strength != 10 || stats.Energy != 10 || stats.Dexterity != 10 || stats.Vitality != 10 {
		t.Errorf("expected every attribute restored to its base of 10, got Str=%d Ene=%d Dex=%d Vit=%d",
			stats.Strength, stats.Energy, stats.Dexterity, stats.Vitality)
	}

	if stats.StatsPoints != 4 {
		t.Errorf("expected all 4 points back in StatsPoints, got %d", stats.StatsPoints)
	}
}

func TestRespecAllAttributePointsShrinksMaxHealthAndMana(t *testing.T) {
	stats := &HeroStatsState{
		StatsPoints: 2,
		Vitality:    10, Health: 50, MaxHealth: 50, LifePerVit: 4,
		Energy: 10, Mana: 30, MaxMana: 30, ManaPerEne: 3,
	}

	if err := stats.SpendAttributePoint(AttributeVitality); err != nil {
		t.Fatalf("test setup: SpendAttributePoint(Vitality) failed: %v", err)
	}

	if err := stats.SpendAttributePoint(AttributeEnergy); err != nil {
		t.Fatalf("test setup: SpendAttributePoint(Energy) failed: %v", err)
	}

	stats.RespecAllAttributePoints()

	if stats.MaxHealth != 50 || stats.Health != 50 {
		t.Errorf("expected MaxHealth=Health=50, got MaxHealth=%d Health=%d", stats.MaxHealth, stats.Health)
	}

	if stats.MaxMana != 30 || stats.Mana != 30 {
		t.Errorf("expected MaxMana=Mana=30, got MaxMana=%d Mana=%d", stats.MaxMana, stats.Mana)
	}
}
