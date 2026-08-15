package d2hero

import "testing"

func TestRollTreasureClassFailsOnNilOrEmpty(t *testing.T) {
	if _, err := RollTreasureClass(nil); err == nil {
		t.Error("expected an error for a nil TreasureClass")
	}

	if _, err := RollTreasureClass(&TreasureClass{Picks: 1}); err == nil {
		t.Error("expected an error for a TreasureClass with no entries")
	}
}

func TestRollTreasureClassFailsOnZeroTotalWeight(t *testing.T) {
	tc := &TreasureClass{
		Picks:   1,
		Entries: []TreasureClassEntry{{ItemCode: ItemBatonInitie, Weight: 0}},
	}

	if _, err := RollTreasureClass(tc); err == nil {
		t.Error("expected an error when every entry has weight 0")
	}
}

// TestRollTreasureClassAlwaysPicksTheOnlyPositiveWeightEntry is
// deterministic despite RollTreasureClass being randomized: with every
// other entry at weight 0, there's only one possible outcome per pick.
func TestRollTreasureClassAlwaysPicksTheOnlyPositiveWeightEntry(t *testing.T) {
	tc := &TreasureClass{
		Picks: 5,
		Entries: []TreasureClassEntry{
			{ItemCode: ItemBatonInitie, Weight: 1},
			{ItemCode: ItemRobeDuNovice, Weight: 0},
			{ItemCode: "", Weight: 0},
		},
	}

	drops, err := RollTreasureClass(tc)
	if err != nil {
		t.Fatalf("RollTreasureClass failed: %v", err)
	}

	if len(drops) != 5 {
		t.Fatalf("expected 5 drops, got %d", len(drops))
	}

	for _, drop := range drops {
		if drop != ItemBatonInitie {
			t.Errorf("expected every pick to be %q, got %q", ItemBatonInitie, drop)
		}
	}
}

// TestRollTreasureClassAllNoDropYieldsNoDrops is deterministic for the
// same reason: with only a "" (NoDrop) entry, every pick must land on it.
func TestRollTreasureClassAllNoDropYieldsNoDrops(t *testing.T) {
	tc := &TreasureClass{
		Picks:   3,
		Entries: []TreasureClassEntry{{ItemCode: "", Weight: 1}},
	}

	drops, err := RollTreasureClass(tc)
	if err != nil {
		t.Fatalf("RollTreasureClass failed: %v", err)
	}

	if len(drops) != 0 {
		t.Errorf("expected no drops when every entry is NoDrop, got %v", drops)
	}
}

func TestRollTreasureClassZeroPicksYieldsNoDrops(t *testing.T) {
	tc := &TreasureClass{
		Picks:   0,
		Entries: []TreasureClassEntry{{ItemCode: ItemBatonInitie, Weight: 1}},
	}

	drops, err := RollTreasureClass(tc)
	if err != nil {
		t.Fatalf("RollTreasureClass failed: %v", err)
	}

	if len(drops) != 0 {
		t.Errorf("expected no drops for Picks: 0, got %v", drops)
	}
}

// TestTCBasicMonsterRollsWithoutError is a structural regression test:
// nothing here depends on the actual random outcome, just that
// TCBasicMonster is well-formed enough for RollTreasureClass to never
// fail against it (positive weights, at least one entry) and that every
// non-empty drop it can ever produce is one of Devil's own known item
// codes, not something malformed.
func TestTCBasicMonsterRollsWithoutError(t *testing.T) {
	known := map[string]bool{
		ItemPotionDeMana:           true,
		ItemPendentifArcane:        true,
		ItemAnneauDuDebut:          true,
		ItemRobeDuNovice:           true,
		ItemCeintureDeCuirRunique:  true,
		ItemEssenceDeBossOrdinaire: true,
	}

	for i := 0; i < 100; i++ {
		drops, err := RollTreasureClass(TCBasicMonster)
		if err != nil {
			t.Fatalf("RollTreasureClass(TCBasicMonster) failed: %v", err)
		}

		for _, code := range drops {
			if !known[code] {
				t.Errorf("unexpected item code %q in TCBasicMonster's drops", code)
			}
		}
	}
}

// TestTCBasicMonsterCanDropNothing guards the "mostly nothing" weighting
// TCBasicMonster's own doc comment describes -- NoDrop's weight (100) is
// by far the largest of the pool (145 total), so across enough rolls at
// least one should come back empty. Not a strict probability assertion,
// just a sanity check that NoDrop entries aren't accidentally excluded
// from ever winning.
func TestTCBasicMonsterCanDropNothing(t *testing.T) {
	for i := 0; i < 200; i++ {
		drops, err := RollTreasureClass(TCBasicMonster)
		if err != nil {
			t.Fatalf("RollTreasureClass(TCBasicMonster) failed: %v", err)
		}

		if len(drops) == 0 {
			return
		}
	}

	t.Error("expected at least one empty (NoDrop) result across 200 rolls")
}
