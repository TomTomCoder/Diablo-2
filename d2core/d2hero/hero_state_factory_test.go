package d2hero

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2inventory"
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

// TestSaveThenLoadRoundTripsARealDevilCharacter exercises the actual
// Save/LoadHeroState pair (not a hand-rolled marshal) against a character
// shaped like a real one at this point in the session: all 4 Devil
// equipment slots filled, several learned skills including a multi-point
// one (Maîtrise élémentaire via InvestSkillPoint), a belt with a potion,
// and the newer HeroStatsState fields (StatsPoints, a spent attribute
// counter, LifePerVit/ManaPerEne). ROADMAP.md Phase 1 flagged this as
// "reste à vérifier ... une fois qu'il y a plus d'un seul type d'objet en
// jeu" -- that's now true, so it's worth actually checking.
func TestSaveThenLoadRoundTripsARealDevilCharacter(t *testing.T) {
	original := &HeroState{
		HeroName:   "Devil",
		Gold:       123,
		LeftSkill:  SkillTraitDeFeu,
		RightSkill: SkillEclatDeGlace,
		Equipment: d2inventory.CharacterEquipment{
			RightHand: &d2inventory.InventoryItemWeapon{ItemCode: ItemBatonApprenti, ItemName: "Bâton de l'Apprenti"},
			Amulet:    &d2inventory.InventoryItemMisc{ItemCode: ItemPendentifArcane, ItemName: "Pendentif Arcane"},
			Ring:      &d2inventory.InventoryItemMisc{ItemCode: ItemAnneauDuDebut, ItemName: "Anneau du Début"},
			Torso:     &d2inventory.InventoryItemArmor{ItemCode: ItemRobeDuNovice, ItemName: "Robe du Novice"},
		},
		Skills: map[int]*HeroSkill{
			SkillTraitDeFeu:          NewDevilHeroSkill(SkillTraitDeFeu),
			SkillEclatDeGlace:        NewDevilHeroSkill(SkillEclatDeGlace),
			SkillMaitriseElementaire: NewDevilHeroSkill(SkillMaitriseElementaire),
		},
		Belt: []string{ItemPotionDeMana, "", "", ""},
		Stats: &HeroStatsState{
			Level: 18, StatsPoints: 3, SkillPoints: 1,
			Vitality: 12, VitalitySpent: 2, LifePerVit: 4,
			MaxHealth: 58, Health: 58,
			ManaPerEne: 3,
		},
		DiscoveredRecipes:   map[string]bool{RecipeUpgradeBatonApprenti: true},
		RespecPartielUsedAt: map[d2enum.DifficultyType]bool{d2enum.DifficultyNormal: true},
	}

	// a second point invested in Maîtrise élémentaire, so SkillPoints != 1
	// is actually exercised by the round-trip.
	if err := original.InvestSkillPoint(SkillMaitriseElementaire); err != nil {
		t.Fatalf("test setup: InvestSkillPoint failed: %v", err)
	}

	factory := &HeroStateFactory{}
	original.FilePath = filepath.Join(t.TempDir(), "test.od2")

	if err := factory.Save(original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded := factory.LoadHeroState(original.FilePath)
	if loaded == nil {
		t.Fatal("expected LoadHeroState to succeed")
	}

	if loaded.HeroName != "Devil" || loaded.Gold != 123 {
		t.Errorf("expected HeroName/Gold to survive, got %q/%d", loaded.HeroName, loaded.Gold)
	}

	if loaded.LeftSkill != SkillTraitDeFeu || loaded.RightSkill != SkillEclatDeGlace {
		t.Errorf("expected LeftSkill/RightSkill to survive, got %d/%d", loaded.LeftSkill, loaded.RightSkill)
	}

	equip := loaded.Equipment
	if equip.RightHand.GetItemCode() != ItemBatonApprenti || equip.Amulet.GetItemCode() != ItemPendentifArcane ||
		equip.Ring.GetItemCode() != ItemAnneauDuDebut || equip.Torso.GetItemCode() != ItemRobeDuNovice {
		t.Errorf("expected all 4 equipped Devil items to survive, got RightHand=%q Amulet=%q Ring=%q Torso=%q",
			equip.RightHand.GetItemCode(), equip.Amulet.GetItemCode(), equip.Ring.GetItemCode(), equip.Torso.GetItemCode())
	}

	if got := loaded.Skills[SkillMaitriseElementaire]; got == nil || got.SkillPoints != 2 {
		t.Errorf("expected Maîtrise élémentaire at 2 invested points to survive, got %v", got)
	}

	if got := loaded.Skills[SkillEclatDeGlace]; got == nil || got.SkillPoints != 1 {
		t.Errorf("expected Éclat de glace at 1 point to survive, got %v", got)
	}

	if len(loaded.Belt) != 4 || loaded.Belt[0] != ItemPotionDeMana {
		t.Errorf("expected the belt's contents to survive, got %v", loaded.Belt)
	}

	if loaded.Stats == nil {
		t.Fatal("expected Stats to survive")
	}

	if loaded.Stats.StatsPoints != 3 || loaded.Stats.VitalitySpent != 2 || loaded.Stats.LifePerVit != 4 {
		t.Errorf("expected StatsPoints/VitalitySpent/LifePerVit to survive, got %d/%d/%d",
			loaded.Stats.StatsPoints, loaded.Stats.VitalitySpent, loaded.Stats.LifePerVit)
	}

	if loaded.Stats.MaxHealth != 58 || loaded.Stats.Health != 58 {
		t.Errorf("expected MaxHealth/Health to survive, got %d/%d", loaded.Stats.MaxHealth, loaded.Stats.Health)
	}

	if !loaded.HasDiscoveredRecipe(RecipeUpgradeBatonApprenti) {
		t.Error("expected the Codex's discovered recipe to survive the round-trip")
	}

	if !loaded.RespecPartielUsedAt[d2enum.DifficultyNormal] {
		t.Error("expected RespecPartielUsedAt to survive the round-trip")
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
