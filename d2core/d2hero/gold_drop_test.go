package d2hero

import "testing"

func TestRollGoldDropStaysInRange(t *testing.T) {
	for i := 0; i < 200; i++ {
		got := RollGoldDrop()
		if got < minGoldDrop || got > maxGoldDrop {
			t.Fatalf("RollGoldDrop() = %d, want in [%d, %d]", got, minGoldDrop, maxGoldDrop)
		}
	}
}
