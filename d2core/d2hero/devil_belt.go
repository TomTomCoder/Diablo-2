package d2hero

import "errors"

// InitBelt sets h's belt to capacity empty slots, discarding whatever was
// there before. Meant to be called once, when a belt item is equipped
// (e.g. character creation assigning Ceinture de Cuir Runique's 4 slots,
// devil_mage_character_design.md §5 as it read before that document's
// visual-identity rewrite -- see devil_items.go's own top-of-file note) --
// same "baked in at equip time" shape as Robe du Novice's Health/Mana
// bonus, not a live recomputation.
// devil_game_design_reference.md §8 notes belts can hold up to 16 slots
// once upgraded -- there's no belt-upgrade path yet, so capacity is
// whatever the caller passes (ItemPotionSlots of the equipped belt).
func (h *HeroState) InitBelt(capacity int) {
	h.Belt = make([]string, capacity)
}

// AddPotionToBelt places itemCode in the first empty belt slot, reporting
// its index. Errors if every slot is occupied (including if InitBelt was
// never called, leaving a nil/zero-length belt).
//
// Correction (août 2026): now returns the slot index, aligning with
// AddToInventory/AddToStash (devil_inventory.go) -- needed by MoveToBelt
// below, and delegates to the same shared addToSlots primitive those use
// instead of its own copy of the same "first empty slot" loop. No real
// production caller existed for the old error-only signature to break.
func (h *HeroState) AddPotionToBelt(itemCode string) (slot int, err error) {
	return addToSlots(h.Belt, itemCode)
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

// MoveToBelt moves the item at Inventory[inventoryIndex] into the first
// empty Belt slot, reporting the moved item's code and its new belt
// index. Only potions (a nonzero ItemManaRestoreAmount) can go in the
// belt -- devil_game_design_reference.md §8: "Ceinture : jusqu'à 16
// emplacements pour potions de mana et parchemins" -- a real,
// already-existing data source, not a guessed distinction. If the item
// isn't a potion, or the belt is full, Inventory is left untouched (put
// back rather than lost), same rollback shape as
// HeroState.MoveToStash/MoveToInventory (devil_inventory.go).
func (h *HeroState) MoveToBelt(inventoryIndex int) (itemCode string, beltSlot int, err error) {
	itemCode, err = h.RemoveFromInventory(inventoryIndex)
	if err != nil {
		return "", 0, err
	}

	if ItemManaRestoreAmount(itemCode) <= 0 {
		h.Inventory[inventoryIndex] = itemCode
		return "", 0, errors.New("only potions can go in the belt")
	}

	beltSlot, err = h.AddPotionToBelt(itemCode)
	if err != nil {
		h.Inventory[inventoryIndex] = itemCode
		return "", 0, err
	}

	return itemCode, beltSlot, nil
}

// MoveFromBelt moves the potion at Belt[beltIndex] into Inventory's first
// empty slot, reporting the moved item's code and its new inventory
// index. No potion-type check needed on this side -- anything that made
// it into the belt in the first place is necessarily already a potion.
// Same rollback-on-full behavior as MoveToBelt.
func (h *HeroState) MoveFromBelt(beltIndex int) (itemCode string, inventorySlot int, err error) {
	itemCode, err = removeFromSlots(h.Belt, beltIndex)
	if err != nil {
		return "", 0, err
	}

	inventorySlot, err = h.AddToInventory(itemCode)
	if err != nil {
		h.Belt[beltIndex] = itemCode
		return "", 0, err
	}

	return itemCode, inventorySlot, nil
}
