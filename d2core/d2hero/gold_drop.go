package d2hero

import "math/rand"

// minGoldDrop/maxGoldDrop is the gold a killable monster drops on death.
//
// ponytail: one flat range for every monster -- no monster-level-based
// Treasure Class data exists yet (devil_game_design_reference.md §4.2,
// ROADMAP.md Phase 5). Sidesteps the same wall item drops hit: a dropped
// item needs a real sprite (its CommonRecord's FlippyFile) to render as a
// pickup entity, and no Devil-specific item sprites exist yet
// (ROADMAP.md Phase 6) -- gold has no such visual dependency.
const (
	minGoldDrop = 5
	maxGoldDrop = 20
)

// RollGoldDrop returns a random gold amount in [minGoldDrop, maxGoldDrop].
//
// nolint:gosec // not concerned with crypto-strong randomness
func RollGoldDrop() int {
	return minGoldDrop + rand.Intn(maxGoldDrop-minGoldDrop+1)
}
