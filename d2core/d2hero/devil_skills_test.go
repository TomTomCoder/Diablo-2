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
