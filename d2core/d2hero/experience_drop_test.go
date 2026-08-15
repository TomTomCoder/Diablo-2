package d2hero

import "testing"

func TestRollExperienceDropStaysInRange(t *testing.T) {
	for i := 0; i < 200; i++ {
		got := RollExperienceDrop()
		if got < minExperienceDrop || got > maxExperienceDrop {
			t.Fatalf("RollExperienceDrop() = %d, want in [%d, %d]", got, minExperienceDrop, maxExperienceDrop)
		}
	}
}
