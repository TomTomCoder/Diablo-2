package d2player

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"
)

// fakeSurface is, as far as a repo-wide search could confirm, the first
// d2interface.Surface test double in this codebase. It only records what
// DevilHUD.Render actually calls (PushTranslation/Pop/DrawRect/DrawTextf);
// everything else is a trivial no-op, just enough to satisfy the
// interface.
type fakeSurface struct {
	rects []int    // width of each DrawRect call, in call order
	texts []string // formatted text of each DrawTextf call, in call order
}

func (f *fakeSurface) Renderer() d2interface.Renderer                     { return nil }
func (f *fakeSurface) Clear(color.Color)                                  {}
func (f *fakeSurface) DrawLine(int, int, color.Color)                     {}
func (f *fakeSurface) GetSize() (width, height int)                       { return 0, 0 }
func (f *fakeSurface) GetDepth() int                                      { return 0 }
func (f *fakeSurface) PopN(int)                                           {}
func (f *fakeSurface) PushColor(color.Color)                              {}
func (f *fakeSurface) PushEffect(d2enum.DrawEffect)                       {}
func (f *fakeSurface) PushFilter(d2enum.Filter)                           {}
func (f *fakeSurface) PushSkew(float64, float64)                          {}
func (f *fakeSurface) PushScale(float64, float64)                         {}
func (f *fakeSurface) PushBrightness(float64)                             {}
func (f *fakeSurface) PushSaturation(float64)                             {}
func (f *fakeSurface) Render(d2interface.Surface)                         {}
func (f *fakeSurface) RenderSection(d2interface.Surface, image.Rectangle) {}
func (f *fakeSurface) ReplacePixels([]byte)                               {}
func (f *fakeSurface) Screenshot() *image.RGBA                            { return nil }
func (f *fakeSurface) PushTranslation(int, int)                           {}
func (f *fakeSurface) Pop()                                               {}

func (f *fakeSurface) DrawRect(width, _ int, _ color.Color) {
	f.rects = append(f.rects, width)
}

func (f *fakeSurface) DrawTextf(format string, params ...interface{}) {
	f.texts = append(f.texts, fmt.Sprintf(format, params...))
}

func TestDevilHUDRenderNilPlayerIsNoop(t *testing.T) {
	surface := &fakeSurface{}
	hud := &DevilHUD{}

	// must not panic with no local player resolved yet.
	hud.Render(surface, nil)

	if len(surface.rects) != 0 || len(surface.texts) != 0 {
		t.Errorf("expected no draw calls for a nil player, got rects=%v texts=%v", surface.rects, surface.texts)
	}
}

func TestDevilHUDRenderNilStatsIsNoop(t *testing.T) {
	surface := &fakeSurface{}
	hud := &DevilHUD{}

	hud.Render(surface, &d2mapentity.Player{})

	if len(surface.rects) != 0 || len(surface.texts) != 0 {
		t.Errorf("expected no draw calls for a player with nil Stats, got rects=%v texts=%v", surface.rects, surface.texts)
	}
}

func TestDevilHUDRenderDrawsHealthAndManaBarsProportionally(t *testing.T) {
	surface := &fakeSurface{}
	hud := &DevilHUD{}

	player := &d2mapentity.Player{
		Stats: &d2hero.HeroStatsState{
			Health: 50, MaxHealth: 100,
			Mana: 25, MaxMana: 100,
		},
	}

	hud.Render(surface, player)

	if len(surface.rects) != 2 {
		t.Fatalf("expected 2 DrawRect calls (health + mana), got %d", len(surface.rects))
	}

	if surface.rects[0] != devilHUDBarWidth/2 {
		t.Errorf("expected health bar width %d (50%%), got %d", devilHUDBarWidth/2, surface.rects[0])
	}

	if surface.rects[1] != devilHUDBarWidth/4 {
		t.Errorf("expected mana bar width %d (25%%), got %d", devilHUDBarWidth/4, surface.rects[1])
	}
}

func TestDevilHUDRenderShowsEquippedSkillNames(t *testing.T) {
	surface := &fakeSurface{}
	hud := &DevilHUD{}

	player := &d2mapentity.Player{
		Stats:      &d2hero.HeroStatsState{MaxHealth: 1, MaxMana: 1},
		LeftSkill:  d2hero.NewDevilHeroSkill(d2hero.SkillTraitDeFeu),
		RightSkill: nil,
	}

	hud.Render(surface, player)

	foundLeft, foundRight := false, false

	for _, text := range surface.texts {
		if text == "Left: "+d2hero.DevilSkills[d2hero.SkillTraitDeFeu].Name {
			foundLeft = true
		}

		if text == "Right: -" {
			foundRight = true
		}
	}

	if !foundLeft {
		t.Errorf("expected the left skill's real name among %v", surface.texts)
	}

	if !foundRight {
		t.Errorf("expected \"Right: -\" for an unequipped slot among %v", surface.texts)
	}
}

func TestHealthBarWidthZeroMaxHealthIsZero(t *testing.T) {
	if got := healthBarWidth(&d2hero.HeroStatsState{Health: 10, MaxHealth: 0}); got != 0 {
		t.Errorf("expected 0 for MaxHealth 0, got %d", got)
	}
}

func TestManaBarWidthZeroMaxManaIsZero(t *testing.T) {
	if got := manaBarWidth(&d2hero.HeroStatsState{Mana: 10, MaxMana: 0}); got != 0 {
		t.Errorf("expected 0 for MaxMana 0, got %d", got)
	}
}

func TestSkillNameUnknownSkillIDIsDash(t *testing.T) {
	skill := &d2hero.HeroSkill{SkillRecord: &d2records.SkillRecord{ID: -1}}

	if got := skillName(skill); got != "-" {
		t.Errorf("expected \"-\" for an unrecognized skill ID, got %q", got)
	}
}
