package d2player

import "github.com/OpenDiablo2/OpenDiablo2/d2core/d2hero"

type inputCallbackListener interface {
	OnPlayerMove(x, y float64)
	OnPlayerCast(skillID int, x, y float64)
	OnInvestSkillPoint(skillID int)
	OnEquipSkill(slot d2hero.SkillSlot, skillID int)
}
