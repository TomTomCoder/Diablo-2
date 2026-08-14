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

	// Belt holds potion item codes by slot, "" meaning empty -- Devil's
	// own potion belt (devil_game_design_reference.md §8), sized by
	// whichever belt item granted it (see InitBelt in devil_belt.go). Not
	// a d2inventory.CharacterEquipment field: it's slot *contents*, not an
	// equipped item itself.
	Belt []string `json:"belt"`

	// Inventory/Stash hold Devil item codes by slot, "" meaning empty --
	// devil_game_design_reference.md §8's "Inventaire : grille portée sur
	// le personnage (taille contrainte)" / "Coffre (stash) : accessible en
	// ville uniquement". See devil_inventory.go's own top-of-file note for
	// why these are flat slot lists rather than a true spatial grid, and
	// InitStorage for how they're sized.
	Inventory []string `json:"inventory"`
	Stash     []string `json:"stash"`

	// DiscoveredRecipes tracks which Cube de Nexus recipe IDs (see
	// devil_crafting.go's DevilCraftingRecipe.ID) h has ever crafted --
	// devil_game_design_reference.md §8's "Codex progressif": "chaque
	// recette découverte (par quête ou utilisation) est enregistrée et
	// consultable à tout moment". Only "par utilisation" is modeled here
	// (Craft marks a recipe discovered on success) -- "par quête" needs a
	// quest system that doesn't exist yet. A value is only ever true;
	// presence in the map is what DiscoverRecipe/HasDiscoveredRecipe check.
	DiscoveredRecipes map[string]bool `json:"discoveredRecipes"`

	// RespecPartielUsedAt tracks which difficulties h has already spent its
	// one-time "Respec partiel" at (devil_game_design_reference.md §10: "1
	// fois par difficulté (quête Den of Nexus)"). A value is only ever
	// true; presence is what RespecSkills checks. Doesn't track the quest
	// gate half of the same rule -- see RespecSkills' own doc comment.
	RespecPartielUsedAt map[d2enum.DifficultyType]bool `json:"respecPartielUsedAt"`
}

// LearnSkill spends one skill point to add skillID to h.Skills, if h is
// eligible: CanLearnSkill (level/palier gate), has a skill point to spend,
// and doesn't already know it.
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

// InvestSkillPoint spends one skill point to add another point to a skill h
// already knows -- e.g. Maîtrise élémentaire's "+% dégâts élémentaires par
// point" (devil_game_design_reference.md §7), or any of the design's
// "chaque point dans X augmente Y" synergies (§7 "Synergies"). Unlike
// LearnSkill, which only ever grants the first point, this requires the
// skill to already be known.
func (h *HeroState) InvestSkillPoint(skillID int) error {
	if h.Stats == nil {
		return errors.New("hero has no stats")
	}

	skill, known := h.Skills[skillID]
	if !known {
		return errors.New("skill not yet learned")
	}

	if h.Stats.SkillPoints <= 0 {
		return errors.New("no skill points available")
	}

	skill.SkillPoints++

	if skill.Shallow != nil {
		skill.Shallow.SkillPoints = skill.SkillPoints
	}

	h.Stats.SkillPoints--

	return nil
}

// resetAllSkillsAndAttributes clears every skill h has learned, refunding
// the skill points spent on them, and refunds every attribute point ever
// spent via HeroStatsState.SpendAttributePoint
// (HeroStatsState.RespecAllAttributePoints). LeftSkill/RightSkill are reset
// to 0 (no skill equipped) since they'd otherwise reference a skill that no
// longer exists in h.Skills. Reports how many *skill* points were refunded
// (the attribute refund is reflected directly in h.Stats.StatsPoints).
//
// Shared by RespecSkills ("Respec partiel", gated to once per difficulty)
// and RespecComplet ("Respec complet", no such gate) -- both reset exactly
// the same scope (§10: "Tous les points de compétences et d'attributs" /
// "Remise à zéro totale" read as the same reset, just via a different
// access route), only the gating around the reset differs.
func (h *HeroState) resetAllSkillsAndAttributes() (pointsRefunded int) {
	for _, skill := range h.Skills {
		if skill != nil {
			pointsRefunded += skill.SkillPoints
		}
	}

	h.Skills = make(map[int]*HeroSkill)
	h.LeftSkill = 0
	h.RightSkill = 0

	if h.Stats != nil {
		h.Stats.SkillPoints += pointsRefunded
		h.Stats.RespecAllAttributePoints()
	}

	return pointsRefunded
}

