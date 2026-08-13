package d2hero

import "math/rand"

// minExperienceDrop/maxExperienceDrop is the experience a killable monster
// grants on death.
//
// ponytail: one flat range for every monster -- same placeholder shape as
// RollGoldDrop, no monster-level-based experience data exists yet.
const (
	minExperienceDrop = 10
	maxExperienceDrop = 30
)

// RollExperienceDrop returns a random experience amount in
// [minExperienceDrop, maxExperienceDrop].
//
// nolint:gosec // not concerned with crypto-strong randomness
func RollExperienceDrop() int {
	return minExperienceDrop + rand.Intn(maxExperienceDrop-minExperienceDrop+1)
}
