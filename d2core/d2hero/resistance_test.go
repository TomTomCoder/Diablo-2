package d2hero

import "testing"

func TestCapResistance(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{in: -20, want: 0}, // never negative, unlike Diablo 2
		{in: 0, want: 0},
		{in: 40, want: 40},
		{in: 75, want: 75},
		{in: 200, want: 75}, // capped
	}

	for _, c := range cases {
		if got := CapResistance(c.in); got != c.want {
			t.Errorf("CapResistance(%d): expected %d, got %d", c.in, c.want, got)
		}
	}
}

func TestMitigateDamage(t *testing.T) {
	cases := []struct {
		amount, resist, want int
	}{
		{amount: 10, resist: 0, want: 10},
		{amount: 10, resist: 50, want: 5},
		{amount: 10, resist: 75, want: 3}, // 10 - (10*75)/100 = 10-7, integer division
		{amount: 1, resist: 75, want: 1},  // (1*75)/100 truncates to 0 -- no reduction at this scale
	}

	for _, c := range cases {
		if got := MitigateDamage(c.amount, c.resist); got != c.want {
			t.Errorf("MitigateDamage(%d, %d%%): expected %d, got %d", c.amount, c.resist, c.want, got)
		}
	}
}
