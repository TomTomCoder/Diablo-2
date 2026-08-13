package d2hero

import "testing"

func TestApplyDamageWithManaShieldInactive(t *testing.T) {
	stats := &HeroStatsState{Health: 10, MaxHealth: 10, Mana: 10, MaxMana: 10, ManaShieldActive: false}

	stats.ApplyDamageWithManaShield(4)

	if stats.Health != 6 {
		t.Errorf("expected damage to hit Health directly when the shield is off, got HP %d", stats.Health)
	}

	if stats.Mana != 10 {
		t.Errorf("expected Mana untouched when the shield is off, got %d", stats.Mana)
	}
}

func TestApplyDamageWithManaShieldAbsorbsFully(t *testing.T) {
	stats := &HeroStatsState{Health: 10, MaxHealth: 10, Mana: 10, MaxMana: 10, ManaShieldActive: true}

	if died := stats.ApplyDamageWithManaShield(6); died {
		t.Error("did not expect death: mana fully absorbed the hit")
	}

	if stats.Mana != 4 {
		t.Errorf("expected Mana to drop by the full damage amount (6), got %d", stats.Mana)
	}

	if stats.Health != 10 {
		t.Errorf("expected Health untouched when mana fully absorbs the hit, got %d", stats.Health)
	}
}

func TestApplyDamageWithManaShieldOverflowsToHealth(t *testing.T) {
	stats := &HeroStatsState{Health: 10, MaxHealth: 10, Mana: 3, MaxMana: 10, ManaShieldActive: true}

	stats.ApplyDamageWithManaShield(6)

	if stats.Mana != 0 {
		t.Errorf("expected all available Mana (3) drained, got %d", stats.Mana)
	}

	if stats.Health != 7 {
		t.Errorf("expected the remaining 3 damage to fall through to Health (10-3=7), got %d", stats.Health)
	}
}

func TestApplyDamageWithManaShieldNoManaFallsThrough(t *testing.T) {
	stats := &HeroStatsState{Health: 10, MaxHealth: 10, Mana: 0, MaxMana: 10, ManaShieldActive: true}

	stats.ApplyDamageWithManaShield(4)

	if stats.Health != 6 {
		t.Errorf("expected full damage to Health when there's no Mana to absorb with, got HP %d", stats.Health)
	}
}
