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

func TestMagicResistanceForDifficulty(t *testing.T) {
	record := &MonStatRecord{
		ResistanceMagicNormal:    10,
		ResistanceMagicNightmare: 30,
		ResistanceMagicHell:      50,
	}

	cases := []struct {
		name       string
		difficulty d2enum.DifficultyType
		want       int
	}{
		{"Normal", d2enum.DifficultyNormal, 10},
		{"Nightmare", d2enum.DifficultyNightmare, 30},
		{"Hell", d2enum.DifficultyHell, 50},
	}

	for _, c := range cases {
		if got := record.MagicResistanceForDifficulty(c.difficulty); got != c.want {
			t.Errorf("%s: expected %d, got %d", c.name, c.want, got)
		}
	}
}

func TestAttackDamageRangeForDifficulty(t *testing.T) {
	record := &MonStatRecord{
		DamageMinA1Normal:    3,
		DamageMaxA1Normal:    7,
		DamageMinA1Nightmare: 10,
		DamageMaxA1Nightmare: 20,
		DamageMinA1Hell:      30,
		DamageMaxA1Hell:      50,
	}

	cases := []struct {
		name       string
		difficulty d2enum.DifficultyType
		wantMin    int
		wantMax    int
	}{
		{"Normal", d2enum.DifficultyNormal, 3, 7},
		{"Nightmare", d2enum.DifficultyNightmare, 10, 20},
		{"Hell", d2enum.DifficultyHell, 30, 50},
	}

	for _, c := range cases {
		gotMin, gotMax := record.AttackDamageRangeForDifficulty(c.difficulty)
		if gotMin != c.wantMin || gotMax != c.wantMax {
			t.Errorf("%s: expected (%d, %d), got (%d, %d)", c.name, c.wantMin, c.wantMax, gotMin, gotMax)
		}
	}
}

func TestColdSensitivityForDifficulty(t *testing.T) {
	record := &MonStatRecord{
		ColdSensitivityNormal:    100,
		ColdSensitivityNightmare: 50,
		ColdSensitivityHell:      0,
	}

	cases := []struct {
		name       string
		difficulty d2enum.DifficultyType
		want       int
	}{
		{"Normal", d2enum.DifficultyNormal, 100},
		{"Nightmare", d2enum.DifficultyNightmare, 50},
		{"Hell", d2enum.DifficultyHell, 0},
	}

	for _, c := range cases {
		if got := record.ColdSensitivityForDifficulty(c.difficulty); got != c.want {
			t.Errorf("%s: expected %d, got %d", c.name, c.want, got)
		}
	}
}

func TestAiDistanceForDifficulty(t *testing.T) {
	record := &MonStatRecord{
		AiDistanceNormal:    12,
		AiDistanceNightmare: 18,
		AiDistanceHell:      24,
	}

	cases := []struct {
		name       string
		difficulty d2enum.DifficultyType
		want       int
	}{
		{"Normal", d2enum.DifficultyNormal, 12},
		{"Nightmare", d2enum.DifficultyNightmare, 18},
		{"Hell", d2enum.DifficultyHell, 24},
	}

	for _, c := range cases {
		if got := record.AiDistanceForDifficulty(c.difficulty); got != c.want {
			t.Errorf("%s: expected %d, got %d", c.name, c.want, got)
		}
	}
}

func TestAiDelayForDifficulty(t *testing.T) {
	record := &MonStatRecord{
		AiDelayNormal:    9,
		AiDelayNightmare: 6,
		AiDelayHell:      3,
	}

	cases := []struct {
		name       string
		difficulty d2enum.DifficultyType
		want       int
	}{
		{"Normal", d2enum.DifficultyNormal, 9},
		{"Nightmare", d2enum.DifficultyNightmare, 6},
		{"Hell", d2enum.DifficultyHell, 3},
	}

	for _, c := range cases {
		if got := record.AiDelayForDifficulty(c.difficulty); got != c.want {
			t.Errorf("%s: expected %d, got %d", c.name, c.want, got)
		}
	}
}