// RespecSkills implements the design's "Respec partiel"
// (devil_game_design_reference.md §10: "1 fois par difficulté (quête Den
// of Nexus)" / "Tous les points de compétences et d'attributs") in full --
// see resetAllSkillsAndAttributes for the reset itself.
//
// Errors if h has already used its one-time Respec partiel at h.Difficulty
// (RespecPartielUsedAt) -- the design's own "1 fois par difficulté" limit.
// Doesn't enforce the quest gate ("quête Den of Nexus") alongside it: no
// quest system exists yet, so that half stays undocumented as a real gap
// rather than faked.
func (h *HeroState) RespecSkills() (pointsRefunded int, err error) {
	if h.RespecPartielUsedAt[h.Difficulty] {
		return 0, errors.New("respec partiel already used at this difficulty")
	}

	pointsRefunded = h.resetAllSkillsAndAttributes()

	if h.RespecPartielUsedAt == nil {
		h.RespecPartielUsedAt = make(map[d2enum.DifficultyType]bool)
	}

	h.RespecPartielUsedAt[h.Difficulty] = true

	return pointsRefunded, nil
}

// RespecComplet implements the design's "Respec complet"
// (devil_game_design_reference.md §10: "Combinaison de 4 essences de boss
// dans le Cube de Nexus" / "Remise à zéro totale") -- the same full reset
// as RespecSkills ("Respec partiel"), but with no once-per-difficulty
// limit, matching how the design's own table only states that restriction
// for the partial variant.
//
// ponytail: the design's own trigger (combining 4 boss-essence items in
// the Cube de Nexus) needs boss-essence items that don't exist yet -- no
// boss content is implemented at all (ROADMAP.md Phase 4). This is the
// reset mechanism ready for whichever trigger comes first, same
// "mechanism before trigger" sequencing already used for LearnSkill/
// UsePotion/InvestSkillPoint in their own time.
func (h *HeroState) RespecComplet() (pointsRefunded int) {
	return h.resetAllSkillsAndAttributes()
}

// RespecSingleSkill removes skillID from h.Skills and refunds its skill
// points, without touching any other learned skill. Implements the "1
// compétence" half of the design's "Glyphe d'oubli"
// (devil_game_design_reference.md §10: "1 compétence ou 1 point
// d'attribut") -- see RespecSingleAttributePoint for the other half, "1
// point d'attribut". A single use of the item is one or the other, not
// both, hence two separate methods rather than one that does both.
//
// ponytail: no item/drop/craft delivers this yet -- same as LearnSkill
// originally, this is the business logic ready for whichever trigger comes
// first (ROADMAP.md Phase 5: Glyphe d'oubli, Cube de Nexus).
func (h *HeroState) RespecSingleSkill(skillID int) (pointsRefunded int, err error) {
	skill, known := h.Skills[skillID]
	if !known || skill == nil {
		return 0, errors.New("skill not learned")
	}

	pointsRefunded = skill.SkillPoints

	delete(h.Skills, skillID)

	if h.LeftSkill == skillID {
		h.LeftSkill = 0
	}

	if h.RightSkill == skillID {
		h.RightSkill = 0
	}

	if h.Stats != nil {
		h.Stats.SkillPoints += pointsRefunded
	}

	return pointsRefunded, nil
}

// RespecSingleAttributePoint refunds one point previously spent (via
// HeroStatsState.SpendAttributePoint) on attr, without touching any skill
// or any other attribute. Implements the "1 point d'attribut" half of the
// design's "Glyphe d'oubli" (devil_game_design_reference.md §10) -- see
// RespecSingleSkill for the other half, "1 compétence".
func (h *HeroState) RespecSingleAttributePoint(attr Attribute) error {
	if h.Stats == nil {
		return errors.New("hero has no stats")
	}

	return h.Stats.RefundAttributePoint(attr)
}

// DiscoverRecipe marks recipeID as discovered in h's Codex
// (DiscoveredRecipes), if it isn't already. Reports whether this was a new
// discovery (false if recipeID was already known).
func (h *HeroState) DiscoverRecipe(recipeID string) (newlyDiscovered bool) {
	if h.DiscoveredRecipes == nil {
		h.DiscoveredRecipes = make(map[string]bool)
	}

	if h.DiscoveredRecipes[recipeID] {
		return false
	}

	h.DiscoveredRecipes[recipeID] = true

	return true
}

// HasDiscoveredRecipe reports whether recipeID is in h's Codex.
func (h *HeroState) HasDiscoveredRecipe(recipeID string) bool {
	return h.DiscoveredRecipes[recipeID]
}
