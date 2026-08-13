package d2hero

import "testing"

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
