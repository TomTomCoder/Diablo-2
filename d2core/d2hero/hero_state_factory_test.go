package d2hero

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadHeroStateHydratesDevilSkillsWithoutPanicking is a regression test:
// LoadHeroState used to assume every persisted skill ID lives in D2's own
// skills.txt, causing a nil-pointer panic on load for any Devil skill (i.e.
// every saved character, since even the starting Trait de feu is one).
// A HeroStateFactory with no asset manager is deliberate: Devil skills must
// never touch it, which is exactly what the fix guarantees.
func TestLoadHeroStateHydratesDevilSkillsWithoutPanicking(t *testing.T) {
	original := &HeroState{
		HeroName: "Test",
		Stats:    &HeroStatsState{},
		Skills: map[int]*HeroSkill{
			SkillTraitDeFeu: NewDevilHeroSkill(SkillTraitDeFeu),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal fixture: %v", err)
	}

	path := filepath.Join(t.TempDir(), "test.od2")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	factory := &HeroStateFactory{}

	loaded := factory.LoadHeroState(path)
	if loaded == nil {
		t.Fatal("expected LoadHeroState to succeed")
	}

	skill := loaded.Skills[SkillTraitDeFeu]
	if skill == nil {
		t.Fatal("expected Trait de feu to survive the round-trip")
	}

	if skill.SkillRecord == nil {
		t.Error("expected SkillRecord to be populated after load")
	}

	if skill.SkillPoints != 1 {
		t.Errorf("expected SkillPoints 1, got %d", skill.SkillPoints)
	}
}

func TestHydrateSkillsSkipsDevilSkillsWithoutPanicking(t *testing.T) {
	skills := map[int]*HeroSkill{
		SkillTraitDeFeu: NewDevilHeroSkill(SkillTraitDeFeu),
	}

	// must not panic with a nil asset manager -- Devil skills never touch it.
	HydrateSkills(skills, nil)

	if skills[SkillTraitDeFeu].SkillRecord == nil {
		t.Error("expected the Devil skill's own SkillRecord (from NewDevilHeroSkill) to remain untouched")
	}
}
