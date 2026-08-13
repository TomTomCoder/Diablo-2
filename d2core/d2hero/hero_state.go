package d2hero

import (
	"errors"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2inventory"
)

// HeroState stores the state of the player
type HeroState struct {
	HeroName   string                         `json:"heroName"`
	HeroType   d2enum.Hero                    `json:"heroType"`
	Act        int                            `json:"act"`
	FilePath   string                         `json:"-"`
	Equipment  d2inventory.CharacterEquipment `json:"equipment"`
	Stats      *HeroStatsState                `json:"stats"`
	Skills     map[int]*HeroSkill             `json:"skills"`
	X          float64                        `json:"x"`
	Y          float64                        `json:"y"`
	LeftSkill  int                            `json:"leftSkill"`
	RightSkill int                            `json:"rightSkill"`
	Gold       int                            `json:"Gold"`
	Difficulty d2enum.DifficultyType          `json:"difficulty"`
}

// LearnSkill spends one skill point to add skillID to h.Skills, if h is
// eligible: CanLearnSkill (level/palier gate), has a skill point to spend,
// and doesn't already know it.
//
// ponytail: nothing currently grants SkillPoints (no leveling/XP system
// exists yet), so this has no real caller yet either -- it's the business
// logic ready for whichever trigger comes first: a debug command, or the
// real client request once one exists. See ROADMAP.md Phase 2.
func (h *HeroState) LearnSkill(skillID int) error {
	if h.Stats == nil {
		return errors.New("hero has no stats")
	}

	if !CanLearnSkill(skillID, h.Stats.Level) {
		return errors.New("hero's level is too low for this skill")
	}

	if h.Stats.SkillPoints <= 0 {
		return errors.New("no skill points available")
	}

	if _, already := h.Skills[skillID]; already {
		return errors.New("skill already learned")
	}

	skill := NewDevilHeroSkill(skillID)
	if skill == nil {
		return errors.New("unknown skill")
	}

	if h.Skills == nil {
		h.Skills = make(map[int]*HeroSkill)
	}

	h.Skills[skillID] = skill
	h.Stats.SkillPoints--

	return nil
}
