package d2hero

import "testing"

// TestNewTraitDeFeuSkill guards against the crash this skill was built to
// avoid in the first place: renderLeftSkill/renderRightSkill (d2player/
// hud.go) and the OnPlayerCast call sites (d2player/game_controls.go) all
// dereference LeftSkill/RightSkill's embedded SkillRecord and
// SkillDescriptionRecord directly, with no nil-check on those embedded
// pointers themselves.
func TestNewTraitDeFeuSkill(t *testing.T) {
	skill := NewTraitDeFeuSkill()

	if skill == nil {
		t.Fatal("NewTraitDeFeuSkill returned nil")
	}

	if skill.SkillRecord == nil {
		t.Fatal("SkillRecord is nil -- ID/Charclass access would panic")
	}

	if skill.SkillRecord.ID != SkillTraitDeFeu {
		t.Errorf("expected ID %d, got %d", SkillTraitDeFeu, skill.SkillRecord.ID)
	}

	if skill.SkillDescriptionRecord == nil {
		t.Fatal("SkillDescriptionRecord is nil -- IconCel access would panic")
	}

	if skill.Shallow == nil || skill.Shallow.SkillID != SkillTraitDeFeu {
		t.Error("Shallow.SkillID must match SkillTraitDeFeu for save/load to round-trip the ID")
	}
}

// TestNewDevilHeroSkillPopulatesTreePositionAndEquipFlags is a regression
// test for a real bug: SkillPage/SkillColumn/SkillRow/Leftskill used to be
// left at their Go zero value here, which meant (page 0, column 0, row 0,
// Leftskill false) for every Devil skill -- d2game/d2player/skilltree.go's
// setTab() only shows an icon whose SkillPage matches the open tab, so
// every tab rendered empty, and skill_select_panel.go's left-click popup
// never offered any Devil skill at all (Leftskill always false).
func TestNewDevilHeroSkillPopulatesTreePositionAndEquipFlags(t *testing.T) {
	skill := NewDevilHeroSkill(SkillTraitDeFeu)

	wantPage, wantColumn, wantRow := SkillGridPosition(SkillTraitDeFeu)

	if skill.SkillPage != wantPage {
		t.Errorf("expected SkillPage %d, got %d", wantPage, skill.SkillPage)
	}

	if skill.SkillColumn != wantColumn || skill.SkillRow != wantRow {
		t.Errorf("expected (column %d, row %d), got (%d, %d)", wantColumn, wantRow, skill.SkillColumn, skill.SkillRow)
	}

	if !skill.Leftskill {
		t.Error("expected Leftskill true -- Devil draws no left/right distinction between skills")
	}

	if skill.Passive {
		t.Error("Trait de feu is an active skill, expected Passive false")
	}

	if skill.ListRow != wantPage {
		t.Errorf("expected ListRow to match SkillPage (%d) so the equip popup groups by tree, got %d", wantPage, skill.ListRow)
	}
}

// TestNewDevilHeroSkillMarksPassiveSkillsCorrectly guards the other half of
// the same fix: skill_select_panel.go excludes a skill from the equip
// popup via `skill.ListRow == -1 || skill.Passive` -- DevilSkillDef.Passive
// must actually reach the constructed HeroSkill for that exclusion to work.
func TestNewDevilHeroSkillMarksPassiveSkillsCorrectly(t *testing.T) {
	skill := NewDevilHeroSkill(SkillMaitriseElementaire)

	if !skill.Passive {
		t.Error("expected Maîtrise élémentaire to be marked Passive")
	}
}

func TestMaitriseElementaireDamagePercent(t *testing.T) {
	if got := MaitriseElementaireDamagePercent(0); got != 0 {
		t.Errorf("expected 0%% with no points invested, got %d", got)
	}

	one := MaitriseElementaireDamagePercent(1)
	if one <= 0 {
		t.Fatalf("expected a positive bonus for 1 point invested, got %d", one)
	}

	if got := MaitriseElementaireDamagePercent(3); got != one*3 {
		t.Errorf("expected the bonus to scale linearly with points invested (%d), got %d", one*3, got)
	}
}

func TestResonanceMagiqueDamagePercent(t *testing.T) {
	if got := ResonanceMagiqueDamagePercent(0); got != 0 {
		t.Errorf("expected 0%% with no points invested, got %d", got)
	}

	one := ResonanceMagiqueDamagePercent(1)
	if one <= 0 {
		t.Fatalf("expected a positive bonus for 1 point invested, got %d", one)
	}

	if got := ResonanceMagiqueDamagePercent(3); got != one*3 {
		t.Errorf("expected the bonus to scale linearly with points invested (%d), got %d", one*3, got)
	}
}

func TestRegenerationAccellereePercent(t *testing.T) {
	if got := RegenerationAccellereePercent(0); got != 0 {
		t.Errorf("expected 0%% with no points invested, got %d", got)
	}

	one := RegenerationAccellereePercent(1)
	if one <= 0 {
		t.Fatalf("expected a positive bonus for 1 point invested, got %d", one)
	}

	if got := RegenerationAccellereePercent(3); got != one*3 {
		t.Errorf("expected the bonus to scale linearly with points invested (%d), got %d", one*3, got)
	}
}
