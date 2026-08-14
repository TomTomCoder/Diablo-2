package d2player

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"
)

const (
	devilHUDOffsetX    = 10
	devilHUDOffsetY    = 10
	devilHUDBarWidth   = 150
	devilHUDBarHeight  = 12
	devilHUDLineHeight = 16

	devilHUDHealthColor = 0xc82828ff // opaque red
	devilHUDManaColor   = 0x2848c8ff // opaque blue
)

// DevilHUD is a minimal, asset-independent HUD for Devil's own character
// state (health, mana, equipped skills) -- unlike HUD (hud.go), which
// renders stock Diablo II orb/panel .DC6 sprites that don't exist in this
// environment (no real Diablo II installation, ROADMAP.md Phase 6) and
// have no fallback if missing (d2asset.LoadAnimation errors, and nothing
// downstream guards against the nil *d2ui.Sprite that leaves behind).
// This draws plain rectangles and text via Surface.DrawRect/DrawTextf --
// both already asset-free themselves (DrawTextf's glyph table is an
// embedded bitmap, d2common/d2util/assets.CreateTextImage, not loaded
// from any .DC6/MPQ file).
//
// ponytail: no panel background, no icons, no skill *bar* to click yet --
// bars and text only. Good enough to prove real character state is
// visible at all; a real visual treatment is Phase 6/7 territory once
// real art exists, and clickable skill slots are a separate follow-up
// (adding input handling here risks colliding with GameControls' own
// move/cast click routing, not attempted in this pass).
type DevilHUD struct{}

// Render draws player's health/mana bars and equipped skill names in the
// screen's top-left corner. A no-op if player or player.Stats is nil (no
// local player resolved yet).
func (h *DevilHUD) Render(target d2interface.Surface, player *d2mapentity.Player) {
	if player == nil || player.Stats == nil {
		return
	}

	stats := player.Stats

	target.PushTranslation(devilHUDOffsetX, devilHUDOffsetY)
	defer target.Pop()

	target.DrawRect(healthBarWidth(stats), devilHUDBarHeight, d2util.Color(devilHUDHealthColor))

	target.PushTranslation(0, devilHUDLineHeight)
	target.DrawTextf("HP %d/%d", stats.Health, stats.MaxHealth)
	target.Pop()

	target.PushTranslation(0, devilHUDLineHeight*2) //nolint:gomnd // line position, not a magic constant
	target.DrawRect(manaBarWidth(stats), devilHUDBarHeight, d2util.Color(devilHUDManaColor))
	target.Pop()

	target.PushTranslation(0, devilHUDLineHeight*3) //nolint:gomnd // line position, not a magic constant
	target.DrawTextf("MP %d/%d", stats.Mana, stats.MaxMana)
	target.Pop()

	target.PushTranslation(0, devilHUDLineHeight*4) //nolint:gomnd // line position, not a magic constant
	target.DrawTextf("Left: %s", skillName(player.LeftSkill))
	target.Pop()

	target.PushTranslation(0, devilHUDLineHeight*5) //nolint:gomnd // line position, not a magic constant
	target.DrawTextf("Right: %s", skillName(player.RightSkill))
	target.Pop()
}

func healthBarWidth(stats *d2hero.HeroStatsState) int {
	if stats.MaxHealth <= 0 {
		return 0
	}

	return stats.Health * devilHUDBarWidth / stats.MaxHealth
}

func manaBarWidth(stats *d2hero.HeroStatsState) int {
	if stats.MaxMana <= 0 {
		return 0
	}

	return stats.Mana * devilHUDBarWidth / stats.MaxMana
}

// skillName returns skill's Devil display name, or "-" if unequipped or
// unknown.
func skillName(skill *d2hero.HeroSkill) string {
	if skill == nil || skill.SkillRecord == nil {
		return "-"
	}

	if def, ok := d2hero.DevilSkills[skill.SkillRecord.ID]; ok {
		return def.Name
	}

	return "-"
}
