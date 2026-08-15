package d2hero

import "errors"

// UseGlypheDOubliOnSkill consumes one ItemGlypheDOubli from h's inventory
// and, if found, forgets and refunds skillID (RespecSingleSkill) --
// implements a real (if simplified) rare-item gate for the "1 compétence"
// half of "Glyphe d'oubli" (§10), the same practice as
// UseRespecEssence/ItemEssenceDeBossOrdinaire for Respec complet. See
// ItemGlypheDOubli's own doc comment. UseGlypheDOubliOnAttribute below is
// the "1 point d'attribut" half -- a single use of the item is one or the
// other, matching RespecSingleSkill/RespecSingleAttributePoint's own
// "not both" split.
//
// Errors (and consumes nothing) if h has no glyph in inventory, or
// skillID wasn't learned (RespecSingleSkill's own failure mode).
func (h *HeroState) UseGlypheDOubliOnSkill(skillID int) (pointsRefunded int, err error) {
	slot, found := h.consumeFromInventory(ItemGlypheDOubli, 1)
	if !found {
		return 0, errors.New("no glyphe d'oubli in inventory")
	}

	pointsRefunded, err = h.RespecSingleSkill(skillID)
	if err != nil {
		// The glyph must not be lost on a failed respec (unknown/
		// unlearned skillID) -- restore it to the exact slot
		// consumeFromInventory cleared, not appended (Inventory is a
		// fixed-size 40-slot slice, see d2game/d2player/devilInventoryPanel's
		// own architecture notes on why that size must never change) --
		// same rollback discipline as MoveToStash/MoveToBelt.
		h.Inventory[slot] = ItemGlypheDOubli
		return 0, err
	}

	return pointsRefunded, nil
}

// UseGlypheDOubliOnAttribute consumes one ItemGlypheDOubli from h's
// inventory and, if found, refunds one point from attr
// (RespecSingleAttributePoint) -- the "1 point d'attribut" half of
// "Glyphe d'oubli", see UseGlypheDOubliOnSkill's own doc comment.
//
// Errors (and consumes nothing) if h has no glyph in inventory, or attr
// has no points spent on it (RespecSingleAttributePoint's own failure
// mode, via HeroStatsState.RefundAttributePoint).
func (h *HeroState) UseGlypheDOubliOnAttribute(attr Attribute) error {
	slot, found := h.consumeFromInventory(ItemGlypheDOubli, 1)
	if !found {
		return errors.New("no glyphe d'oubli in inventory")
	}

	if err := h.RespecSingleAttributePoint(attr); err != nil {
		// Same rollback discipline as UseGlypheDOubliOnSkill: restore to
		// the exact slot cleared, never append.
		h.Inventory[slot] = ItemGlypheDOubli
		return err
	}

	return nil
}
