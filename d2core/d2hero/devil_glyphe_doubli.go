package d2hero

import "errors"

// UseGlypheDOubliOnSkill consumes one ItemGlypheDOubli from h's inventory
// and, if found, forgets and refunds skillID (RespecSingleSkill) --
// implements a real (if simplified) rare-item gate for the "1 compétence"
// half of "Glyphe d'oubli" (§10), the same practice as
// UseRespecEssence/ItemEssenceDeBossOrdinaire for Respec complet. See
// ItemGlypheDOubli's own doc comment.
//
// The "1 point d'attribut" half (RespecSingleAttributePoint) isn't
// covered by a matching UseGlypheDOubliOnAttribute here -- not because it
// couldn't use the exact same item/gate, but for time, see ROADMAP.md.
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
