package d2hero

import (
	"errors"
	"math/rand"
)

// TreasureClassEntry is one weighted possibility within a TreasureClass:
// either a real item code, or "" meaning "nothing" -- Diablo 2's own
// NoDrop expressed as just another weighted entry rather than a separate
// field, so RollTreasureClass has only one weighted pick to make per
// pick, not a special-cased NoDrop check alongside it.
type TreasureClassEntry struct {
	ItemCode string
	Weight   int
}

// TreasureClass is Devil's version of Diablo 2's own Treasure Classes
// (devil_game_design_reference.md §4.2): "Chaque TC contient une liste
// d'objets avec des probabilités associées, un nombre de picks et une
// valeur NoDrop."
//
// ponytail: real D2 Treasure Classes can nest (an entry can itself name
// another TC, picked recursively) and are selected by the zone/monster's
// level. Neither is modeled here: nesting has no concrete use yet with
// only one flat pool, and there's no per-monster-level data to select a
// TC by (MonsterArchetypes is still an empty registry -- see its own
// comment). This is the flat weighted-pick-with-NoDrop core, ready to
// grow into the nested/level-gated version once that data exists.
type TreasureClass struct {
	ID      string
	Entries []TreasureClassEntry
	Picks   int
}

// RollTreasureClass makes tc.Picks independent weighted picks across
// tc.Entries and returns the non-empty item codes picked (a "" entry,
// i.e. NoDrop, contributes nothing). Each pick is independent and can
// itself land on NoDrop, so the result can be shorter than tc.Picks or
// even empty.
//
// nolint:gosec // not concerned with crypto-strong randomness
func RollTreasureClass(tc *TreasureClass) ([]string, error) {
	if tc == nil || len(tc.Entries) == 0 {
		return nil, errors.New("treasure class has no entries")
	}

	totalWeight := 0
	for _, entry := range tc.Entries {
		totalWeight += entry.Weight
	}

	if totalWeight <= 0 {
		return nil, errors.New("treasure class has no positive weight")
	}

	var drops []string

	for i := 0; i < tc.Picks; i++ {
		roll := rand.Intn(totalWeight)
		cumulative := 0

		for _, entry := range tc.Entries {
			cumulative += entry.Weight
			if roll < cumulative {
				if entry.ItemCode != "" {
					drops = append(drops, entry.ItemCode)
				}

				break
			}
		}
	}

	return drops, nil
}
