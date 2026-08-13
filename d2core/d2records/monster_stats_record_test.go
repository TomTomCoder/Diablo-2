package d2records

import (
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
)

func TestHPRangeForDifficulty(t *testing.T) {
	record := &MonStatRecord{
		MinHPNormal:    10,
		MaxHPNormal:    20,
		MinHPNightmare: 100,
		MaxHPNightmare: 200,
		MinHPHell:      1000,
		MaxHPHell:      2000,
	}

	cases := []struct {
		name       string
		difficulty d2enum.DifficultyType
		wantMin    int
		wantMax    int
	}{
		{"Normal", d2enum.DifficultyNormal, 10, 20},
		{"Nightmare", d2enum.DifficultyNightmare, 100, 200},
		{"Hell", d2enum.DifficultyHell, 1000, 2000},
	}

	for _, c := range cases {
		gotMin, gotMax := record.HPRangeForDifficulty(c.difficulty)
		if gotMin != c.wantMin || gotMax != c.wantMax {
			t.Errorf("%s: expected (%d, %d), got (%d, %d)", c.name, c.wantMin, c.wantMax, gotMin, gotMax)
		}
	}
}
