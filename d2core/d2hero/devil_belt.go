package d2hero

import "errors"

// InitBelt sets h's belt to capacity empty slots, discarding whatever was
// there before. Meant to be called once, when a belt item is equipped
// (e.g. character creation assigning Ceinture de Cuir Runique's 4 slots,
// devil_mage_character_design.md §5) -- same "baked in at equip time"
// shape as Robe du Novice's Health/Mana bonus, not a live recomputation.
// devil_game_design_reference.md §8 notes belts can hold up to 16 slots
// once upgraded -- there's no belt-upgrade path yet, so capacity is
// whatever the caller passes (ItemPotionSlots of the equipped belt).
func (h *HeroState) InitBelt(capacity int) {
	h.Belt = make([]string, capacity)
}

// AddPotionToBelt places itemCode in the first empty belt slot. Reports an
// error if every slot is occupied (including if InitBelt was never called,
// leaving a nil/zero-length belt).
func (h *HeroState) AddPotionToBelt(itemCode string) error {
	for i, slot := range h.Belt {
		if slot == "" {
			h.Belt[i] = itemCode
			return nil
		}
	}

	return errors.New("belt is full")
}

// UsePotion consumes the potion in belt slot index: restores Mana
// (ItemManaRestoreAmount) and clears the slot. Reports an error for an
// out-of-range or empty slot, or a hero with no stats to restore Mana to.
func (h *HeroState) UsePotion(index int) error {
	if index < 0 || index >= len(h.Belt) {
		return errors.New("belt slot out of range")
	}

	itemCode := h.Belt[index]
	if itemCode == "" {
		return errors.New("belt slot is empty")
	}

	if h.Stats == nil {
		return errors.New("hero has no stats")
	}

	h.Stats.RestoreMana(ItemManaRestoreAmount(itemCode))
	h.Belt[index] = ""

	return nil
}
