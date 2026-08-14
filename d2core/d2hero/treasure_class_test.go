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
