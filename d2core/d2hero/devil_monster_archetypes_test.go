package d2hero

import (
	"testing"
	"time"
)

func TestMonsterArchetypeLookupsFallBackForUnregisteredKey(t *testing.T) {
	const unregisteredKey = "not-a-real-monster"

	if got, want := MonsterAggroRadiusSubtiles(unregisteredKey, 8), 8.0; got != want {
		t.Errorf("expected fallback aggro radius %v, got %v", want, got)
	}

	if got, want := MonsterAttackRangeSubtiles(unregisteredKey, 2), 2.0; got != want {
		t.Errorf("expected fallback attack range %v, got %v", want, got)
	}

	if got, want := MonsterAttackCooldown(unregisteredKey, time.Second), time.Second; got != want {
		t.Errorf("expected fallback cooldown %v, got %v", want, got)
	}

	if got, want := MonsterAttackDamage(unregisteredKey, 3), 3; got != want {
		t.Errorf("expected fallback damage %d, got %d", want, got)
	}
}

func TestMonsterArchetypeLookupsUseRegisteredOverride(t *testing.T) {
	const key = "test-only-archetype"

	MonsterArchetypes[key] = &MonsterArchetypeDef{
		Key:                 key,
		AggroRadiusSubtiles: 20,
		AttackRangeSubtiles: 5,
		AttackCooldown:      3 * time.Second,
		AttackDamage:        50,
	}
	defer delete(MonsterArchetypes, key)

	if got, want := MonsterAggroRadiusSubtiles(key, 8), 20.0; got != want {
		t.Errorf("expected overridden aggro radius %v, got %v", want, got)
	}

	if got, want := MonsterAttackRangeSubtiles(key, 2), 5.0; got != want {
		t.Errorf("expected overridden attack range %v, got %v", want, got)
	}

	if got, want := MonsterAttackCooldown(key, time.Second), 3*time.Second; got != want {
		t.Errorf("expected overridden cooldown %v, got %v", want, got)
	}

	if got, want := MonsterAttackDamage(key, 3), 50; got != want {
		t.Errorf("expected overridden damage %d, got %d", want, got)
	}
}

func TestRollDamageInRangeStaysWithinBounds(t *testing.T) {
	for i := 0; i < 50; i++ {
		got := RollDamageInRange(3, 7, 99)
		if got < 3 || got > 7 {
			t.Fatalf("expected a roll within [3, 7], got %d", got)
		}
	}
}

func TestRollDamageInRangeFallsBackForEmptyRange(t *testing.T) {
	if got, want := RollDamageInRange(0, 0, 5), 5; got != want {
		t.Errorf("expected fallback %d for an empty (0, 0) range, got %d", want, got)
	}

	if got, want := RollDamageInRange(5, 3, 5), 5; got != want {
		t.Errorf("expected fallback %d when min > max, got %d", want, got)
	}
}
